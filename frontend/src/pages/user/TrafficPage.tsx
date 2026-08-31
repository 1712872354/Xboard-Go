import { useState } from 'react'
import ReactECharts from 'echarts-for-react'
import { ArrowUp, ArrowDown, Activity } from 'lucide-react'
import { useTrafficStats, useTrafficHistory, useTrafficDaily } from '@/hooks/useTraffic'
import { formatBytes, formatDate } from '@/lib/utils'
import type { TrafficLog } from '@/types'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Skeleton } from '@/components/ui/skeleton'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'

export default function TrafficPage() {
  const [page, setPage] = useState(1)
  const { data: stats, isLoading: statsLoading } = useTrafficStats()
  const { data: daily, isLoading: dailyLoading } = useTrafficDaily(30)
  const { data: history, isLoading: historyLoading } = useTrafficHistory(page, 10)

  const totalPages = history ? Math.ceil(history.total / history.page_size) : 1

  const chartOption = daily
    ? {
        tooltip: {
          trigger: 'axis' as const,
          formatter: (params: any) => {
            const [up, down] = params
            return `${up.axisValue}<br/>上传: ${formatBytes(up.value)}<br/>下载: ${formatBytes(down.value)}`
          },
        },
        legend: { data: ['上传', '下载'], bottom: 0 },
        grid: { left: '3%', right: '4%', bottom: '12%', top: '5%', containLabel: true },
        xAxis: {
          type: 'category' as const,
          boundaryGap: false,
          data: daily.map((d) => d.date),
        },
        yAxis: {
          type: 'value' as const,
          axisLabel: {
            formatter: (v: number) => formatBytes(v),
          },
        },
        series: [
          {
            name: '上传',
            type: 'line',
            smooth: true,
            areaStyle: { opacity: 0.3 },
            data: daily.map((d) => d.upload),
          },
          {
            name: '下载',
            type: 'line',
            smooth: true,
            areaStyle: { opacity: 0.3 },
            data: daily.map((d) => d.download),
          },
        ],
      }
    : null

  if (statsLoading) {
    return (
      <div className="space-y-6">
        <Skeleton className="h-8 w-32" />
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
          {Array.from({ length: 4 }).map((_, i) => (
            <Skeleton key={i} className="h-28" />
          ))}
        </div>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-bold">流量统计</h1>

      {/* Stats Cards */}
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between pb-2">
            <CardTitle className="text-sm font-medium">今日上传</CardTitle>
            <ArrowUp className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{formatBytes(stats?.today_upload || 0)}</div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between pb-2">
            <CardTitle className="text-sm font-medium">今日下载</CardTitle>
            <ArrowDown className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{formatBytes(stats?.today_download || 0)}</div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between pb-2">
            <CardTitle className="text-sm font-medium">本月流量</CardTitle>
            <Activity className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {formatBytes((stats?.month_upload || 0) + (stats?.month_download || 0))}
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between pb-2">
            <CardTitle className="text-sm font-medium">已用流量</CardTitle>
            <Activity className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {formatBytes(stats?.total_used || 0)}
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between pb-2">
            <CardTitle className="text-sm font-medium">流量限制</CardTitle>
            <Activity className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {formatBytes(stats?.traffic_limit || 0)}
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between pb-2">
            <CardTitle className="text-sm font-medium">剩余流量</CardTitle>
            <Activity className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {formatBytes(stats?.remaining || 0)}
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Traffic Chart */}
      <Card>
        <CardHeader>
          <CardTitle>每日流量趋势</CardTitle>
          <CardDescription>最近30天流量使用情况</CardDescription>
        </CardHeader>
        <CardContent>
          {dailyLoading ? (
            <Skeleton className="h-[300px] w-full" />
          ) : chartOption ? (
            <ReactECharts option={chartOption} style={{ height: 300 }} />
          ) : (
            <p className="text-center text-muted-foreground py-8">暂无数据</p>
          )}
        </CardContent>
      </Card>

      {/* Traffic Log Table */}
      <Card>
        <CardHeader>
          <CardTitle>流量日志</CardTitle>
          <CardDescription>详细的流量使用记录</CardDescription>
        </CardHeader>
        <CardContent>
          {historyLoading ? (
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
                      <TableHead>节点</TableHead>
                      <TableHead>上传</TableHead>
                      <TableHead>下载</TableHead>
                      <TableHead>时间</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {history?.list?.length ? (
                      history.list.map((log: TrafficLog) => (
                        <TableRow key={log.id}>
                          <TableCell>{log.node_name || `节点 #${log.node_id}`}</TableCell>
                          <TableCell>{formatBytes(log.upload)}</TableCell>
                          <TableCell>{formatBytes(log.download)}</TableCell>
                          <TableCell className="text-sm text-muted-foreground">
                            {formatDate(log.recorded_at)}
                          </TableCell>
                        </TableRow>
                      ))
                    ) : (
                      <TableRow>
                        <TableCell colSpan={4} className="text-center py-8 text-muted-foreground">
                          暂无流量记录
                        </TableCell>
                      </TableRow>
                    )}
                  </TableBody>
                </Table>
              </div>

              {totalPages > 1 && (
                <div className="flex items-center justify-between mt-4">
                  <p className="text-sm text-muted-foreground">
                    共 {history?.total || 0} 条记录
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
    </div>
  )
}
