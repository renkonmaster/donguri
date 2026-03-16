import { join } from 'node:path';
import vue from '@vitejs/plugin-vue';
import tailwindcss from '@tailwindcss/vite';
import { defineConfig } from 'vite';

export default defineConfig({
  plugins: [
    vue(),
    tailwindcss(),
  ],
  resolve: {
    alias: {
      '@': join(import.meta.dirname, 'src'),
    },
  },
  server: {
    allowedHosts: true,
    proxy: {
      '/api': 'http://localhost:8080',
    },
  },
});
