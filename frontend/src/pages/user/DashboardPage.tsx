import { useQuery } from '@tanstack/react-query'
import { Link } from 'react-router-dom'
import {
  ArrowUpRight, ArrowDownRight, Package, ShoppingCart,
  Ticket, Users, Calendar, Wallet, Activity,
} from 'lucide-react'
import { useUserProfile } from '@/hooks/useAuth'
import { useOrders } from '@/hooks/useOrders'
import api from '@/lib/api'
import { formatBytes, formatDate, formatCurrency } from '@/lib/utils'
import type { Notice, Order, Plan } from '@/types'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Skeleton } from '@/components/ui/skeleton'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Separator } from '@/components/ui/separator'
import { Progress } from '@/components/ui/progress'

const orderStatusMap: Record<number, { label: string; variant: 'default' | 'secondary' | 'destructive' | 'outline' }> = {
  0: { label: '待支付', variant: 'secondary' },
  1: { label: '已支付', variant: 'default' },
  2: { label: '已取消', variant: 'outline' },
  3: { label: '已退款', variant: 'destructive' },
}

export default function DashboardPage() {
  const { data: user, isLoading: userLoading } = useUserProfile()
  const { data: ordersData, isLoading: ordersLoading } = useOrders(1, 5)

  const { data: notices, isLoading: noticesLoading } = useQuery({
    queryKey: ['user', 'notices'],
    queryFn: async () => (await api.get('/notices')) as unknown as Notice[],
  })

  const { data: plans } = useQuery({
    queryKey: ['plans'],
    queryFn: async () => (await api.get('/plans')) as unknown as Plan[],
  })

  const currentPlan = plans?.find((p) => p.id === user?.plan_id)

  if (userLoading) {
    return (
      <div className="space-y-6">
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
          {Array.from({ length: 4 }).map((_, i) => (
            <Card key={i}>
              <CardHeader className="pb-2">
                <Skeleton className="h-4 w-24" />
              </CardHeader>
              <CardContent>
                <Skeleton className="h-8 w-32" />
              </CardContent>
            </Card>
          ))}
        </div>
      </div>
    )
  }

  const trafficUsed = user?.used_traffic || 0
  const trafficTotal = user?.traffic_limit || 0
  const trafficPercent = trafficTotal > 0 ? Math.min((trafficUsed / trafficTotal) * 100, 100) : 0
  const isExpired = user?.expired_at ? new Date(user.expired_at) < new Date() : false
  const daysLeft = user?.expired_at
    ? Math.max(0, Math.ceil((new Date(user.expired_at).getTime() - Date.now()) / (1000 * 60 * 60 * 24)))
    : null

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-bold">仪表盘</h1>

      {/* 套餐信息卡片 */}
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <div>
              <CardTitle className="flex items-center gap-2">
                <Package className="h-5 w-5" />
                {currentPlan?.name || '未订阅套餐'}
              </CardTitle>
              <CardDescription>
                {user?.expired_at
                  ? isExpired
                    ? '套餐已过期'
                    : `剩余 ${daysLeft} 天 · 到期时间：${formatDate(user.expired_at)}`
                  : '未订阅套餐'}
              </CardDescription>
            </div>
            <Badge variant={isExpired ? 'destructive' : 'default'}>
              {isExpired ? '已过期' : '有效'}
            </Badge>
          </div>
        </CardHeader>
        <CardContent className="space-y-4">
          {/* 流量进度条 */}
          <div className="space-y-2">
            <div className="flex items-center justify-between text-sm">
              <span className="text-muted-foreground">流量使用</span>
              <span className="font-medium">
                {formatBytes(trafficUsed)} / {trafficTotal > 0 ? formatBytes(trafficTotal) : '无限制'}
              </span>
            </div>
            <Progress value={trafficPercent} className="h-2" />
            <p className="text-xs text-muted-foreground text-right">
              {trafficPercent.toFixed(1)}%
            </p>
          </div>

          {/* 统计卡片 */}
          <div className="grid gap-4 md:grid-cols-4">
            <div className="flex items-center gap-3 rounded-lg border p-3">
              <Activity className="h-4 w-4 text-muted-foreground" />
              <div>
                <p className="text-xs text-muted-foreground">已用流量</p>
                <p className="text-sm font-medium">{formatBytes(trafficUsed)}</p>
              </div>
            </div>
            <div className="flex items-center gap-3 rounded-lg border p-3">
              <Calendar className="h-4 w-4 text-muted-foreground" />
              <div>
                <p className="text-xs text-muted-foreground">到期时间</p>
                <p className="text-sm font-medium">
                  {user?.expired_at ? formatDate(user.expired_at, { month: '2-digit', day: '2-digit' }) : '无限制'}
                </p>
              </div>
            </div>
            <div className="flex items-center gap-3 rounded-lg border p-3">
              <Wallet className="h-4 w-4 text-muted-foreground" />
              <div>
                <p className="text-xs text-muted-foreground">账户余额</p>
                <p className="text-sm font-medium">{formatCurrency(user?.balance || 0)}</p>
              </div>
            </div>
            <div className="flex items-center gap-3 rounded-lg border p-3">
              <ArrowUpRight className="h-4 w-4 text-muted-foreground" />
              <div>
                <p className="text-xs text-muted-foreground">佣金收入</p>
                <p className="text-sm font-medium">{formatCurrency(user?.commission || 0)}</p>
              </div>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Quick Actions */}
      <div className="grid gap-4 grid-cols-2 md:grid-cols-4">
        <Button asChild variant="outline" className="h-auto flex-col gap-2 py-4">
          <Link to="/user/plans">
            <Package className="h-5 w-5" />
            <span className="text-sm">查看套餐</span>
          </Link>
        </Button>
        <Button asChild variant="outline" className="h-auto flex-col gap-2 py-4">
          <Link to="/user/orders">
            <ShoppingCart className="h-5 w-5" />
            <span className="text-sm">我的订单</span>
          </Link>
        </Button>
        <Button asChild variant="outline" className="h-auto flex-col gap-2 py-4">
          <Link to="/user/tickets">
            <Ticket className="h-5 w-5" />
            <span className="text-sm">提交工单</span>
          </Link>
        </Button>
        <Button asChild variant="outline" className="h-auto flex-col gap-2 py-4">
          <Link to="/user/invite">
            <Users className="h-5 w-5" />
            <span className="text-sm">邀请好友</span>
          </Link>
        </Button>
      </div>

      <div className="grid gap-6 lg:grid-cols-2">
        {/* Recent Orders */}
        <Card>
          <CardHeader>
            <CardTitle>最近订单</CardTitle>
            <CardDescription>最近5笔订单记录</CardDescription>
          </CardHeader>
          <CardContent>
            {ordersLoading ? (
              <div className="space-y-3">
                {Array.from({ length: 3 }).map((_, i) => (
                  <Skeleton key={i} className="h-12 w-full" />
                ))}
              </div>
            ) : ordersData?.list?.length ? (
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>订单号</TableHead>
                    <TableHead>金额</TableHead>
                    <TableHead>状态</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {ordersData.list.map((order: Order) => (
                    <TableRow key={order.id}>
                      <TableCell className="font-mono text-xs">
                        {order.trade_no.slice(0, 12)}...
                      </TableCell>
                      <TableCell>{formatCurrency(order.amount)}</TableCell>
                      <TableCell>
                        <Badge variant={orderStatusMap[order.status]?.variant || 'secondary'}>
                          {orderStatusMap[order.status]?.label || '未知'}
                        </Badge>
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            ) : (
              <p className="text-sm text-muted-foreground text-center py-4">暂无订单</p>
            )}
          </CardContent>
        </Card>

        {/* Recent Notices */}
        <Card>
          <CardHeader>
            <CardTitle>最新公告</CardTitle>
            <CardDescription>系统公告与通知</CardDescription>
          </CardHeader>
          <CardContent>
            {noticesLoading ? (
              <div className="space-y-3">
                {Array.from({ length: 3 }).map((_, i) => (
                  <Skeleton key={i} className="h-12 w-full" />
                ))}
              </div>
            ) : notices?.length ? (
              <div className="space-y-4">
                {notices.slice(0, 5).map((notice) => (
                  <div key={notice.id}>
                    <div className="flex items-center justify-between">
                      <p className="text-sm font-medium">{notice.title}</p>
                      <span className="text-xs text-muted-foreground">
                        {formatDate(notice.created_at, { month: '2-digit', day: '2-digit' })}
                      </span>
                    </div>
                    <p className="text-xs text-muted-foreground line-clamp-1 mt-1">
                      {notice.content.replace(/<[^>]*>/g, '')}
                    </p>
                    <Separator className="mt-3" />
                  </div>
                ))}
              </div>
            ) : (
              <p className="text-sm text-muted-foreground text-center py-4">暂无公告</p>
            )}
          </CardContent>
        </Card>
      </div>
    </div>
  )
}
