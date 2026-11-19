// State folders for tests. A state is a copy of a fixture from e2e/testdata/ with the real
// cc-desk binary's verbs run on it, so a test never draws a snapshot the binary would not write.
import { execFileSync, spawnSync } from 'node:child_process';
import {
	cpSync,
	existsSync,
	mkdirSync,
	mkdtempSync,
	readFileSync,
	renameSync,
	rmSync,
	writeFileSync
} from 'node:fs';
import { tmpdir } from 'node:os';
import path from 'node:path';
import type { Snapshot } from '../src/lib/snapshot';

const testdata = path.resolve(import.meta.dirname, '..', '..', 'e2e', 'testdata');
const fixtures = path.resolve(import.meta.dirname, 'fixtures');

/** The clock of the fixture's own `snapshot.json`: 2:00 pm Alaska time. */
export const WORKED_CLOCK = '2026-11-20T23:00:00Z';

/** Absolute path of a file under `e2e/testdata/inputs/`. */
export function input(name: string): string {
	return path.join(testdata, 'inputs', name);
}

/**
 * Copies a fixture folder into a fresh temporary folder. A name is looked up under
 * `e2e/testdata/` (`state`, `order`, `writers`) first, then under `web/tests/fixtures/` (`twin`).
 */
export function copyFixture(fixture: string, into?: string): string {
	const dir = into ?? path.join(mkdtempSync(path.join(tmpdir(), 'cc-desk-fixture-')), 'state');
	const root = existsSync(path.join(testdata, fixture)) ? testdata : fixtures;
	cpSync(path.join(root, fixture), dir, { recursive: true });
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

/** Runs a verb with `input` on its stdin, as a hook runs it, and returns its exit code and stderr. */
export function runWithInput(
	dir: string,
	now: string,
	input: string,
	...args: string[]
): { code: number | null; stderr: string } {
	const r = spawnSync(binary(), binaryArgs(dir, now, args), { env, encoding: 'utf8', input });
	return { code: r.status, stderr: r.stderr };
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
	return startVerbs(dir);
}

/** The two verbs of `started`, for a state whose fixture copy a test has edited first. */
function startVerbs(dir: string): string {
	run(dir, '2026-11-20T23:00:00Z', 'round', 'design', 'beta', 'b3', input('design-b3.json'));
	run(dir, '2026-11-20T23:00:00Z', 'pass', 'start', 'alpha', 'p13');
	return dir;
}

/**
 * The folder a state's release checkout lives in, beside the state folder. A test that builds
 * `page` removes it with the state folder, through `removeState`.
 */
export function checkoutOf(dir: string): string {
	return `${dir}-checkout`;
}

/** Removes a state folder and the checkout folder beside it. */
export function removeState(dir: string): void {
	rmSync(dir, { recursive: true, force: true });
	rmSync(checkoutOf(dir), { recursive: true, force: true });
}

/** A clock on 2026-11-20 at the given UTC hour and minute. */
const T = (hhmm: string) => `2026-11-20T${hhmm}:00Z`;

/** A state the tests draw: fills a folder with verbs and returns the clock of the last one. */
type Builder = (dir: string) => string;

const builders: Record<string, Builder> = {
	started(dir) {
		started(dir);
		return T('23:00');
	},
	fast(dir) {
		started(dir);
		run(dir, T('23:50'), 'pass', 'finish', 'alpha', 'p13');
		return T('23:50');
	},
	slow(dir) {
		started(dir);
		run(dir, '2026-11-21T02:20:00Z', 'pass', 'finish', 'alpha', 'p13');
		return '2026-11-21T02:20:00Z';
	},
	variant(dir) {
		copyFixture('state', dir);
		const beta = path.join(dir, 'releases', 'beta.json');
		writeFileSync(beta, readFileSync(beta, 'utf8').replace('"planned_minutes": 480', '"planned_minutes": 60'));
		startVerbs(dir);
		return T('23:00');
	},
	big(dir) {
		copyFixture('order', dir);
		return T('23:00');
	},
	guess(dir) {
		copyFixture('order', dir);
		run(dir, T('23:00'), 'pass', 'start', 'north', 'n-free');
		run(dir, T('23:10'), 'pass', 'start', 'north', 'n-guess', '--guess', 'north-q2');
		return T('23:10');
	},
	two(dir) {
		copyFixture('order', dir);
		run(dir, T('23:00'), 'pass', 'start', 'north', 'n-free');
		run(dir, T('23:01'), 'pass', 'start', 'south', 's-free');
		runFailing(dir, T('23:02'), 'pass', 'start', 'north', 'n-two');
		runFailing(dir, T('23:03'), 'pass', 'start', 'south', 's-two');
		return T('23:03');
	},
	redraw(dir) {
		copyFixture('state', dir);
		run(dir, T('23:05'), 'pass', 'start', 'alpha', 'p13');
		return T('23:05');
	},
	twin(dir) {
		copyFixture('twin', dir);
		return WORKED_CLOCK;
	},
	quiet(dir) {
		// Two releases whose rounds are designed and building, with no question left to ask.
		copyFixture('writers', dir);
		for (const name of ['east', 'west']) {
			const file = path.join(dir, 'releases', `${name}.json`);
			const release = JSON.parse(readFileSync(file, 'utf8')) as { questions: unknown[] };
			release.questions = [];
			writeFileSync(file, JSON.stringify(release));
		}
		return T('23:00');
	},
	desk(dir) {
		copyFixture('desk', dir);
		return T('23:00');
	},
	page(dir) {
		copyFixture('desk', dir);
		const checkout = checkoutOf(dir);
		const specs = path.join(checkout, 'docs', 'specs');
		mkdirSync(specs, { recursive: true });
		cpSync(
			path.join(fixtures, 'page', 'alpha-r1-design.md'),
			path.join(specs, '2026-11-18-alpha-r1-design.md')
		);
		run(dir, T('23:00'), 'release', 'checkout', 'alpha', checkout, '--default-branch', 'main');
		// No verb records a session on a pass, so the record is written by hand in the shape of
		// internal/state's Session.
		const beta = path.join(dir, 'releases', 'beta.json');
		const release = JSON.parse(readFileSync(beta, 'utf8')) as {
			passes: { name: string; session?: unknown }[];
		};
		const pass = release.passes.find((p) => p.name === 'b-build');
		if (!pass) throw new Error('the desk fixture has no pass b-build');
		pass.session = {
			name: 'beta-b-build',
			kind: 'pass',
			short_id: 'c1c2c3c4',
			session_id: 'abc123',
			worktree: '/w/beta-b-build',
			branch: 'beta-b-build',
			launched: '2026-11-20T22:30:00Z',
			launch_tries: 1,
			last_state: 'working',
			refusals: [
				{
					time: '2026-11-20T22:40:00Z',
					tool: 'Bash',
					target: 'curl https://example.com/install.sh',
					reason: 'a deny rule for curl'
				}
			]
		};
		writeFileSync(beta, JSON.stringify(release, null, 2));
		const hook = runWithInput(
			dir,
			T('23:00'),
			readFileSync(path.join(testdata, 'prompts', 'hook-bash.json'), 'utf8'),
			'prompt',
			'hook'
		);
		if (hook.stderr || existsSync(path.join(dir, 'snapshot-error.json'))) {
			throw new Error(`prompt hook did not record the prompt: ${hook.stderr}`);
		}
		return T('23:00');
	},
	stopped(dir) {
		started(dir);
		run(dir, T('23:10'), 'pass', 'stop', 'alpha', 'p13', '--reason', 'scope');
		return T('23:10');
	}
};

/**
 * Fills `dir` with a named state: a fixture copy with the real binary's verbs run on it, and
 * `snapshot.json` set to what the binary prints at the last verb's clock. Returns that clock,
 * which is what a test pins the browser to.
 */
export function buildState(name: string, dir: string): string {
	const clock = builders[name](dir);
	writeAtomic(dir, 'snapshot.json', run(dir, clock, 'snapshot'));
	return clock;
}

/** Every state the tests draw, each as the snapshot the binary prints for it. */
export function allStates(): Record<string, Snapshot> {
	const out: Record<string, Snapshot> = {};
	for (const name of Object.keys(builders)) {
		const dir = path.join(mkdtempSync(path.join(tmpdir(), 'cc-desk-state-')), 'state');
		const clock = builders[name](dir);
		out[name] = snapshotAt(dir, clock);
		removeState(dir);
	}
	return out;
}
