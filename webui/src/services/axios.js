import axios from 'axios'
import { state } from './state'

const api = axios.create({
  baseURL: 'http://localhost:3000/v1',
})

api.interceptors.request.use((config) => {
  if (state.userId) {
    config.headers = config.headers || {}
    config.headers.Authorization = `Bearer ${state.userId}`
  }
  return config
})

export default api
