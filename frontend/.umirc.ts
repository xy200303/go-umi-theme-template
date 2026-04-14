import path from 'node:path';
import tailwindcss from '@tailwindcss/postcss';
import { defineConfig } from 'umi';

export default defineConfig({
  npmClient: 'npm',
  hash: false,
  mfsu: false,
  esbuildMinifyIIFE: true,
  outputPath: '../backend/web',
  extraPostCSSPlugins: [
    tailwindcss({
      base: __dirname
    })
  ],
  history: {
    type: 'browser'
  },
  alias: {
    '@': path.resolve(__dirname, 'src')
  },
  proxy: {
    '/api': {
      target: 'http://127.0.0.1:8080',
      changeOrigin: true
    }
  },
  routes: [
    { path: '/', component: '@/pages/index' },
    { path: '/about', component: '@/pages/about' },
    { path: '/blog', component: '@/pages/blog' },
    { path: '/login', component: '@/pages/login' },
    { path: '/register', component: '@/pages/register' },
    { path: '/profile', component: '@/pages/profile/index' },
    { path: '/profile/interfaces', component: '@/pages/profile/interfaces' },
    { path: '/profile/gateway-logs', component: '@/pages/profile/gateway-logs' },
    { path: '/profile/chat-records', component: '@/pages/profile/chat-records' },
    { path: '/admin', component: '@/pages/admin' },
    { path: '/admin/home', component: '@/pages/admin/home' },
    { path: '/admin/system/users', component: '@/pages/admin/system/users' },
    { path: '/admin/system/files', component: '@/pages/admin/system/files' },
    { path: '/admin/system/config', component: '@/pages/admin/system/config' },
    { path: '/admin/system/role', component: '@/pages/admin/system/role' },
    { path: '/admin/system/audit', component: '@/pages/admin/system/audit' },
    { path: '*', component: '@/pages/404' }
  ]
});
