import { defineConfig, loadEnv } from 'vite'
import { devtools } from '@tanstack/devtools-vite'
import { tanstackRouter } from '@tanstack/router-plugin/vite'
import viteReact from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'

const config = defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), '')
  return {
    resolve: { tsconfigPaths: true },
    plugins: [
      devtools(),
      tailwindcss(),
      tanstackRouter({ target: 'react', autoCodeSplitting: true }),
      viteReact(),
    ],
    server: {
      host: true, // Required for Docker container port mapping to work
      proxy: {
        "/api": {
          target: env.BACKEND_URL, // Changed from localhost:8000 to the Docker service name
          changeOrigin: true
        }
      }
    }
  }
})

export default config
