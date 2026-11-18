// The terminal link: the handler's reading of an address, the address the page builds, and the
// button that copies the command. Every expected string is written out here.
import { test, expect } from '@playwright/test';
import { spawnSync } from 'node:child_process';
import { chmodSync, existsSync, mkdirSync, mkdtempSync, readFileSync, rmSync, writeFileSync } from 'node:fs';
import { tmpdir } from 'node:os';
import path from 'node:path';

const handler = path.resolve(import.meta.dirname, '..', 'desktop', 'cc-desk-open');

interface Run {
	status: number | null;
	stderr: string;
	/** The lines the stub ptyxis recorded, or null when it never ran. */
	record: string[] | null;
}

/**
 * Runs the handler with stubs in a folder of its own and removes the folder after.
 * `withCcDesk: false` leaves the stub cc-desk out of `bin/`.
 */
function runHandler(
	args: string[],
	opts: { env?: Record<string, string>; withCcDesk?: boolean } = {}
): Run & { bin: string } {
	const dir = mkdtempSync(path.join(tmpdir(), 'cc-desk-link-'));
	try {
		const bin = path.join(dir, 'bin');
		const recordFile = path.join(dir, 'record');
		mkdirSync(bin);
		const ptyxis = path.join(bin, 'ptyxis');
		writeFileSync(ptyxis, `#!/bin/sh\nfor a in "$@"; do printf '%s\\n' "$a" >> '${recordFile}'; done\n`);
		chmodSync(ptyxis, 0o755);
		if (opts.withCcDesk !== false) {
			const cc = path.join(bin, 'cc-desk');
			writeFileSync(cc, '#!/bin/sh\nexit 0\n');
			chmodSync(cc, 0o755);
		}
		const res = spawnSync('bash', [handler, ...args], {
			env: { PATH: `${bin}:/usr/bin:/bin`, ...opts.env },
			encoding: 'utf8'
		});
		return {
			status: res.status,
			stderr: res.stderr,
			record: existsSync(recordFile) ? readFileSync(recordFile, 'utf8').trimEnd().split('\n') : null,
			bin
		};
	} finally {
		rmSync(dir, { recursive: true, force: true });
	}
}

test.describe('the handler', () => {
	test('cc-desk://alpha/r1 runs the command in a terminal', () => {
		const run = runHandler(['cc-desk://alpha/r1']);
		expect(run.status).toBe(0);
		expect(run.record).toEqual(['-x', `${run.bin}/cc-desk alpha r1`]);
	});

	const refused = [
		'cc-desk://alpha/r1;x',
		'cc-desk://alpha/r1%3Bx',
		'cc-desk://Alpha/r1',
		'cc-desk://alpha/r1/x',
		'cc-desk://alpha/../r1',
		'cc-desk://alpha/r1?x=1',
		'cc-desk://alpha/r1/',
		'cc-desk://alpha',
		'cc-desk:alpha/r1',
		'cc-desk:///alpha/r1',
		'cc-desk://alpha/r1#x',
		'cc-desk://alpha/r1 x',
		'cc-desk://alph%C3%A4/r1',
		'cc-desk://ALPHA/R1',
		'cc-desk://-x/r1',
		'cc-desk://alpha/--help'
	];
	for (const address of refused) {
		test(`refuses ${address}`, () => {
			const run = runHandler([address]);
			expect(run.status).toBe(2);
			expect(run.record).toBeNull();
			expect(run.stderr).toContain('cc-desk-open: refused');
		});
	}

	test('refuses an accented letter under a UTF-8 locale', () => {
		const run = runHandler(['cc-desk://alphä/r1'], { env: { LANG: 'en_US.UTF-8' } });
		expect(run.status).toBe(2);
		expect(run.record).toBeNull();
	});

	test('refuses no address', () => {
		const run = runHandler([]);
		expect(run.status).toBe(2);
		expect(run.record).toBeNull();
	});

	test('refuses two addresses', () => {
		const run = runHandler(['cc-desk://alpha/r1', 'cc-desk://alpha/r2']);
		expect(run.status).toBe(2);
		expect(run.record).toBeNull();
	});

	test('refuses when cc-desk is not on the path', () => {
		const run = runHandler(['cc-desk://alpha/r1'], { withCcDesk: false });
		expect(run.status).toBe(2);
		expect(run.record).toBeNull();
		expect(run.stderr).toContain('cc-desk');
	});
});
