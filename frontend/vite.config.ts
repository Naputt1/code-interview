import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'
import { tanstackRouter } from '@tanstack/router-plugin/vite'
import path from 'node:path'

const apiTarget = process.env.API_PROXY_TARGET ?? 'http://127.0.0.1:8080'

// https://vite.dev/config/
export default defineConfig({
  plugins: [
    // Must run before react() so generated routes are picked up.
    tanstackRouter({ target: 'react', autoCodeSplitting: true }),
    react(),
    tailwindcss(),
  ],
  resolve: {
    alias: {
      '@': path.resolve(import.meta.dirname, './src'),
    },
  },
  server: {
    host: true,
    port: Number(process.env.PORT ?? 5173),
    proxy: {
      // Single origin in dev: the browser calls /api and Vite forwards to Go.
      '/api': { target: apiTarget, changeOrigin: true },
    },
  },
})
