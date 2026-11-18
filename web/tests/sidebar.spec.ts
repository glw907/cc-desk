import { expect, test } from './app';
import { block, show } from './page';

test.describe('the sidebar on started', () => {
	test.beforeEach(async ({ app, page }) => {
		await show(app, page, 'started');
	});

	// Worked by hand from the printed snapshot: alpha's rounds are 1200 and 360 planned minutes
	// with the first done, 1200 / 1560 = 76.9; beta's are 90, 480, and 90 with the first done,
	// 90 / 660 = 13.6.
	test('each release block counts its rounds and its percent', async ({ page }) => {
		await expect(block(page, 'alpha/v1')).toContainText('1 of 2 rounds done');
		await expect(block(page, 'alpha/v1')).toContainText('77% of the work');
		await expect(block(page, 'beta/v2')).toContainText('1 of 3 rounds done');
		await expect(block(page, 'beta/v2')).toContainText('14% of the work');
	});

	test('a triangle sits under alpha r2 and beta b3, and under no other segment', async ({ page }) => {
		const alpha = block(page, 'alpha/v1').locator('.scale').first().locator('.seg');
		await expect(alpha).toHaveCount(2);
		await expect(alpha.nth(0)).not.toHaveClass(/asks/);
		await expect(alpha.nth(1)).toHaveClass(/asks/);

		const beta = block(page, 'beta/v2').locator('.scale').first().locator('.seg');
		await expect(beta).toHaveCount(3);
		await expect(beta.nth(0)).not.toHaveClass(/asks/);
		await expect(beta.nth(1)).not.toHaveClass(/asks/);
		await expect(beta.nth(2)).toHaveClass(/asks/);
	});

	test('the building round hangs under alpha, and beta has none', async ({ page }) => {
		const round = block(page, 'alpha/v1').locator('[data-round="alpha/r2"]');
		await expect(round).toContainText('0 of 3 passes done');
		await expect(round.locator('.figures .state')).toHaveText('Building');
		await expect(block(page, 'beta/v2').locator('[data-round]')).toHaveCount(0);
	});

	test('Waiting for you reads Nothing', async ({ page }) => {
		const waits = page.getByRole('region', { name: 'Waiting for you' });
		await expect(waits).toContainText('Nothing');
		await expect(waits.getByRole('link')).toHaveCount(0);
	});
});

test('on big, Waiting for you lists the two rounds and links to each', async ({ app, page }) => {
	await show(app, page, 'big');
	// From the printed snapshot: north/n1 has 2 questions and south/s1 has 1.
	const north = page.getByRole('link', { name: 'north/n1 2 questions' });
	await expect(north).toHaveAttribute('href', '/north/n1');
	const south = page.getByRole('link', { name: 'south/s1 1 question' });
	await expect(south).toHaveAttribute('href', '/south/s1');
});
