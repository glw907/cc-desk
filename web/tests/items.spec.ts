import type { Page } from '@playwright/test';
import { expect, test } from './app';
import { show } from './page';

// Every expected string below is written out here, never taken from the app's formatters. The
// clock is 2026-11-20T23:00:00Z, 2:00 pm in Alaska, and the browser's zone is Europe/London.

/** A figure on an item page, found by its label. */
const figure = (page: Page, label: string) => page.getByLabel(label, { exact: true });

/** The en dash a null figure shows, with "None" for a screen reader. */
async function expectEmpty(page: Page, label: string) {
	await expect(figure(page, label)).toContainText('–');
	await expect(figure(page, label).locator('.sr-only')).toHaveText('None');
}

test.describe('a round page on twin', () => {
	test('alder/r1 shows its own figures, its reason, its passes, and its command', async ({ app, page }) => {
		await show(app, page, 'twin', '/alder/r1');
		await expect(page.getByRole('heading', { level: 1, name: 'r1' })).toBeVisible();
		await expect(figure(page, 'Planned')).toHaveText('100 minutes');
		await expect(figure(page, 'Work before it')).toHaveText('0 minutes');
		await expect(figure(page, 'Needed by')).toHaveText('Today, 2:00 pm');
		await expectEmpty(page, 'Actual');
		await expect(page.locator('.state-line .state')).toHaveText('Not started');
		await expect(page.getByText('Nothing. Routine upkeep.')).toBeVisible();
		await expect(page.getByRole('link', { name: 'a-one' })).toHaveAttribute('href', '/alder/r1/a-one');
		await expect(page.getByRole('link', { name: 'a-two' })).toHaveAttribute('href', '/alder/r1/a-two');
		await expect(page.locator('main code')).toHaveText('cc-desk alder r1');
	});

	test('birch/r1 has the same name and other figures', async ({ app, page }) => {
		await show(app, page, 'twin', '/birch/r1');
		await expect(page.getByRole('heading', { level: 1, name: 'r1' })).toBeVisible();
		await expect(figure(page, 'Planned')).toHaveText('90 minutes');
		await expect(figure(page, 'Work before it')).toHaveText('120 minutes');
		await expect(figure(page, 'Needed by')).toHaveText('Today, 4:00 pm');
		await expect(page.getByText('Keep the long names?')).toBeVisible();
		await expect(page.locator('main code')).toHaveText('cc-desk birch r1');
	});

	test('birch/r0 shows an answered question with Ruled on', async ({ app, page }) => {
		await show(app, page, 'twin', '/birch/r0');
		await expect(page.getByRole('heading', { level: 1, name: 'r0' })).toBeVisible();
		await expect(page.getByRole('listitem').filter({ hasText: 'Start with the setup round?' })).toContainText(
			'Ruled on'
		);
	});

	test('birch/r2 reads Not designed yet with its forecast', async ({ app, page }) => {
		await show(app, page, 'twin', '/birch/r2');
		await expect(page.getByRole('heading', { level: 1, name: 'r2' })).toBeVisible();
		await expect(page.getByText('Not designed yet.')).toBeVisible();
		await expect(figure(page, 'Expected design')).toHaveText('Today, about 5:30 pm');
		await expect(figure(page, 'Planned')).toHaveText('30 minutes');
		await expect(figure(page, 'Work before it')).toHaveText('210 minutes');
	});
});

test.describe('a round page on started', () => {
	test('alpha/r1 is done, with actual minutes and empty fields', async ({ app, page }) => {
		await show(app, page, 'started', '/alpha/r1');
		await expect(page.getByRole('heading', { level: 1, name: 'r1' })).toBeVisible();
		await expect(page.locator('main .state').first()).toHaveText('Done');
		await expect(figure(page, 'Actual')).toHaveText('1,800 minutes');
		await expectEmpty(page, 'Work before it');
		await expectEmpty(page, 'Needed by');
		await expect(page.getByText('Nothing. No reason recorded')).toBeVisible();
	});

	test('beta/b1 reads Nothing with its reason', async ({ app, page }) => {
		await show(app, page, 'started', '/beta/b1');
		await expect(page.getByText('Nothing. Routine cleanup.')).toBeVisible();
	});

	test('beta/b2 shows Ruled on and Carried over', async ({ app, page }) => {
		await show(app, page, 'started', '/beta/b2');
		await expect(page.getByRole('listitem').filter({ hasText: 'Keep the old format?' })).toContainText(
			'Ruled on'
		);
		await expect(page.getByRole('listitem').filter({ hasText: 'Name the tool?' })).toContainText(
			'Carried over'
		);
	});
});

test.describe('a pass page', () => {
	test('alpha/r2/p14 shows its figures, a link to its round, and the round command', async ({ app, page }) => {
		await show(app, page, 'started', '/alpha/r2/p14');
		await expect(page.getByRole('heading', { level: 1, name: 'p14' })).toBeVisible();
		await expect(figure(page, 'Planned')).toHaveText('200 minutes');
		await expect(figure(page, 'Work before it')).toHaveText('100 minutes');
		await expect(figure(page, 'Needed by')).toHaveText('Today, 4:00 pm');
		await expect(page.locator('main .state')).toHaveText('Not started');
		await expect(page.locator('main .doc-meta').getByRole('link')).toHaveAttribute('href', '/alpha/r2');
		await expect(page.locator('main code')).toHaveText('cc-desk alpha r2');
	});

	test('alpha/r2/p15 shows its written question', async ({ app, page }) => {
		await show(app, page, 'started', '/alpha/r2/p15');
		await expect(page.getByText('Which storage should the report use?')).toBeVisible();
	});

	test('a stopped pass shows its reason and no since', async ({ app, page }) => {
		await show(app, page, 'stopped', '/alpha/r2/p13');
		await expect(page.locator('main .state')).toHaveText('Stopped');
		await expect(page.getByText('A change of scope')).toBeVisible();
		await expect(page.locator('main')).not.toContainText('since');
	});
});
