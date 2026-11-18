import { expect, test } from './app';
import { started } from './state';

test.describe('the server only reads', () => {
	test.beforeEach(async ({ app }) => {
		started(app.stateDir);
		await app.start();
	});

	for (const method of ['POST', 'PUT']) {
		for (const route of ['/', '/api/state']) {
			test(`${method} ${route} is refused with 405`, async ({ app }) => {
				// SvelteKit's own form check answers 403, ahead of the hook, to a write with no
				// content type. A JSON write from the page's origin reaches the hook and gets the 405.
				const res = await fetch(app.url + route, {
					method,
					headers: { origin: app.url, 'content-type': 'application/json' },
					body: '{}'
				});
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
