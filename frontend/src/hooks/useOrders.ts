import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import api from '@/lib/api'
import type { Order, PaginatedResponse } from '@/types'
import { toast } from 'sonner'

export const useOrders = (page = 1, pageSize = 20) =>
  useQuery({
    queryKey: ['user', 'orders', page, pageSize],
    queryFn: async () =>
      (await api.get('/orders', { params: { page, page_size: pageSize } })) as unknown as PaginatedResponse<Order>,
  })

export const useCreateOrder = () => {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (data: { plan_id: number; coupon_code?: string }) =>
      (await api.post('/orders', data)) as unknown as Order,
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['user', 'orders'] })
      toast.success('订单创建成功')
    },
  })
}

export const useCancelOrder = () => {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (id: number) => await api.post(`/orders/${id}/cancel`),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['user', 'orders'] })
      toast.success('订单已取消')
    },
  })
}

export const useAdminOrders = (page = 1, pageSize = 20, status?: number) =>
  useQuery({
    queryKey: ['admin', 'orders', page, pageSize, status],
    queryFn: async () =>
      (await api.get('/admin/orders', { params: { page, page_size: pageSize, status } })) as unknown as PaginatedResponse<Order>,
  })

export const useConfirmPayment = () => {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (tradeNo: string) =>
      await api.post('/admin/orders/confirm-payment', { trade_no: tradeNo }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['admin', 'orders'] })
      toast.success('支付确认成功')
    },
  })
}
