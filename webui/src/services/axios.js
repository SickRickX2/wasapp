import axios from 'axios'
import { state } from './state'

const api = axios.create({
  baseURL: typeof __API_URL__ !== 'undefined' ? __API_URL__ : 'http://localhost:3000',
})

api.interceptors.request.use((config) => {
  if (state.userId) {
    config.headers = config.headers || {}
    config.headers.Authorization = `Bearer ${state.userId}`
  }
  return config
})

export default api
