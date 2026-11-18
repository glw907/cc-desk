import { expect, test } from './app';

test('the built app opens and names itself', async ({ app, page }) => {
	await app.start();
	await page.goto(app.url);
	await expect(page.getByText('cc-desk')).toBeVisible();
});
