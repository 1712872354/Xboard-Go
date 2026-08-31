import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'
import api from '@/lib/api'
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import { Switch } from '@/components/ui/switch'
import { Skeleton } from '@/components/ui/skeleton'
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription, DialogFooter } from '@/components/ui/dialog'
import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/tooltip'
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuSeparator, DropdownMenuTrigger } from '@/components/ui/dropdown-menu'
import { formatDate } from '@/lib/utils'
import {
  Plus, Pencil, Trash2, RefreshCw, Copy, Terminal, Key, Eye, EyeOff,
  MoreHorizontal, HardDrive, Server, Wifi, WifiOff, Activity, Cpu, MemoryStick, Disc,
} from 'lucide-react'

interface ServerMachine {
  id: number
  name: string
  notes: string
  address: string
  port: number
  token: string
  status: number
  cpu: number
  memory: number
  disk: number
  uptime: number
  node_count: number
  last_check_at: string | null
  created_at: string
}

// ─── Overview Cards ────────────────────────────────────────
function OverviewCards({ machines }: { machines: ServerMachine[] }) {
  const online = machines.filter((m) => m.status === 1).length
  const offline = machines.filter((m) => m.status === 0).length
  const highLoad = machines.filter((m) => m.cpu > 80 || m.memory > 80 || m.disk > 90).length
  const totalNodes = machines.reduce((sum, m) => sum + (m.node_count ?? 0), 0)

  const cards = [
    { title: '服务器总数', value: machines.length, hint: `共承载 ${totalNodes} 个节点`, icon: HardDrive, color: 'bg-primary/10 text-primary' },
    { title: '在线服务器', value: online, hint: '最近 5 分钟内正常心跳', icon: Wifi, color: 'bg-emerald-500/10 text-emerald-500' },
    { title: '离线/失联', value: offline, hint: '需要检查心跳或节点代理', icon: WifiOff, color: 'bg-red-500/10 text-red-500' },
    { title: '高负载', value: highLoad, hint: 'CPU、内存或磁盘接近阈值', icon: Activity, color: 'bg-amber-500/10 text-amber-500' },
  ]

  return (
    <div className="grid gap-4 grid-cols-2 lg:grid-cols-4">
      {cards.map((c) => (
        <Card key={c.title}>
          <CardContent className="flex items-center gap-3 pt-6">
            <div className={`rounded-lg p-2 ${c.color}`}><c.icon className="h-5 w-5" /></div>
            <div>
              <p className="text-2xl font-bold">{c.value}</p>
              <p className="text-xs text-muted-foreground">{c.title}</p>
            </div>
          </CardContent>
        </Card>
      ))}
    </div>
  )
}

// ─── Status Dot ────────────────────────────────────────────
function StatusDot({ status }: { status: number }) {
  const map: Record<number, { color: string; label: string }> = {
    0: { color: 'bg-red-500', label: '离线' },
    1: { color: 'bg-emerald-500', label: '在线' },
  }
  const s = map[status] ?? map[0]
  return (
    <span className="inline-flex items-center gap-1.5 text-xs">
      <span className={`h-2 w-2 rounded-full ${s.color}`} />
      {s.label}
    </span>
  )
}

// ─── Load Bar ──────────────────────────────────────────────
function LoadBar({ value, label, icon: Icon }: { value: number; label: string; icon: React.ElementType }) {
  const color = value > 90 ? 'bg-red-500' : value > 70 ? 'bg-amber-500' : 'bg-emerald-500'
  return (
    <div className="flex items-center gap-1.5">
      <Icon className="h-3 w-3 text-muted-foreground" />
      <div className="w-12 h-1.5 rounded-full bg-muted overflow-hidden">
        <div className={`h-full rounded-full ${color}`} style={{ width: `${Math.min(value, 100)}%` }} />
      </div>
      <span className="font-mono text-[11px] text-muted-foreground w-8">{value.toFixed(0)}%</span>
    </div>
  )
}

// ─── Time Since ────────────────────────────────────────────
function timeSince(dateStr: string | null): { text: string; stale: boolean } {
  if (!dateStr) return { text: '从未上报', stale: true }
  const diff = Date.now() - new Date(dateStr).getTime()
  const mins = Math.floor(diff / 60000)
  if (mins < 1) return { text: '刚刚', stale: false }
  if (mins < 60) return { text: `${mins}分钟前`, stale: mins > 5 }
  const hours = Math.floor(mins / 60)
  if (hours < 24) return { text: `${hours}小时前`, stale: true }
  return { text: `${Math.floor(hours / 24)}天前`, stale: true }
}

// ─── Create/Edit Form ──────────────────────────────────────
function MachineForm({ machine, onClose, onSave }: {
  machine?: ServerMachine; onClose: () => void; onSave: (data: { name: string; notes: string; status: number }) => void
}) {
  const [name, setName] = useState(machine?.name ?? '')
  const [notes, setNotes] = useState(machine?.notes ?? '')
  const [status, setStatus] = useState(machine?.status ?? 1)
  const [saving, setSaving] = useState(false)

  const handleSubmit = async () => {
    if (!name.trim()) { toast.error('请输入服务器名称'); return }
    setSaving(true)
    await onSave({ name, notes, status })
    setSaving(false)
  }

  return (
    <div className="space-y-4">
      <div className="space-y-2">
        <Label>服务器名称</Label>
        <Input value={name} onChange={(e) => setName(e.target.value)} placeholder="例如 HK-01" />
      </div>
      <div className="space-y-2">
        <Label>备注</Label>
        <Textarea value={notes} onChange={(e) => setNotes(e.target.value)} placeholder="关于此服务器的可选备注" rows={3} />
      </div>
      <div className="flex items-center justify-between rounded-lg border p-4">
        <div className="space-y-0.5">
          <Label>启用服务器</Label>
          <p className="text-sm text-muted-foreground">禁用后 xboard-node 将不再使用此服务器</p>
        </div>
        <Switch checked={status === 1} onCheckedChange={(v) => setStatus(v ? 1 : 0)} />
      </div>
      <DialogFooter>
        <Button variant="outline" onClick={onClose}>取消</Button>
        <Button onClick={handleSubmit} disabled={saving}>{saving ? '保存中...' : (machine ? '更新' : '创建')}</Button>
      </DialogFooter>
    </div>
  )
}

// ─── Token Dialog ──────────────────────────────────────────
function TokenDialog({ open, onOpenChange, machineId, machineName }: {
  open: boolean; onOpenChange: (v: boolean) => void; machineId: number; machineName: string
}) {
  const [token, setToken] = useState<string | null>(null)
  const [showToken, setShowToken] = useState(false)
  const [loading, setLoading] = useState(false)

  const fetchToken = async () => {
    setLoading(true)
    try {
      const res: any = await api.get(`/admin/server-machines/${machineId}`)
      setToken(res.token || res.data?.token || '')
    } catch { toast.error('获取令牌失败') }
    setLoading(false)
  }

  const resetToken = async () => {
    if (!confirm('确认重置 Token？旧 Token 将立即失效。')) return
    try {
      const res: any = await api.post(`/admin/server-machines/${machineId}/reset-token`)
      setToken(res.token || res.data?.token || '')
      toast.success('Token 已重置')
    } catch { toast.error('重置失败') }
  }

  const copyToken = () => {
    if (token) { navigator.clipboard.writeText(token); toast.success('已复制') }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>服务器 Token — {machineName}</DialogTitle>
          <DialogDescription>此 Token 用于 xboard-node 向面板认证，请妥善保管。</DialogDescription>
        </DialogHeader>
        <div className="space-y-4">
          {token === null ? (
            <Button onClick={fetchToken} disabled={loading} className="w-full">
              <Key className="mr-2 h-4 w-4" />{loading ? '加载中...' : '查看 Token'}
            </Button>
          ) : (
            <div className="space-y-2">
              <div className="flex items-center gap-2">
                <Input
                  type={showToken ? 'text' : 'password'}
                  value={token}
                  readOnly
                  className="font-mono text-xs"
                />
                <Button variant="outline" size="icon" onClick={() => setShowToken(!showToken)}>
                  {showToken ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
                </Button>
                <Button variant="outline" size="icon" onClick={copyToken}>
                  <Copy className="h-4 w-4" />
                </Button>
              </div>
              <Button variant="destructive" size="sm" onClick={resetToken}>
                <RefreshCw className="mr-2 h-4 w-4" />重置 Token
              </Button>
            </div>
          )}
        </div>
      </DialogContent>
    </Dialog>
  )
}

// ─── Install Command Dialog ────────────────────────────────
function InstallCommandDialog({ open, onOpenChange, machineId, machineName }: {
  open: boolean; onOpenChange: (v: boolean) => void; machineId: number; machineName: string
}) {
  const [command, setCommand] = useState<string | null>(null)
  const [loading, setLoading] = useState(false)

  const fetchCommand = async () => {
    setLoading(true)
    try {
      const res: any = await api.get(`/admin/server-machines/${machineId}/install-command`)
      setCommand(res.command || res.data?.command || '')
    } catch { toast.error('获取安装命令失败') }
    setLoading(false)
  }

  const copyCommand = () => {
    if (command) { navigator.clipboard.writeText(command); toast.success('已复制') }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-lg">
        <DialogHeader>
          <DialogTitle>安装 xboard-node — {machineName}</DialogTitle>
          <DialogDescription>在目标服务器上执行此命令，即可安装 xboard-node 并接入当前服务器记录。</DialogDescription>
        </DialogHeader>
        <div className="space-y-4">
          {command === null ? (
            <Button onClick={fetchCommand} disabled={loading} className="w-full">
              <Terminal className="mr-2 h-4 w-4" />{loading ? '生成中...' : '生成安装命令'}
            </Button>
          ) : (
            <div className="space-y-2">
              <div className="relative rounded-md border bg-muted p-3">
                <pre className="text-xs font-mono whitespace-pre-wrap break-all">{command}</pre>
              </div>
              <p className="text-xs text-muted-foreground">需要 root 或 sudo 权限，且目标服务器需为支持 systemd 的 Linux。</p>
              <Button onClick={copyCommand} className="w-full">
                <Copy className="mr-2 h-4 w-4" />复制安装命令
              </Button>
            </div>
          )}
        </div>
      </DialogContent>
    </Dialog>
  )
}

// ─── Detail Dialog ─────────────────────────────────────────
function MachineDetailDialog({ open, onOpenChange, machine }: {
  open: boolean; onOpenChange: (v: boolean) => void; machine: ServerMachine | null
}) {
  if (!machine) return null

  const { data: detail } = useQuery({
    queryKey: ['admin', 'server-machine', machine.id],
    queryFn: async () => await api.get(`/admin/server-machines/${machine.id}`) as unknown as ServerMachine & { nodes: any[] },
    enabled: open,
  })

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-2xl max-h-[80vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <HardDrive className="h-4 w-4" />{machine.name}
          </DialogTitle>
          <DialogDescription>
            <StatusDot status={machine.status} />
            {machine.notes && <span className="ml-2 text-muted-foreground">· {machine.notes}</span>}
          </DialogDescription>
        </DialogHeader>
        <div className="space-y-6">
          {/* 负载信息 */}
          <div className="space-y-3">
            <h4 className="text-sm font-medium">负载状态</h4>
            <div className="grid grid-cols-3 gap-4">
              <div className="space-y-1.5">
                <div className="flex items-center gap-1.5 text-xs text-muted-foreground"><Cpu className="h-3 w-3" />CPU</div>
                <div className="flex items-center gap-2">
                  <div className="flex-1 h-2 rounded-full bg-muted overflow-hidden">
                    <div className={`h-full rounded-full ${machine.cpu > 80 ? 'bg-red-500' : machine.cpu > 50 ? 'bg-amber-500' : 'bg-emerald-500'}`} style={{ width: `${machine.cpu}%` }} />
                  </div>
                  <span className="font-mono text-sm">{machine.cpu.toFixed(1)}%</span>
                </div>
              </div>
              <div className="space-y-1.5">
                <div className="flex items-center gap-1.5 text-xs text-muted-foreground"><MemoryStick className="h-3 w-3" />内存</div>
                <div className="flex items-center gap-2">
                  <div className="flex-1 h-2 rounded-full bg-muted overflow-hidden">
                    <div className={`h-full rounded-full ${machine.memory > 80 ? 'bg-red-500' : machine.memory > 50 ? 'bg-amber-500' : 'bg-emerald-500'}`} style={{ width: `${machine.memory}%` }} />
                  </div>
                  <span className="font-mono text-sm">{machine.memory.toFixed(1)}%</span>
                </div>
              </div>
              <div className="space-y-1.5">
                <div className="flex items-center gap-1.5 text-xs text-muted-foreground"><Disc className="h-3 w-3" />磁盘</div>
                <div className="flex items-center gap-2">
                  <div className="flex-1 h-2 rounded-full bg-muted overflow-hidden">
                    <div className={`h-full rounded-full ${machine.disk > 90 ? 'bg-red-500' : machine.disk > 70 ? 'bg-amber-500' : 'bg-emerald-500'}`} style={{ width: `${machine.disk}%` }} />
                  </div>
                  <span className="font-mono text-sm">{machine.disk.toFixed(1)}%</span>
                </div>
              </div>
            </div>
          </div>

          {/* 关联节点 */}
          <div className="space-y-3">
            <h4 className="text-sm font-medium">关联节点</h4>
            {detail?.nodes && detail.nodes.length > 0 ? (
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead className="text-[11px]">ID</TableHead>
                    <TableHead className="text-[11px]">名称</TableHead>
                    <TableHead className="text-[11px]">类型</TableHead>
                    <TableHead className="text-[11px]">地址</TableHead>
                    <TableHead className="text-[11px]">状态</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {detail.nodes.map((n: any) => (
                    <TableRow key={n.id}>
                      <TableCell className="font-mono text-xs">{n.id}</TableCell>
                      <TableCell className="text-sm">{n.name}</TableCell>
                      <TableCell><Badge variant="outline" className="text-xs">{n.type}</Badge></TableCell>
                      <TableCell className="font-mono text-xs">{n.address}:{n.port}</TableCell>
                      <TableCell><StatusDot status={n.status} /></TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            ) : (
              <p className="text-sm text-muted-foreground">暂无绑定节点。</p>
            )}
          </div>

          {/* 服务器信息 */}
          <div className="space-y-3">
            <h4 className="text-sm font-medium">服务器信息</h4>
            <div className="grid grid-cols-2 gap-2 text-xs">
              <div><span className="text-muted-foreground">创建时间:</span> {formatDate(machine.created_at)}</div>
              <div><span className="text-muted-foreground">最后心跳:</span> {timeSince(machine.last_check_at).text}</div>
            </div>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  )
}

// ─── Main Page ─────────────────────────────────────────────
export default function ServerMachinesPage() {
  const qc = useQueryClient()
  const [dialogOpen, setDialogOpen] = useState(false)
  const [editingMachine, setEditingMachine] = useState<ServerMachine | undefined>()
  const [tokenMachine, setTokenMachine] = useState<ServerMachine | null>(null)
  const [installMachine, setInstallMachine] = useState<ServerMachine | null>(null)
  const [detailMachine, setDetailMachine] = useState<ServerMachine | null>(null)
  const [searchQuery, setSearchQuery] = useState('')

  const { data: machinesData, isLoading, refetch } = useQuery({
    queryKey: ['admin', 'server-machines'],
    queryFn: async () => await api.get('/admin/server-machines') as unknown as { list: ServerMachine[]; total: number },
  })

  const machines = machinesData?.list ?? []

  const filteredMachines = searchQuery.trim()
    ? machines.filter((m) => m.name.toLowerCase().includes(searchQuery.toLowerCase()) || (m.notes ?? '').toLowerCase().includes(searchQuery.toLowerCase()))
    : machines

  const createMutation = useMutation({
    mutationFn: async (data: Partial<ServerMachine>) => await api.post('/admin/server-machines', data),
    onSuccess: () => { qc.invalidateQueries({ queryKey: ['admin', 'server-machines'] }); setDialogOpen(false); toast.success('服务器创建成功') },
    onError: () => toast.error('创建失败'),
  })

  const updateMutation = useMutation({
    mutationFn: async ({ id, ...data }: Partial<ServerMachine> & { id: number }) => await api.put(`/admin/server-machines/${id}`, data),
    onSuccess: () => { qc.invalidateQueries({ queryKey: ['admin', 'server-machines'] }); setDialogOpen(false); toast.success('服务器更新成功') },
    onError: () => toast.error('更新失败'),
  })

  const deleteMutation = useMutation({
    mutationFn: async (id: number) => await api.delete(`/admin/server-machines/${id}`),
    onSuccess: () => { qc.invalidateQueries({ queryKey: ['admin', 'server-machines'] }); toast.success('服务器删除成功') },
    onError: () => toast.error('删除失败'),
  })

  const handleSave = async (formData: { name: string; notes: string; status: number }) => {
    if (editingMachine) {
      await updateMutation.mutateAsync({ id: editingMachine.id, ...formData })
    } else {
      await createMutation.mutateAsync(formData)
    }
  }

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold">服务器管理</h1>
        <p className="text-muted-foreground">用于查看服务器健康、负载与承载节点，并从运维视角快捷发起节点操作。</p>
      </div>

      <OverviewCards machines={machines} />

      <Card>
        <CardHeader className="flex flex-row items-center justify-between">
          <div>
            <CardTitle>服务器列表</CardTitle>
            <CardDescription>适合集中查看服务器在线情况、承载节点数量与资源压力。</CardDescription>
          </div>
          <div className="flex items-center gap-2">
            <div className="relative">
              <Input
                placeholder="搜索服务器名称、备注..."
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                className="w-48"
              />
            </div>
            <Button variant="outline" size="sm" onClick={() => refetch()}>
              <RefreshCw className="mr-2 h-4 w-4" />刷新
            </Button>
            <Button size="sm" onClick={() => { setEditingMachine(undefined); setDialogOpen(true) }}>
              <Plus className="mr-2 h-4 w-4" />添加服务器
            </Button>
          </div>
        </CardHeader>
        <CardContent>
          {isLoading ? (
            <div className="space-y-3">{Array.from({ length: 4 }).map((_, i) => <Skeleton key={i} className="h-14" />)}</div>
          ) : filteredMachines.length === 0 ? (
            <div className="flex flex-col items-center justify-center py-16 text-muted-foreground">
              <Server className="mb-3 h-10 w-10" />
              <p className="text-sm">{searchQuery ? '没有匹配的服务器' : '暂无服务器'}</p>
            </div>
          ) : (
            <div className="rounded-md border">
              <Table>
                <TableHeader>
                  <TableRow className="bg-muted/50">
                    <TableHead className="text-[11px] font-semibold uppercase tracking-wider text-muted-foreground w-12">ID</TableHead>
                    <TableHead className="text-[11px] font-semibold uppercase tracking-wider text-muted-foreground">服务器名称</TableHead>
                    <TableHead className="text-[11px] font-semibold uppercase tracking-wider text-muted-foreground w-16">状态</TableHead>
                    <TableHead className="text-[11px] font-semibold uppercase tracking-wider text-muted-foreground w-16">节点数</TableHead>
                    <TableHead className="text-[11px] font-semibold uppercase tracking-wider text-muted-foreground">负载</TableHead>
                    <TableHead className="text-[11px] font-semibold uppercase tracking-wider text-muted-foreground">最后心跳</TableHead>
                    <TableHead className="text-[11px] font-semibold uppercase tracking-wider text-muted-foreground text-right">操作</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {filteredMachines.map((m) => {
                    const hb = timeSince(m.last_check_at)
                    return (
                      <TableRow key={m.id} className="group">
                        {/* ID */}
                        <TableCell className="font-mono text-[12px] text-foreground/80">{m.id}</TableCell>
                        {/* 名称 */}
                        <TableCell>
                          <button className="text-sm font-medium hover:text-primary hover:underline cursor-pointer text-left" onClick={() => setDetailMachine(m)}>
                            {m.name}
                          </button>
                          {m.notes && <p className="text-xs text-muted-foreground truncate max-w-[200px]">{m.notes}</p>}
                        </TableCell>
                        {/* 状态 */}
                        <TableCell><StatusDot status={m.status} /></TableCell>
                        {/* 节点数 */}
                        <TableCell>
                          <Badge variant={m.node_count > 0 ? 'secondary' : 'outline'} className="text-xs">
                            {m.node_count > 0 ? `${m.node_count} 个节点` : '暂无承载'}
                          </Badge>
                        </TableCell>
                        {/* 负载 */}
                        <TableCell>
                          {m.cpu > 0 || m.memory > 0 || m.disk > 0 ? (
                            <Tooltip>
                              <TooltipTrigger>
                                <div className="flex items-center gap-3">
                                  <LoadBar value={m.cpu} label="CPU" icon={Cpu} />
                                </div>
                              </TooltipTrigger>
                              <TooltipContent>
                                <div className="text-xs space-y-1 min-w-[140px]">
                                  <div className="flex justify-between"><span>CPU:</span><span className="font-mono">{m.cpu.toFixed(1)}%</span></div>
                                  <div className="flex justify-between"><span>内存:</span><span className="font-mono">{m.memory.toFixed(1)}%</span></div>
                                  <div className="flex justify-between"><span>磁盘:</span><span className="font-mono">{m.disk.toFixed(1)}%</span></div>
                                </div>
                              </TooltipContent>
                            </Tooltip>
                          ) : (
                            <span className="text-xs text-muted-foreground">暂无数据</span>
                          )}
                        </TableCell>
                        {/* 最后心跳 */}
                        <TableCell>
                          <span className={`text-xs ${hb.stale ? 'text-destructive' : 'text-muted-foreground'}`}>{hb.text}</span>
                        </TableCell>
                        {/* 操作 */}
                        <TableCell className="text-right">
                          <DropdownMenu>
                            <DropdownMenuTrigger asChild>
                              <Button variant="ghost" size="icon" className="h-7 w-7">
                                <MoreHorizontal className="h-4 w-4" />
                              </Button>
                            </DropdownMenuTrigger>
                            <DropdownMenuContent align="end" className="w-40">
                              <DropdownMenuItem onClick={() => { setEditingMachine(m); setDialogOpen(true) }}>
                                <Pencil className="mr-2 h-4 w-4" />编辑
                              </DropdownMenuItem>
                              <DropdownMenuItem onClick={() => setTokenMachine(m)}>
                                <Key className="mr-2 h-4 w-4" />查看 Token
                              </DropdownMenuItem>
                              <DropdownMenuItem onClick={() => setInstallMachine(m)}>
                                <Terminal className="mr-2 h-4 w-4" />安装命令
                              </DropdownMenuItem>
                              <DropdownMenuItem onClick={() => setDetailMachine(m)}>
                                <Eye className="mr-2 h-4 w-4" />查看详情
                              </DropdownMenuItem>
                              <DropdownMenuSeparator />
                              <DropdownMenuItem className="text-destructive focus:text-destructive" onClick={() => {
                                if (confirm('确认删除服务器？关联节点将自动解绑（不会被删除）。')) deleteMutation.mutate(m.id)
                              }}>
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
        </CardContent>
      </Card>

      {/* Create/Edit Dialog */}
      <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{editingMachine ? '编辑服务器' : '新建服务器'}</DialogTitle>
            <DialogDescription>{editingMachine ? '修改服务器名称、备注或启用状态。' : '当你希望一台服务器承载多个节点时，再创建服务器记录。'}</DialogDescription>
          </DialogHeader>
          <MachineForm machine={editingMachine} onClose={() => setDialogOpen(false)} onSave={handleSave} />
        </DialogContent>
      </Dialog>

      {/* Token Dialog */}
      {tokenMachine && (
        <TokenDialog
          open={!!tokenMachine}
          onOpenChange={() => setTokenMachine(null)}
          machineId={tokenMachine.id}
          machineName={tokenMachine.name}
        />
      )}

      {/* Install Command Dialog */}
      {installMachine && (
        <InstallCommandDialog
          open={!!installMachine}
          onOpenChange={() => setInstallMachine(null)}
          machineId={installMachine.id}
          machineName={installMachine.name}
        />
      )}

      {/* Detail Dialog */}
      <MachineDetailDialog open={!!detailMachine} onOpenChange={() => setDetailMachine(null)} machine={detailMachine} />
    </div>
  )
}
