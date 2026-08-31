import { useQuery } from '@tanstack/react-query'
import api from '@/lib/api'
import type { TrafficStats, TrafficHistoryItem, TrafficLog, PaginatedResponse } from '@/types'

export const useTrafficStats = () =>
  useQuery({
    queryKey: ['user', 'traffic', 'stats'],
    queryFn: async () => (await api.get('/user/traffic/stats')) as unknown as TrafficStats,
  })

export const useTrafficHistory = (page = 1, pageSize = 20) =>
  useQuery({
    queryKey: ['user', 'traffic', 'history', page, pageSize],
    queryFn: async () =>
      (await api.get('/user/traffic/history', { params: { page, page_size: pageSize } })) as unknown as PaginatedResponse<TrafficLog>,
  })

export const useTrafficDaily = (days = 7) =>
  useQuery({
    queryKey: ['user', 'traffic', 'daily', days],
    queryFn: async () => (await api.get('/user/traffic/daily', { params: { days } })) as unknown as TrafficHistoryItem[],
  })
