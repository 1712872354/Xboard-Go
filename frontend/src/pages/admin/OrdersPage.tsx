import { useState } from 'react'
import { useQuery, useQueryClient, useMutation } from '@tanstack/react-query'
import { toast } from 'sonner'
import api from '@/lib/api'
import { Card, CardContent, CardHeader } from '@/components/ui/card'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Skeleton } from '@/components/ui/skeleton'
import { Separator } from '@/components/ui/separator'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { formatCurrency, formatDate, formatBytes } from '@/lib/utils'
import { CheckCircle, Search, Eye, Loader2 } from 'lucide-react'
import type { Order, PaginatedResponse } from '@/types'

const statusOptions = [
  { value: 'all', label: '全部' },
  { value: '0', label: '待支付' },
  { value: '1', label: '已支付' },
  { value: '2', label: '已取消' },
  { value: '3', label: '已退款' },
]

const statusMap: Record<number, { label: string; variant: 'default' | 'secondary' | 'success' | 'destructive' | 'warning' }> = {
  0: { label: '待支付', variant: 'warning' },
  1: { label: '已支付', variant: 'success' },
  2: { label: '已取消', variant: 'secondary' },
  3: { label: '已退款', variant: 'destructive' },
}

export default function OrdersPage() {
  const [page, setPage] = useState(1)
  const [statusFilter, setStatusFilter] = useState<string>('all')
  const [searchKeyword, setSearchKeyword] = useState('')
  const [confirmTradeNo, setConfirmTradeNo] = useState<string | null>(null)
  const [viewOrder, setViewOrder] = useState<Order | null>(null)

  const qc = useQueryClient()
  const statusNum = statusFilter === 'all' ? undefined : Number(statusFilter)

  const { data, isLoading } = useQuery({
    queryKey: ['admin', 'orders', page, statusNum, searchKeyword],
    queryFn: async () =>
      (await api.get('/admin/orders', {
        params: { page, page_size: 20, status: statusNum, keyword: searchKeyword || undefined },
      })) as unknown as PaginatedResponse<Order>,
  })

  const confirmPayment = useMutation({
    mutationFn: async (tradeNo: string) =>
      await api.post('/admin/orders/confirm-payment', { trade_no: tradeNo }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['admin', 'orders'] })
      toast.success('支付确认成功')
      setConfirmTradeNo(null)
    },
  })

  // Fetch order details when viewing
  const { data: orderDetail, isLoading: detailLoading } = useQuery({
    queryKey: ['admin', 'order', viewOrder?.id],
    queryFn: async () => (await api.get(`/orders/${viewOrder?.id}`)) as unknown as Order,
    enabled: !!viewOrder?.id,
  })

  const totalPages = data ? Math.ceil(data.total / data.page_size) : 1

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold">订单管理</h1>
      </div>

      <Card>
        <CardHeader>
          <div className="flex items-center gap-4">
            <Select value={statusFilter} onValueChange={setStatusFilter}>
              <SelectTrigger className="w-40">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                {statusOptions.map((o) => (
                  <SelectItem key={o.value} value={o.value}>{o.label}</SelectItem>
                ))}
              </SelectContent>
            </Select>
            <div className="flex items-center gap-2">
              <Input
                placeholder="搜索订单号..."
                value={searchKeyword}
                onChange={(e) => setSearchKeyword(e.target.value)}
                onKeyDown={(e) => e.key === 'Enter' && qc.invalidateQueries({ queryKey: ['admin', 'orders'] })}
                className="max-w-sm"
              />
              <Button
                size="sm"
                variant="outline"
                onClick={() => qc.invalidateQueries({ queryKey: ['admin', 'orders'] })}
              >
                <Search className="h-4 w-4" />
              </Button>
            </div>
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
                    <TableHead>订单号</TableHead>
                    <TableHead>用户</TableHead>
                    <TableHead>套餐</TableHead>
                    <TableHead>金额</TableHead>
                    <TableHead>优惠</TableHead>
                    <TableHead>实付</TableHead>
                    <TableHead>状态</TableHead>
                    <TableHead>时间</TableHead>
                    <TableHead className="text-right">操作</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {data?.list.map((o) => (
                    <TableRow key={o.id}>
                      <TableCell className="font-mono text-xs">{o.trade_no}</TableCell>
                      <TableCell className="text-sm">{o.user_id}</TableCell>
                      <TableCell>{o.plan_name ?? o.plan?.name ?? '-'}</TableCell>
                      <TableCell>{formatCurrency(o.amount)}</TableCell>
                      <TableCell>
                        {o.discount && o.discount > 0 ? (
                          <span className="text-green-600">-{formatCurrency(o.discount)}</span>
                        ) : o.coupon_code ? (
                          <span className="text-xs text-muted-foreground">{o.coupon_code}</span>
                        ) : (
                          '-'
                        )}
                      </TableCell>
                      <TableCell className="font-medium">{formatCurrency(o.actual_amount || o.amount)}</TableCell>
                      <TableCell>
                        <Badge variant={statusMap[o.status]?.variant}>{statusMap[o.status]?.label}</Badge>
                      </TableCell>
                      <TableCell className="text-xs">{formatDate(o.created_at)}</TableCell>
                      <TableCell className="text-right">
                        <div className="flex justify-end gap-1">
                          <Button size="icon" variant="ghost" onClick={() => setViewOrder(o)} title="查看详情">
                            <Eye className="h-4 w-4" />
                          </Button>
                          {o.status === 0 && (
                            <Button size="icon" variant="ghost" onClick={() => setConfirmTradeNo(o.trade_no)} title="确认支付">
                              <CheckCircle className="h-4 w-4 text-green-600" />
                            </Button>
                          )}
                        </div>
                      </TableCell>
                    </TableRow>
                  ))}
                  {data?.list.length === 0 && (
                    <TableRow>
                      <TableCell colSpan={9} className="text-center py-8 text-muted-foreground">
                        暂无订单记录
                      </TableCell>
                    </TableRow>
                  )}
                </TableBody>
              </Table>

              {totalPages > 1 && (
                <div className="mt-4 flex items-center justify-between">
                  <p className="text-sm text-muted-foreground">共 {data?.total ?? 0} 条</p>
                  <div className="flex gap-2">
                    <Button size="sm" variant="outline" disabled={page <= 1} onClick={() => setPage(page - 1)}>上一页</Button>
                    <span className="flex items-center text-sm">{page} / {totalPages}</span>
                    <Button size="sm" variant="outline" disabled={page >= totalPages} onClick={() => setPage(page + 1)}>下一页</Button>
                  </div>
                </div>
              )}
            </>
          )}
        </CardContent>
      </Card>

      {/* Order Detail Dialog */}
      <Dialog open={!!viewOrder} onOpenChange={() => setViewOrder(null)}>
        <DialogContent className="sm:max-w-lg">
          <DialogHeader>
            <DialogTitle>订单详情</DialogTitle>
            <DialogDescription>订单号：{viewOrder?.trade_no}</DialogDescription>
          </DialogHeader>
          {detailLoading ? (
            <div className="space-y-3">{Array.from({ length: 6 }).map((_, i) => <Skeleton key={i} className="h-8" />)}</div>
          ) : orderDetail ? (
            <div className="space-y-4">
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <Label className="text-muted-foreground">订单号</Label>
                  <p className="text-sm font-mono">{orderDetail.trade_no}</p>
                </div>
                <div>
                  <Label className="text-muted-foreground">状态</Label>
                  <p><Badge variant={statusMap[orderDetail.status]?.variant}>{statusMap[orderDetail.status]?.label}</Badge></p>
                </div>
                <div>
                  <Label className="text-muted-foreground">套餐</Label>
                  <p className="text-sm">{orderDetail.plan?.name || orderDetail.plan_name || `#${orderDetail.plan_id}`}</p>
                </div>
                <div>
                  <Label className="text-muted-foreground">支付方式</Label>
                  <p className="text-sm">{orderDetail.payment_method || '-'}</p>
                </div>
              </div>
              <Separator />
              <div className="space-y-2">
                <div className="flex justify-between text-sm">
                  <span className="text-muted-foreground">套餐原价</span>
                  <span>{formatCurrency(orderDetail.amount)}</span>
                </div>
                {orderDetail.discount && orderDetail.discount > 0 && (
                  <div className="flex justify-between text-sm">
                    <span className="text-muted-foreground">优惠券折扣</span>
                    <span className="text-green-600">-{formatCurrency(orderDetail.discount)}</span>
                  </div>
                )}
                {orderDetail.coupon_code && (
                  <div className="flex justify-between text-sm">
                    <span className="text-muted-foreground">优惠券码</span>
                    <span>{orderDetail.coupon_code}</span>
                  </div>
                )}
                <Separator />
                <div className="flex justify-between font-medium">
                  <span>实付金额</span>
                  <span className="text-lg">{formatCurrency(orderDetail.actual_amount || orderDetail.amount)}</span>
                </div>
              </div>
              <Separator />
              <div className="grid grid-cols-2 gap-4 text-sm">
                <div>
                  <Label className="text-muted-foreground">创建时间</Label>
                  <p>{formatDate(orderDetail.created_at)}</p>
                </div>
                <div>
                  <Label className="text-muted-foreground">支付时间</Label>
                  <p>{orderDetail.paid_at ? formatDate(orderDetail.paid_at) : '-'}</p>
                </div>
              </div>
            </div>
          ) : null}
          <DialogFooter>
            <Button variant="outline" onClick={() => setViewOrder(null)}>关闭</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Confirm Payment Dialog */}
      <Dialog open={confirmTradeNo !== null} onOpenChange={() => setConfirmTradeNo(null)}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>确认支付</DialogTitle>
            <DialogDescription>确认订单 {confirmTradeNo} 已支付？</DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button variant="outline" onClick={() => setConfirmTradeNo(null)}>取消</Button>
            <Button
              onClick={() => confirmTradeNo && confirmPayment.mutate(confirmTradeNo)}
              disabled={confirmPayment.isPending}
            >
              {confirmPayment.isPending && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
              确认
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}
