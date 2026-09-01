import { useState, useMemo, useEffect } from 'react'
import { toast } from 'sonner'
import {
  useServers, useCreateServer, useUpdateServer, useDeleteServer,
  useResetServerToken, getInstallCommand, useLoadHistory,
} from '@/hooks/useServers'
import { useAdminNodes, useUpdateNode } from '@/hooks/useNodes'
import type { ServerMachine } from '@/types'
import { Card, CardContent } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import { Switch } from '@/components/ui/switch'
import { Badge } from '@/components/ui/badge'
import { Progress } from '@/components/ui/progress'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import {
  Dialog, DialogContent, DialogDescription, DialogFooter,
  DialogHeader, DialogTitle, DialogTrigger,
} from '@/components/ui/dialog'
import {
  AlertDialog, AlertDialogAction, AlertDialogCancel, AlertDialogContent,
  AlertDialogDescription, AlertDialogFooter, AlertDialogHeader, AlertDialogTitle,
} from '@/components/ui/alert-dialog'
import {
  DropdownMenu, DropdownMenuContent, DropdownMenuItem,
  DropdownMenuSeparator, DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { Skeleton } from '@/components/ui/skeleton'
import ReactECharts from 'echarts-for-react'
import {
  Plus, MoreHorizontal, Pencil, Trash2, Key, Copy, Server,
  Cpu, HardDrive, Clock, Search, Eye, RotateCcw, Terminal,
  Users, Activity, ArrowRight, X, Filter, BarChart3,
} from 'lucide-react'

// ─── 负载阈值 ──────────────────────────────────────────
const LOAD_THRESHOLDS = {
  cpu: { warn: 70, danger: 85 },
  mem: { warn: 75, danger: 90 },
  disk: { warn: 80, danger: 90 },
}

function getLoadColor(value: number, thresholds: { warn: number; danger: number }) {
  if (value >= thresholds.danger) return 'bg-red-500'
  if (value >= thresholds.warn) return 'bg-amber-500'
  return 'bg-emerald-500'
}

// ─── 状态计算 ──────────────────────────────────────────
function getServerStatus(machine: ServerMachine): { label: string; variant: 'default' | 'secondary' | 'destructive' | 'outline'; dotClass: string } {
  if (machine.is_active === false) return { label: '已禁用', variant: 'secondary', dotClass: 'bg-slate-400' }
  if (!machine.last_seen_at) return { label: '从未上报', variant: 'outline', dotClass: 'bg-slate-400' }
  // last_seen_at 是 Unix 时间戳（秒）
  const diff = Date.now() / 1000 - machine.last_seen_at
  if (diff < 300) return { label: '在线', variant: 'default', dotClass: 'bg-emerald-500' }
  return { label: '离线', variant: 'destructive', dotClass: 'bg-red-500' }
}

function timeSince(timestamp: number | string | null | undefined): string {
  if (!timestamp) return '从未'
  // 支持 Unix 时间戳（秒）和 ISO 字符串
  const timeMs = typeof timestamp === 'number' ? timestamp * 1000 : new Date(timestamp).getTime()
  const diff = Date.now() - timeMs
  const mins = Math.floor(diff / 60000)
  if (mins < 1) return '刚刚'
  if (mins < 60) return `${mins}分钟前`
  const hours = Math.floor(mins / 60)
  if (hours < 24) return `${hours}小时前`
  return `${Math.floor(hours / 24)}天前`
}

// ─── 统计卡片 ──────────────────────────────────────────
function StatCard({ title, value, icon: Icon, color }: {
  title: string; value: number | string; icon: React.ElementType; color: string
}) {
  return (
    <Card>
      <CardContent className="flex items-center gap-3 pt-6">
        <div className={`rounded-lg p-2 ${color}`}><Icon className="h-5 w-5" /></div>
        <div>
          <p className="text-2xl font-bold">{value}</p>
          <p className="text-xs text-muted-foreground">{title}</p>
        </div>
      </CardContent>
    </Card>
  )
}

// ─── 服务器表单 ─────────────────────────────────────────
function MachineFormDialog({ machine, onSave, children }: {
  machine?: ServerMachine
  onSave: (data: Partial<ServerMachine>) => void
  children: React.ReactNode
}) {
  const [open, setOpen] = useState(false)
  const [name, setName] = useState(machine?.name || '')
  const [notes, setNotes] = useState(machine?.remark || '')
  const [isActive, setIsActive] = useState(machine?.is_active !== false)
  const [errors, setErrors] = useState<Record<string, string>>({})

  const handleSave = () => {
    const newErrors: Record<string, string> = {}
    if (!name.trim()) newErrors.name = '请输入服务器名称'
    if (Object.keys(newErrors).length > 0) { setErrors(newErrors); return }
    setErrors({})
    onSave({ name: name.trim(), notes: notes.trim(), is_active: isActive })
    setOpen(false)
  }

  return (
    <Dialog open={open} onOpenChange={(v) => { setOpen(v); if (!v) { setName(machine?.name || ''); setNotes(machine?.remark || ''); setIsActive(machine?.is_active !== false); setErrors({}) } }}>
      <DialogTrigger asChild>{children}</DialogTrigger>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{machine ? '编辑服务器' : '新建服务器'}</DialogTitle>
          <DialogDescription>{machine ? '修改服务器名称、备注或启用状态' : '当你希望一台服务器承载多个节点时，再创建服务器记录'}</DialogDescription>
        </DialogHeader>
        <div className="space-y-4 py-4">
          <div className="space-y-2">
            <Label>服务器名称</Label>
            <Input value={name} onChange={(e) => { setName(e.target.value); setErrors((p) => ({ ...p, name: '' })) }} placeholder="例如 HK-01" />
            {errors.name && <p className="text-xs text-destructive">{errors.name}</p>}
          </div>
          <div className="space-y-2">
            <Label>备注</Label>
            <Textarea value={notes} onChange={(e) => setNotes(e.target.value)} placeholder="关于此服务器的可选备注" rows={2} />
          </div>
          <div className="flex items-center gap-2">
            <Switch checked={isActive} onCheckedChange={setIsActive} />
            <Label>启用服务器</Label>
            <p className="text-xs text-muted-foreground">禁用后 xboard-node 将不再使用此服务器</p>
          </div>
        </div>
        <DialogFooter>
          <Button variant="outline" onClick={() => setOpen(false)}>取消</Button>
          <Button onClick={handleSave}>{machine ? '更新' : '提交'}</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}

// ─── Token 管理弹窗 ────────────────────────────────────
function TokenDialog({ open, onOpenChange, token, machineName }: {
  open: boolean; onOpenChange: (v: boolean) => void; token: string; machineName: string
}) {
  const [showToken, setShowToken] = useState(true)
  const [countdown, setCountdown] = useState(180)

  useEffect(() => {
    if (open) {
      setShowToken(true)
      setCountdown(180)
      const timer = setInterval(() => {
        setCountdown((prev) => {
          if (prev <= 1) { clearInterval(timer); setShowToken(false); return 0 }
          return prev - 1
        })
      }, 1000)
      return () => clearInterval(timer)
    }
  }, [open])

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>服务器 Token — {machineName}</DialogTitle>
          <DialogDescription>此 Token 用于 xboard-node 向面板认证，请妥善保管</DialogDescription>
        </DialogHeader>
        <div className="space-y-3">
          <div className="flex items-center gap-2">
            <Input
              readOnly
              value={showToken ? token : '••••••••••••••••••••••••••••••••'}
              className="font-mono text-sm"
            />
            <Button variant="outline" size="icon" onClick={() => setShowToken(!showToken)}>
              <Eye className="h-4 w-4" />
            </Button>
            <Button variant="outline" size="icon" onClick={() => { navigator.clipboard.writeText(token); toast.success('已复制') }}>
              <Copy className="h-4 w-4" />
            </Button>
          </div>
          {showToken && countdown > 0 && (
            <p className="text-xs text-muted-foreground">{countdown} 秒后自动隐藏</p>
          )}
        </div>
      </DialogContent>
    </Dialog>
  )
}

// ─── 安装命令弹窗 ──────────────────────────────────────
function InstallCommandDialog({ open, onOpenChange, machineId, machineName }: {
  open: boolean; onOpenChange: (v: boolean) => void; machineId: number; machineName: string
}) {
  const [cmd, setCmd] = useState('')
  const [loading, setLoading] = useState(false)

  useEffect(() => {
    if (open && machineId) {
      setLoading(true)
      getInstallCommand(machineId).then(setCmd).catch(() => setCmd('获取失败')).finally(() => setLoading(false))
    }
  }, [open, machineId])

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-2xl">
        <DialogHeader>
          <DialogTitle>安装 xboard-node — {machineName}</DialogTitle>
          <DialogDescription>在目标服务器上执行此命令，即可用 machine mode 安装 xboard-node 并接入当前服务器记录</DialogDescription>
        </DialogHeader>
        {loading ? (
          <Skeleton className="h-20 w-full" />
        ) : (
          <div className="relative">
            <pre className="overflow-x-auto rounded-md bg-muted p-4 text-xs">{cmd}</pre>
            <Button variant="outline" size="sm" className="absolute top-2 right-2" onClick={() => { navigator.clipboard.writeText(cmd); toast.success('已复制') }}>
              <Copy className="mr-1 h-3 w-3" />复制
            </Button>
          </div>
        )}
        <p className="text-xs text-muted-foreground">需要 root 或 sudo 权限，且目标服务器需为支持 systemd 的 Linux</p>
      </DialogContent>
    </Dialog>
  )
}

// ─── 关联节点弹窗 ──────────────────────────────────────
function RelatedNodesDialog({ open, onOpenChange, machineId, machineName }: {
  open: boolean; onOpenChange: (v: boolean) => void; machineId: number; machineName: string
}) {
  const { data: allNodes } = useAdminNodes(1, 1000)
  const updateNode = useUpdateNode()

  const relatedNodes = useMemo(() => {
    if (!allNodes?.list) return []
    return allNodes.list.filter((n) => {
      try {
        const info = JSON.parse(n.server_info || '{}')
        return info.machine_id === machineId
      } catch { return false }
    })
  }, [allNodes, machineId])

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-3xl max-h-[80vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>关联节点 — {machineName}</DialogTitle>
          <DialogDescription>该服务器机器下部署的所有节点</DialogDescription>
        </DialogHeader>
        {relatedNodes.length === 0 ? (
          <div className="flex flex-col items-center justify-center py-10 text-muted-foreground">
            <Server className="mb-2 h-8 w-8" /><p className="text-sm">暂无关联节点</p>
          </div>
        ) : (
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>ID</TableHead>
                <TableHead>名称</TableHead>
                <TableHead>类型</TableHead>
                <TableHead>地址</TableHead>
                <TableHead>状态</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {relatedNodes.map((n) => (
                <TableRow key={n.id}>
                  <TableCell className="font-mono text-xs">{n.id}</TableCell>
                  <TableCell className="font-medium">{n.name}</TableCell>
                  <TableCell><Badge variant="outline" className="text-xs">{n.type}</Badge></TableCell>
                  <TableCell className="font-mono text-xs">{n.host}:{n.port}</TableCell>
                  <TableCell>
                    <span className={`inline-flex items-center gap-1 text-xs ${n.status === 1 ? 'text-emerald-600' : 'text-red-500'}`}>
                      <span className={`h-1.5 w-1.5 rounded-full ${n.status === 1 ? 'bg-emerald-500' : 'bg-red-500'}`} />
                      {n.status === 1 ? '运行中' : '未运行'}
                    </span>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        )}
      </DialogContent>
    </Dialog>
  )
}

// ─── 服务器详情弹窗（含负载趋势图）────────────────────
function ServerDetailDialog({ open, onOpenChange, machine }: {
  open: boolean; onOpenChange: (v: boolean) => void; machine: ServerMachine | null
}) {
  const { data: historyData, isLoading: historyLoading } = useLoadHistory(machine?.id ?? null, 1, 100)
  const { data: allNodes } = useAdminNodes(1, 1000)

  const relatedNodes = useMemo(() => {
    if (!allNodes?.list || !machine) return []
    return allNodes.list.filter((n) => {
      try {
        const info = JSON.parse(n.server_info || '{}')
        return info.machine_id === machine.id
      } catch { return false }
    })
  }, [allNodes, machine])

  const history = historyData?.list ?? []

  const loadChartOption = useMemo(() => {
    if (history.length === 0) return null
    const sorted = [...history].sort((a, b) => a.recorded_at - b.recorded_at)
    const times = sorted.map((h) => {
      const d = new Date(h.recorded_at * 1000)
      return `${d.getMonth() + 1}/${d.getDate()} ${d.getHours()}:${String(d.getMinutes()).padStart(2, '0')}`
    })
    return {
      tooltip: { trigger: 'axis' },
      legend: { data: ['CPU', '内存', '磁盘'], bottom: 0 },
      grid: { left: 40, right: 20, top: 10, bottom: 40 },
      xAxis: { type: 'category', data: times, axisLabel: { fontSize: 10 } },
      yAxis: { type: 'value', max: 100, axisLabel: { formatter: '{value}%', fontSize: 10 } },
      series: [
        { name: 'CPU', type: 'line', smooth: true, data: sorted.map((h) => h.cpu?.toFixed(1) ?? 0), itemStyle: { color: '#f59e0b' } },
        { name: '内存', type: 'line', smooth: true, data: sorted.map((h) => h.mem_total > 0 ? ((h.mem_used / h.mem_total) * 100).toFixed(1) : 0), itemStyle: { color: '#3b82f6' } },
        { name: '磁盘', type: 'line', smooth: true, data: sorted.map((h) => h.disk_total > 0 ? ((h.disk_used / h.disk_total) * 100).toFixed(1) : 0), itemStyle: { color: '#10b981' } },
      ],
    }
  }, [history])

  const netChartOption = useMemo(() => {
    if (history.length === 0) return null
    const sorted = [...history].sort((a, b) => a.recorded_at - b.recorded_at)
    const times = sorted.map((h) => {
      const d = new Date(h.recorded_at * 1000)
      return `${d.getMonth() + 1}/${d.getDate()} ${d.getHours()}:${String(d.getMinutes()).padStart(2, '0')}`
    })
    return {
      tooltip: { trigger: 'axis', formatter: (params: any[]) => params.map((p: any) => `${p.seriesName}: ${formatBytesShort(p.value)}/s`).join('<br/>') },
      legend: { data: ['入站速率', '出站速率'], bottom: 0 },
      grid: { left: 60, right: 20, top: 10, bottom: 40 },
      xAxis: { type: 'category', data: times, axisLabel: { fontSize: 10 } },
      yAxis: { type: 'value', axisLabel: { formatter: (v: number) => formatBytesShort(v), fontSize: 10 } },
      series: [
        { name: '入站速率', type: 'line', smooth: true, data: sorted.map((h) => h.net_in_speed ?? 0), itemStyle: { color: '#3b82f6' }, areaStyle: { opacity: 0.1 } },
        { name: '出站速率', type: 'line', smooth: true, data: sorted.map((h) => h.net_out_speed ?? 0), itemStyle: { color: '#f59e0b' }, areaStyle: { opacity: 0.1 } },
      ],
    }
  }, [history])

  if (!machine) return null

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-4xl max-h-[85vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <BarChart3 className="h-4 w-4" />服务器详情 — {machine.name}
            <span className="text-sm text-muted-foreground font-mono">SID:{machine.id}</span>
          </DialogTitle>
          <DialogDescription>{machine.remark || '服务器负载趋势与关联节点'}</DialogDescription>
        </DialogHeader>

        <Tabs defaultValue="chart">
          <TabsList>
            <TabsTrigger value="chart">负载趋势</TabsTrigger>
            <TabsTrigger value="nodes">关联节点 ({relatedNodes.length})</TabsTrigger>
          </TabsList>
          <TabsContent value="chart" className="space-y-4">
            {historyLoading ? (
              <Skeleton className="h-[300px] w-full" />
            ) : history.length === 0 ? (
              <div className="flex flex-col items-center justify-center py-16 text-muted-foreground">
                <BarChart3 className="mb-3 h-10 w-10" />
                <p className="text-sm">暂无负载历史数据</p>
                <p className="text-xs mt-1">xboard-node 上报后将自动显示趋势图</p>
              </div>
            ) : (
              <>
                <div>
                  <p className="text-sm font-medium mb-2">系统负载 (%)</p>
                  <ReactECharts option={loadChartOption} style={{ height: 250 }} opts={{ renderer: 'svg' }} />
                </div>
                <div>
                  <p className="text-sm font-medium mb-2">网络速率</p>
                  <ReactECharts option={netChartOption} style={{ height: 250 }} opts={{ renderer: 'svg' }} />
                </div>
              </>
            )}
          </TabsContent>
          <TabsContent value="nodes">
            {relatedNodes.length === 0 ? (
              <div className="flex flex-col items-center justify-center py-10 text-muted-foreground">
                <Server className="mb-2 h-8 w-8" /><p className="text-sm">暂无关联节点</p>
              </div>
            ) : (
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>ID</TableHead>
                    <TableHead>名称</TableHead>
                    <TableHead>类型</TableHead>
                    <TableHead>地址</TableHead>
                    <TableHead>状态</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {relatedNodes.map((n) => (
                    <TableRow key={n.id}>
                      <TableCell className="font-mono text-xs">{n.id}</TableCell>
                      <TableCell className="font-medium">{n.name}</TableCell>
                      <TableCell><Badge variant="outline" className="text-xs">{n.type}</Badge></TableCell>
                      <TableCell className="font-mono text-xs">{n.host}:{n.port}</TableCell>
                      <TableCell>
                        <span className={`inline-flex items-center gap-1 text-xs ${n.status === 1 ? 'text-emerald-600' : 'text-red-500'}`}>
                          <span className={`h-1.5 w-1.5 rounded-full ${n.status === 1 ? 'bg-emerald-500' : 'bg-red-500'}`} />
                          {n.status === 1 ? '运行中' : '未运行'}
                        </span>
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            )}
          </TabsContent>
        </Tabs>
      </DialogContent>
    </Dialog>
  )
}

function formatBytesShort(bytes: number): string {
  if (bytes < 1024) return `${bytes}B`
  if (bytes < 1048576) return `${(bytes / 1024).toFixed(1)}KB`
  if (bytes < 1073741824) return `${(bytes / 1048576).toFixed(1)}MB`
  return `${(bytes / 1073741824).toFixed(2)}GB`
}

// ─── 主页面 ────────────────────────────────────────────
export default function ServerMachinesPage() {
  const { data, isLoading } = useServers()
  const createServer = useCreateServer()
  const updateServer = useUpdateServer()
  const deleteServer = useDeleteServer()
  const resetToken = useResetServerToken()

  const [search, setSearch] = useState('')
  const [statusFilter, setStatusFilter] = useState('all')
  const [nodeFilter, setNodeFilter] = useState('all')
  const [deleteId, setDeleteId] = useState<number | null>(null)
  const [resetTokenId, setResetTokenId] = useState<number | null>(null)
  const [tokenMachine, setTokenMachine] = useState<ServerMachine | null>(null)
  const [installMachine, setInstallMachine] = useState<ServerMachine | null>(null)
  const [nodesMachine, setNodesMachine] = useState<ServerMachine | null>(null)
  const [detailMachine, setDetailMachine] = useState<ServerMachine | null>(null)

  const machines = data?.list || []

  const filteredMachines = useMemo(() => {
    let result = machines
    if (search.trim()) {
      const q = search.toLowerCase()
      result = result.filter((m) => m.name.toLowerCase().includes(q) || (m.remark || '').toLowerCase().includes(q) || String(m.id).includes(q))
    }
    if (statusFilter !== 'all') {
      result = result.filter((m) => {
        const s = getServerStatus(m)
        if (statusFilter === 'online') return s.label === '在线'
        if (statusFilter === 'offline') return s.label === '离线'
        if (statusFilter === 'inactive') return s.label === '已禁用'
        if (statusFilter === 'never') return s.label === '从未上报'
        return true
      })
    }
    if (nodeFilter !== 'all') {
      result = result.filter((m) => {
        if (nodeFilter === 'has_nodes') return (m.nodes_count ?? 0) > 0
        if (nodeFilter === 'idle') return (m.nodes_count ?? 0) === 0
        if (nodeFilter === 'high_load') return (m.cpu ?? 0) > LOAD_THRESHOLDS.cpu.warn || (m.memory ?? 0) > LOAD_THRESHOLDS.mem.warn || (m.disk ?? 0) > LOAD_THRESHOLDS.disk.warn
        return true
      })
    }
    return result
  }, [machines, search, statusFilter, nodeFilter])

  const stats = useMemo(() => {
    const total = machines.length
    let online = 0, offline = 0, highLoad = 0, totalNodes = 0
    machines.forEach((m) => {
      const s = getServerStatus(m)
      if (s.label === '在线') online++
      else if (s.label === '离线') offline++
      if ((m.cpu ?? 0) > LOAD_THRESHOLDS.cpu.warn) highLoad++
      totalNodes += m.nodes_count ?? 0
    })
    return { total, online, offline, highLoad, totalNodes }
  }, [machines])

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold tracking-tight">服务器管理</h1>
        <p className="text-muted-foreground">用于查看服务器健康、负载与承载节点，并从运维视角快捷发起节点操作</p>
      </div>

      {/* 概览统计 */}
      <div className="grid gap-4 grid-cols-2 lg:grid-cols-5">
        <StatCard title="服务器总数" value={stats.total} icon={Server} color="bg-primary/10 text-primary" />
        <StatCard title="在线服务器" value={stats.online} icon={Activity} color="bg-emerald-500/10 text-emerald-500" />
        <StatCard title="离线/失联" value={stats.offline} icon={Clock} color="bg-red-500/10 text-red-500" />
        <StatCard title="高负载" value={stats.highLoad} icon={Cpu} color="bg-amber-500/10 text-amber-500" />
        <StatCard title="托管节点" value={stats.totalNodes} icon={Users} color="bg-violet-500/10 text-violet-500" />
      </div>

      {/* 工具栏 */}
      <div className="flex items-center gap-3 flex-wrap">
        <div className="relative flex-1 min-w-[200px] max-w-sm">
          <Search className="absolute left-2.5 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
          <Input placeholder="搜索服务器名称、备注、SID..." value={search} onChange={(e) => setSearch(e.target.value)} className="pl-9" />
        </div>
        <Select value={statusFilter} onValueChange={setStatusFilter}>
          <SelectTrigger className="w-[120px]"><SelectValue placeholder="全部状态" /></SelectTrigger>
          <SelectContent>
            <SelectItem value="all">全部状态</SelectItem>
            <SelectItem value="online">在线</SelectItem>
            <SelectItem value="offline">离线</SelectItem>
            <SelectItem value="inactive">已禁用</SelectItem>
            <SelectItem value="never">从未上报</SelectItem>
          </SelectContent>
        </Select>
        <Select value={nodeFilter} onValueChange={setNodeFilter}>
          <SelectTrigger className="w-[120px]"><SelectValue placeholder="节点筛选" /></SelectTrigger>
          <SelectContent>
            <SelectItem value="all">全部</SelectItem>
            <SelectItem value="has_nodes">已承载节点</SelectItem>
            <SelectItem value="idle">空闲服务器</SelectItem>
            <SelectItem value="high_load">高负载</SelectItem>
          </SelectContent>
        </Select>
        <div className="ml-auto flex items-center gap-2">
          {(search || statusFilter !== 'all' || nodeFilter !== 'all') && (
            <Button variant="ghost" size="sm" className="h-8 gap-1 text-muted-foreground" onClick={() => { setSearch(''); setStatusFilter('all'); setNodeFilter('all') }}>
              <X className="h-3.5 w-3.5" />重置
            </Button>
          )}
          <span className="text-sm text-muted-foreground">在线: {stats.online}/{stats.total}</span>
          <MachineFormDialog onSave={(data) => createServer.mutate(data, { onSuccess: () => toast.success('创建成功') })}>
            <Button><Plus className="mr-2 h-4 w-4" />新建服务器</Button>
          </MachineFormDialog>
        </div>
      </div>

      {/* 表格 */}
      {isLoading ? (
        <div className="space-y-2">{Array.from({ length: 4 }).map((_, i) => <Skeleton key={i} className="h-14 w-full" />)}</div>
      ) : filteredMachines.length === 0 ? (
        <div className="flex flex-col items-center justify-center py-16 text-muted-foreground">
          <Server className="mb-3 h-10 w-10" />
          <p className="text-sm">{search || statusFilter !== 'all' ? '没有匹配的服务器' : '暂无服务器'}</p>
        </div>
      ) : (
        <div className="rounded-md border">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead className="w-14">ID</TableHead>
                <TableHead>名称</TableHead>
                <TableHead className="w-24">状态</TableHead>
                <TableHead className="w-[160px]">负载</TableHead>
                <TableHead className="w-20 text-center">节点数</TableHead>
                <TableHead className="w-32">最后心跳</TableHead>
                <TableHead className="w-20 text-right">操作</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {filteredMachines.map((m) => {
                const status = getServerStatus(m)
                return (
                  <TableRow key={m.id}>
                    <TableCell className="font-mono text-xs text-muted-foreground">{m.id}</TableCell>
                    <TableCell>
                      <div className="flex items-center gap-2">
                        <span className={`h-2 w-2 rounded-full ${status.dotClass}`} />
                        <div>
                          <div className="flex items-center gap-1.5">
                            <p className="font-medium">{m.name}</p>
                            <span className="text-[10px] text-muted-foreground font-mono">SID:{m.id}</span>
                            {(m.cpu ?? 0) > LOAD_THRESHOLDS.cpu.warn && (
                              <span className="inline-flex items-center rounded-md border border-amber-500/25 bg-amber-500/10 px-1.5 py-0 text-[10px] font-medium text-amber-600">高负载</span>
                            )}
                          </div>
                          {m.remark && <p className="text-xs text-muted-foreground truncate max-w-[200px]">{m.remark}</p>}
                        </div>
                      </div>
                    </TableCell>
                    <TableCell>
                      <Badge variant={status.variant} className="text-xs">{status.label}</Badge>
                    </TableCell>
                    <TableCell>
                      <div className="space-y-1">
                        <div className="flex items-center gap-2">
                          <Cpu className="h-3 w-3 text-muted-foreground" />
                          <Progress value={m.cpu ?? 0} className="h-1.5 flex-1" />
                          <span className="text-xs font-mono w-8 text-right">{(m.cpu ?? 0).toFixed(0)}%</span>
                        </div>
                        <div className="flex items-center gap-2">
                          <HardDrive className="h-3 w-3 text-muted-foreground" />
                          <Progress value={m.memory ?? 0} className="h-1.5 flex-1" />
                          <span className="text-xs font-mono w-8 text-right">{(m.memory ?? 0).toFixed(0)}%</span>
                        </div>
                        <div className="flex items-center gap-2">
                          <HardDrive className="h-3 w-3 text-muted-foreground" />
                          <Progress value={m.disk ?? 0} className="h-1.5 flex-1" />
                          <span className="text-xs font-mono w-8 text-right">{(m.disk ?? 0).toFixed(0)}%</span>
                        </div>
                      </div>
                    </TableCell>
                    <TableCell className="text-center">
                      <Button variant="ghost" size="sm" className="h-6 px-2 text-xs gap-1" onClick={() => setNodesMachine(m)}>
                        <Users className="h-3 w-3" />{m.nodes_count ?? 0}
                      </Button>
                    </TableCell>
                    <TableCell>
                      <span className="text-xs text-muted-foreground">{timeSince(m.last_seen_at)}</span>
                    </TableCell>
                    <TableCell className="text-right">
                      <DropdownMenu>
                        <DropdownMenuTrigger asChild>
                          <Button variant="ghost" size="icon" className="h-7 w-7"><MoreHorizontal className="h-4 w-4" /></Button>
                        </DropdownMenuTrigger>
                        <DropdownMenuContent align="end" className="w-44">
                          <DropdownMenuItem onClick={() => setDetailMachine(m)}>
                            <BarChart3 className="mr-2 h-4 w-4" />查看详情
                          </DropdownMenuItem>
                          <MachineFormDialog machine={m} onSave={(data) => updateServer.mutate({ id: m.id, ...data }, { onSuccess: () => toast.success('更新成功') })}>
                            <DropdownMenuItem onSelect={(e) => e.preventDefault()}>
                              <Pencil className="mr-2 h-4 w-4" />编辑
                            </DropdownMenuItem>
                          </MachineFormDialog>
                          <DropdownMenuItem onClick={() => setNodesMachine(m)}>
                            <Users className="mr-2 h-4 w-4" />查看关联节点
                          </DropdownMenuItem>
                          <DropdownMenuItem onClick={() => setInstallMachine(m)}>
                            <Terminal className="mr-2 h-4 w-4" />安装命令
                          </DropdownMenuItem>
                          <DropdownMenuItem onClick={() => setTokenMachine(m)}>
                            <Key className="mr-2 h-4 w-4" />查看 Token
                          </DropdownMenuItem>
                          <DropdownMenuItem onClick={() => setResetTokenId(m.id)}>
                            <RotateCcw className="mr-2 h-4 w-4" />重置 Token
                          </DropdownMenuItem>
                          <DropdownMenuSeparator />
                          <DropdownMenuItem className="text-destructive focus:text-destructive" onClick={() => setDeleteId(m.id)}>
                            <Trash2 className="mr-2 h-4 w-4" />删除
                          </DropdownMenuItem>
                        </DropdownMenuContent>
                      </DropdownMenu>
                    </TableCell>
                  </TableRow>
                )
              })}
            </TableBody>
          </Table>
        </div>
      )}

      {/* 删除确认 */}
      <AlertDialog open={deleteId !== null} onOpenChange={() => setDeleteId(null)}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>确认删除服务器？</AlertDialogTitle>
            <AlertDialogDescription>关联节点将自动解绑（不会被删除），此操作不可撤销</AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>取消</AlertDialogCancel>
            <AlertDialogAction onClick={() => { if (deleteId !== null) deleteServer.mutate(deleteId, { onSuccess: () => { toast.success('删除成功'); setDeleteId(null) } }) }}>删除</AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>

      {/* 重置 Token 确认 */}
      <AlertDialog open={resetTokenId !== null} onOpenChange={() => setResetTokenId(null)}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>确认重置 Token？</AlertDialogTitle>
            <AlertDialogDescription>旧 Token 将立即失效，xboard-node 需要重新配置新 Token</AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>取消</AlertDialogCancel>
            <AlertDialogAction onClick={() => { if (resetTokenId !== null) resetToken.mutate(resetTokenId, { onSuccess: () => { toast.success('Token 已重置'); setResetTokenId(null) } }) }}>确认重置</AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>

      {/* Token 弹窗 */}
      {tokenMachine && (
        <TokenDialog
          open={!!tokenMachine}
          onOpenChange={(v) => { if (!v) setTokenMachine(null) }}
          token={tokenMachine.token || ''}
          machineName={tokenMachine.name}
        />
      )}

      {/* 安装命令弹窗 */}
      {installMachine && (
        <InstallCommandDialog
          open={!!installMachine}
          onOpenChange={(v) => { if (!v) setInstallMachine(null) }}
          machineId={installMachine.id}
          machineName={installMachine.name}
        />
      )}

      {/* 关联节点弹窗 */}
      {nodesMachine && (
        <RelatedNodesDialog
          open={!!nodesMachine}
          onOpenChange={(v) => { if (!v) setNodesMachine(null) }}
          machineId={nodesMachine.id}
          machineName={nodesMachine.name}
        />
      )}

      {/* 服务器详情弹窗 */}
      {detailMachine && (
        <ServerDetailDialog
          open={!!detailMachine}
          onOpenChange={(v) => { if (!v) setDetailMachine(null) }}
          machine={detailMachine}
        />
      )}
    </div>
  )
}
