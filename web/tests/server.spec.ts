import { expect, test } from './app';
import { buildState, started } from './state';

test.describe('the server only reads', () => {
	test.beforeEach(async ({ app }) => {
		started(app.stateDir);
		await app.start();
	});

	for (const method of ['POST', 'PUT', 'PATCH', 'DELETE']) {
		for (const route of ['/', '/api/state']) {
			test(`${method} ${route} is refused with 405`, async ({ app }) => {
				// A bare write, with no origin or content type, still gets the hook's 405.
				const res = await fetch(app.url + route, { method });
				expect(res.status).toBe(405);
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
