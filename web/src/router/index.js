import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const routes = [
  { path: '/setup',    component: () => import('@/views/Setup.vue'),    meta: { public: true } },
  { path: '/login',    component: () => import('@/views/Login.vue'),    meta: { public: true } },
  {
    path: '/',
    component: () => import('@/views/Layout.vue'),
    children: [
      { path: '',          redirect: '/dashboard' },
      { path: 'dashboard', component: () => import('@/views/Dashboard.vue') },
      { path: 'proxy',     component: () => import('@/views/Proxy.vue') },
      { path: 'core',      component: () => import('@/views/Core.vue') },
      { path: 'config',    component: () => import('@/views/Config.vue') },
      { path: 'dns',       component: () => import('@/views/sections/DNS.vue') },
      { path: 'inbounds',  component: () => import('@/views/sections/Inbounds.vue') },
      { path: 'outbounds', component: () => import('@/views/sections/Outbounds.vue') },
      { path: 'route',     component: () => import('@/views/sections/Route.vue') },
      { path: 'rulesets',  component: () => import('@/views/sections/RuleSets.vue') },
      { path: 'providers', component: () => import('@/views/Providers.vue') },
      { path: 'logs',      component: () => import('@/views/Logs.vue') },
    ]
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

router.beforeEach(async (to) => {
  if (to.meta.public) return true
  const auth = useAuthStore()
  if (!auth.token) {
    const status = await auth.checkSetup()
    return status.setup_done ? '/login' : '/setup'
  }
  return true
})

export default router
