import { expect, test } from './app';
import { show } from './page';

for (const address of ['/alpha/zz', '/zz/r1', '/alpha/r1/p13', '/a/b/c/d']) {
	test(`${address} shows not found, with a link to the front page`, async ({ app, page }) => {
		await show(app, page, 'started', address);
		await expect(page.getByRole('heading', { level: 1, name: 'Not found' })).toBeVisible();
		await expect(page.locator('main code')).toHaveText(address);
		await page.getByRole('link', { name: 'Go to the front page' }).click();
		await expect(page).toHaveURL(app.url + '/');
		await expect(page.locator('main [data-release]').first()).toBeVisible();
	});
}
