import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// https://vite.dev/config/
export default defineConfig({
  plugins: [react()],
   server: {
    host: '172.16.0.196',
    port: 5173,
    watch: {
      usePolling: true
    }
  }
})
