// The test fixture: one built app per test, on a free port, reading a state folder of its own.
import { test as base, expect, type Page } from '@playwright/test';
import { spawn, type ChildProcess } from 'node:child_process';
import { mkdtempSync, rmSync } from 'node:fs';
import { createServer } from 'node:net';
import { tmpdir } from 'node:os';
import path from 'node:path';

const buildDir = path.resolve(import.meta.dirname, '..', 'build');

export interface App {
	/** The state folder the server reads. The test fills it before `start()`. */
	stateDir: string;
	/** The address of the running server, set by `start()`. */
	url: string;
	/** What the server has written to stdout so far, since the last start. */
	stdout: string;
	/** What the server has written to stderr so far, since the last start. */
	stderr: string;
	/**
	 * Starts the built app, from `web/build` unless `appDir` names another copy of it. A taken port
	 * moves the app to a fresh one once.
	 */
	start(appDir?: string): Promise<void>;
	/** Stops the server and waits for it to exit. */
	stop(): Promise<void>;
}

async function freePort(): Promise<number> {
	return new Promise((resolve, reject) => {
		const probe = createServer();
		probe.once('error', reject);
		probe.listen(0, '127.0.0.1', () => {
			const { port } = probe.address() as { port: number };
			probe.close(() => resolve(port));
		});
	});
}

async function answers(url: string, child: ChildProcess): Promise<boolean> {
	for (let i = 0; i < 100; i++) {
		if (child.exitCode !== null) return false;
		try {
			await fetch(url, { method: 'HEAD' });
			return true;
		} catch {
			await new Promise((r) => setTimeout(r, 100));
		}
	}
	return false;
}

export const test = base.extend<{ app: App }>({
	app: async ({ page }: { page: Page }, use) => {
		const stateDir = mkdtempSync(path.join(tmpdir(), 'cc-desk-state-'));
		let child: ChildProcess | undefined;
		let port = 0;
		const app: App = {
			stateDir,
			url: '',
			stdout: '',
			stderr: '',
			async start(appDir = buildDir) {
				for (const attempt of [0, 1]) {
					const wanted = attempt === 0 && port !== 0 ? port : await freePort();
					app.stdout = '';
					app.stderr = '';
					child = spawn('node', [appDir], {
						env: { ...process.env, HOST: '127.0.0.1', PORT: String(wanted), CC_DESK_STATE_DIR: stateDir, SHUTDOWN_TIMEOUT: '1' },
						stdio: ['ignore', 'pipe', 'pipe']
					});
					child.stdout?.on('data', (chunk: Buffer) => (app.stdout += chunk.toString()));
					child.stderr?.on('data', (chunk: Buffer) => (app.stderr += chunk.toString()));
					const url = `http://127.0.0.1:${wanted}`;
					if (await answers(url, child)) {
						const moved = port !== 0 && wanted !== port;
						port = wanted;
						app.url = url;
						if (moved) await page.goto(url);
						return;
					}
					await app.stop();
				}
				throw new Error('the app did not start');
			},
			async stop() {
				const running = child;
				child = undefined;
				if (!running || running.exitCode !== null) return;
				const exited = new Promise<void>((r) => running.once('exit', () => r()));
				running.kill('SIGTERM');
				await exited;
			}
		};
		await use(app);
		await app.stop();
		rmSync(stateDir, { recursive: true, force: true });
	}
});

export { expect };
