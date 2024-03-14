import { fileURLToPath, URL } from "node:url";

import { defineConfig } from "vite";
import vue from "@vitejs/plugin-vue";
import svgLoader from 'vite-svg-loader';

// https://vitejs.dev/config/
export default defineConfig({
	plugins: [
		vue(),
		svgLoader({
		defaultImport: 'component',
		}),
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
});
