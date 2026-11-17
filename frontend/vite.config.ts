/// <reference types="vitest/config" />
import tailwindcss from '@tailwindcss/vite'
import react from '@vitejs/plugin-react'
import { defineConfig } from 'vite'

export default defineConfig({
  plugins: [react(), tailwindcss()],
  resolve: { alias: { '@': '/src' } },
  // Same origin in dev: the session cookie (Path=/api, SameSite=Strict) just works.
  // changeOrigin stays false so Host matches the browser's Origin (the API skips
  // the CORS allowlist for same-origin requests).
  server: { proxy: { '/api': { target: 'http://localhost:8080', changeOrigin: false } } },
  test: {
    environment: 'jsdom',
    globals: true,
    setupFiles: ['./src/test/setup.ts'],
    include: ['src/**/*.test.{ts,tsx}'],
    css: false,
  },
})
