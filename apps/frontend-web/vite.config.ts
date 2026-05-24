import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';

export default defineConfig({
  plugins: [react()],
  server: {
    port: 5173,
    proxy: {
      '/api/auth': 'http://127.0.0.1:8080',
      '/api/config': 'http://127.0.0.1:8080',
      '/api/oauth': 'http://127.0.0.1:8081',
      '/api/ad-accounts': 'http://127.0.0.1:8081',
      '/api/ad-sync': 'http://127.0.0.1:8081',
      '/api/leads': 'http://127.0.0.1:8090',
      '/api/conflicts': 'http://127.0.0.1:8090',
      '/api/attributions': 'http://127.0.0.1:8090',
      '/api/etl': 'http://127.0.0.1:8091',
      '/api/analytics': 'http://127.0.0.1:8091'
    }
  }
});
