import { useQuery } from '@tanstack/react-query'
import api from '@/lib/api'
import type { DashboardOverview, DailyIncome, DailyUserGrowth, NodeStats, Order, User } from '@/types'

export interface ComprehensiveStats {
  today_income: number
  day_income_growth: number
  current_month_income: number
  last_month_income: number
  month_income_growth: number
  current_month_commission_payout: number
  last_month_commission_payout: number
  commission_growth: number
  commission_pending_total: number
  current_month_new_users: number
  total_users: number
  active_users: number
  user_growth: number
  online_users: number
  online_devices: number
  ticket_pending_total: number
  online_nodes: number
  today_traffic: { upload: number; download: number; total: number }
  month_traffic: { upload: number; download: number; total: number }
  total_traffic: { upload: number; download: number; total: number }
}

export const useDashboardOverview = () =>
  useQuery({
    queryKey: ['admin', 'dashboard', 'overview'],
    queryFn: async () => (await api.get('/admin/dashboard/overview')) as unknown as DashboardOverview,
  })

export const useRecentOrders = (limit = 10) =>
  useQuery({
    queryKey: ['admin', 'dashboard', 'recent-orders', limit],
    queryFn: async () => (await api.get('/admin/dashboard/recent-orders', { params: { limit } })) as unknown as Order[],
  })

export const useRecentUsers = (limit = 10) =>
  useQuery({
    queryKey: ['admin', 'dashboard', 'recent-users', limit],
    queryFn: async () => (await api.get('/admin/dashboard/recent-users', { params: { limit } })) as unknown as User[],
  })

export const useIncomeStats = (days = 7) =>
  useQuery({
    queryKey: ['admin', 'dashboard', 'income-stats', days],
    queryFn: async () => (await api.get('/admin/dashboard/income-stats', { params: { days } })) as unknown as DailyIncome[],
  })

export const useUserGrowthStats = (days = 7) =>
  useQuery({
    queryKey: ['admin', 'dashboard', 'user-growth', days],
    queryFn: async () => (await api.get('/admin/dashboard/user-growth', { params: { days } })) as unknown as DailyUserGrowth[],
  })

export const useNodeStats = () =>
  useQuery({
    queryKey: ['admin', 'dashboard', 'node-stats'],
    queryFn: async () => (await api.get('/admin/dashboard/node-stats')) as unknown as NodeStats,
  })

export const useComprehensiveStats = () =>
  useQuery({
    queryKey: ['admin', 'dashboard', 'comprehensive-stats'],
    queryFn: async () => (await api.get('/admin/dashboard/comprehensive-stats')) as unknown as ComprehensiveStats,
  })

export interface NodeTrafficRankingItem {
  node_id: number
  node_name: string
  upload: number
  download: number
  total: number
}

export const useNodeTrafficRanking = (days = 7) =>
  useQuery({
    queryKey: ['admin', 'dashboard', 'node-traffic-ranking', days],
    queryFn: async () =>
      (await api.get('/admin/dashboard/node-traffic-ranking', { params: { days } })) as unknown as NodeTrafficRankingItem[],
  })

export interface UserTrafficRankingItem {
  user_id: number
  user_email: string
  upload: number
  download: number
  total: number
}

export const useUserTrafficRanking = (days = 7) =>
  useQuery({
    queryKey: ['admin', 'dashboard', 'user-traffic-ranking', days],
    queryFn: async () =>
      (await api.get('/admin/dashboard/user-traffic-ranking', { params: { days } })) as unknown as UserTrafficRankingItem[],
  })
