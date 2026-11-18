import { chmodSync, existsSync, mkdtempSync, rmSync, writeFileSync } from 'node:fs';
import { tmpdir } from 'node:os';
import path from 'node:path';
import { expect, test } from './app';
import { hashTree } from './hash';
import { buildState, started } from './state';

test.describe('the server only reads', () => {
	test.beforeEach(async ({ app }) => {
		started(app.stateDir);
		await app.start();
	});

	for (const method of ['POST', 'PUT', 'PATCH', 'DELETE']) {
		for (const route of ['/', '/api/state', '/reports/rep-0001']) {
			test(`${method} ${route} from the page's origin is refused with 405`, async ({ app }) => {
				const res = await fetch(app.url + route, { method, headers: { origin: app.url } });
				expect(res.status).toBe(405);
				expect(res.headers.get('allow')).toBe('GET, HEAD');
			});

			test(`${method} ${route} with no origin is refused with 403`, async ({ app }) => {
				// Kit's origin check answers a form-typed or untyped write before the hook runs.
				expect((await fetch(app.url + route, { method })).status).toBe(403);
			});
		}
	}

	test('HEAD answers 200', async ({ app }) => {
		expect((await fetch(app.url + '/', { method: 'HEAD' })).status).toBe(200);
		expect((await fetch(app.url + '/api/state', { method: 'HEAD' })).status).toBe(200);
	});

	test('GET /api/state answers the four parts', async ({ app }) => {
		const body = await (await fetch(app.url + '/api/state')).json();
		expect(Object.keys(body).sort()).toEqual(['error', 'lastGood', 'noState', 'snapshot']);
		expect(body.noState).toBe(false);
		expect(body.error).toBeNull();
		expect(body.lastGood).toBe('2026-11-20T23:00:00Z');
		expect(body.snapshot.next_session.state).toBe('scheduled');
	});
});

test.describe('the page state', () => {
	test.beforeEach(async ({ app }) => {
		buildState('page', app.stateDir);
		await app.start();
	});

	test('GET /reports/rep-0001 answers 200', async ({ app }) => {
		expect((await fetch(app.url + '/reports/rep-0001')).status).toBe(200);
	});

	test("beta's b2 round page links its report", async ({ page, app }) => {
		await page.goto(app.url + '/beta/b2');
		const link = page.locator('main').getByRole('link', { name: 'Round b2: the report verbs' });
		await expect(link).toHaveAttribute('href', '/reports/rep-0001');
		await expect(page.getByText('Not read yet')).toBeVisible();
	});

	test('the report page shows its calls, its misses, and the folded detail', async ({ page, app }) => {
		await page.goto(app.url + '/reports/rep-0001');
		await expect(page.getByRole('heading', { level: 1 })).toHaveText('Round b2: the report verbs');
		await expect(page.getByRole('heading', { name: 'Calls made along the way' })).toBeVisible();
		await expect(page.getByText('Where the report lives.')).toBeVisible();
		await expect(page.getByRole('heading', { name: 'What the plan missed' })).toBeVisible();
		await expect(page.getByText('The plan said every refusal exits 1.')).toBeVisible();
		await expect(page.locator('details')).not.toHaveAttribute('open', '');
	});
});

/** A form post, the way the page's own form sends it. */
const formPost = (origin: string | null) => ({
	method: 'POST',
	headers: { 'content-type': 'application/x-www-form-urlencoded', ...(origin ? { origin } : {}) },
	body: ''
});

test.describe('Mark as read', () => {
	test.beforeEach(async ({ app }) => {
		buildState('page', app.stateDir);
		await app.start();
	});

	test('one click changes only the report and the snapshot, then the report leaves the list', async ({
		app,
		page
	}) => {
		const before = hashTree(app.stateDir);
		await page.goto(app.url + '/reports/rep-0001');
		await page.getByRole('button', { name: 'Mark as read' }).click();

		const reports = page.getByRole('region', { name: 'Reports to read' });
		await expect(reports.getByRole('link')).toHaveCount(0, { timeout: 30_000 });
		await expect(page.getByRole('button', { name: 'Mark as read' })).toHaveCount(0);

		const after = hashTree(app.stateDir);
		expect(Object.keys(after).sort()).toEqual(Object.keys(before).sort());
		const changed = Object.keys(after).filter((file) => after[file] !== before[file]);
		expect(changed.sort()).toEqual(['reports.json', 'snapshot.json']);
		expect(after['usage.json']).toBe(before['usage.json']);

		// The report stays on its round's page, now read. The test asserts no time it shows.
		await page.goto(app.url + '/beta/b2');
		await expect(page.locator('main').getByRole('link', { name: 'Round b2: the report verbs' })).toBeVisible();
		await expect(page.getByText('Not read yet')).toHaveCount(0);
	});

	test('a post from another origin is refused with 403 and changes nothing', async ({ app }) => {
		const before = hashTree(app.stateDir);
		const foreign = await fetch(app.url + '/reports/rep-0001?/read', formPost('http://evil.example'));
		expect(foreign.status).toBe(403);
		const bare = await fetch(app.url + '/reports/rep-0001?/read', { method: 'POST' });
		expect(bare.status).toBe(403);
		expect(hashTree(app.stateDir)).toEqual(before);
	});
});

test.describe('Mark as read with a binary that fails', () => {
	let marker = '';

	test.beforeEach(async ({ app }) => {
		const dir = mkdtempSync(path.join(tmpdir(), 'cc-desk-stub-'));
		marker = path.join(dir, 'ran');
		const stub = path.join(dir, 'cc-desk');
		writeFileSync(stub, `#!/bin/sh\ntouch '${marker}'\necho 'the stub refuses' >&2\nexit 1\n`);
		chmodSync(stub, 0o755);
		buildState('page', app.stateDir);
		app.env = { CC_DESK_BINARY: stub };
		await app.start();
	});

	test.afterEach(() => {
		rmSync(path.dirname(marker), { recursive: true, force: true });
	});

	test('a failed command shows the alert and keeps the report listed', async ({ app, page }) => {
		await page.goto(app.url + '/reports/rep-0001');
		await page.getByRole('button', { name: 'Mark as read' }).click();
		await expect(page.getByRole('alert')).toHaveText("Mark as read couldn't be recorded: the stub refuses");
		await expect(page.getByRole('button', { name: 'Mark as read' })).toBeVisible();
		await expect(
			page.getByRole('region', { name: 'Reports to read' }).getByRole('link')
		).toHaveCount(1);
		expect(existsSync(marker)).toBe(true);
	});

	test('a name with a dash or a slash is refused before any command runs', async ({ app }) => {
		// The control: a good name reaches the command, which fails as the stub is made to, and
		// the marker proves the stub can run.
		const good = await fetch(app.url + '/reports/rep-0001?/read', {
			...formPost(app.url),
			headers: { ...formPost(app.url).headers, accept: 'application/json', 'x-sveltekit-action': 'true' }
		});
		expect(good.status).toBe(500);
		expect(existsSync(marker)).toBe(true);
		rmSync(marker);

		for (const address of ['/reports/-x?/read', '/reports/a/b?/read']) {
			const res = await fetch(app.url + address, formPost(app.url));
			expect(res.status, address).toBe(405);
			expect(res.headers.get('allow'), address).toBe('GET, HEAD');
		}
		expect(existsSync(marker)).toBe(false);
	});

	test('a write with a JSON body or no body is refused with 415 and runs no command', async ({ app }) => {
		const json = await fetch(app.url + '/reports/rep-0001?/read', {
			method: 'POST',
			headers: { 'content-type': 'application/json', origin: app.url },
			body: '{}'
		});
		expect(json.status).toBe(415);
		const none = await fetch(app.url + '/reports/rep-0001?/read', {
			method: 'POST',
			headers: { origin: app.url }
		});
		expect(none.status).toBe(415);
		expect(existsSync(marker)).toBe(false);
	});
});
