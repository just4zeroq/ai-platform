import { create } from 'zustand'
import { apiGet, apiPost } from '../api/client'

interface User {
  id: number
  username: string
  email: string
  display_name?: string
}

interface AuthState {
  token: string | null
  user: User | null
  loading: boolean
  setAuth: (token: string, user: User) => void
  login: (username: string, password: string) => Promise<void>
  register: (username: string, password: string, email: string) => Promise<void>
  fetchProfile: () => Promise<void>
  logout: () => void
}

export const useAuthStore = create<AuthState>((set, get) => ({
  token: localStorage.getItem('token'),
  user: null,
  loading: false,

  setAuth: (token, user) => {
    localStorage.setItem('token', token)
    set({ token, user })
  },

  login: async (username, password) => {
    set({ loading: true })
    try {
      const res = await apiPost<{ user_id: number; username: string; token: string }>(
        '/user/login', { username, password }
      )
      const user = { id: res.user_id, username: res.username, email: '' }
      localStorage.setItem('token', res.token)
      set({ token: res.token, user, loading: false })
    } catch (e) {
      set({ loading: false })
      throw e
    }
  },

  register: async (username, password, email) => {
    set({ loading: true })
    try {
      const res = await apiPost<{ user_id: number; username: string; token: string }>(
        '/user/register', { username, password, email }
      )
      const user = { id: res.user_id, username: res.username, email: '' }
      localStorage.setItem('token', res.token)
      set({ token: res.token, user, loading: false })
    } catch (e) {
      set({ loading: false })
      throw e
    }
  },

  fetchProfile: async () => {
    const { token } = get()
    if (!token) return
    try {
      const user = await apiGet<User>('/user/profile', token)
      set({ user })
    } catch {
      localStorage.removeItem('token')
      set({ token: null, user: null })
    }
  },

  logout: () => {
    localStorage.removeItem('token')
    set({ token: null, user: null })
  },
}))
