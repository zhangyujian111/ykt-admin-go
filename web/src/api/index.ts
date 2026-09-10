import axios from 'axios'

export const api = axios.create({
  baseURL: '/api',
  timeout: 15000
})

api.interceptors.request.use((config: any) => {
  const t = localStorage.getItem('admin_token')
  if (t) config.headers['Authorization'] = `Bearer ${t}`
  return config
})

api.interceptors.response.use(
  (r) => r,
  (err) => {
    if (err?.response?.status === 401) {
      localStorage.removeItem('admin_token')
      if (location.pathname !== '/login') location.href = '/login'
    }
    return Promise.reject(err)
  }
)