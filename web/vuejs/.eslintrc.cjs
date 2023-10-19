module.exports = {
  root: true,
  env: {
      node: true,
      es2021: true,
      browser: true,
  },
  extends: [
      'plugin:vue/vue3-recommended',
      'eslint:recommended',
      '@vue/typescript/recommended',
      'prettier',
  ],
  parserOptions: {
      ecmaVersion: 'latest',
  },
  plugins: [],
  rules: {
      // override/add rules settings here, such as:
      // 'vue/no-unused-vars': 'error'
      '@typescript-eslint/no-inferrable-types': 'off',
      'vue/component-tags-order': [
          'error',
          {
              order: ['script', 'template', 'style'],
          },
      ],
  },
};
