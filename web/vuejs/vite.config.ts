/**
 * Copyright (c) 2025 worldiety GmbH
 *
 * This file is part of the NAGO Low-Code Platform.
 * Licensed under the terms specified in the LICENSE file.
 *
 * SPDX-License-Identifier: Custom-License
 */
import legacy from '@vitejs/plugin-legacy';
import vue from '@vitejs/plugin-vue';
import { URL, fileURLToPath } from 'node:url';
import path from 'path';
import { visualizer } from 'rollup-plugin-visualizer';
import { PluginOption, defineConfig } from 'vite';
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
	legacy({
		targets: ['defaults', 'not IE 11'],
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
			},
			chunkSizeWarningLimit: 600,
		},
	};
});
