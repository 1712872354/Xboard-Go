import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { useNavigate } from 'react-router-dom'
import api from '@/lib/api'
import { useAuthStore } from '@/stores/auth'
import type { LoginResponse, User } from '@/types'
import { toast } from 'sonner'

export const useLogin = (redirectTo?: string) => {
  const queryClient = useQueryClient()
  const navigate = useNavigate()
  const setTokens = useAuthStore((s) => s.setTokens)
  const setUser = useAuthStore((s) => s.setUser)

  return useMutation({
    mutationFn: async (data: { email: string; password: string }) => {
      return (await api.post('/auth/login', data)) as unknown as LoginResponse
    },
    onSuccess: (data) => {
      setTokens(data.access_token, data.refresh_token)
      setUser(data.user)
      queryClient.invalidateQueries({ queryKey: ['user', 'profile'] })
      toast.success('登录成功')
      const target = redirectTo || '/user/dashboard'
      navigate(target, { replace: true })
    },
  })
}

export const useRegister = () => {
  return useMutation({
    mutationFn: async (data: { email: string; password: string }) => {
      return await api.post('/auth/register', data)
    },
  })
}

export const useLogout = () => {
  const navigate = useNavigate()
  const logout = useAuthStore((s) => s.logout)
  const queryClient = useQueryClient()

  return () => {
    logout()
    queryClient.clear()
    navigate('/user/login')
  }
}

export const useUserProfile = () => {
  const token = useAuthStore((s) => s.token)

  return useQuery({
    queryKey: ['user', 'profile'],
    queryFn: async () => (await api.get('/user/profile')) as unknown as User,
    enabled: !!token,
  })
}
