import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'
import api from '@/lib/api'
import type { AuditLog, PaginatedResponse } from '@/types'
import { Card, CardContent, CardHeader } from '@/components/ui/card'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Skeleton } from '@/components/ui/skeleton'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { formatDate } from '@/lib/utils'
import { Trash2, X } from 'lucide-react'

const useAdminAuditLogs = (page = 1, pageSize = 20, action?: string) =>
  useQuery({
    queryKey: ['admin', 'audit-logs', page, pageSize, action],
    queryFn: async () => (await api.get('/admin/audit-logs', { params: { page, page_size: pageSize, action: action || undefined } })) as unknown as PaginatedResponse<AuditLog>,
  })

const useDeleteAuditLog = () => {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (id: number) => await api.delete(`/admin/audit-logs/${id}`),
    onSuccess: () => { qc.invalidateQueries({ queryKey: ['admin', 'audit-logs'] }); toast.success('日志删除成功') },
  })
}

const actionOptions = [
  { value: 'all', label: '全部' },
  { value: 'login', label: '登录' },
  { value: 'create', label: '创建' },
  { value: 'update', label: '更新' },
  { value: 'delete', label: '删除' },
  { value: 'export', label: '导出' },
]

const actionVariant: Record<string, 'default' | 'secondary' | 'success' | 'destructive' | 'warning' | 'outline'> = {
  login: 'default',
  create: 'success',
  update: 'warning',
  delete: 'destructive',
  export: 'secondary',
}

export default function AuditLogsPage() {
  const [page, setPage] = useState(1)
  const [actionFilter, setActionFilter] = useState('all')
  const [deleteId, setDeleteId] = useState<number | null>(null)

  const actionVal = actionFilter === 'all' ? undefined : actionFilter
  const { data, isLoading } = useAdminAuditLogs(page, 20, actionVal)
  const deleteLog = useDeleteAuditLog()

  const handleDelete = () => {
    if (deleteId !== null) deleteLog.mutate(deleteId, { onSuccess: () => setDeleteId(null) })
  }

  const totalPages = data ? Math.ceil(data.total / data.page_size) : 1

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold">审计日志</h1>
      </div>

      <Card>
        <CardHeader>
          <div className="flex items-center gap-2">
            <Select value={actionFilter} onValueChange={(v) => { setActionFilter(v); setPage(1) }}>
              <SelectTrigger className="w-40">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                {actionOptions.map((o) => (
                  <SelectItem key={o.value} value={o.value}>{o.label}</SelectItem>
                ))}
              </SelectContent>
            </Select>
            {actionFilter !== 'all' && (
              <Button variant="ghost" size="sm" className="h-9 px-2" onClick={() => { setActionFilter('all'); setPage(1) }}>
                <X className="h-4 w-4" />
              </Button>
            )}
          </div>
        </CardHeader>
        <CardContent>
          {isLoading ? (
            <div className="space-y-2">{Array.from({ length: 10 }).map((_, i) => <Skeleton key={i} className="h-12" />)}</div>
          ) : (
            <>
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>用户</TableHead>
                    <TableHead>操作</TableHead>
                    <TableHead>资源</TableHead>
                    <TableHead>详情</TableHead>
                    <TableHead>IP</TableHead>
                    <TableHead>时间</TableHead>
                    <TableHead className="text-right">操作</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {data?.list.map((log) => (
                    <TableRow key={log.id}>
                      <TableCell>{log.username}</TableCell>
                      <TableCell>
                        <Badge variant={actionVariant[log.action] ?? 'outline'}>{log.action}</Badge>
                      </TableCell>
                      <TableCell>{log.resource}{log.resource_id ? `#${log.resource_id}` : ''}</TableCell>
                      <TableCell className="max-w-48 truncate text-xs text-muted-foreground">{log.detail || '-'}</TableCell>
                      <TableCell className="font-mono text-xs">{log.ip}</TableCell>
                      <TableCell className="text-xs">{formatDate(log.created_at)}</TableCell>
                      <TableCell className="text-right">
                        <Button size="icon" variant="ghost" onClick={() => setDeleteId(log.id)}>
                          <Trash2 className="h-4 w-4 text-destructive" />
                        </Button>
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
              <div className="mt-4 flex items-center justify-between">
                <p className="text-sm text-muted-foreground">共 {data?.total ?? 0} 条</p>
                <div className="flex gap-2">
                  <Button size="sm" variant="outline" disabled={page <= 1} onClick={() => setPage(page - 1)}>上一页</Button>
                  <span className="flex items-center text-sm">{page} / {totalPages}</span>
                  <Button size="sm" variant="outline" disabled={page >= totalPages} onClick={() => setPage(page + 1)}>下一页</Button>
                </div>
              </div>
            </>
          )}
        </CardContent>
      </Card>

      <Dialog open={deleteId !== null} onOpenChange={() => setDeleteId(null)}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>确认删除</DialogTitle>
            <DialogDescription>确定要删除该审计日志吗？此操作不可撤销。</DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button variant="outline" onClick={() => setDeleteId(null)}>取消</Button>
            <Button variant="destructive" onClick={handleDelete} disabled={deleteLog.isPending}>确认删除</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}
