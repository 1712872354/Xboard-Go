import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import api from '@/lib/api'
import type { ServerRoute, PaginatedResponse } from '@/types'

export const useServerRoutes = (page = 1, pageSize = 50, groupId?: number) =>
  useQuery({
    queryKey: ['admin', 'server-routes', page, pageSize, groupId],
    queryFn: async () =>
      (await api.get('/admin/server-routes', { params: { page, page_size: pageSize, group_id: groupId } })) as unknown as PaginatedResponse<ServerRoute>,
  })

export const useCreateServerRoute = () => {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (data: Partial<ServerRoute>) => await api.post('/admin/server-routes', data),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['admin', 'server-routes'] })
    },
  })
}

export const useUpdateServerRoute = () => {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async ({ id, ...data }: Partial<ServerRoute> & { id: number }) =>
      await api.put(`/admin/server-routes/${id}`, data),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['admin', 'server-routes'] })
    },
  })
}

export const useDeleteServerRoute = () => {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (id: number) => await api.delete(`/admin/server-routes/${id}`),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['admin', 'server-routes'] })
    },
  })
}
