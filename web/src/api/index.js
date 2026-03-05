import axios from 'axios'
import { useAuthStore } from '@/stores/auth'

const http = axios.create({ baseURL: '/api' })

http.interceptors.request.use(cfg => {
  const auth = useAuthStore()
  if (auth.token) cfg.headers.Authorization = `Bearer ${auth.token}`
  return cfg
})

http.interceptors.response.use(
  r => r.data,
  err => {
    if (err.response?.status === 401) {
      useAuthStore().logout()
      window.location.href = '/login'
    }
    return Promise.reject(err.response?.data?.error || err.message)
  }
)

export default http

export const authApi = {
  status:  ()      => http.get('/auth/status'),
  setup:   (pwd)   => http.post('/auth/setup',  { password: pwd }),
  login:   (pwd)   => http.post('/auth/login',  { password: pwd }),
}

export const systemApi = {
  info:   () => http.get('/system/info'),
  status: () => http.get('/system/status'),
}

export const coreApi = {
  info:             ()           => http.get('/core/info'),
  download:         (v, arch)    => http.post('/core/download', { version: v, arch }),
  downloadProgress: ()           => http.get('/core/download/progress'),
  start:            ()           => http.post('/core/start'),
  stop:             ()           => http.post('/core/stop'),
  restart:          ()           => http.post('/core/restart'),
  logs:             (n=200)      => http.get(`/core/logs?lines=${n}`),
}

export const configApi = {
  getRaw:        ()              => http.get('/config/raw'),
  setRaw:        (config)        => http.put('/config/raw',  { config }),
  getSections:   ()              => http.get('/config/sections'),
  setSection:    (s, data)       => http.put(`/config/sections/${s}`, data),
  validate:      ()              => http.post('/config/validate'),
}

export const providersApi = {
  get:    ()      => http.get('/providers'),
  set:    (list)  => http.put('/providers', { providers: list }),
  update: (tag)   => http.post(`/providers/${tag}/update`),
}

export const proxyApi = {
  getMode: ()    => http.get('/proxy/mode'),
  apply:   (cfg) => http.post('/proxy/apply', cfg),
  stop:    ()    => http.post('/proxy/stop'),
  status:  ()    => http.get('/proxy/status'),
}

export const rulesetsApi = {
  get:    () => http.get('/rulesets'),
  update: () => http.post('/rulesets/update'),
}
