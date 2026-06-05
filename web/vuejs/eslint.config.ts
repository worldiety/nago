import js from "@eslint/js";
import globals from "globals";
import tseslint from "typescript-eslint";
import pluginVue from "eslint-plugin-vue";
import { defineConfig } from "eslint/config";
import eslintConfigPrettier from "eslint-config-prettier/flat";

export default defineConfig([
	{
		ignores: ['**/*.d.ts', './src/shared/proto'],
	},
	{
		files: ['**/*.{js,mjs,cjs,ts,mts,cts,vue}'],
		plugins: { js },
		extends: ['js/recommended'],
		languageOptions: { globals: globals.browser },
	},
	tseslint.configs.recommended,
	pluginVue.configs['flat/recommended'],
	{ files: ['**/*.vue'], languageOptions: { parserOptions: { parser: tseslint.parser } } },
	eslintConfigPrettier,
	{
		rules: {
			'@typescript-eslint/no-explicit-any': 'off',
			'vue/multi-word-component-names': 'off',
			'vue/require-default-prop': 'off',
			'vue/valid-v-for': 'off', // TODO: Enable this rule, once NAGO is not dependant on wrong for loops anymore..
			'vue/require-v-for-key': 'off', // TODO: Enable this rule, once NAGO is not dependant on wrong for loops anymore..
		},
	},
]);
