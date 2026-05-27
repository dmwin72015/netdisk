/// <reference types="vitest/config" />
import { sveltekit } from '@sveltejs/kit/vite';
import tailwindcss from '@tailwindcss/vite';
import { defineConfig } from 'vite';
import { paraglideVitePlugin } from '@inlang/paraglide-js';

export default defineConfig({
	plugins: [
		tailwindcss(),
		sveltekit(),
		paraglideVitePlugin({
			project: './project.inlang',
			outdir: './src/lib/paraglide',
			strategy: ['cookie', 'baseLocale']
		})
	],
	server: {
		port: 5173,
		proxy: {
			'/api': { target: 'http://localhost:8080', changeOrigin: true },
			'/hls': { target: 'http://localhost:8080', changeOrigin: true }
		}
	},
	test: {
		include: ['src/**/*.test.ts'],
		globals: true,
	}
});
