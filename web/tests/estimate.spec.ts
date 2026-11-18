import { expect, test } from './app';
import { block, board, show, widget } from './page';

test.describe('the estimate, in Alaska time, from a browser in London', () => {
	test('started: 8:00 pm, 6:00 pm to 9:30 pm, 2 questions, 2 rounds, alpha and beta', async ({ app, page }) => {
		await show(app, page, 'started');
		const w = widget(page);
		await expect(w.getByTestId('session-state')).toHaveText('Today, about 8:00 pm');
		await expect(w.getByTestId('session-range')).toHaveText('6:00 pm to 9:30 pm');
		await expect(w).toContainText('2 questions from 2 rounds, in alpha and beta');
		await expect(w).not.toContainText('Rough');
		await expect(page).toHaveTitle('Today, about 8:00 pm · cc-desk');
	});

	test('started: the front page states the same time, range, count, and projects', async ({ app, page }) => {
		await show(app, page, 'started');
		await expect(page.getByRole('heading', { level: 2, name: 'Your next decision session: Today, about 8:00 pm' })).toBeVisible();
		await expect(page.getByTestId('front-range')).toContainText('Likely range: 6:00 pm to 9:30 pm.');
		await expect(page.getByRole('main')).toContainText('2 questions from 2 rounds, in alpha and beta.');
	});

	test('started: Rough sits beside beta in the sidebar and on its board, and beside alpha nowhere', async ({ app, page }) => {
		await show(app, page, 'started');
		await expect(block(page, 'beta/v2')).toContainText('Rough');
		await expect(board(page, 'beta/v2')).toContainText('Rough');
		await expect(block(page, 'alpha/v1')).not.toContainText('Rough');
		await expect(board(page, 'alpha/v1')).not.toContainText('Rough');
	});

	test('fast: 6:30 pm with 4:30 pm to 7:00 pm', async ({ app, page }) => {
		await show(app, page, 'fast');
		await expect(widget(page).getByTestId('session-state')).toHaveText('Today, about 6:30 pm');
		await expect(widget(page).getByTestId('session-range')).toHaveText('4:30 pm to 7:00 pm');
	});

	test('slow: 9:30 pm with 8:00 pm to midnight', async ({ app, page }) => {
		await show(app, page, 'slow');
		await expect(widget(page).getByTestId('session-state')).toHaveText('Today, about 9:30 pm');
		await expect(widget(page).getByTestId('session-range')).toHaveText('8:00 pm to midnight');
	});

	test('variant: Rough in the widget, and the missing range as a dash read as None', async ({ app, page }) => {
		await show(app, page, 'variant');
		const w = widget(page);
		await expect(w.getByTestId('session-state')).toHaveText('Today, about 3:00 pm');
		await expect(w).toContainText('Rough');
		const range = w.getByTestId('session-range');
		await expect(range.locator('[aria-hidden="true"]')).toHaveText('–');
		await expect(range.locator('.sr-only')).toHaveText('None');
	});
});
