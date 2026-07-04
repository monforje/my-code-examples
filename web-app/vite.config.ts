import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import path from 'path'

export default defineConfig({
  plugins: [react()],
  build: {
    rollupOptions: {
      output: {
        manualChunks(id) {
          if (id.includes("node_modules/react") || id.includes("node_modules/react-dom") || id.includes("node_modules/react-router")) return "react";
          if (id.includes("node_modules/@tanstack/react-query")) return "query";
          if (id.includes("node_modules/react-markdown")) return "markdown";
          if (id.includes("node_modules/@iconify")) return "iconify";
        },
      },
    },
  },
  server: {
    allowedHosts: ['codurity.ai', 'codurity.local'],
    port: 80,
    host: "0.0.0.0",
    proxy: {
      '/api/v1': {
        target: 'http://codurity.ai',
        changeOrigin: true,
      },
      '/uploads': {
        target: 'http://codurity.ai',
        changeOrigin: true,
      },
    },
  },
  resolve: {
    alias: {
      '@app': path.resolve(__dirname, 'src/app'),
      '@pages': path.resolve(__dirname, 'src/pages'),
      '@widgets': path.resolve(__dirname, 'src/widgets'),
      '@features': path.resolve(__dirname, 'src/features'),
      '@entities': path.resolve(__dirname, 'src/entities'),
      '@shared': path.resolve(__dirname, 'src/shared'),
    },
  },
})
