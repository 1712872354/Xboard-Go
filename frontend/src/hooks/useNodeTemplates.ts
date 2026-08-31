import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import api from '@/lib/api'
import type { PaginatedResponse } from '@/types'
import { toast } from 'sonner'

export interface NodeTemplate {
  id: number
  name: string
  type: string
  server_info: string
  description: string
  created_at: string
  updated_at: string
}

export const useNodeTemplates = (page = 1, pageSize = 50) =>
  useQuery({
    queryKey: ['admin', 'node-templates', page, pageSize],
    queryFn: async () =>
      (await api.get('/admin/node-templates', { params: { page, page_size: pageSize } })) as unknown as PaginatedResponse<NodeTemplate>,
  })

export const useAllNodeTemplates = () =>
  useQuery({
    queryKey: ['admin', 'node-templates', 'all'],
    queryFn: async () => {
      const res = (await api.get('/admin/node-templates', { params: { page: 1, page_size: 100 } })) as unknown as PaginatedResponse<NodeTemplate>
      return res.list
    },
  })

export const useCreateNodeTemplate = () => {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (data: Partial<NodeTemplate>) => await api.post('/admin/node-templates', data),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['admin', 'node-templates'] })
      toast.success('模板创建成功')
    },
  })
}

export const useUpdateNodeTemplate = () => {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async ({ id, ...data }: Partial<NodeTemplate> & { id: number }) =>
      await api.put(`/admin/node-templates/${id}`, data),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['admin', 'node-templates'] })
      toast.success('模板更新成功')
    },
  })
}

export const useDeleteNodeTemplate = () => {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (id: number) => await api.delete(`/admin/node-templates/${id}`),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['admin', 'node-templates'] })
      toast.success('模板已删除')
    },
  })
}
