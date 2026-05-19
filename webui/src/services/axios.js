import axios from 'axios'
import { state } from './state'

function resolveApiBaseUrl() {
  const configuredUrl = typeof __API_URL__ !== 'undefined' ? __API_URL__ : ''
  const cleanedConfiguredUrl = typeof configuredUrl === 'string' ? configuredUrl.trim() : ''
  if (cleanedConfiguredUrl) {
    return cleanedConfiguredUrl.replace(/\/$/, '')
  }

  if (typeof window !== 'undefined' && window.location) {
    return `${window.location.protocol}//${window.location.host}`
  }

  return ''
}

export function resolveApiUrl(path) {
  if (!path) return null
  if (/^https?:\/\//i.test(path)) return path

  const baseUrl = resolveApiBaseUrl()
  if (!baseUrl) return path
  return `${baseUrl}${path.startsWith('/') ? '' : '/'}${path}`
}

const api = axios.create({
  baseURL: resolveApiBaseUrl(),
})

api.interceptors.request.use((config) => {
  if (state.userId) {
    config.headers = config.headers || {}
    config.headers.Authorization = `Bearer ${state.userId}`
  }
  return config
})

export default api
