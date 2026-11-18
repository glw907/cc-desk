import { expect, test } from './app';
import { started } from './state';

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
