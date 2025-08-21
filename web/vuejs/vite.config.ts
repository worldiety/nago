/**
 * Copyright (c) 2025 worldiety GmbH
 *
 * This file is part of the NAGO Low-Code Platform.
 * Licensed under the terms specified in the LICENSE file.
 *
 * SPDX-License-Identifier: Custom-License
 */
import vue from '@vitejs/plugin-vue';
import { URL, fileURLToPath } from 'node:url';
import path from 'path';
import { visualizer } from 'rollup-plugin-visualizer';
import type { PluginOption } from 'vite';
import { defineConfig } from 'vite';
import { viteStaticCopy } from 'vite-plugin-static-copy';
import vueDevTools from 'vite-plugin-vue-devtools';
import svgLoader from 'vite-svg-loader';

const basePlugins: PluginOption[] = [
	vue(),
	svgLoader({
		defaultImport: 'component',
	}),
	vueDevTools(),
];
const legacyPlugins: PluginOption[] = [
	viteStaticCopy({
		targets: [
			{
				src: 'node_modules/es5-polyfill/dist/polyfill.min.js',
				dest: 'assets',
			},
			{
				src: 'node_modules/core-js-bundle/minified.js',
				dest: 'assets',
			},
			{
				src: 'node_modules/regenerator-runtime/runtime.js',
				dest: 'assets',
			},
			{
				src: 'node_modules/bowser/es5.js',
				dest: 'assets',
			},
		],
	}),
];
const modernPlugins: PluginOption[] = [
	visualizer({
		filename: 'dist/bundle-report.html',
		gzipSize: true,
	}),
];

// https://vitejs.dev/config/
export default defineConfig(({ mode }) => {
	const isModern = mode === 'modern';

	return {
		plugins: [...basePlugins, ...(isModern ? modernPlugins : legacyPlugins)],
		server: {
			port: 8090,
			host: true,
			proxy: {
				'/api': 'http://localhost:3000',
				'/wire': {
					target: 'http://localhost:3000',
					ws: true,
				},
			},
		},
		resolve: {
			alias: {
				'@': isModern ? fileURLToPath(new URL('./src', import.meta.url)) : path.resolve(__dirname, 'src'),
			},
		},
		base: isModern ? '/modern/' : '/legacy/',
		build: {
			target: isModern ? 'esnext' : undefined,
			outDir: `dist/${isModern ? 'modern' : 'legacy'}`,
			manifest: `manifest.json`,
			rollupOptions: {
				input: 'src/main.ts',
				output: {
					format: isModern ? 'es' : 'iife',
				},
			},
			chunkSizeWarningLimit: isModern ? 600 : 5_000, // 600 KB, 5 MB
		},
	};
});
