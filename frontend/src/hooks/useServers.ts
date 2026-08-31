import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'
import api from '@/lib/api'
import type { ServerMachine, PaginatedResponse } from '@/types'

export const useServers = (page = 1, pageSize = 50) =>
  useQuery({
    queryKey: ['admin', 'servers', page, pageSize],
    queryFn: async () =>
      (await api.get('/admin/server-machines', { params: { page, page_size: pageSize } })) as unknown as PaginatedResponse<ServerMachine>,
  })

export const useAllServers = () =>
  useQuery({
    queryKey: ['admin', 'servers', 'all'],
    queryFn: async () => (await api.get('/admin/server-machines/all')) as unknown as ServerMachine[],
  })

export const useCreateServer = () => {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (data: Partial<ServerMachine>) => await api.post('/admin/server-machines', data),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['admin', 'servers'] })
    },
  })
}

export const useUpdateServer = () => {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async ({ id, ...data }: Partial<ServerMachine> & { id: number }) =>
      await api.put(`/admin/server-machines/${id}`, data),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['admin', 'servers'] })
    },
  })
}

export const useUpdateServerStatus = () => {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async ({ id, status }: { id: number; status: number }) =>
      await api.put(`/admin/server-machines/${id}/status`, { status }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['admin', 'servers'] })
    },
  })
}

export const useDeleteServer = () => {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (id: number) => await api.delete(`/admin/server-machines/${id}`),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['admin', 'servers'] })
    },
  })
}

export const useResetServerToken = () => {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (id: number) => await api.post(`/admin/server-machines/${id}/reset-token`),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['admin', 'servers'] })
      toast.success('Token 已重新生成')
    },
  })
}

export const getInstallCommand = async (id: number): Promise<string> => {
  const res = await api.get(`/admin/server-machines/${id}/install-command`)
  return res as unknown as string
}
