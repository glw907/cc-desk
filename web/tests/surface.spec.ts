// The page is a reading surface: it holds no control, refuses every write, writes no file.
import { chmodSync, cpSync, mkdtempSync, readdirSync, readFileSync, rmSync, writeFileSync } from 'node:fs';
import { tmpdir } from 'node:os';
import path from 'node:path';
import { expect, test } from './app';
import { crawl, openDrawn } from './crawl';
import { hashTree } from './hash';
import { buildState } from './state';
import type { Snapshot } from '../src/lib/snapshot';

/** Sets the mode of `root`, every folder under it, and every file under it. */
function chmodTree(root: string, fileMode: number, dirMode: number): void {
	const walk = (dir: string) => {
		chmodSync(dir, dirMode);
		for (const entry of readdirSync(dir, { withFileTypes: true })) {
			const full = path.join(dir, entry.name);
			if (entry.isDirectory()) walk(full);
			else chmodSync(full, fileMode);
		}
	};
	walk(root);
}

const readOnly = (root: string) => chmodTree(root, 0o444, 0o555);
const writable = (root: string) => chmodTree(root, 0o644, 0o755);

test.describe('a reading surface', () => {
	test('the walk finds no control, every write is 405, both folders end as they began', async ({ app, page }) => {
		const appCopy = mkdtempSync(path.join(tmpdir(), 'cc-desk-app-'));
		cpSync(path.resolve(import.meta.dirname, '..', 'build'), appCopy, { recursive: true });
		const clock = buildState('started', app.stateDir);
		const snapshot = JSON.parse(readFileSync(path.join(app.stateDir, 'snapshot.json'), 'utf8')) as Snapshot;
		const stateBefore = hashTree(app.stateDir);
		const appBefore = hashTree(appCopy);
		try {
			readOnly(app.stateDir);
			readOnly(appCopy);
			// A run as root ignores the write bits, so prove them before relying on them.
			for (const root of [app.stateDir, appCopy]) {
				expect(() => writeFileSync(path.join(root, 'probe.txt'), 'x'), `${root} must refuse a write`).toThrow(/EACCES/);
			}

			await page.clock.setFixedTime(new Date(clock));
			await app.start(appCopy);

			// 1. Every page the links reach holds no control but "Copy command".
			const visited = await crawl(page, app.url, async (address) => {
				for (const selector of ['form', 'input', 'select', 'textarea', '[contenteditable]']) {
					const found = await page.locator(selector).count();
					expect(found, `${address} holds ${found} ${selector}`).toBe(0);
				}
				for (const button of await page.getByRole('button').all()) {
					const name = await button.evaluate((el) => el.getAttribute('aria-label') ?? el.textContent?.trim() ?? '');
					expect(name, `${address} holds a button named "${name}"`).toBe('Copy command');
				}
			});

			// 2. The walk reached every address the snapshot names.
			const named = ['/', '/rulings'];
			for (const release of snapshot.releases) {
				for (const round of release.rounds) {
					named.push(`/${release.project}/${round.name}`);
					for (const pass of round.passes) named.push(`/${release.project}/${round.name}/${pass.name}`);
				}
			}
			expect(named.length).toBeGreaterThan(20);
			for (const address of named) expect(visited, `the walk missed ${address}`).toContain(address);

			// 3. Every write method is 405 with the allow header, on every address and on a built asset.
			await openDrawn(page, app.url, '/');
			const assetHref = await page.locator('link[rel="modulepreload"]').first().getAttribute('href');
			expect(assetHref, 'the page links no built asset').toContain('_app/immutable/');
			const asset = '/' + (assetHref as string).replace(/^\.\//, '');
			const addresses = [...visited, '/api/state', '/nothing/names/this', asset];
			for (const address of addresses) {
				for (const method of ['POST', 'PUT', 'PATCH', 'DELETE', 'OPTIONS']) {
					// The page's own origin, so Kit's origin check passes and the hook answers.
					const res = await page.request.fetch(app.url + address, { method, headers: { origin: app.url } });
					expect(res.status(), `${method} ${address}`).toBe(405);
					expect(res.headers()['allow'], `${method} ${address} allow`).toBe('GET, HEAD');
				}
			}

			// 4. The poll runs against the read-only folders: three answers from one open page.
			let asked = 0;
			page.on('request', (r) => {
				if (new URL(r.url()).pathname === '/api/state') asked++;
			});
			await openDrawn(page, app.url, '/');
			await expect.poll(() => asked, { timeout: 60_000 }).toBeGreaterThanOrEqual(3);

			// 5. The server said nothing wrong, and both folders are what they were.
			await app.stop();
			expect(app.stderr).toBe('');
			expect(app.stdout.split('\n').filter((line) => /error/i.test(line))).toEqual([]);
			expect(hashTree(app.stateDir)).toEqual(stateBefore);
			expect(hashTree(appCopy)).toEqual(appBefore);
		} finally {
			// A failed walk would leave read-only folders the fixture cannot remove.
			writable(app.stateDir);
			writable(appCopy);
			rmSync(appCopy, { recursive: true, force: true });
		}
	});

	test('the server source holds no call that writes a file', () => {
		const src = path.resolve(import.meta.dirname, '..', 'src');
		const files: string[] = [];
		const walk = (dir: string) => {
			for (const entry of readdirSync(dir, { withFileTypes: true })) {
				const full = path.join(dir, entry.name);
				if (entry.isDirectory()) walk(full);
				else if (/\.(ts|js|svelte)$/.test(entry.name)) files.push(full);
			}
		};
		walk(src);
		expect(files.length).toBeGreaterThan(10);

		const allowedNames = new Set(['readFile', 'stat', 'access', 'realpath']);
		// The one file that runs a command, for "Mark as read". It may import child_process and
		// call execFile; every other name below stays banned there as everywhere.
		const runsCommand = path.join('lib', 'server', 'command.ts');
		const importFrom = /(?:import|export)\s+([^;]*?)\s+from\s+['"](?:node:)?(fs|fs\/promises|child_process)['"]/g;
		const requireOf = /require\(\s*['"](?:node:)?(fs|fs\/promises|child_process)['"]\s*\)/;
		const wholeModule = /^(\*\s+as\s+\w+|\w+)$/;
		const names = [
			'writeFile', 'writeFileSync', 'appendFile', 'createWriteStream', 'mkdir', 'mkdirSync', 'rename',
			'unlink', 'rm', 'rmSync', 'rmdir', 'copyFile', 'cp', 'open', 'spawn', 'exec', 'execSync',
			'execFile', 'fork'
		];
		const call = new RegExp(`\\b(${names.join('|')})\\s*\\(`);
		const found: string[] = [];
		for (const file of files) {
			const text = readFileSync(file, 'utf8');
			const rel = path.relative(src, file);
			const importsFs = /['"](?:node:)?(?:fs|fs\/promises|child_process)['"]/.test(text);
			for (const m of text.matchAll(importFrom)) {
				const clause = m[1].trim();
				if (wholeModule.test(clause)) {
					found.push(`${rel}: a default or namespace import of ${m[2]}`);
					continue;
				}
				for (const raw of clause.replace(/[{}]/g, '').split(',')) {
					const name = raw.trim().split(/\s+as\s+/)[0];
					if (rel === runsCommand && m[2] === 'child_process' && name === 'execFile') continue;
					if (name && !allowedNames.has(name)) found.push(`${rel}: imports ${name} from ${m[2]}`);
				}
			}
			const required = requireOf.exec(text);
			if (required) found.push(`${rel}: requires ${required[1]}`);
			text.split('\n').forEach((line, i) => {
				const hit = call.exec(line);
				if (hit && hit[1] === 'execFile' && rel === runsCommand) return;
				// open( is a file call only where the file imports node:fs; window.open is not.
				if (hit && (hit[1] !== 'open' || importsFs)) found.push(`${rel}:${i + 1}: calls ${hit[1]}(`);
			});
		}
		expect(found).toEqual([]);
	});
});
