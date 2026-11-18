import { expect, test } from './app';
import { show } from './page';

// The clock is 2026-11-20T23:00:00Z, 2:00 pm in Alaska. The browser's zone is Europe/London.

test('twin lists two rulings, the newest first', async ({ app, page }) => {
	await show(app, page, 'twin', '/rulings');
	await expect(page.getByRole('heading', { level: 1, name: 'Rulings' })).toBeVisible();
	const rows = page.locator('tbody tr');
	await expect(rows).toHaveCount(2);

	const first = rows.nth(0).locator('td');
	await expect(first.nth(0)).toHaveText('Today, 11:00 am');
	await expect(first.nth(1)).toContainText('–');
	await expect(first.nth(1).locator('.sr-only')).toHaveText('None');
	await expect(first.nth(2)).toHaveText('Hold the release a day.');

	const second = rows.nth(1).locator('td');
	await expect(second.nth(0)).toHaveText('Thu, Nov 19, 9:00 am');
	await expect(second.nth(1)).toHaveText('Start with the setup round?');
	await expect(second.nth(2)).toHaveText('Yes, setup first.');
});

test('started lists one ruling', async ({ app, page }) => {
	await show(app, page, 'started', '/rulings');
	const rows = page.locator('tbody tr');
	await expect(rows).toHaveCount(1);
	const cells = rows.nth(0).locator('td');
	await expect(cells.nth(0)).toHaveText('Wed, Nov 18, 7:00 am');
	await expect(cells.nth(1)).toHaveText('Keep the old format?');
	await expect(cells.nth(2)).toHaveText('Drop the old format.');
});

test('the rail links to the rulings page', async ({ app, page }) => {
	await show(app, page, 'started');
	await page.getByRole('navigation', { name: 'Status' }).getByRole('link', { name: 'Rulings' }).click();
	await expect(page).toHaveURL(app.url + '/rulings');
});
