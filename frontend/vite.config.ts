import { defineConfig } from 'vitest/config';
import { sveltekit } from '@sveltejs/kit/vite';
import tailwindcss from '@tailwindcss/vite';

export default defineConfig({
	plugins: [tailwindcss(), sveltekit()],
	build: {
		rollupOptions: {
			// This tells Rollup to treat these files as external
			// so it doesn't try to bundle them.
			external: ['/convex/rooms.ts']
		}
	},
	test: {
		include: ['src/**/*.{test,spec}.{js,ts}']
	}
});
