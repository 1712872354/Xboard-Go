import { create } from 'zustand'
import { persist } from 'zustand/middleware'
import type { User } from '@/types'

interface AuthState {
  user: User | null
  token: string | null
  refreshToken: string | null
  isAuthenticated: boolean

  setTokens: (access: string, refresh: string) => void
  setUser: (user: User) => void
  logout: () => void
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set) => ({
      user: null,
      token: null,
      refreshToken: null,
      isAuthenticated: false,

      setTokens: (access, refresh) => {
        set({ token: access, refreshToken: refresh, isAuthenticated: true })
      },

      setUser: (user) => set({ user }),

      logout: () => {
        set({ user: null, token: null, refreshToken: null, isAuthenticated: false })
      },
    }),
    {
      name: 'auth-storage',
      partialize: (state) => ({
        user: state.user,
        token: state.token,
        refreshToken: state.refreshToken,
        isAuthenticated: state.isAuthenticated,
      }),
      version: 2,
      migrate: (persisted: unknown, version: number) => {
        if (version < 2) {
          return { user: null, token: null, refreshToken: null, isAuthenticated: false }
        }
        return persisted
      },
    },
  ),
)
