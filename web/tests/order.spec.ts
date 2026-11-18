import { readFileSync } from 'node:fs';
import path from 'node:path';
import { expect, test } from './app';
import { show, widget } from './page';
import { writeAtomic } from './state';

test('big: only the 400-minute dependent pass is left, so the widget reads Ready now', async ({ app, page }) => {
	await show(app, page, 'big');
	await expect(widget(page).getByTestId('session-state')).toHaveText('Ready now');
	await expect(page).toHaveTitle('Ready now · cc-desk');
});

test('guess: the front page lists the guess, and the total comes from the snapshot', async ({ app, page }) => {
	await show(app, page, 'guess');
	const main = page.getByRole('main');
	await expect(main.locator('[data-guess="north/n-guess"]')).toContainText('200 minutes');
	await expect(main).toContainText('200 of 300 minutes');

	const file = path.join(app.stateDir, 'snapshot.json');
	const edited = readFileSync(file, 'utf8').replace(/"guess_minutes_total": ?200/, '"guess_minutes_total": 250');
	expect(edited).toMatch(/"guess_minutes_total": ?250/);
	writeAtomic(app.stateDir, 'snapshot.json', edited);
	await expect(main).toContainText('250 of 300 minutes', { timeout: 5000 });
});

test('two: a pass held by the two-pass limit reads Waiting with that reason, in its row', async ({ app, page }) => {
	await show(app, page, 'two');
	for (const pass of ['north/n-two', 'south/s-two']) {
		const row = page.locator(`main [data-pass="${pass}"]`);
		await expect(row).toContainText('Waiting');
		await expect(row).toContainText('Two passes already running');
	}
});
