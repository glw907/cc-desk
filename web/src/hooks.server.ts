import type { Handle } from '@sveltejs/kit/hooks';

// The page only reads: every method but GET and HEAD is refused before a route sees it.
export const handle: Handle = async ({ event, resolve }) => {
	if (event.request.method !== 'GET' && event.request.method !== 'HEAD') {
		return new Response('Method Not Allowed', { status: 405, headers: { allow: 'GET, HEAD' } });
	}
	return resolve(event);
};
