import { useState } from 'react'
import { toast } from 'sonner'
import { Loader2, CreditCard, XCircle, ExternalLink, Check } from 'lucide-react'
import { useOrders, useCancelOrder } from '@/hooks/useOrders'
import { usePaymentMethods, useCreatePayment } from '@/hooks/usePayment'
import { formatDate, formatCurrency, cn } from '@/lib/utils'
import type { Order } from '@/types'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Skeleton } from '@/components/ui/skeleton'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import {
  Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle,
} from '@/components/ui/dialog'

const statusMap: Record<number, { label: string; variant: 'default' | 'secondary' | 'destructive' | 'outline' }> = {
  0: { label: '待支付', variant: 'secondary' },
  1: { label: '已支付', variant: 'default' },
  2: { label: '已取消', variant: 'outline' },
  3: { label: '已退款', variant: 'destructive' },
}

export default function OrdersPage() {
  const [page, setPage] = useState(1)
  const [cancelId, setCancelId] = useState<number | null>(null)
  const [payOrder, setPayOrder] = useState<Order | null>(null)
  const [selectedMethod, setSelectedMethod] = useState('')

  const { data, isLoading } = useOrders(page, 10)
  const cancelOrder = useCancelOrder()
  const { data: paymentMethods } = usePaymentMethods()
  const createPayment = useCreatePayment()

  const handleCancel = () => {
    if (!cancelId) return
    cancelOrder.mutate(cancelId, {
      onSuccess: () => {
        toast.success('订单已取消')
        setCancelId(null)
      },
      onError: (err: any) => {
        toast.error(err?.message || '取消失败')
      },
    })
  }

  const handlePay = () => {
    if (!payOrder || !selectedMethod) return

    createPayment.mutate(
      { order_id: payOrder.id, payment_method: selectedMethod },
      {
        onSuccess: (res: any) => {
          if (res.payment_url) {
            window.open(res.payment_url, '_blank')
            toast.success('已跳转到支付页面')
          } else {
            toast.success('支付创建成功')
          }
          setPayOrder(null)
        },
        onError: (err: any) => {
          toast.error(err?.message || '创建支付失败')
        },
      },
    )
  }

  const openPayDialog = (order: Order) => {
    setPayOrder(order)
    setSelectedMethod(paymentMethods?.[0]?.id || '')
  }

  const totalPages = data ? Math.ceil(data.total / data.page_size) : 1

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-bold">我的订单</h1>

      <Card>
        <CardHeader>
          <CardTitle>订单列表</CardTitle>
          <CardDescription>查看和管理您的订单</CardDescription>
        </CardHeader>
        <CardContent>
          {isLoading ? (
            <div className="space-y-3">
              {Array.from({ length: 5 }).map((_, i) => (
                <Skeleton key={i} className="h-12 w-full" />
              ))}
            </div>
          ) : (
            <>
              <div className="overflow-x-auto">
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>订单号</TableHead>
                      <TableHead>套餐</TableHead>
                      <TableHead>金额</TableHead>
                      <TableHead>优惠</TableHead>
                      <TableHead>实付</TableHead>
                      <TableHead>状态</TableHead>
                      <TableHead>创建时间</TableHead>
                      <TableHead className="text-right">操作</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {data?.list?.length ? (
                      data.list.map((order: Order) => (
                        <TableRow key={order.id}>
                          <TableCell className="font-mono text-xs">
                            {order.trade_no}
                          </TableCell>
                          <TableCell>{order.plan_name || order.plan?.name || '-'}</TableCell>
                          <TableCell>{formatCurrency(order.amount)}</TableCell>
                          <TableCell>
                            {order.discount && order.discount > 0 ? (
                              <span className="text-green-600">-{formatCurrency(order.discount)}</span>
                            ) : order.coupon_code ? (
                              <span className="text-xs text-muted-foreground">{order.coupon_code}</span>
                            ) : (
                              '-'
                            )}
                          </TableCell>
                          <TableCell className="font-medium">
                            {formatCurrency(order.actual_amount || order.amount)}
                          </TableCell>
                          <TableCell>
                            <Badge variant={statusMap[order.status]?.variant || 'secondary'}>
                              {statusMap[order.status]?.label || '未知'}
                            </Badge>
                          </TableCell>
                          <TableCell className="text-sm text-muted-foreground">
                            {formatDate(order.created_at)}
                          </TableCell>
                          <TableCell className="text-right">
                            <div className="flex justify-end gap-2">
                              {order.status === 0 && (
                                <>
                                  <Button size="sm" onClick={() => openPayDialog(order)}>
                                    <CreditCard className="mr-1 h-3 w-3" />
                                    支付
                                  </Button>
                                  <Button
                                    size="sm"
                                    variant="outline"
                                    onClick={() => setCancelId(order.id)}
                                  >
                                    <XCircle className="mr-1 h-3 w-3" />
                                    取消
                                  </Button>
                                </>
                              )}
                            </div>
                          </TableCell>
                        </TableRow>
                      ))
                    ) : (
                      <TableRow>
                        <TableCell colSpan={8} className="text-center py-8 text-muted-foreground">
                          暂无订单记录
                        </TableCell>
                      </TableRow>
                    )}
                  </TableBody>
                </Table>
              </div>

              {totalPages > 1 && (
                <div className="flex items-center justify-between mt-4">
                  <p className="text-sm text-muted-foreground">
                    共 {data?.total || 0} 条记录
                  </p>
                  <div className="flex gap-2">
                    <Button
                      variant="outline"
                      size="sm"
                      disabled={page <= 1}
                      onClick={() => setPage((p) => p - 1)}
                    >
                      上一页
                    </Button>
                    <Button
                      variant="outline"
                      size="sm"
                      disabled={page >= totalPages}
                      onClick={() => setPage((p) => p + 1)}
                    >
                      下一页
                    </Button>
                  </div>
                </div>
              )}
            </>
          )}
        </CardContent>
      </Card>

      {/* Payment Dialog */}
      <Dialog open={!!payOrder} onOpenChange={() => setPayOrder(null)}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>选择支付方式</DialogTitle>
            <DialogDescription>
              订单号：{payOrder?.trade_no} · 实付：{formatCurrency(payOrder?.actual_amount || payOrder?.amount || 0)}
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-4">
            {paymentMethods && paymentMethods.length > 0 ? (
              <div className="space-y-2">
                {paymentMethods.filter((m) => m.enabled).map((method) => (
                  <button
                    key={method.id}
                    className={cn(
                      "flex items-center w-full rounded-lg border p-3 transition-colors",
                      selectedMethod === method.id
                        ? "border-primary bg-primary/5"
                        : "hover:bg-muted/50"
                    )}
                    onClick={() => setSelectedMethod(method.id)}
                  >
                    <div className={cn(
                      "mr-3 h-4 w-4 rounded-full border-2",
                      selectedMethod === method.id
                        ? "border-primary bg-primary"
                        : "border-muted-foreground"
                    )}>
                      {selectedMethod === method.id && (
                        <Check className="h-3 w-3 text-primary-foreground" />
                      )}
                    </div>
                    <span className="font-medium">{method.name}</span>
                  </button>
                ))}
              </div>
            ) : (
              <p className="text-sm text-muted-foreground text-center py-4">
                暂无可用支付方式
              </p>
            )}
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setPayOrder(null)}>
              取消
            </Button>
            <Button
              onClick={handlePay}
              disabled={!selectedMethod || createPayment.isPending}
            >
              {createPayment.isPending ? (
                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
              ) : (
                <ExternalLink className="mr-2 h-4 w-4" />
              )}
              去支付
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Cancel Dialog */}
      <Dialog open={cancelId !== null} onOpenChange={() => setCancelId(null)}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>确认取消订单</DialogTitle>
            <DialogDescription>
              确定要取消此订单吗？此操作无法撤销。
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button variant="outline" onClick={() => setCancelId(null)}>
              返回
            </Button>
            <Button variant="destructive" onClick={handleCancel} disabled={cancelOrder.isPending}>
              {cancelOrder.isPending && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
              确认取消
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}
