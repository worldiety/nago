/**
 * Copyright (c) 2025 worldiety GmbH
 *
 * This file is part of the NAGO Low-Code Platform.
 * Licensed under the terms specified in the LICENSE file.
 *
 * SPDX-License-Identifier: Custom-License
 */

module.exports = {
    root: true,
    env: {
        node: true,
        es2021: true,
        browser: true,
    },
    extends: ['plugin:vue/vue3-recommended', 'eslint:recommended', '@vue/typescript/recommended', 'prettier'],
    parserOptions: {
        ecmaVersion: 2021,
        project: ['./tsconfig.json'],
    },
    plugins: [],
    rules: {
        // allow to always explicitly define a type
        '@typescript-eslint/no-inferrable-types': 'off',
        // allow empty functions (for development), but show a warning
        '@typescript-eslint/no-empty-function': 'warn',
        // allow unused vars (for development), but show a warning
        '@typescript-eslint/no-unused-vars': 'warn',
        // automatically add type import
        '@typescript-eslint/consistent-type-imports': 'error',
        'vue/multi-word-component-names':'off'
    },
};

