import type { Config } from 'stylelint';

export default {
  extends: [
    '@stylistic/stylelint-config',
    'stylelint-config-standard-vue',
    'stylelint-config-tailwindcss',
  ],
  plugins: [
    'stylelint-plugin-use-baseline',
  ],
  rules: {
    'plugin/use-baseline': [true, {
      available: 'widely',
      // Tailwind CSS のエンジンである Lightning CSS が自動で対応できる機能は無視する
      ignoreSelectors: [
        'nesting',
        'dir',
        'autofill',
        'fullscreen',
        'selection',
      ],
      ignoreProperties: {
        '/^mask/': [],
        'backdrop-filter': [],
        'hyphens': [],
        'box-decoration-break': [],
        'user-select': [],
        'text-size-adjust': [],
        'print-color-adjust': [],
        'hyphenate-character': [],
        'hyphenate-limit-chars': [],
        'text-emphasis-position': [],
        'text-emphasis-style': [],
        'initial-letter': [],
        'background-clip': ['text'],
        'background-image': ['/^image-set/'],
        'content': ['/^image-set/'],
        'width': ['stretch'],
        'min-width': ['stretch'],
        'max-width': ['stretch'],
        'height': ['stretch'],
        'min-height': ['stretch'],
        'max-height': ['stretch'],
        'image-rendering': ['crisp-edges'],
        'float': ['inline-start', 'inline-end'],
        'clear': ['inline-start', 'inline-end'],
      },
      ignoreAtRules: [
        'custom-media',
      ],
      ignoreFunctions: [
        'image-set',
      ],
    }],
  },
  reportDescriptionlessDisables: true,
} satisfies Config;
