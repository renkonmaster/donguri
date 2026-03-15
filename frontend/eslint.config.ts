import { defineConfig } from 'eslint/config';
import type { Linter } from 'eslint';
import pluginVue from 'eslint-plugin-vue';
import tseslint from 'typescript-eslint';
import baselineJs from 'eslint-plugin-baseline-js';
import vueScopedCSS from 'eslint-plugin-vue-scoped-css';
import eslintCommentsConfigs from '@eslint-community/eslint-plugin-eslint-comments/configs';
import stylistic from '@stylistic/eslint-plugin';

export default defineConfig(
  pluginVue.configs['flat/recommended'] as Linter.Config[],
  tseslint.configs.recommended as Linter.Config[],
  vueScopedCSS.configs['flat/recommended'] as Linter.Config[],
  eslintCommentsConfigs.recommended as Linter.Config,
  stylistic.configs.customize({ semi: true }) as Linter.Config,
  {
    files: ['**/*.vue'],
    languageOptions: {
      parserOptions: {
        parser: tseslint.parser,
      },
    },
  },
  {
    plugins: {
      'baseline-js': { rules: baselineJs.rules },
    },
    rules: {
      'max-len': ['error', {
        code: 120,
        ignoreUrls: true,
      }],
      'baseline-js/use-baseline': ['error', {
        available: 'widely',
        // Vue が自動で対応できる機能は無視する
        ignoreFeatures: [
          'top-level-await',
        ],
      }],
      '@eslint-community/eslint-comments/disable-enable-pair': ['error', {
        allowWholeFile: true,
      }],
      '@eslint-community/eslint-comments/require-description': 'error',
    },
  },
  {
    files: ['**/*.vue'],
    rules: {
      'max-len': 'off',
      'vue/max-len': ['error', {
        code: 120,
        ignoreUrls: true,
        ignoreStrings: true,
        ignoreTemplateLiterals: true,
        ignoreRegExpLiterals: true,
        ignoreHTMLAttributeValues: true,
      }],
      'vue/block-order': ['error', {
        order: ['script', 'template', 'style'],
      }],
    },
  },
  {
    ignores: ['dist/**', 'node_modules/**'],
  },
);
