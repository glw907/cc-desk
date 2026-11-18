// An axe-core scan with the WCAG 2.2 AA rules, on every page the page draws, in both themes.
// No rule is disabled and no element is left out of a scan.
import { rmSync, writeFileSync } from 'node:fs';
import path from 'node:path';
import AxeBuilder from '@axe-core/playwright';
import type { Page } from '@playwright/test';
import { expect, test } from './app';
import { crawl, openDrawn } from './crawl';
import { buildState } from './state';

const TAGS = ['wcag2a', 'wcag2aa', 'wcag21a', 'wcag21aa', 'wcag22aa'];
const THEMES = ['light', 'dark'] as const;

/** Scans the drawn page and returns one line per violating node: rule, address, first target. */
async function violations(page: Page, address: string): Promise<string[]> {
	const result = await new AxeBuilder({ page }).withTags(TAGS).analyze();
	return result.violations.flatMap((v) =>
		v.nodes.map((n) => `${v.id} on ${address}: ${String(n.target[0])}`)
	);
}

test('the two themes paint the page two different backgrounds', async ({ app, page }) => {
	const clock = buildState('started', app.stateDir);
	await page.clock.setFixedTime(new Date(clock));
	await app.start();
	const backgrounds: string[] = [];
	for (const colorScheme of THEMES) {
		await page.emulateMedia({ colorScheme });
		await openDrawn(page, app.url, '/');
		backgrounds.push(await page.evaluate(() => getComputedStyle(document.body).backgroundColor));
	}
	expect(backgrounds[0], 'light and dark share a background').not.toBe(backgrounds[1]);
});

/** One scan test per state and theme, so no test holds dozens of scans. */
for (const theme of THEMES) {
	// The states whose every reachable page is scanned.
	for (const state of ['started', 'twin', 'page']) {
		test(`every page on ${state} passes the scan in ${theme}`, async ({ app, page }) => {
			test.slow();
			const clock = buildState(state, app.stateDir);
			await page.clock.setFixedTime(new Date(clock));
			await page.emulateMedia({ colorScheme: theme });
			await app.start();
			const found: string[] = [];
			await crawl(page, app.url, async (address) => {
				// A folded design body or report table is scanned only once it is open.
				await page.locator('details').evaluateAll((all) => all.forEach((d) => ((d as HTMLDetailsElement).open = true)));
				found.push(...(await violations(page, address)));
			});
			expect(found).toEqual([]);
		});
	}

	// The states whose front page alone is scanned.
	for (const state of ['stopped', 'two', 'guess']) {
		test(`the front page on ${state} passes the scan in ${theme}`, async ({ app, page }) => {
			const clock = buildState(state, app.stateDir);
			await page.clock.setFixedTime(new Date(clock));
			await page.emulateMedia({ colorScheme: theme });
			await app.start();
			await openDrawn(page, app.url, '/');
			expect(await violations(page, '/')).toEqual([]);
		});
	}

	test(`the stale banner passes the scan in ${theme}`, async ({ app, page }) => {
		const clock = buildState('started', app.stateDir);
		writeFileSync(
			path.join(app.stateDir, 'snapshot-error.json'),
			JSON.stringify({ message: 'releases/alpha.json: invalid character' })
		);
		await page.clock.setFixedTime(new Date(clock));
		await page.emulateMedia({ colorScheme: theme });
		await app.start();
		await openDrawn(page, app.url, '/');
		await expect(page.locator('[data-banner="stale"]')).toBeVisible();
		expect(await violations(page, '/')).toEqual([]);
	});

	test(`No state yet passes the scan in ${theme}`, async ({ app, page }) => {
		const clock = buildState('started', app.stateDir);
		rmSync(path.join(app.stateDir, 'snapshot.json'));
		await page.clock.setFixedTime(new Date(clock));
		await page.emulateMedia({ colorScheme: theme });
		await app.start();
		await openDrawn(page, app.url, '/');
		await expect(page.locator('[data-banner="no-state"]')).toBeVisible();
		expect(await violations(page, '/')).toEqual([]);
	});

	for (const address of ['/alpha/zz', '/a/b/c/d']) {
		test(`${address}, which names nothing, passes the scan in ${theme}`, async ({ app, page }) => {
			const clock = buildState('started', app.stateDir);
			await page.clock.setFixedTime(new Date(clock));
			await page.emulateMedia({ colorScheme: theme });
			await app.start();
			await openDrawn(page, app.url, address);
			expect(await violations(page, address)).toEqual([]);
		});
	}
}
