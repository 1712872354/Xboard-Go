import { useState, useMemo } from 'react'
import { useQuery } from '@tanstack/react-query'
import api from '@/lib/api'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Skeleton } from '@/components/ui/skeleton'
import { formatDate } from '@/lib/utils'
import { RefreshCw, Download, X, BarChart3, Users, Clock } from 'lucide-react'

interface TrafficResetLog {
  id: number
  user_id: number
  user_email: string
  reset_type: string
  trigger_source: string
  old_upload: number
  old_download: number
  old_total: number
  new_total: number
  reset_time: string
  created_at: string
}

const resetTypeMap: Record<string, string> = {
  first_day_month: '每月1号',
  monthly: '按月重置',
  first_day_year: '每年1月1号',
  yearly: '按年重置',
  manual: '手动重置',
}

const resetTypeOptions = [
  { value: 'all', label: '全部类型' },
  { value: 'monthly_first', label: '每月1号' },
  { value: 'monthly', label: '按月重置' },
  { value: 'yearly_first', label: '每年1月1号' },
  { value: 'yearly', label: '按年重置' },
  { value: 'manual', label: '手动重置' },
]

const sourceMap: Record<string, { label: string; variant: 'default' | 'secondary' | 'success' | 'destructive' }> = {
  auto: { label: '自动', variant: 'default' },
  manual: { label: '手动', variant: 'secondary' },
  cron: { label: '定时', variant: 'success' },
}

const triggerSourceOptions = [
  { value: 'all', label: '全部来源' },
  { value: 'auto', label: '自动触发' },
  { value: 'manual', label: '手动触发' },
  { value: 'cron', label: '定时任务' },
]

function formatTraffic(bytes: number): string {
  if (bytes >= 1073741824) return `${(bytes / 1073741824).toFixed(2)} GB`
  if (bytes >= 1048576) return `${(bytes / 1048576).toFixed(2)} MB`
  if (bytes >= 1024) return `${(bytes / 1024).toFixed(2)} KB`
  return `${bytes} B`
}

export default function TrafficResetLogsPage() {
  const [userEmail, setUserEmail] = useState('')
  const [resetType, setResetType] = useState('all')
  const [triggerSource, setTriggerSource] = useState('all')
  const [startDate, setStartDate] = useState('')
  const [endDate, setEndDate] = useState('')
  const [page, setPage] = useState(1)

  const hasFilters = userEmail || resetType !== 'all' || triggerSource !== 'all' || startDate || endDate

  const stats = useMemo(() => {
    const list = data?.list ?? []
    if (list.length === 0) return null
    const uniqueUsers = new Set(list.map((l) => l.user_id)).size
    const latestTime = list.reduce((max, l) => l.reset_time > max ? l.reset_time : max, list[0].reset_time)
    return { total: data?.total ?? 0, uniqueUsers, latestTime }
  }, [data])

  const resetFilters = () => {
    setUserEmail('')
    setResetType('all')
    setTriggerSource('all')
    setStartDate('')
    setEndDate('')
    setPage(1)
  }

  const { data, isLoading, refetch } = useQuery({
    queryKey: ['admin', 'traffic-reset-logs', page, userEmail, resetType, triggerSource, startDate, endDate],
    queryFn: async () => {
      const params: Record<string, string> = { page: page.toString(), per_page: '20' }
      if (userEmail) params.user_email = userEmail
      if (resetType !== 'all') params.reset_type = resetType
      if (triggerSource !== 'all') params.trigger_source = triggerSource
      if (startDate) params.start_date = startDate
      if (endDate) params.end_date = endDate
      return await api.get('/admin/traffic/logs', { params }) as unknown as { list: TrafficResetLog[]; total: number }
    },
  })

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold">流量重置日志</h1>
          <p className="text-muted-foreground">查看用户流量重置历史记录</p>
        </div>
        <div className="flex gap-2">
          <Button variant="outline" size="sm" onClick={() => {}}>
            <Download className="mr-2 h-4 w-4" />
            导出
          </Button>
          <Button variant="outline" size="sm" onClick={() => refetch()}>
            <RefreshCw className="mr-2 h-4 w-4" />
            刷新
          </Button>
        </div>
      </div>

      {stats && (
        <div className="grid gap-4 sm:grid-cols-3">
          <Card>
            <CardContent className="pt-6">
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-sm text-muted-foreground">总重置次数</p>
                  <p className="text-2xl font-bold mt-1">{stats.total}</p>
                </div>
                <BarChart3 className="h-8 w-8 text-muted-foreground" />
              </div>
            </CardContent>
          </Card>
          <Card>
            <CardContent className="pt-6">
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-sm text-muted-foreground">涉及用户数</p>
                  <p className="text-2xl font-bold mt-1">{stats.uniqueUsers}</p>
                </div>
                <Users className="h-8 w-8 text-muted-foreground" />
              </div>
            </CardContent>
          </Card>
          <Card>
            <CardContent className="pt-6">
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-sm text-muted-foreground">最近重置</p>
                  <p className="text-sm font-medium mt-1">{formatDate(stats.latestTime)}</p>
                </div>
                <Clock className="h-8 w-8 text-muted-foreground" />
              </div>
            </CardContent>
          </Card>
        </div>
      )}

      <Card>
        <CardHeader className="space-y-4">
          <div className="flex items-center justify-between">
            <CardTitle>日志列表</CardTitle>
          </div>
          <div className="flex flex-wrap items-center gap-3">
            <Input
              placeholder="搜索用户邮箱..."
              value={userEmail}
              onChange={(e) => { setUserEmail(e.target.value); setPage(1) }}
              className="w-48"
            />
            <Select value={resetType} onValueChange={(v) => { setResetType(v); setPage(1) }}>
              <SelectTrigger className="w-32">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                {resetTypeOptions.map((o) => (
                  <SelectItem key={o.value} value={o.value}>{o.label}</SelectItem>
                ))}
              </SelectContent>
            </Select>
            <Select value={triggerSource} onValueChange={(v) => { setTriggerSource(v); setPage(1) }}>
              <SelectTrigger className="w-32">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                {triggerSourceOptions.map((o) => (
                  <SelectItem key={o.value} value={o.value}>{o.label}</SelectItem>
                ))}
              </SelectContent>
            </Select>
            <Input
              type="date"
              value={startDate}
              onChange={(e) => { setStartDate(e.target.value); setPage(1) }}
              className="w-40"
            />
            <span className="text-sm text-muted-foreground">至</span>
            <Input
              type="date"
              value={endDate}
              onChange={(e) => { setEndDate(e.target.value); setPage(1) }}
              className="w-40"
            />
            {hasFilters && (
              <Button variant="ghost" size="sm" onClick={resetFilters}>
                <X className="mr-1 h-3.5 w-3.5" />
                重置
              </Button>
            )}
          </div>
        </CardHeader>
        <CardContent>
          {isLoading ? (
            <div className="space-y-3">{Array.from({ length: 10 }).map((_, i) => <Skeleton key={i} className="h-10" />)}</div>
          ) : (
            <>
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>ID</TableHead>
                    <TableHead>用户</TableHead>
                    <TableHead>重置类型</TableHead>
                    <TableHead>触发方式</TableHead>
                    <TableHead>重置前流量</TableHead>
                    <TableHead>重置后流量</TableHead>
                    <TableHead>重置时间</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {data?.list?.length === 0 ? (
                    <TableRow><TableCell colSpan={7} className="h-24 text-center text-muted-foreground">暂无记录</TableCell></TableRow>
                  ) : data?.list?.map((log) => (
                    <TableRow key={log.id}>
                      <TableCell className="font-mono text-xs">{log.id}</TableCell>
                      <TableCell className="text-sm">{log.user_email || `ID: ${log.user_id}`}</TableCell>
                      <TableCell><Badge variant="outline">{resetTypeMap[log.reset_type] ?? log.reset_type}</Badge></TableCell>
                      <TableCell>
                        <Badge variant={sourceMap[log.trigger_source]?.variant ?? 'secondary'}>
                          {sourceMap[log.trigger_source]?.label ?? log.trigger_source}
                        </Badge>
                      </TableCell>
                      <TableCell className="font-mono text-xs">{formatTraffic(log.old_total)}</TableCell>
                      <TableCell className="font-mono text-xs">{formatTraffic(log.new_total)}</TableCell>
                      <TableCell className="text-xs text-muted-foreground">{formatDate(log.reset_time)}</TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>

              {data && data.total > 20 && (
                <div className="flex items-center justify-between mt-4">
                  <p className="text-sm text-muted-foreground">共 {data.total} 条记录</p>
                  <div className="flex gap-2">
                    <Button variant="outline" size="sm" disabled={page <= 1} onClick={() => setPage(page - 1)}>上一页</Button>
                    <span className="flex items-center px-3 text-sm">第 {page} 页</span>
                    <Button variant="outline" size="sm" disabled={data.total <= page * 20} onClick={() => setPage(page + 1)}>下一页</Button>
                  </div>
                </div>
              )}
            </>
          )}
        </CardContent>
      </Card>
    </div>
  )
}
