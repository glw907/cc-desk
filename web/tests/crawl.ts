// Walks the built page the way a reader does: open an address, wait for it to draw, follow every
// link on this site once. The walk and the scan share it so both see a page only once drawn.
import type { Page } from '@playwright/test';

/** What a page holds once it has drawn: a release in the sidebar, a heading, or the no-state banner. */
const DRAWN = 'nav [data-release], h1, [data-banner="no-state"]';

/**
 * Opens `address` on the app at `base` and waits for the page to draw: the first `/api/state`
 * response, then one of the elements a drawn page holds. The page's browser clock must already be
 * pinned, as `show()` does.
 */
export async function openDrawn(page: Page, base: string, address: string): Promise<void> {
	const answered = page.waitForResponse((r) => new URL(r.url()).pathname === '/api/state');
	await page.goto(base + address);
	await answered;
	await page.locator(DRAWN).first().waitFor();
}

/** Every address on this site that a page links to, as path and query, with no fragment. */
async function linksOf(page: Page, base: string): Promise<string[]> {
	const hrefs = await page.locator('a[href]').evaluateAll((as) => as.map((a) => (a as HTMLAnchorElement).href));
	const out = new Set<string>();
	for (const href of hrefs) {
		const url = new URL(href);
		if (url.origin === new URL(base).origin) out.add(url.pathname + url.search);
	}
	return [...out];
}

/**
 * Visits `/` and then every address reachable by links, each once, calling `onPage` after each page
 * has drawn. Returns the addresses visited, in order.
 */
export async function crawl(
	page: Page,
	base: string,
	onPage: (address: string) => Promise<void>
): Promise<string[]> {
	const seen = new Set<string>(['/']);
	const queue = ['/'];
	const visited: string[] = [];
	while (queue.length > 0) {
		const address = queue.shift() as string;
		await openDrawn(page, base, address);
		visited.push(address);
		await onPage(address);
		for (const next of await linksOf(page, base)) {
			if (!seen.has(next)) {
				seen.add(next);
				queue.push(next);
			}
		}
	}
	return visited;
}
