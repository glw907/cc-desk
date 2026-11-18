// State folders for tests. A state is a copy of a fixture from e2e/testdata/ with the real
// cc-desk binary's verbs run on it, so a test never draws a snapshot the binary would not write.
import { execFileSync, spawnSync } from 'node:child_process';
import { cpSync, mkdtempSync, readFileSync, renameSync, writeFileSync } from 'node:fs';
import { tmpdir } from 'node:os';
import path from 'node:path';
import type { Snapshot } from '../src/lib/snapshot';

const testdata = path.resolve(import.meta.dirname, '..', '..', 'e2e', 'testdata');

/** The clock of the fixture's own `snapshot.json`: 2:00 pm Alaska time. */
export const WORKED_CLOCK = '2026-11-20T23:00:00Z';

/** Absolute path of a file under `e2e/testdata/inputs/`. */
export function input(name: string): string {
	return path.join(testdata, 'inputs', name);
}

/** Copies a fixture folder (`state`, `order`, or `writers`) into a fresh temporary folder. */
export function copyFixture(fixture: string, into?: string): string {
	const dir = into ?? path.join(mkdtempSync(path.join(tmpdir(), 'cc-desk-fixture-')), 'state');
	cpSync(path.join(testdata, fixture), dir, { recursive: true });
	return dir;
}

function binary(): string {
	const bin = process.env.CC_DESK_BINARY;
	if (!bin) throw new Error('CC_DESK_BINARY is unset: global setup did not build cc-desk');
	return bin;
}

function binaryArgs(dir: string, now: string, args: string[]): string[] {
	return [
		'--state-dir',
		dir,
		'--now',
		now,
		'--usage-command',
		path.join(testdata, 'usage40.sh'),
		...args
	];
}

const env = { ...process.env, PATH: '/usr/bin:/bin' };

/** Runs a verb and returns what it printed. A non-zero exit throws. */
export function run(dir: string, now: string, ...args: string[]): string {
	return execFileSync(binary(), binaryArgs(dir, now, args), { env, encoding: 'utf8' });
}

/** Runs a verb that is expected to fail and returns its exit code and stderr. */
export function runFailing(
	dir: string,
	now: string,
	...args: string[]
): { code: number | null; stderr: string } {
	const r = spawnSync(binary(), binaryArgs(dir, now, args), { env, encoding: 'utf8' });
	return { code: r.status, stderr: r.stderr };
}

/** What the `snapshot` verb prints at a clock, parsed. */
export function snapshotAt(dir: string, now: string): Snapshot {
	return JSON.parse(run(dir, now, 'snapshot')) as Snapshot;
}

/**
 * Replaces a file in the state folder the way the binary does: write a temporary name in the
 * same folder, then rename it over the target, so the server never reads half a file.
 */
export function writeAtomic(dir: string, name: string, text: string): void {
	const tmp = path.join(dir, `.${path.basename(name)}.tmp`);
	writeFileSync(tmp, text);
	renameSync(tmp, path.join(dir, name));
}

/** Fills `dir` with the `state` fixture, beta's b3 designed and alpha's p13 started, both at 23:00Z. */
export function started(dir: string): string {
	copyFixture('state', dir);
	run(dir, '2026-11-20T23:00:00Z', 'round', 'design', 'beta', 'b3', input('design-b3.json'));
	run(dir, '2026-11-20T23:00:00Z', 'pass', 'start', 'alpha', 'p13');
	return dir;
}

/** The usual clock for the second pass-start verb in a state. */
const T = (hhmm: string) => `2026-11-20T${hhmm}:00Z`;

/**
 * Every state the tests draw, each printed by the real binary from a fresh copy of a fixture,
 * with the verbs the plan's table lists. Each is a snapshot at the clock of its last verb.
 */
export function allStates(): Record<string, Snapshot> {
	const out: Record<string, Snapshot> = {};

	const base = started(copyFixture('state'));
	out.started = snapshotAt(base, T('23:00'));

	const fast = started(copyFixture('state'));
	run(fast, T('23:50'), 'pass', 'finish', 'alpha', 'p13');
	out.fast = snapshotAt(fast, T('23:50'));

	const slow = started(copyFixture('state'));
	run(slow, '2026-11-21T02:20:00Z', 'pass', 'finish', 'alpha', 'p13');
	out.slow = snapshotAt(slow, '2026-11-21T02:20:00Z');

	const variant = copyFixture('state');
	const beta = path.join(variant, 'releases', 'beta.json');
	writeFileSync(beta, readFileSync(beta, 'utf8').replace('"planned_minutes": 480', '"planned_minutes": 60'));
	started(variant);
	out.variant = snapshotAt(variant, T('23:00'));

	out.big = snapshotAt(copyFixture('order'), T('23:00'));

	const guess = copyFixture('order');
	run(guess, T('23:00'), 'pass', 'start', 'north', 'n-free');
	run(guess, T('23:10'), 'pass', 'start', 'north', 'n-guess', '--guess', 'north-q2');
	out.guess = snapshotAt(guess, T('23:10'));

	const two = copyFixture('order');
	run(two, T('23:00'), 'pass', 'start', 'north', 'n-free');
	run(two, T('23:01'), 'pass', 'start', 'south', 's-free');
	runFailing(two, T('23:02'), 'pass', 'start', 'north', 'n-two');
	runFailing(two, T('23:03'), 'pass', 'start', 'south', 's-two');
	out.two = snapshotAt(two, T('23:03'));

	const redraw = copyFixture('state');
	run(redraw, T('23:05'), 'pass', 'start', 'alpha', 'p13');
	out.redraw = snapshotAt(redraw, T('23:05'));

	const stopped = started(copyFixture('state'));
	run(stopped, T('23:10'), 'pass', 'stop', 'alpha', 'p13', '--reason', 'scope');
	out.stopped = snapshotAt(stopped, T('23:10'));

	return out;
}
