import adapter from '@sveltejs/adapter-node';
import { sveltekit } from '@sveltejs/kit/vite';
import tailwindcss from '@tailwindcss/vite';
import { defineConfig } from 'vite';

// The page has no form action, and hooks.server.ts refuses every write with a 405. SvelteKit's
// origin check would answer a bare write with a 403 first, so it is off here.
export default defineConfig({
	plugins: [tailwindcss(), sveltekit({ adapter: adapter(), csrf: { trustedOrigins: ['*'] } })]
});
