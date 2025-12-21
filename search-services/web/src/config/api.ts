/**
 * API Configuration
 * Base URL for backend API requests
 */
export const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:28080'

export const apiConfig = {
  baseURL: API_BASE_URL,
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json',
  },
}
