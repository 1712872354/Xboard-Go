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
  // 后端可能返回 { command: "..." } 或直接返回字符串
  if (typeof res === 'string') return res
  if (res && typeof res === 'object' && 'command' in res) return (res as any).command
  return String(res)
}

export interface LoadHistoryItem {
  id: number
  machine_id: number
  cpu: number
  mem_total: number
  mem_used: number
  disk_total: number
  disk_used: number
  net_in_speed: number
  net_out_speed: number
  recorded_at: number
  created_at: string
}

export const useLoadHistory = (machineId: number | null, page = 1, pageSize = 50) =>
  useQuery({
    queryKey: ['admin', 'servers', machineId, 'load-history', page, pageSize],
    queryFn: async () =>
      (await api.get(`/admin/server-machines/${machineId}/load-history`, {
        params: { page, page_size: pageSize },
      })) as unknown as { list: LoadHistoryItem[]; total: number; page: number; page_size: number },
    enabled: !!machineId,
  })
