import { useState } from 'react'
import {
  useDashboardOverview,
  useRecentOrders,
  useRecentUsers,
  useIncomeStats,
  useUserGrowthStats,
  useNodeStats,
  useComprehensiveStats,
} from '@/hooks/useDashboard'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Badge } from '@/components/ui/badge'
import { Skeleton } from '@/components/ui/skeleton'
import { formatCurrency, formatDate, formatBytes } from '@/lib/utils'
import {
  Users, ShoppingCart, DollarSign, Server,
  TrendingUp, UserPlus, Ticket, CreditCard, Activity, Wifi,
} from 'lucide-react'
import ReactECharts from 'echarts-for-react'
import { useThemeStore } from '@/stores/theme'

function StatCard({ title, value, subtitle, icon: Icon, loading, trend }: {
  title: string
  value: string | number
  subtitle?: string
  icon: React.ElementType
  loading: boolean
  trend?: 'up' | 'down' | 'neutral'
}) {
  if (loading) return <Skeleton className="h-[120px] rounded-xl" />
  return (
    <Card className="relative overflow-hidden">
      <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
        <CardTitle className="text-sm font-medium text-muted-foreground">{title}</CardTitle>
        <div className="rounded-lg bg-primary/10 p-2">
          <Icon className="h-4 w-4 text-primary" />
        </div>
      </CardHeader>
      <CardContent>
        <div className="text-2xl font-bold tracking-tight">{value}</div>
        {subtitle && (
          <p className="mt-1 flex items-center gap-1 text-xs text-muted-foreground">
            {trend === 'up' && <TrendingUp className="h-3 w-3 text-emerald-500" />}
            {trend === 'down' && <TrendingUp className="h-3 w-3 text-red-500 rotate-180" />}
            {subtitle}
          </p>
        )}
      </CardContent>
    </Card>
  )
}

const orderStatusMap: Record<number, { label: string; variant: 'default' | 'secondary' | 'success' | 'destructive' | 'warning' }> = {
  0: { label: '待支付', variant: 'warning' },
  1: { label: '已支付', variant: 'success' },
  2: { label: '已取消', variant: 'secondary' },
  3: { label: '已退款', variant: 'destructive' },
}

function chartTheme(isDark: boolean) {
  return {
    backgroundColor: 'transparent',
    textStyle: { color: isDark ? '#a1a1aa' : '#71717a' },
    axisLine: { lineStyle: { color: isDark ? '#27272a' : '#e4e4e7' } },
    splitLine: { lineStyle: { color: isDark ? '#27272a' : '#f4f4f5' } },
  }
}

export default function DashboardPage() {
  const isDark = useThemeStore((s) => s.theme === 'dark')
  const { data: overview, isLoading: loadingOverview } = useDashboardOverview()
  const { data: compStats, isLoading: loadingComp } = useComprehensiveStats()
  const { data: incomeData, isLoading: loadingIncome } = useIncomeStats(7)
  const { data: growthData, isLoading: loadingGrowth } = useUserGrowthStats(7)
  const { data: nodeStats, isLoading: loadingNodeStats } = useNodeStats()
  const { data: recentOrders, isLoading: loadingOrders } = useRecentOrders(5)
  const { data: recentUsers, isLoading: loadingUsers } = useRecentUsers(5)

  const theme = chartTheme(isDark)
  const primaryColor = isDark ? '#818cf8' : '#6366f1'
  const emeraldColor = isDark ? '#34d399' : '#10b981'

  const incomeChartOption = {
    ...theme,
    tooltip: {
      trigger: 'axis' as const,
      backgroundColor: isDark ? '#27272a' : '#fff',
      borderColor: isDark ? '#3f3f46' : '#e4e4e7',
      textStyle: { color: isDark ? '#fafafa' : '#18181b', fontSize: 12 },
      formatter: (params: Array<{ name: string; value: number }>) => {
        const p = params[0]
        return `<div style="font-size:12px"><b>${p.name}</b><br/>收入: ¥${(p.value / 100).toFixed(2)}</div>`
      },
    },
    xAxis: {
      type: 'category' as const,
      data: incomeData?.map((i) => i.date.slice(5)) ?? [],
      axisLine: theme.axisLine,
      axisTick: { show: false },
      axisLabel: { color: theme.textStyle.color, fontSize: 11 },
    },
    yAxis: {
      type: 'value' as const,
      splitLine: theme.splitLine,
      axisLabel: {
        color: theme.textStyle.color,
        fontSize: 11,
        formatter: (v: number) => `¥${(v / 100).toFixed(0)}`,
      },
    },
    series: [{
      name: '收入',
      type: 'line',
      smooth: true,
      symbol: 'circle',
      symbolSize: 6,
      data: incomeData?.map((i) => i.amount) ?? [],
      lineStyle: { color: primaryColor, width: 2 },
      itemStyle: { color: primaryColor },
      areaStyle: {
        color: {
          type: 'linear' as const, x: 0, y: 0, x2: 0, y2: 1,
          colorStops: [
            { offset: 0, color: isDark ? 'rgba(129,140,248,0.2)' : 'rgba(99,102,241,0.15)' },
            { offset: 1, color: 'rgba(0,0,0,0)' },
          ],
        },
      },
    }],
    grid: { left: 55, right: 16, bottom: 28, top: 16 },
  }

  const growthChartOption = {
    ...theme,
    tooltip: {
      trigger: 'axis' as const,
      backgroundColor: isDark ? '#27272a' : '#fff',
      borderColor: isDark ? '#3f3f46' : '#e4e4e7',
      textStyle: { color: isDark ? '#fafafa' : '#18181b', fontSize: 12 },
    },
    xAxis: {
      type: 'category' as const,
      data: growthData?.map((i) => i.date.slice(5)) ?? [],
      axisLine: theme.axisLine,
      axisTick: { show: false },
      axisLabel: { color: theme.textStyle.color, fontSize: 11 },
    },
    yAxis: {
      type: 'value' as const,
      splitLine: theme.splitLine,
      axisLabel: { color: theme.textStyle.color, fontSize: 11 },
      minInterval: 1,
    },
    series: [{
      name: '新用户',
      type: 'bar',
      barWidth: '50%',
      data: growthData?.map((i) => i.new_users) ?? [],
      itemStyle: {
        color: {
          type: 'linear' as const, x: 0, y: 0, x2: 0, y2: 1,
          colorStops: [
            { offset: 0, color: emeraldColor },
            { offset: 1, color: isDark ? 'rgba(52,211,153,0.3)' : 'rgba(16,185,129,0.3)' },
          ],
        },
        borderRadius: [4, 4, 0, 0],
      },
    }],
    grid: { left: 40, right: 16, bottom: 28, top: 16 },
  }

  const nodeChartOption = {
    ...theme,
    tooltip: {
      trigger: 'item' as const,
      backgroundColor: isDark ? '#27272a' : '#fff',
      borderColor: isDark ? '#3f3f46' : '#e4e4e7',
      textStyle: { color: isDark ? '#fafafa' : '#18181b', fontSize: 12 },
      formatter: '{b}: {c} ({d}%)',
    },
    legend: {
      bottom: 0,
      textStyle: { color: theme.textStyle.color, fontSize: 11 },
      itemWidth: 10,
      itemHeight: 10,
      itemGap: 16,
    },
    series: [{
      type: 'pie',
      radius: ['45%', '72%'],
      center: ['50%', '42%'],
      avoidLabelOverlap: true,
      label: { show: false },
      emphasis: {
        label: { show: true, fontSize: 13, fontWeight: 'bold' },
        scaleSize: 6,
      },
      data: [
        { name: '在线', value: nodeStats?.online ?? 0, itemStyle: { color: '#22c55e' } },
        { name: '离线', value: nodeStats?.offline ?? 0, itemStyle: { color: '#ef4444' } },
        { name: '维护', value: nodeStats?.maintenance ?? 0, itemStyle: { color: '#f59e0b' } },
      ],
    }],
  }

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold tracking-tight">仪表盘</h1>
        <p className="text-muted-foreground">系统运行概览</p>
      </div>

      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
        <StatCard
          title="今日收入"
          value={formatCurrency(compStats?.today_income ?? 0)}
          subtitle={compStats?.day_income_growth ? `较昨日 ${compStats.day_income_growth > 0 ? '+' : ''}${compStats.day_income_growth.toFixed(1)}%` : undefined}
          icon={DollarSign}
          loading={loadingComp}
          trend={compStats?.day_income_growth && compStats.day_income_growth > 0 ? 'up' : compStats?.day_income_growth && compStats.day_income_growth < 0 ? 'down' : undefined}
        />
        <StatCard
          title="本月收入"
          value={formatCurrency(compStats?.current_month_income ?? 0)}
          subtitle={compStats?.month_income_growth ? `较上月 ${compStats.month_income_growth > 0 ? '+' : ''}${compStats.month_income_growth.toFixed(1)}%` : undefined}
          icon={ShoppingCart}
          loading={loadingComp}
          trend={compStats?.month_income_growth && compStats.month_income_growth > 0 ? 'up' : compStats?.month_income_growth && compStats.month_income_growth < 0 ? 'down' : undefined}
        />
        <StatCard
          title="总用户"
          value={compStats?.total_users ?? overview?.total_users ?? 0}
          subtitle={`本月新增 ${compStats?.current_month_new_users ?? 0}`}
          icon={Users}
          loading={loadingComp || loadingOverview}
          trend={compStats?.user_growth && compStats.user_growth > 0 ? 'up' : undefined}
        />
        <StatCard
          title="在线用户"
          value={compStats?.online_users ?? 0}
          subtitle={`在线设备 ${compStats?.online_devices ?? 0}`}
          icon={Wifi}
          loading={loadingComp}
        />
      </div>

      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
        <StatCard
          title="节点"
          value={overview?.total_nodes ?? 0}
          subtitle={`在线 ${compStats?.online_nodes ?? overview?.online_nodes ?? 0}`}
          icon={Server}
          loading={loadingOverview || loadingComp}
        />
        <StatCard
          title="待处理工单"
          value={compStats?.ticket_pending_total ?? overview?.open_tickets ?? 0}
          icon={Ticket}
          loading={loadingComp || loadingOverview}
        />
        <StatCard
          title="待结算佣金"
          value={compStats?.commission_pending_total ?? 0}
          subtitle={`本月发放 ${formatCurrency(compStats?.current_month_commission_payout ?? 0)}`}
          icon={CreditCard}
          loading={loadingComp}
        />
        <StatCard
          title="今日流量"
          value={formatBytes(compStats?.today_traffic?.total ?? 0)}
          subtitle={`本月 ${formatBytes(compStats?.month_traffic?.total ?? 0)}`}
          icon={Activity}
          loading={loadingComp}
        />
      </div>

      <div className="grid gap-4 xl:grid-cols-3">
        <Card className="xl:col-span-2">
          <CardHeader className="flex flex-row items-center justify-between">
            <CardTitle className="text-base">收入趋势</CardTitle>
            <span className="text-xs text-muted-foreground">近 7 天</span>
          </CardHeader>
          <CardContent>
            {loadingIncome
              ? <Skeleton className="h-[280px] rounded-lg" />
              : <ReactECharts option={incomeChartOption} style={{ height: 280 }} opts={{ renderer: 'svg' }} />
            }
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="flex flex-row items-center justify-between">
            <CardTitle className="text-base">节点状态</CardTitle>
            <Badge variant="outline" className="font-normal">
              共 {nodeStats?.total ?? 0}
            </Badge>
          </CardHeader>
          <CardContent>
            {loadingNodeStats
              ? <Skeleton className="h-[280px] rounded-lg" />
              : <ReactECharts option={nodeChartOption} style={{ height: 280 }} opts={{ renderer: 'svg' }} />
            }
          </CardContent>
        </Card>
      </div>

      <div className="grid gap-4 lg:grid-cols-7">
        <Card className="lg:col-span-4">
          <CardHeader className="flex flex-row items-center justify-between">
            <CardTitle className="text-base">最近订单</CardTitle>
            <span className="text-xs text-muted-foreground">最近 5 笔</span>
          </CardHeader>
          <CardContent>
            {loadingOrders ? (
              <div className="space-y-3">{Array.from({ length: 5 }).map((_, i) => <Skeleton key={i} className="h-12" />)}</div>
            ) : (
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>订单号</TableHead>
                    <TableHead>金额</TableHead>
                    <TableHead>状态</TableHead>
                    <TableHead className="text-right">时间</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {recentOrders?.length === 0 ? (
                    <TableRow><TableCell colSpan={4} className="h-24 text-center text-muted-foreground">暂无订单</TableCell></TableRow>
                  ) : recentOrders?.map((o) => (
                    <TableRow key={o.id}>
                      <TableCell className="font-mono text-xs">{o.trade_no}</TableCell>
                      <TableCell className="font-medium">{formatCurrency(o.amount)}</TableCell>
                      <TableCell>
                        <Badge variant={orderStatusMap[o.status]?.variant} className="text-xs">
                          {orderStatusMap[o.status]?.label}
                        </Badge>
                      </TableCell>
                      <TableCell className="text-right text-xs text-muted-foreground">{formatDate(o.created_at, { month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit' })}</TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            )}
          </CardContent>
        </Card>
        <Card className="lg:col-span-3">
          <CardHeader className="flex flex-row items-center justify-between">
            <CardTitle className="text-base">最近用户</CardTitle>
            <span className="text-xs text-muted-foreground">最近 5 位</span>
          </CardHeader>
          <CardContent>
            {loadingUsers ? (
              <div className="space-y-3">{Array.from({ length: 5 }).map((_, i) => <Skeleton key={i} className="h-12" />)}</div>
            ) : (
              <div className="space-y-4">
                {recentUsers?.length === 0 ? (
                  <p className="py-8 text-center text-sm text-muted-foreground">暂无用户</p>
                ) : recentUsers?.map((u) => (
                  <div key={u.id} className="flex items-center gap-3">
                    <div className="flex h-9 w-9 shrink-0 items-center justify-center rounded-full bg-primary/10 text-sm font-medium text-primary">
                      {u.email?.[0]?.toUpperCase()}
                    </div>
                    <div className="flex-1 min-w-0">
                      <p className="text-sm font-medium truncate">{u.email}</p>
                      <p className="text-xs text-muted-foreground">{formatDate(u.created_at)}</p>
                    </div>
                    <Badge variant={u.status === 1 ? 'success' : 'destructive'} className="text-xs shrink-0">
                      {u.status === 1 ? '正常' : '禁用'}
                    </Badge>
                  </div>
                ))}
              </div>
            )}
          </CardContent>
        </Card>
      </div>

      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
        <Card>
          <CardContent className="flex items-center gap-3 pt-6">
            <div className="rounded-lg bg-amber-500/10 p-2"><Ticket className="h-5 w-5 text-amber-500" /></div>
            <div>
              <p className="text-2xl font-bold">{overview?.open_tickets ?? 0}</p>
              <p className="text-xs text-muted-foreground">待处理工单</p>
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="flex items-center gap-3 pt-6">
            <div className="rounded-lg bg-violet-500/10 p-2"><CreditCard className="h-5 w-5 text-violet-500" /></div>
            <div>
              <p className="text-2xl font-bold">{overview?.unused_redeems ?? 0}</p>
              <p className="text-xs text-muted-foreground">未使用卡密</p>
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="flex items-center gap-3 pt-6">
            <div className="rounded-lg bg-blue-500/10 p-2"><UserPlus className="h-5 w-5 text-blue-500" /></div>
            <div>
              <p className="text-2xl font-bold">{growthData?.reduce((s, d) => s + d.new_users, 0) ?? 0}</p>
              <p className="text-xs text-muted-foreground">本周新增用户</p>
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="flex items-center gap-3 pt-6">
            <div className="rounded-lg bg-emerald-500/10 p-2"><DollarSign className="h-5 w-5 text-emerald-500" /></div>
            <div>
              <p className="text-2xl font-bold">{formatCurrency(incomeData?.reduce((s, d) => s + d.amount, 0) ?? 0)}</p>
              <p className="text-xs text-muted-foreground">本周收入</p>
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  )
}
