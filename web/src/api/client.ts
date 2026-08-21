import axios, { type AxiosInstance, type AxiosRequestConfig } from 'axios'
import { ElMessage } from 'element-plus'

// API response envelope mirroring the backend.
export interface ApiEnvelope<T = unknown> {
  code: number
  message: string
  data?: T
  request_id?: string
}

export interface PageResult<T> {
  items: T[]
  total: number
  page: number
  page_size: number
  total_page: number
}

const baseURL = import.meta.env.VITE_API_BASE || '/api/v1'

// Singleton Axios client with auth + refresh interceptor.
const client: AxiosInstance = axios.create({
  baseURL,
  timeout: 30000,
})

client.interceptors.request.use((config) => {
  const token = useAccessToken()
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

let refreshing: Promise<string> | null = null

client.interceptors.response.use(
  (response) => {
    const env = response.data as ApiEnvelope | undefined
    if (env && typeof env.code === 'number') {
      if (env.code !== 0) {
        ElMessage.error(env.message || '请求失败')
        return Promise.reject(new Error(env.message))
      }
      return { ...response, data: env.data }
    }
    return response
  },
  async (error) => {
    const status = error.response?.status
    if (status === 401 && !error.config.__retry) {
      error.config.__retry = true
      try {
        refreshing ||= refreshTokens()
        const token = await refreshing
        refreshing = null
        error.config.headers.Authorization = `Bearer ${token}`
        return client.request(error.config)
      } catch (e) {
        refreshing = null
        clearSession()
        if (routerAvailable()) routerPush('/login')
        return Promise.reject(e)
      }
    }
    const msg = error.response?.data?.message || error.message || '网络错误'
    ElMessage.error(msg)
    return Promise.reject(error)
  }
)

async function refreshTokens(): Promise<string> {
  const rt = localStorage.getItem('refresh_token')
  if (!rt) throw new Error('no refresh token')
  const res = await axios.post<ApiEnvelope<{ access_token: string; refresh_token: string }>>(
    `${baseURL}/auth/refresh`,
    { refresh_token: rt }
  )
  if (res.data.code !== 0) throw new Error(res.data.message)
  setSession(res.data.data!.access_token, res.data.data!.refresh_token)
  return res.data.data!.access_token
}

// --- session storage helpers ---
function useAccessToken(): string | null {
  return localStorage.getItem('access_token')
}

function setSession(access: string, refresh: string) {
  localStorage.setItem('access_token', access)
  localStorage.setItem('refresh_token', refresh)
}

function clearSession() {
  localStorage.removeItem('access_token')
  localStorage.removeItem('refresh_token')
}

// router import is lazy to avoid a cycle at module-eval time.
let router: any = null
export function registerRouter(r: any) {
  router = r
}
function routerAvailable() {
  return !!router
}
function routerPush(path: string) {
  if (router) router.push(path)
}

export const api = {
  setSession,
  clearSession,
  get<T>(url: string, config?: AxiosRequestConfig) {
    return client.get<unknown, { data: T }>(url, config).then((r) => r.data)
  },
  post<T>(url: string, body?: unknown, config?: AxiosRequestConfig) {
    return client.post<unknown, { data: T }>(url, body, config).then((r) => r.data)
  },
  put<T>(url: string, body?: unknown, config?: AxiosRequestConfig) {
    return client.put<unknown, { data: T }>(url, body, config).then((r) => r.data)
  },
  patch<T>(url: string, body?: unknown, config?: AxiosRequestConfig) {
    return client.patch<unknown, { data: T }>(url, body, config).then((r) => r.data)
  },
  delete<T>(url: string, config?: AxiosRequestConfig) {
    return client.delete<unknown, { data: T }>(url, config).then((r) => r.data)
  },
}
