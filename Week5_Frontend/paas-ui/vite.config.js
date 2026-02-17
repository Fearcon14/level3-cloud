import { fileURLToPath, URL } from 'node:url'

import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import vueDevTools from 'vite-plugin-vue-devtools'

// https://vite.dev/config/
export default defineConfig({
  plugins: [
    vue(),
    vueDevTools(),
  ],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url))
    },
  },

  server: {
    proxy: {
      '/api': {
        // In production, UI and API are behind the same ingress; /api is routed to the API.
        // For local dev, set VITE_API_PROXY_TARGET (e.g. your cluster LB IP or http://localhost:8080).
        target: process.env.VITE_API_PROXY_TARGET || 'http://localhost:8080',
        changeOrigin: true,
        secure: false,
        headers: process.env.VITE_API_PROXY_TARGET
          ? { Host: 'paas-api' }
          : undefined
      }
    }
  }
})
