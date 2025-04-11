import js from '@eslint/js'
import typescript from '@typescript-eslint/parser'
import { defineConfig } from 'eslint/config'
import globals from 'globals'

export default defineConfig([
  {
    files: ['**/*.{ts,tsx}'],
    languageOptions: {
      parser: typescript,
      globals: globals.browser,
    },
  },
  {
    files: ['**/*.{js,mjs,cjs}'],
    languageOptions: {
      globals: globals.browser,
    },
  },
  {
    files: ['**/*.{js,ts,mjs,cjs,tsx}'],
    plugins: { js },
    extends: ['js/recommended'],
    rules: {
      eqeqeq: ['error', 'always'],
    },
  },
])
