import { create } from 'zustand'
import type { User } from '../types'
import { getMe, logout as apiLogout } from '../api/auth'

interface AuthState {
  user: User | null
  loading: boolean
  checkAuth: () => Promise<void>
  logout: () => Promise<void>
}

export const useAuthStore = create<AuthState>((set) => ({
  user: null,
  loading: true,
  checkAuth: async () => {
    try {
      const user = await getMe()
      set({ user, loading: false })
    } catch {
      set({ user: null, loading: false })
    }
  },
  logout: async () => {
    await apiLogout()
    set({ user: null })
  },
}))
