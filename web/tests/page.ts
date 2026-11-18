// Opens the built app on a named state with the browser's clock pinned to that state's clock.
import type { Page } from '@playwright/test';
import type { App } from './app';
import { buildState } from './state';

/**
 * Fills the app's state folder with `name`, starts the server, pins the browser's clock to the
 * state's own clock before the page loads, and opens `address`, the front page by default. The zone is
 * `Europe/London` from the config, so a page that read the browser's zone would show the wrong hour.
 */
export async function show(
	app: App,
	page: Page,
	name: string,
	address = '/'
): Promise<string> {
	const clock = buildState(name, app.stateDir);
	await page.clock.setFixedTime(new Date(clock));
	await app.start();
	await page.goto(app.url + address);
	return clock;
}

/** The sidebar's forecast box. */
export const widget = (page: Page) => page.getByRole('region', { name: 'Next decision session', exact: true });

/** The sidebar block of one release, as `project/release`. */
export const block = (page: Page, release: string) => page.locator(`nav [data-release="${release}"]`);

/** The front page's board of one release, as `project/release`. */
export const board = (page: Page, release: string) => page.locator(`main [data-release="${release}"]`);
