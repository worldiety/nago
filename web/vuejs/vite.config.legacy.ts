/**
 * Copyright (c) 2025 worldiety GmbH
 *
 * This file is part of the NAGO Low-Code Platform.
 * Licensed under the terms specified in the LICENSE file.
 *
 * SPDX-License-Identifier: Custom-License
 */

import { fileURLToPath, URL } from "node:url";

import { defineConfig } from "vite";
import vue from "@vitejs/plugin-vue";
import svgLoader from 'vite-svg-loader';
import vueDevTools from 'vite-plugin-vue-devtools';

// https://vitejs.dev/config/
export default defineConfig({
	plugins: [
		vue(),
		svgLoader({
			defaultImport: 'component',
		}),
		vueDevTools(),
	],
	server: {
		port: 8090,
		host: true,
		proxy: {
			'/api': 'http://localhost:3000',
			'/wire': 'http://localhost:3000',
		},
	},
	resolve: {
		alias: {
			"@": fileURLToPath(new URL("./src", import.meta.url))
		}
	},
	build: {
		target: 'es2015',
		outDir: 'dist/legacy',
		manifest: true,
	},
});
