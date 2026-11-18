import { readFileSync, writeFileSync } from 'node:fs';
import path from 'node:path';
import type { Page } from '@playwright/test';
import { expect, test, type App } from './app';
import { widget } from './page';
import { copyFixture, run, runFailing, writeAtomic, WORKED_CLOCK } from './state';

// The browser's clock stays at 23:00Z (2:00 pm Alaska time) in every test here, so a page that
// read the clock for its answer, rather than the server's snapshot, would not redraw.
const T0 = WORKED_CLOCK;
const T5 = '2026-11-20T23:05:00Z';

/** Opens the page on a fresh `state` copy, with the browser's clock at 23:00Z. */
async function open(app: App, page: Page) {
	copyFixture('state', app.stateDir);
	await page.clock.setFixedTime(new Date(T0));
	await app.start();
	await page.goto(app.url);
}

test('a change to the state redraws the page within 5 seconds', async ({ app, page }) => {
	await open(app, page);
	const w = widget(page);
	await expect(w.getByTestId('session-state')).toHaveText('Today, about 8:00 pm');
	await expect(w.getByTestId('session-range')).toHaveText('6:00 pm to 9:30 pm');
	const row = page.locator('main [data-pass="alpha/p13"]');
	// The design opens only a running round to its passes, so p13 has no row until it starts.
	await expect(row).toHaveCount(0);

	run(app.stateDir, T5, 'pass', 'start', 'alpha', 'p13');

	await expect(w.getByTestId('session-state')).toHaveText('Today, about 8:05 pm', { timeout: 5000 });
	await expect(w.getByTestId('session-range')).toHaveText('6:05 pm to 9:35 pm');
	await expect(row).toContainText('Building');
});

test('an error file shows Stale over the last good snapshot', async ({ app, page }) => {
	await open(app, page);
	await expect(widget(page).getByTestId('session-state')).toHaveText('Today, about 8:00 pm');

	const alpha = path.join(app.stateDir, 'releases', 'alpha.json');
	writeFileSync(alpha, readFileSync(alpha, 'utf8').replace('{', '{,'));
	runFailing(app.stateDir, T5, 'pass', 'start', 'alpha', 'p13');

	const banner = page.locator('[data-banner="stale"]');
	await expect(banner).toContainText('Stale', { timeout: 5000 });
	await expect(banner).toContainText(
		"releases/alpha.json: invalid character ',' looking for beginning of object key string"
	);
	await expect(banner).toContainText('Today, 2:00 pm');
	await expect(widget(page).getByTestId('session-state')).toHaveText('Today, about 8:00 pm');
});

test('mending the file and taking a snapshot clears the Stale banner', async ({ app, page }) => {
	await open(app, page);
	const alpha = path.join(app.stateDir, 'releases', 'alpha.json');
	const good = readFileSync(alpha, 'utf8');
	writeFileSync(alpha, good.replace('{', '{,'));
	runFailing(app.stateDir, T5, 'pass', 'start', 'alpha', 'p13');
	await expect(page.locator('[data-banner="stale"]')).toBeVisible({ timeout: 5000 });

	writeFileSync(alpha, good);
	run(app.stateDir, T0, 'snapshot');

	await expect(page.locator('[data-banner="stale"]')).toHaveCount(0, { timeout: 5000 });
});

test('a snapshot.json that is not JSON shows Stale over the last snapshot', async ({ app, page }) => {
	await open(app, page);
	await expect(widget(page).getByTestId('session-state')).toHaveText('Today, about 8:00 pm');

	writeAtomic(app.stateDir, 'snapshot.json', 'this is not JSON');

	const banner = page.locator('[data-banner="stale"]');
	await expect(banner).toContainText('Stale', { timeout: 5000 });
	await expect(banner).toContainText("snapshot.json isn't valid JSON");
	await expect(banner).toContainText('Today, 2:00 pm');
	await expect(widget(page).getByTestId('session-state')).toHaveText('Today, about 8:00 pm');
});

test('with no snapshot the page reads No state yet, and nothing else until a bad write', async ({ app, page }) => {
	copyFixture('order', app.stateDir);
	await page.clock.setFixedTime(new Date(T0));
	await app.start();
	await page.goto(app.url);
	await expect(page).toHaveTitle('No state yet · cc-desk');
	const banner = page.locator('[data-banner="no-state"]');
	await expect(banner).toHaveText('No state yet');
	await expect(page.locator('[data-banner="stale"]')).toHaveCount(0);

	writeAtomic(app.stateDir, 'snapshot.json', 'this is not JSON');

	await expect(banner).toContainText("snapshot.json isn't valid JSON", { timeout: 5000 });
	await expect(banner).toContainText('No state yet');
});

test('a stopped server shows Not connected, and a restarted one clears it', async ({ app, page }) => {
	await open(app, page);
	await expect(widget(page).getByTestId('session-state')).toHaveText('Today, about 8:00 pm');

	await app.stop();

	const banner = page.locator('[data-banner="not-connected"]');
	await expect(banner).toContainText('Not connected', { timeout: 10000 });
	await expect(banner).toContainText('Today, 2:00 pm');
	await expect(widget(page).getByTestId('session-state')).toHaveText('Today, about 8:00 pm');

	await app.start();

	await expect(banner).toHaveCount(0, { timeout: 10000 });
});
