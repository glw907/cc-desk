// The terminal link: the handler's reading of an address, the address the page builds, and the
// button that copies the command. Every expected string is written out here.
import { test, expect } from './app';
import { show } from './page';
import { terminalLink } from '../src/lib/link';
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
	test('cc-desk://a.b/r_1 runs the command in a terminal', () => {
		const run = runHandler(['cc-desk://a.b/r_1']);
		expect(run.status).toBe(0);
		expect(run.record).toEqual(['-x', `${run.bin}/cc-desk a.b r_1`]);
	});

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
		'cc-desk://.a/r1',
		'cc-desk://alpha/_r1',
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

test.describe('the address', () => {
	const good: [string, string, string][] = [
		['alpha', 'r1', 'cc-desk://alpha/r1'],
		['a.b', 'r1', 'cc-desk://a.b/r1'],
		['a_b', 'r_1', 'cc-desk://a_b/r_1']
	];
	for (const [project, round, want] of good) {
		test(`${project} and ${round} give ${want}`, () => {
			expect(terminalLink(project, round)).toBe(want);
		});
	}

	const bad: [string, string][] = [
		['Alpha', 'r1'],
		['.a', 'r1'],
		['_a', 'r1'],
		['a/b', 'r1'],
		['-x', 'r1'],
		['', 'r1'],
		['alpha;x', 'r1'],
		['alpha', 'R1'],
		['alpha', '.1'],
		['alpha', 'r/1'],
		['alpha', '-x'],
		['alpha', ''],
		['alpha', 'r1;x']
	];
	for (const [project, round] of bad) {
		test(`"${project}" and "${round}" give no address`, () => {
			expect(terminalLink(project, round)).toBeNull();
		});
	}
});

test.describe('the pages', () => {
	test('a round page and a pass page show the link and the command', async ({ app, page }) => {
		await show(app, page, 'twin', '/alder/r1');
		const link = page.getByRole('main').getByRole('link', { name: 'Open in a terminal', exact: true });
		await expect(link).toHaveCount(1);
		await expect(link).toHaveAttribute('href', 'cc-desk://alder/r1');
		await expect(page.locator('main code')).toHaveText('cc-desk alder r1');

		await page.goto(app.url + '/alder/r1/a-one');
		const passLink = page.getByRole('main').getByRole('link', { name: 'Open in a terminal', exact: true });
		await expect(passLink).toHaveCount(1);
		await expect(passLink).toHaveAttribute('href', 'cc-desk://alder/r1');
		await expect(page.locator('main code')).toHaveText('cc-desk alder r1');
	});

	test('each line under Waiting for you has a link, and its row link keeps its address', async ({ app, page }) => {
		await show(app, page, 'big');
		const waits = page.getByRole('region', { name: 'Waiting for you' });
		await expect(waits.getByRole('link', { name: 'Open north/n1 in a terminal' })).toHaveAttribute(
			'href',
			'cc-desk://north/n1'
		);
		await expect(waits.getByRole('link', { name: 'Open south/s1 in a terminal' })).toHaveAttribute(
			'href',
			'cc-desk://south/s1'
		);
		await expect(waits.getByRole('link', { name: 'north/n1 2 questions' })).toHaveAttribute('href', '/north/n1');
		await expect(waits.getByRole('link', { name: 'south/s1 1 question' })).toHaveAttribute('href', '/south/s1');
	});
});

test.describe('the copy button', () => {
	test('copies the command and says Copied', async ({ app, page, context }) => {
		await context.grantPermissions(['clipboard-read', 'clipboard-write']);
		await show(app, page, 'twin', '/alder/r1');
		const button = page.getByRole('button', { name: 'Copy command', exact: true });
		await button.click();
		await expect(page.getByRole('status').filter({ hasText: 'Copied' })).toHaveCount(1);
		await expect(button).toHaveText('Copy command');
		expect(await page.evaluate(() => navigator.clipboard.readText())).toBe('cc-desk alder r1');
	});
});
