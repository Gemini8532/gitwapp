import { defineConfig, loadEnv } from 'vite'
import react from '@vitejs/plugin-react'

// https://vitejs.dev/config/
export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), '')
  const port = parseInt(env.APP_PORT || '8080')

  return {
    plugins: [react()],
    server: {
      proxy: {
        '/api': {
          target: `http://localhost:${port}`,
          changeOrigin: true,
        },
      },
    },
  }
})
