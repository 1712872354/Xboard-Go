import axios from 'axios'
import { toast } from 'sonner'
import { useAuthStore } from '@/stores/auth'

const api = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || '/api/v1',
  timeout: 30000,
  headers: { 'Content-Type': 'application/json' },
})

let isRefreshing = false
let isLoggingOut = false
let failedQueue: Array<{
  resolve: (value: unknown) => void
  reject: (reason?: unknown) => void
}> = []

const processQueue = (error: unknown | null, token: string | null = null) => {
  failedQueue.forEach((prom) => {
    if (error) prom.reject(error)
    else prom.resolve(token)
  })
  failedQueue = []
}

export const refreshAccessToken = async (): Promise<string> => {
  const refreshToken = useAuthStore.getState().refreshToken
  if (!refreshToken) throw new Error('No refresh token')

  const { data } = await axios.post(
    `${import.meta.env.VITE_API_BASE_URL || '/api/v1'}/auth/refresh`,
    { refresh_token: refreshToken },
  )

  // Backend returns {code, message, data} — check for auth errors
  if (data.code !== 0) {
    throw new Error(data.message || 'Token refresh failed')
  }

  const { access_token, refresh_token } = data.data
  useAuthStore.getState().setTokens(access_token, refresh_token)
  return access_token
}

api.interceptors.request.use((config) => {
  const token = useAuthStore.getState().token
  if (token) config.headers.Authorization = `Bearer ${token}`
  return config
})

api.interceptors.response.use(
  (response) => {
    const res = response.data

    // Success
    if (res.code === 0) return res.data

    // Auth error (backend returns HTTP 200 with code 401)
    if (res.code === 401) {
      if (!isLoggingOut) {
        isLoggingOut = true
        useAuthStore.getState().logout()
        window.location.href = '/user/login'
        setTimeout(() => { isLoggingOut = false }, 2000)
      }
      return Promise.reject(new Error(res.message || '登录已过期'))
    }

    const msg = res.message || '请求失败'
    toast.error(msg)
    return Promise.reject(new Error(msg))
  },
  async (error) => {
    if (!error.response) {
      toast.error('网络连接失败，请检查网络')
      return Promise.reject(error)
    }

    const { status, data } = error.response
    const originalRequest = error.config

    if (status === 401 && !originalRequest._retry) {
      if (isRefreshing) {
        return new Promise((resolve, reject) => {
          failedQueue.push({ resolve, reject })
        }).then((token) => {
          originalRequest.headers.Authorization = `Bearer ${token}`
          return api(originalRequest)
        })
      }

      originalRequest._retry = true
      isRefreshing = true

      try {
        const newToken = await refreshAccessToken()
        originalRequest.headers.Authorization = `Bearer ${newToken}`
        processQueue(null, newToken)
        return api(originalRequest)
      } catch (refreshError) {
        processQueue(refreshError, null)
        if (!isLoggingOut) {
          isLoggingOut = true
          useAuthStore.getState().logout()
          window.location.href = '/user/login'
          setTimeout(() => { isLoggingOut = false }, 2000)
        }
        return Promise.reject(refreshError)
      } finally {
        isRefreshing = false
      }
    }

    const msg = data?.message || `请求失败 (${status})`
    toast.error(msg)
    return Promise.reject(new Error(msg))
  },
)

export default api
