import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import api from '@/lib/api'
import type { ServerGroup, PaginatedResponse } from '@/types'

export const useServerGroups = (page = 1, pageSize = 50) =>
  useQuery({
    queryKey: ['admin', 'server-groups', page, pageSize],
    queryFn: async () =>
      (await api.get('/admin/server-groups', { params: { page, page_size: pageSize } })) as unknown as PaginatedResponse<ServerGroup>,
  })

export const useAllServerGroups = () =>
  useQuery({
    queryKey: ['admin', 'server-groups', 'all'],
    queryFn: async () => (await api.get('/admin/server-groups/all')) as unknown as ServerGroup[],
  })

export const useCreateServerGroup = () => {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (data: Partial<ServerGroup>) => await api.post('/admin/server-groups', data),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['admin', 'server-groups'] })
    },
  })
}

export const useUpdateServerGroup = () => {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async ({ id, ...data }: Partial<ServerGroup> & { id: number }) =>
      await api.put(`/admin/server-groups/${id}`, data),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['admin', 'server-groups'] })
    },
  })
}

export const useDeleteServerGroup = () => {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (id: number) => await api.delete(`/admin/server-groups/${id}`),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['admin', 'server-groups'] })
    },
  })
}
