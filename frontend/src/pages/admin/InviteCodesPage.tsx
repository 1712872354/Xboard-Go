import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { toast } from 'sonner'
import api from '@/lib/api'
import type { InviteCode, CommissionLog, PaginatedResponse } from '@/types'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Skeleton } from '@/components/ui/skeleton'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { formatCurrency, formatDate } from '@/lib/utils'
import { Plus, Pencil, Trash2, Copy, Users, DollarSign, Wallet } from 'lucide-react'

const useAdminInviteCodes = (page = 1, pageSize = 20) =>
  useQuery({
    queryKey: ['admin', 'invite-codes', page, pageSize],
    queryFn: async () => (await api.get('/admin/invite-codes', { params: { page, page_size: pageSize } })) as unknown as PaginatedResponse<InviteCode>,
  })

const useAdminCommissionLogs = (page = 1, pageSize = 20) =>
  useQuery({
    queryKey: ['admin', 'commission-logs', page, pageSize],
    queryFn: async () => (await api.get('/admin/commissions', { params: { page, page_size: pageSize } })) as unknown as PaginatedResponse<CommissionLog>,
  })

const useCreateInviteCode = () => {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (data: Partial<InviteCode>) => await api.post('/admin/invite-codes', data),
    onSuccess: () => { qc.invalidateQueries({ queryKey: ['admin', 'invite-codes'] }); toast.success('邀请码创建成功') },
  })
}

const useUpdateInviteCode = () => {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async ({ id, ...data }: Partial<InviteCode> & { id: number }) => await api.put(`/admin/invite-codes/${id}`, data),
    onSuccess: () => { qc.invalidateQueries({ queryKey: ['admin', 'invite-codes'] }); toast.success('邀请码更新成功') },
  })
}

const useSettleCommission = () => {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (id: number) => await api.post(`/admin/commissions/${id}/settle`),
    onSuccess: () => { qc.invalidateQueries({ queryKey: ['admin', 'commission-logs'] }); toast.success('佣金已结算') },
  })
}

const inviteCodeSchema = z.object({
  user_id: z.coerce.number().min(1, '请输入用户 ID'),
  commission: z.coerce.number().min(0).max(100),
  limit_count: z.coerce.number().min(0),
})

type InviteCodeForm = z.infer<typeof inviteCodeSchema>

export default function InviteCodesPage() {
  const [codePage, setCodePage] = useState(1)
  const [logPage, setLogPage] = useState(1)
  const [dialog, setDialog] = useState<'create' | 'edit' | null>(null)
  const [editItem, setEditItem] = useState<InviteCode | null>(null)

  const { data: codes, isLoading: loadingCodes } = useAdminInviteCodes(codePage)
  const { data: logs, isLoading: loadingLogs } = useAdminCommissionLogs(logPage)
  const createCode = useCreateInviteCode()
  const updateCode = useUpdateInviteCode()
  const settleCommission = useSettleCommission()

  const allCodes = codes?.list ?? []
  const allLogs = logs?.list ?? []
  const pendingCommission = allLogs.filter((l) => l.status === 0).reduce((sum, l) => sum + l.commission, 0)
  const settledCommission = allLogs.filter((l) => l.status === 1).reduce((sum, l) => sum + l.commission, 0)

  const { register, handleSubmit, reset, formState: { errors } } = useForm<InviteCodeForm>({
    resolver: zodResolver(inviteCodeSchema),
  })

  const openCreate = () => {
    reset({ user_id: 0, commission: 10, limit_count: 10 })
    setDialog('create')
  }

  const openEdit = (c: InviteCode) => {
    reset({ user_id: c.user_id, commission: c.commission, limit_count: c.limit_count })
    setEditItem(c)
    setDialog('edit')
  }

  const handleCreate = (data: InviteCodeForm) => {
    createCode.mutate(data, { onSuccess: () => setDialog(null) })
  }

  const handleEdit = (data: InviteCodeForm) => {
    if (!editItem) return
    updateCode.mutate({ id: editItem.id, ...data }, { onSuccess: () => { setDialog(null); setEditItem(null) } })
  }

  const copyCode = (code: string) => {
    navigator.clipboard.writeText(code)
    toast.success('已复制邀请码')
  }

  const codeTotalPages = codes ? Math.ceil(codes.total / codes.page_size) : 1
  const logTotalPages = logs ? Math.ceil(logs.total / logs.page_size) : 1

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-bold">邀请 & 佣金</h1>

      <div className="grid gap-4 sm:grid-cols-3">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">邀请码总数</CardTitle>
            <Users className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent><div className="text-2xl font-bold">{codes?.total ?? 0}</div></CardContent>
        </Card>
        <Card>
          <CardHeader className="flex flex-row items-center justify-between pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">待结算佣金</CardTitle>
            <Wallet className="h-4 w-4 text-amber-500" />
          </CardHeader>
          <CardContent><div className="text-2xl font-bold">{formatCurrency(pendingCommission)}</div></CardContent>
        </Card>
        <Card>
          <CardHeader className="flex flex-row items-center justify-between pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">已结算佣金</CardTitle>
            <DollarSign className="h-4 w-4 text-emerald-500" />
          </CardHeader>
          <CardContent><div className="text-2xl font-bold">{formatCurrency(settledCommission)}</div></CardContent>
        </Card>
      </div>

      <Tabs defaultValue="codes">
        <TabsList>
          <TabsTrigger value="codes">邀请码</TabsTrigger>
          <TabsTrigger value="logs">佣金记录</TabsTrigger>
        </TabsList>

        <TabsContent value="codes">
          <Card>
            <CardHeader>
              <div className="flex justify-end">
                <Button onClick={openCreate}><Plus className="mr-2 h-4 w-4" />新建邀请码</Button>
              </div>
            </CardHeader>
            <CardContent>
              {loadingCodes ? (
                <div className="space-y-2">{Array.from({ length: 5 }).map((_, i) => <Skeleton key={i} className="h-12" />)}</div>
              ) : (
                <>
                  <Table>
                    <TableHeader>
                      <TableRow>
                        <TableHead>邀请码</TableHead>
                        <TableHead>用户 ID</TableHead>
                        <TableHead>佣金 %</TableHead>
                        <TableHead>使用/限制</TableHead>
                        <TableHead>状态</TableHead>
                        <TableHead className="text-right">操作</TableHead>
                      </TableRow>
                    </TableHeader>
                    <TableBody>
                      {allCodes.map((c) => (
                        <TableRow key={c.id}>
                          <TableCell>
                            <div className="flex items-center gap-1">
                              <span className="font-mono text-xs">{c.code}</span>
                              <Button size="icon" variant="ghost" className="h-5 w-5" onClick={() => copyCode(c.code)}>
                                <Copy className="h-3 w-3" />
                              </Button>
                            </div>
                          </TableCell>
                          <TableCell>{c.user_id}</TableCell>
                          <TableCell>{c.commission}%</TableCell>
                          <TableCell>{c.used_count} / {c.limit_count || '不限'}</TableCell>
                          <TableCell>
                            <Badge variant={c.status === 1 ? 'success' : 'secondary'}>{c.status === 1 ? '可用' : '禁用'}</Badge>
                          </TableCell>
                          <TableCell className="text-right">
                            <Button size="icon" variant="ghost" onClick={() => openEdit(c)}>
                              <Pencil className="h-4 w-4" />
                            </Button>
                          </TableCell>
                        </TableRow>
                      ))}
                    </TableBody>
                  </Table>
                  <div className="mt-4 flex items-center justify-between">
                    <p className="text-sm text-muted-foreground">共 {codes?.total ?? 0} 条</p>
                    <div className="flex gap-2">
                      <Button size="sm" variant="outline" disabled={codePage <= 1} onClick={() => setCodePage(codePage - 1)}>上一页</Button>
                      <span className="flex items-center text-sm">{codePage} / {codeTotalPages}</span>
                      <Button size="sm" variant="outline" disabled={codePage >= codeTotalPages} onClick={() => setCodePage(codePage + 1)}>下一页</Button>
                    </div>
                  </div>
                </>
              )}
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="logs">
          <Card>
            <CardHeader />
            <CardContent>
              {loadingLogs ? (
                <div className="space-y-2">{Array.from({ length: 8 }).map((_, i) => <Skeleton key={i} className="h-12" />)}</div>
              ) : (
                <>
                  <Table>
                    <TableHeader>
                      <TableRow>
                        <TableHead>邀请人</TableHead>
                        <TableHead>被邀请人</TableHead>
                        <TableHead>订单金额</TableHead>
                        <TableHead>佣金</TableHead>
                        <TableHead>状态</TableHead>
                        <TableHead>时间</TableHead>
                        <TableHead className="text-right">操作</TableHead>
                      </TableRow>
                    </TableHeader>
                    <TableBody>
                      {allLogs.map((l) => (
                        <TableRow key={l.id}>
                          <TableCell>{l.user_id}</TableCell>
                          <TableCell>{l.from_user_email ?? l.from_user_id}</TableCell>
                          <TableCell>{formatCurrency(l.order_amount)}</TableCell>
                          <TableCell>{formatCurrency(l.commission)}</TableCell>
                          <TableCell>
                            <Badge variant={l.status === 1 ? 'success' : 'warning'}>{l.status === 1 ? '已结算' : '待结算'}</Badge>
                          </TableCell>
                          <TableCell className="text-xs">{formatDate(l.created_at)}</TableCell>
                          <TableCell className="text-right">
                            {l.status === 0 && (
                              <Button size="sm" variant="outline" onClick={() => settleCommission.mutate(l.id)}>
                                结算
                              </Button>
                            )}
                          </TableCell>
                        </TableRow>
                      ))}
                    </TableBody>
                  </Table>
                  <div className="mt-4 flex items-center justify-between">
                    <p className="text-sm text-muted-foreground">共 {logs?.total ?? 0} 条</p>
                    <div className="flex gap-2">
                      <Button size="sm" variant="outline" disabled={logPage <= 1} onClick={() => setLogPage(logPage - 1)}>上一页</Button>
                      <span className="flex items-center text-sm">{logPage} / {logTotalPages}</span>
                      <Button size="sm" variant="outline" disabled={logPage >= logTotalPages} onClick={() => setLogPage(logPage + 1)}>下一页</Button>
                    </div>
                  </div>
                </>
              )}
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>

      <Dialog open={dialog !== null} onOpenChange={() => { setDialog(null); setEditItem(null) }}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{dialog === 'create' ? '新建邀请码' : '编辑邀请码'}</DialogTitle>
          </DialogHeader>
          <form onSubmit={handleSubmit(dialog === 'create' ? handleCreate : handleEdit)} className="space-y-4">
            <div className="space-y-2">
              <Label>用户 ID</Label>
              <Input type="number" {...register('user_id')} />
              {errors.user_id && <p className="text-xs text-destructive">{errors.user_id.message}</p>}
            </div>
            <div className="grid grid-cols-2 gap-4">
              <div className="space-y-2">
                <Label>佣金比例 (%)</Label>
                <Input type="number" {...register('commission')} />
              </div>
              <div className="space-y-2">
                <Label>使用次数限制</Label>
                <Input type="number" {...register('limit_count')} />
              </div>
            </div>
            <DialogFooter>
              <Button type="button" variant="outline" onClick={() => setDialog(null)}>取消</Button>
              <Button type="submit">保存</Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>
    </div>
  )
}
