import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import api from '@/lib/api'
import type { Plan, PaginatedResponse } from '@/types'
import { toast } from 'sonner'

export const usePlans = () =>
  useQuery({
    queryKey: ['plans'],
    queryFn: async () => (await api.get('/plans')) as unknown as Plan[],
  })

export const useAdminPlans = (page = 1, pageSize = 20) =>
  useQuery({
    queryKey: ['admin', 'plans', page, pageSize],
    queryFn: async () =>
      (await api.get('/admin/plans', { params: { page, page_size: pageSize } })) as unknown as PaginatedResponse<Plan>,
  })

export const useCreatePlan = () => {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (data: Partial<Plan>) => await api.post('/admin/plans', data),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['admin', 'plans'] })
      toast.success('套餐创建成功')
    },
  })
}

export const useUpdatePlan = () => {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async ({ id, ...data }: Partial<Plan> & { id: number }) =>
      await api.put(`/admin/plans/${id}`, data),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['admin', 'plans'] })
      toast.success('套餐更新成功')
    },
  })
}

export const useDeletePlan = () => {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (id: number) => await api.delete(`/admin/plans/${id}`),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['admin', 'plans'] })
      toast.success('套餐删除成功')
    },
  })
}
