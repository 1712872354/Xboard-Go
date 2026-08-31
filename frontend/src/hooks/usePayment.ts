import { useQuery, useMutation } from '@tanstack/react-query'
import api from '@/lib/api'
import type { PaymentMethod } from '@/types'

export const usePaymentMethods = () =>
  useQuery({
    queryKey: ['payment', 'methods'],
    queryFn: async () => (await api.get('/payment/methods')) as unknown as PaymentMethod[],
  })

export const useCreatePayment = () =>
  useMutation({
    mutationFn: async (data: { order_id: number; payment_method: string }) =>
      await api.post('/payment/create', data),
  })
