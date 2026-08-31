import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import api from '@/lib/api'
import type { Node, PaginatedResponse, NodeStats } from '@/types'

export const useAdminNodes = (page = 1, pageSize = 20, groupId?: number) =>
  useQuery({
    queryKey: ['admin', 'nodes', page, pageSize, groupId],
    queryFn: async () =>
      (await api.get('/admin/nodes', { params: { page, page_size: pageSize, group_id: groupId || undefined } })) as unknown as PaginatedResponse<Node>,
  })

export const useNodeStats = () =>
  useQuery({
    queryKey: ['admin', 'dashboard', 'node-stats'],
    queryFn: async () => (await api.get('/admin/dashboard/node-stats')) as unknown as NodeStats,
  })

export const useCreateNode = () => {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (data: Partial<Node>) => await api.post('/admin/nodes', data),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['admin', 'nodes'] })
    },
  })
}

export const useUpdateNode = () => {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async ({ id, ...data }: Partial<Node> & { id: number }) =>
      await api.put(`/admin/nodes/${id}`, data),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['admin', 'nodes'] })
    },
  })
}

export const useDeleteNode = () => {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (id: number) => await api.delete(`/admin/nodes/${id}`),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['admin', 'nodes'] })
    },
  })
}

export const useUpdateNodeStatus = () => {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async ({ id, status }: { id: number; status: number }) =>
      await api.put(`/admin/nodes/${id}/status`, { status }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['admin', 'nodes'] })
    },
  })
}
