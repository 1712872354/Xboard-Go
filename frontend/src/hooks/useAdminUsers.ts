import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import api from '@/lib/api'
import type { User, PaginatedResponse } from '@/types'
import { toast } from 'sonner'

export const useAdminUsers = (page = 1, pageSize = 20, keyword = '') =>
  useQuery({
    queryKey: ['admin', 'users', page, pageSize, keyword],
    queryFn: async () =>
      (await api.get('/admin/users', { params: { page, page_size: pageSize, keyword } })) as unknown as PaginatedResponse<User>,
  })

export const useAdminUser = (id: number) =>
  useQuery({
    queryKey: ['admin', 'user', id],
    queryFn: async () => (await api.get(`/admin/users/${id}`)) as unknown as User,
    enabled: !!id,
  })

export const useUpdateUserStatus = () => {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async ({ id, status }: { id: number; status: number }) =>
      await api.put(`/admin/users/${id}/status`, { status }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['admin', 'users'] })
      toast.success('用户状态更新成功')
    },
  })
}

export const useAdminUpdateUser = () => {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async ({ id, data }: { id: number; data: Partial<User> & { password?: string } }) =>
      (await api.put(`/admin/users/${id}`, data)) as unknown as User,
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['admin', 'users'] })
      toast.success('用户信息更新成功')
    },
  })
}

export const useDeleteUser = () => {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (id: number) => await api.delete(`/admin/users/${id}`),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['admin', 'users'] })
      toast.success('用户删除成功')
    },
  })
}

export const useGenerateUsers = () => {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (data: { count: number; prefix?: string; password?: string; plan_id?: number; expired_at?: string }) =>
      await api.post('/admin/users/generate', data),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['admin', 'users'] })
      toast.success('用户生成成功')
    },
  })
}

export const useExportUsers = () =>
  useMutation({
    mutationFn: async (params?: { keyword?: string }) => {
      const res = await api.get('/admin/users/export', { params, responseType: 'blob' })
      return res
    },
  })
