import { defineConfig } from 'vite'

export default defineConfig({
  build: {
    lib: {
      entry: 'internal/static/static/js/src/main.ts',
      name: 'main',
      formats: ['iife'],
      fileName: () => 'main.js',
    },
    outDir: 'internal/static/static/js/dist',
    emptyOutDir: false,
  },
})
