import adapter from '@sveltejs/adapter-node';
import { sveltekit } from '@sveltejs/kit/vite';
import tailwindcss from '@tailwindcss/vite';
import { defineConfig } from 'vite';

// Kit's origin check stays on: a form post from another origin gets a 403 before hooks.server.ts
// runs, which matters for the one write the page takes, "Mark as read". The check runs in a
// production build only, and adapter-node needs ORIGIN set to the page's own address.
export default defineConfig({
	plugins: [tailwindcss(), sveltekit({ adapter: adapter() })]
});
