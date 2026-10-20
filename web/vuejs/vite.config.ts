/**
 * Copyright (c) 2025 worldiety GmbH
 *
 * This file is part of the NAGO Low-Code Platform.
 * Licensed under the terms specified in the LICENSE file.
 *
 * SPDX-License-Identifier: Custom-License
 */
import tailwindcss from '@tailwindcss/vite';
import vue from '@vitejs/plugin-vue';
import { copyFile, mkdir } from 'node:fs/promises';
import { URL, fileURLToPath } from 'node:url';
import path from 'path';
import { visualizer } from 'rollup-plugin-visualizer';
import type { PluginOption } from 'vite';
import { defineConfig } from 'vite';
import vueDevTools from 'vite-plugin-vue-devtools';
import svgLoader from 'vite-svg-loader';

// copyFiles copies the given files into the assets of the build, like the loader of the legacy build expects them.
function copyFiles(files: string[]): PluginOption {
	return {
		name: 'copy-files',
		apply: 'build',
		async writeBundle(options) {
			const dir = path.join(options.dir!, 'assets');
			await mkdir(dir, { recursive: true });
			await Promise.all(files.map((file) => copyFile(file, path.join(dir, path.basename(file)))));
		},
	};
}

const basePlugins: PluginOption[] = [
	tailwindcss(),
	vue(),
	svgLoader({
		defaultImport: 'component',
	}),
	vueDevTools(),
];
const legacyPlugins: PluginOption[] = [
	copyFiles([
		'node_modules/es5-polyfill/dist/polyfill.min.js',
		'node_modules/core-js-bundle/minified.js',
		'node_modules/regenerator-runtime/runtime.js',
		'node_modules/bowser/es5.js',
	]),
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
				// the package neither exports its modules nor declares to be free of side effects, so importing its
				// index would bundle all of its color schemes, while the color picker needs only HCT
				'@material/material-color-utilities/hct/hct.js': path.resolve(
					__dirname,
					'node_modules/@material/material-color-utilities/hct/hct.js'
				),
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
