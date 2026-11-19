import type { Handle } from '@sveltejs/kit/hooks';

const READ = /^\/reports\/rep-\d{4,}$/;

// The page only reads, with one exception: "Mark as read" on a report. Every other method but
// GET and HEAD is refused before a route sees it. Kit's origin check runs first, so a write from
// another origin gets its 403 before this hook runs.
export const handle: Handle = async ({ event, resolve }) => {
	const { method } = event.request;
	if (method === 'GET' || method === 'HEAD') return resolve(event);
	const { pathname, search } = event.url;
	if (method === 'POST' && READ.test(pathname) && search === '?/read') return resolve(event);
	return new Response('Method Not Allowed', { status: 405, headers: { allow: 'GET, HEAD' } });
};
