// Takes the PR's screenshots: the front page, the round page /alpha/r2, the pass page
// /alpha/r2/p13, and /rulings, each in light and dark. Design done item 7 reads its output as
// the before and after sets. It is not a spec: the suite's testMatch skips it, and the gate
// never runs it. Run it by path after a build:
//   CC_DESK_BINARY=<cc-desk binary> node tests/screenshots.ts <output folder> [state]
import { spawn } from 'node:child_process';
import { mkdirSync, mkdtempSync, rmSync } from 'node:fs';
import { createServer } from 'node:net';
import { tmpdir } from 'node:os';
import path from 'node:path';
import { chromium, type Page } from '@playwright/test';
import { buildState } from './state.ts';

const [out, name = 'started'] = process.argv.slice(2);
if (!out) throw new Error('usage: node tests/screenshots.ts <output folder> [state]');
mkdirSync(out, { recursive: true });

const stateDir = mkdtempSync(path.join(tmpdir(), 'cc-desk-shot-'));
const clock = buildState(name, stateDir);

const port = await new Promise<number>((resolve, reject) => {
	const probe = createServer();
	probe.once('error', reject);
	probe.listen(0, '127.0.0.1', () => {
		const { port } = probe.address() as { port: number };
		probe.close(() => resolve(port));
	});
});

const server = spawn('node', [path.resolve(import.meta.dirname, '..', 'build')], {
	env: { ...process.env, HOST: '127.0.0.1', PORT: String(port), CC_DESK_STATE_DIR: stateDir, SHUTDOWN_TIMEOUT: '1' },
	stdio: 'ignore'
});

type Shot = { name: string; address: string; ready: (page: Page) => Promise<void> };

const pages: Shot[] = [
	{
		name: 'front',
		address: '/',
		ready: (page) => page.getByRole('region', { name: 'Next decision session', exact: true }).waitFor()
	},
	...[
		['round', '/alpha/r2'],
		['pass', '/alpha/r2/p13'],
		['rulings', '/rulings']
	].map(([name, address]) => ({
		name,
		address,
		ready: (page: Page) => page.getByRole('main').getByRole('heading', { level: 1 }).waitFor()
	}))
];

const browser = await chromium.launch();
try {
	for (const scheme of ['light', 'dark'] as const) {
		const context = await browser.newContext({
			timezoneId: 'Europe/London',
			locale: 'en-US',
			colorScheme: scheme,
			viewport: { width: 1600, height: 1000 }
		});
		const page = await context.newPage();
		await page.clock.setFixedTime(new Date(clock));
		for (const shot of pages) {
			for (let i = 0; i < 100; i++) {
				try {
					await page.goto(`http://127.0.0.1:${port}${shot.address}`);
					break;
				} catch {
					await new Promise((r) => setTimeout(r, 100));
				}
			}
			await shot.ready(page);
			await page.screenshot({
				path: path.join(out, `${name}-${shot.name}-${scheme}.png`),
				fullPage: true
			});
		}
		await context.close();
	}
} finally {
	await browser.close();
	server.kill('SIGTERM');
	rmSync(stateDir, { recursive: true, force: true });
}
