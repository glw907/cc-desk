// Reads a round's design file and renders it. The file is the checkout's, so its text is data:
// it is rendered with raw HTML off, images off, and only http and https links kept.
import { readFile, realpath, stat } from 'node:fs/promises';
import path from 'node:path';
import MarkdownIt from 'markdown-it';
import { stateDir } from './state.ts';

/** The largest design file the page renders. */
export const MAX_DESIGN_BYTES = 256 * 1024;

/** A project name as `internal/state/events.go` allows it. */
const PROJECT = /^[a-z0-9][a-z0-9._-]*$/;

const md = new MarkdownIt({ html: false, linkify: false });
md.disable('image');
md.validateLink = (url) => /^https?:\/\//i.test(url.trim());

/** The design as markup, or the sentence saying why there is none. */
export type DesignRead = { html: string } | { why: string };

/** Starts the sentence shown when a design can't be read. */
export const NOT_READABLE = "The design file isn't readable here:";

const refuse = (why: string): DesignRead => ({ why: `${NOT_READABLE} ${why}.` });

/** Whether `inner` is `outer` or sits under it. */
function within(outer: string, inner: string): boolean {
	const rel = path.relative(outer, inner);
	return rel === '' || (!rel.startsWith('..') && !path.isAbsolute(rel));
}

/**
 * Reads `design`, a path relative to the project's checkout, and renders it. The real path of the
 * file must sit under the real path of the checkout. The read goes real path, then stat, then
 * read, so a file swapped between those steps could pass; that small race is accepted.
 */
export async function readDesign(
	project: string,
	design: string,
	dir: string = stateDir()
): Promise<DesignRead> {
	if (!PROJECT.test(project)) return refuse('the project name is not one cc-desk records');

	let checkout: string;
	try {
		const release = JSON.parse(await readFile(path.join(dir, 'releases', `${project}.json`), 'utf8')) as {
			checkout?: unknown;
		};
		if (typeof release.checkout !== 'string' || release.checkout === '') {
			return refuse('the release records no checkout');
		}
		checkout = release.checkout;
	} catch {
		return refuse('the release file could not be read');
	}

	try {
		const root = await realpath(checkout);
		const file = await realpath(path.resolve(root, design));
		if (!within(root, file)) return refuse('it lies outside the checkout');
		const info = await stat(file);
		if (!info.isFile()) return refuse('it is not a regular file');
		if (info.size > MAX_DESIGN_BYTES) return refuse('it is larger than 256 KiB');
		return { html: md.render(await readFile(file, 'utf8')) };
	} catch {
		return refuse('the file is missing');
	}
}
