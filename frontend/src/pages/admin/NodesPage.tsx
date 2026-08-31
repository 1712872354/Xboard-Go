import { useState, useEffect, useMemo, useCallback } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { toast } from 'sonner'
import { useAdminNodes, useNodeStats, useCreateNode, useUpdateNode, useDeleteNode, useUpdateNodeStatus } from '@/hooks/useNodes'
import { useServerGroups, useAllServerGroups, useCreateServerGroup, useUpdateServerGroup, useDeleteServerGroup } from '@/hooks/useServerGroups'
import { useServerRoutes, useCreateServerRoute, useUpdateServerRoute, useDeleteServerRoute } from '@/hooks/useServerRoutes'
import { useServers, useAllServers, useCreateServer, useUpdateServer, useUpdateServerStatus, useDeleteServer, useResetServerToken, getInstallCommand } from '@/hooks/useServers'
import { useNodeTemplates, useCreateNodeTemplate, useUpdateNodeTemplate, useDeleteNodeTemplate, type NodeTemplate } from '@/hooks/useNodeTemplates'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Switch } from '@/components/ui/switch'
import { Skeleton } from '@/components/ui/skeleton'
import { Progress } from '@/components/ui/progress'
import { Checkbox } from '@/components/ui/checkbox'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from '@/components/ui/tooltip'
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuSeparator, DropdownMenuTrigger } from '@/components/ui/dropdown-menu'
import {
  Plus, Pencil, Trash2, Server, Wifi, WifiOff, Wrench,
  Globe, Zap, Copy, Monitor, Network, Route, Shield, Search, Eye, Info, Clock,
  Activity, FileCode, ChevronDown, Key, Terminal, RotateCcw, MoreHorizontal, Users, HardDrive,
} from 'lucide-react'
import { TransportSettings } from '@/components/node/TransportSettings'
import { TlsSettings } from '@/components/node/TlsSettings'
import { ProtocolSettings } from '@/components/node/ProtocolSettings'
import { AdvancedSettings } from '@/components/node/AdvancedSettings'
import type { CertSettingsValue } from '@/components/node/CertSettings'
import type { MultiplexSettingsValue } from '@/components/node/MultiplexSettings'
import type { EchSettingsValue } from '@/components/node/EchSettings'
import api from '@/lib/api'
import { formatDate, formatBytes } from '@/lib/utils'
import type { Node, ServerGroup, ServerRoute, ServerMachine } from '@/types'

// ─── Shared ────────────────────────────────────────────
const nodeTypes = [
  { value: 'vmess', label: 'VMess', color: 'bg-blue-500/10 text-blue-600 dark:text-blue-400' },
  { value: 'vless', label: 'VLESS', color: 'bg-violet-500/10 text-violet-600 dark:text-violet-400' },
  { value: 'trojan', label: 'Trojan', color: 'bg-emerald-500/10 text-emerald-600 dark:text-emerald-400' },
  { value: 'shadowsocks', label: 'SS', color: 'bg-amber-500/10 text-amber-600 dark:text-amber-400' },
  { value: 'hysteria2', label: 'Hysteria2', color: 'bg-rose-500/10 text-rose-600 dark:text-rose-400' },
  { value: 'hysteria', label: 'Hysteria', color: 'bg-rose-500/10 text-rose-400 dark:text-rose-300' },
  { value: 'tuic', label: 'TUIC', color: 'bg-cyan-500/10 text-cyan-600 dark:text-cyan-400' },
  { value: 'anytls', label: 'AnyTLS', color: 'bg-pink-500/10 text-pink-600 dark:text-pink-400' },
  { value: 'naive', label: 'Naive', color: 'bg-gray-500/10 text-gray-600 dark:text-gray-400' },
  { value: 'socks', label: 'SOCKS', color: 'bg-gray-500/10 text-gray-600 dark:text-gray-400' },
  { value: 'http', label: 'HTTP', color: 'bg-gray-500/10 text-gray-600 dark:text-gray-400' },
  { value: 'mieru', label: 'Mieru', color: 'bg-teal-500/10 text-teal-600 dark:text-teal-400' },
]
const nodeTypeMap = Object.fromEntries(nodeTypes.map((t) => [t.value, t]))

function NodeTypeBadge({ type }: { type: string }) {
  const info = nodeTypeMap[type]
  return (
    <span className={`inline-flex items-center rounded-md px-2 py-0.5 text-xs font-medium ${info?.color ?? 'bg-muted text-muted-foreground'}`}>
      {info?.label ?? type}
    </span>
  )
}

function StatusDot({ status }: { status: number }) {
  const map: Record<number, { color: string; label: string }> = {
    0: { color: 'bg-red-500', label: '离线' },
    1: { color: 'bg-emerald-500', label: '在线' },
    2: { color: 'bg-amber-500', label: '维护中' },
  }
  const s = map[status] ?? map[0]
  return (
    <span className="inline-flex items-center gap-1.5 text-xs">
      <span className={`h-2 w-2 rounded-full ${s.color}`} />
      {s.label}
    </span>
  )
}

function StatCard({ title, value, icon: Icon, color }: {
  title: string; value: number; icon: React.ElementType; color: string
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

function ConfirmDialog({ open, onOpenChange, title, description, onConfirm, loading }: {
  open: boolean; onOpenChange: (v: boolean) => void; title: string; description: string;
  onConfirm: () => void; loading?: boolean
}) {
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{title}</DialogTitle>
          <DialogDescription>{description}</DialogDescription>
        </DialogHeader>
        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>取消</Button>
          <Button variant="destructive" onClick={onConfirm} disabled={loading}>
            {loading ? '处理中...' : '确认删除'}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}

function Pagination({ page, totalPages, total, onPageChange }: {
  page: number; totalPages: number; total: number; onPageChange: (p: number) => void
}) {
  return (
    <div className="mt-4 flex items-center justify-between">
      <p className="text-sm text-muted-foreground">共 {total} 条</p>
      <div className="flex items-center gap-2">
        <Button size="sm" variant="outline" disabled={page <= 1} onClick={() => onPageChange(page - 1)}>上一页</Button>
        <span className="min-w-[60px] text-center text-sm text-muted-foreground">{page} / {totalPages}</span>
        <Button size="sm" variant="outline" disabled={page >= totalPages} onClick={() => onPageChange(page + 1)}>下一页</Button>
      </div>
    </div>
  )
}

function timeSince(dateStr: string | null): { text: string; stale: boolean } {
  if (!dateStr) return { text: '从未', stale: true }
  const diff = Date.now() - new Date(dateStr).getTime()
  const mins = Math.floor(diff / 60000)
  if (mins < 1) return { text: '刚刚', stale: false }
  if (mins < 60) return { text: `${mins}分钟前`, stale: mins > 5 }
  const hours = Math.floor(mins / 60)
  if (hours < 24) return { text: `${hours}小时前`, stale: true }
  return { text: `${Math.floor(hours / 24)}天前`, stale: true }
}

function formatUptime(seconds: number): string {
  if (!seconds || seconds <= 0) return '-'
  const days = Math.floor(seconds / 86400)
  const hours = Math.floor((seconds % 86400) / 3600)
  const mins = Math.floor((seconds % 3600) / 60)
  if (days > 0) return `${days}天${hours}小时`
  if (hours > 0) return `${hours}小时${mins}分`
  return `${mins}分钟`
}

// ─── Node Detail Dialog ────────────────────────────────
function NodeDetailDialog({ open, onOpenChange, node, groupName }: {
  open: boolean; onOpenChange: (v: boolean) => void; node: Node | null; groupName: string
}) {
  if (!node) return null
  let parsedInfo: Record<string, unknown> = {}
  try { parsedInfo = JSON.parse(node.server_info || '{}') } catch { /* ignore */ }
  const formattedJson = JSON.stringify(parsedInfo, null, 2)

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-lg max-h-[80vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <Info className="h-4 w-4" />节点详情 — {node.name}
          </DialogTitle>
          <DialogDescription>ID: {node.id} | 类型: {node.type} | 地址: {node.address}:{node.port}</DialogDescription>
        </DialogHeader>
        <div className="space-y-4 text-sm">
          <div className="grid grid-cols-2 gap-3">
            <div><span className="text-muted-foreground">分组:</span> {groupName || '-'}</div>
            <div><span className="text-muted-foreground">流量倍率:</span> {node.rate}x</div>
            <div><span className="text-muted-foreground">在线用户:</span> {node.online_users ?? 0}</div>
            <div><span className="text-muted-foreground">父节点ID:</span> {node.parent_id || '-'}</div>
            <div className="flex items-center gap-1"><span className="text-muted-foreground">状态:</span> <StatusDot status={node.status} /></div>
            <div className="flex items-center gap-1"><span className="text-muted-foreground">心跳:</span> {timeSince(node.last_online).text}</div>
            <div className="col-span-2"><span className="text-muted-foreground">创建时间:</span> {formatDate(node.created_at)}</div>
            <div className="col-span-2"><span className="text-muted-foreground">更新时间:</span> {formatDate(node.updated_at)}</div>
          </div>
          {(() => {
            try {
              const info = JSON.parse(node.server_info || '{}')
              const metrics = info.metrics
              if (!metrics) return null
              return (
                <div className="rounded-lg border p-3 space-y-2">
                  <p className="font-medium text-muted-foreground flex items-center gap-1.5"><Activity className="h-4 w-4" />运行指标</p>
                  <div className="grid grid-cols-2 gap-2 text-xs">
                    <div><span className="text-muted-foreground">CPU:</span> {(metrics.cpu ?? 0).toFixed(1)}%</div>
                    <div><span className="text-muted-foreground">内存:</span> {((metrics.mem_used ?? 0) / 1073741824).toFixed(1)}G / {((metrics.mem_total ?? 0) / 1073741824).toFixed(1)}G</div>
                    <div><span className="text-muted-foreground">磁盘:</span> {((metrics.disk_used ?? 0) / 1073741824).toFixed(1)}G / {((metrics.disk_total ?? 0) / 1073741824).toFixed(1)}G</div>
                    {metrics.swap_total > 0 && <div><span className="text-muted-foreground">交换区:</span> {((metrics.swap_used ?? 0) / 1073741824).toFixed(1)}G / {((metrics.swap_total ?? 0) / 1073741824).toFixed(1)}G</div>}
                    <div><span className="text-muted-foreground">运行时长:</span> {formatUptime(metrics.uptime ?? 0)}</div>
                    <div><span className="text-muted-foreground">并发协程:</span> {metrics.goroutines ?? 0}</div>
                    <div><span className="text-muted-foreground">实时连接:</span> {metrics.connections ?? 0}</div>
                    <div><span className="text-muted-foreground">在线用户:</span> {metrics.users ?? 0}</div>
                    {metrics.speed !== undefined && <div><span className="text-muted-foreground">实时速率:</span> {formatBytes(metrics.speed ?? 0)}/s</div>}
                    {metrics.load !== undefined && <div><span className="text-muted-foreground">系统负载:</span> {(metrics.load ?? 0).toFixed(2)}</div>}
                  </div>
                </div>
              )
            } catch { return null }
          })()}
          <div>
            <p className="mb-2 font-medium text-muted-foreground">server_info 配置:</p>
            <div className="relative rounded-md border bg-muted/50 p-3 max-h-[300px] overflow-auto">
              <pre className="text-xs font-mono whitespace-pre-wrap break-all">{formattedJson}</pre>
            </div>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  )
}

// ─── Tab 1: Nodes ──────────────────────────────────────
const nodeSchema = z.object({
  name: z.string().min(1, '请输入名称'),
  type: z.string().min(1, '请选择类型'),
  address: z.string().min(1, '请输入地址'),
  port: z.coerce.number().min(1).max(65535),
  server_info: z.string().optional(),
  group_id: z.coerce.number().optional(),
  rate: z.coerce.number().min(0).max(100).optional(),
  parent_id: z.coerce.number().optional(),
  machine_id: z.coerce.number().optional(),
  listen_address: z.string().optional(),
  server_port: z.coerce.number().optional(),
  show: z.boolean().optional().default(true),
  banned: z.boolean().optional().default(false),
  traffic_limit: z.coerce.number().optional(),
  tags: z.string().optional(),
  code: z.string().optional(),
  route_id: z.coerce.number().optional(),
})
type NodeForm = z.infer<typeof nodeSchema>

const TRANSPORT_TYPES = ['tcp', 'ws', 'grpc', 'h2', 'httpupgrade', 'xhttp']
const PROTOCOLS_WITH_SETTINGS = ['vless', 'shadowsocks', 'hysteria', 'hysteria2', 'tuic', 'anytls', 'mieru']

function parseServerInfo(raw: string): Record<string, any> {
  try { return JSON.parse(raw || '{}') } catch { return {} }
}

function NodeFormDialog({ open, onOpenChange, onSubmit, defaultValues, title, groups, nodes, servers, routes }: {
  open: boolean; onOpenChange: (v: boolean) => void; onSubmit: (data: NodeForm) => void;
  defaultValues?: Partial<NodeForm>; title: string;
  groups: ServerGroup[]; nodes: Node[]; servers: ServerMachine[]; routes?: ServerRoute[]
}) {
  const { register, handleSubmit, reset, setValue, watch, formState: { errors } } = useForm<NodeForm>({
    resolver: zodResolver(nodeSchema),
    defaultValues: { rate: 1, group_id: 0, parent_id: 0, server_info: '{}', ...defaultValues },
  })
  const typeVal = watch('type')
  const groupVal = watch('group_id')
  const [tags, setTags] = useState<string[]>(defaultValues?.tags ? defaultValues.tags.split(',').filter(Boolean) : [])
  const [tagInput, setTagInput] = useState('')

  // Protocol-specific sub-state derived from server_info JSON
  const [transportType, setTransportType] = useState<string>('tcp')
  const [transportSettings, setTransportSettings] = useState<Record<string, any>>({})
  const [tlsMode, setTlsMode] = useState<0 | 1 | 2>(0)
  const [tlsSettings, setTlsSettings] = useState<Record<string, any>>({})
  const [protocolSettings, setProtocolSettings] = useState<Record<string, any>>({})
  const [advancedOpen, setAdvancedOpen] = useState(false)
  const [certValue, setCertValue] = useState<CertSettingsValue>({})
  const [multiplexValue, setMultiplexValue] = useState<MultiplexSettingsValue>({})
  const [echValue, setEchValue] = useState<EchSettingsValue>({})
  const [customOutbounds, setCustomOutbounds] = useState('[]')
  const [customRoutes, setCustomRoutes] = useState('[]')

  // Parse server_info into structured fields when opening / switching to edit
  const syncFromJson = useCallback((raw: string) => {
    const info = parseServerInfo(raw)
    setTransportType(info.network ?? 'tcp')
    setTransportSettings(info.network_settings ?? {})
    setTlsMode((info.tls as 0 | 1 | 2) ?? 0)
    setTlsSettings(info.tls_settings ?? {})
    setCertValue(info.cert_config || {})
    setMultiplexValue(info.multiplex || {})
    setEchValue(info.ech || {})
    setCustomOutbounds(info.custom_outbounds ? JSON.stringify(info.custom_outbounds, null, 2) : '[]')
    setCustomRoutes(info.custom_routes ? JSON.stringify(info.custom_routes, null, 2) : '[]')
    // Everything else is protocol-specific
    const EXCLUDED_KEYS = ['network', 'network_settings', 'tls', 'tls_settings', 'cert_config', 'multiplex', 'ech', 'custom_outbounds', 'custom_routes', 'listen_ip', 'server_port', 'show', 'banned', 'transfer_enable_gb']
    const proto: Record<string, any> = {}
    for (const [k, v] of Object.entries(info)) {
      if (!EXCLUDED_KEYS.includes(k)) {
        proto[k] = v
      }
    }
    setProtocolSettings(proto)
  }, [])

  useEffect(() => {
    if (open && defaultValues) {
      const merged = { rate: 1, group_id: 0, parent_id: 0, server_info: '{}', show: true, banned: false, ...defaultValues }
      reset(merged)
      syncFromJson(merged.server_info ?? '{}')

      // Parse server_info to populate additional form fields
      const info = parseServerInfo(merged.server_info ?? '{}')
      if (info.listen_ip) setValue('listen_address', String(info.listen_ip))
      if (info.server_port) setValue('server_port', Number(info.server_port))
      if (info.show !== undefined) setValue('show', Number(info.show) !== 0)
      if (info.banned !== undefined) setValue('banned', Number(info.banned) === 1)
      if (info.transfer_enable_gb) setValue('traffic_limit', Number(info.transfer_enable_gb))
      if (info.machine_id) setValue('machine_id', Number(info.machine_id))
    }
  }, [open, defaultValues, reset, syncFromJson, setValue])

  // When user edits the raw JSON textarea, re-sync structured fields
  // (kept for future use if raw JSON editing is re-enabled)

  const handleFormSubmit = (d: NodeForm) => {
    // Assemble structured fields into server_info JSON
    const serverInfo: Record<string, any> = {}
    if (transportType && transportType !== 'tcp') {
      serverInfo.network = transportType
      if (transportSettings && Object.keys(transportSettings).length > 0) {
        serverInfo.network_settings = transportSettings
      }
    }
    if (tlsMode > 0) {
      serverInfo.tls = tlsMode
      if (tlsSettings && Object.keys(tlsSettings).length > 0) {
        serverInfo.tls_settings = tlsSettings
      }
    }
    // Merge protocol-specific fields
    Object.assign(serverInfo, protocolSettings)

    // Listen address
    if (d.listen_address) serverInfo.listen_ip = d.listen_address

    // Advanced: cert, multiplex, ECH
    if (certValue && Object.keys(certValue).length > 0) serverInfo.cert_config = certValue
    if (multiplexValue?.enabled) serverInfo.multiplex = multiplexValue
    if (echValue && Object.keys(echValue).filter(k => echValue[k as keyof EchSettingsValue]).length > 0) serverInfo.ech = echValue
    if (customOutbounds && customOutbounds !== '[]') try { serverInfo.custom_outbounds = JSON.parse(customOutbounds) } catch { /* ignore */ }
    if (customRoutes && customRoutes !== '[]') try { serverInfo.custom_routes = JSON.parse(customRoutes) } catch { /* ignore */ }

    // Server port (if different from connection port)
    if (d.server_port) serverInfo.server_port = d.server_port

    // Machine ID
    if (d.machine_id) serverInfo.machine_id = d.machine_id

    // Show/hide, banned
    if (!d.show) serverInfo.show = 0
    if (d.banned) serverInfo.banned = 1

    // Traffic limit
    if (d.traffic_limit) serverInfo.transfer_enable_gb = d.traffic_limit

    // Tags
    if (d.tags) serverInfo.tags = d.tags.split(',').filter(Boolean)

    // Route ID
    if (d.route_id) serverInfo.route_id = d.route_id

    // Custom code
    if (d.code) serverInfo.code = d.code

    const payload = { ...d }
    payload.server_info = JSON.stringify(serverInfo)

    // Remove internal-only fields from the top-level payload
    delete (payload as any).listen_address
    delete (payload as any).server_port
    delete (payload as any).show
    delete (payload as any).banned
    delete (payload as any).traffic_limit
    delete (payload as any).machine_id

    onSubmit(payload)
    reset()
  }

  return (
    <>
    <Dialog open={open} onOpenChange={(v) => { onOpenChange(v); if (!v) reset() }}>
      <DialogContent className="max-w-xl max-h-[90vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>{title}</DialogTitle>
          <DialogDescription>配置代理节点的连接信息</DialogDescription>
        </DialogHeader>
        <form onSubmit={handleSubmit(handleFormSubmit)} className="space-y-5">
          {/* ── Basic Info ── */}
          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-2">
              <Label>节点名称 <span className="text-destructive">*</span></Label>
              <Input placeholder="如：香港 IPLC-01" {...register('name')} />
              {errors.name && <p className="text-xs text-destructive">{errors.name.message}</p>}
            </div>
            <div className="space-y-2">
              <Label>协议类型 <span className="text-destructive">*</span></Label>
              <Select value={typeVal} onValueChange={(v) => setValue('type', v, { shouldValidate: true })}>
                <SelectTrigger><SelectValue placeholder="选择协议" /></SelectTrigger>
                <SelectContent>
                  {nodeTypes.map((t) => <SelectItem key={t.value} value={t.value}>{t.label}</SelectItem>)}
                </SelectContent>
              </Select>
            </div>
          </div>
          <div className="grid grid-cols-3 gap-4">
            <div className="col-span-2 space-y-2">
              <Label>服务器地址 <span className="text-destructive">*</span></Label>
              <Input placeholder="hk01.example.com" {...register('address')} />
            </div>
            <div className="space-y-2">
              <Label>端口 <span className="text-destructive">*</span></Label>
              <Input type="number" placeholder="443" {...register('port')} />
            </div>
          </div>
          <div className="grid grid-cols-3 gap-4">
            <div className="space-y-2">
              <Label>服务端口</Label>
              <Input type="number" placeholder="与连接端口相同时留空" {...register('server_port')} />
            </div>
            <div className="col-span-2 space-y-2">
              <Label>监听地址</Label>
              <Input placeholder="留空使用默认" {...register('listen_address')} />
            </div>
          </div>
          <div className="grid grid-cols-3 gap-4">
            <div className="space-y-2">
              <Label>分组</Label>
              <Select value={String(groupVal ?? 0)} onValueChange={(v) => setValue('group_id', Number(v))}>
                <SelectTrigger><SelectValue placeholder="选择分组" /></SelectTrigger>
                <SelectContent>
                  <SelectItem value="0">无分组</SelectItem>
                  {groups.map((g) => <SelectItem key={g.id} value={String(g.id)}>{g.name}</SelectItem>)}
                </SelectContent>
              </Select>
            </div>
            <div className="space-y-2">
              <Label>流量倍率</Label>
              <Input type="number" step="0.1" placeholder="1.0" {...register('rate')} />
            </div>
            <div className="space-y-2">
              <Label>流量限制 (GB)</Label>
              <Input type="number" placeholder="不限" {...register('traffic_limit')} />
            </div>
          </div>
          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-2">
              <Label>父节点</Label>
              <Select value={String(watch('parent_id') ?? 0)} onValueChange={(v) => setValue('parent_id', Number(v))}>
                <SelectTrigger><SelectValue placeholder="无" /></SelectTrigger>
                <SelectContent>
                  <SelectItem value="0">无（独立节点）</SelectItem>
                  {nodes.map((n) => <SelectItem key={n.id} value={String(n.id)}>{n.name}</SelectItem>)}
                </SelectContent>
              </Select>
            </div>
            <div className="space-y-2">
              <Label>服务器机器</Label>
              <Select value={String(watch('machine_id') ?? 0)} onValueChange={(v) => setValue('machine_id', Number(v))}>
                <SelectTrigger><SelectValue placeholder="无" /></SelectTrigger>
                <SelectContent>
                  <SelectItem value="0">无</SelectItem>
                  {servers.map((s) => <SelectItem key={s.id} value={String(s.id)}>{s.name}</SelectItem>)}
                </SelectContent>
              </Select>
            </div>
          </div>
          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-2">
              <Label>路由组</Label>
              <Select value={String(watch('route_id') ?? 0)} onValueChange={(v) => setValue('route_id', Number(v))}>
                <SelectTrigger><SelectValue placeholder="无" /></SelectTrigger>
                <SelectContent>
                  <SelectItem value="0">无</SelectItem>
                  {(routes ?? []).map((r) => <SelectItem key={r.id} value={String(r.id)}>{r.name}</SelectItem>)}
                </SelectContent>
              </Select>
            </div>
            <div className="space-y-2">
              <Label>自定义节点ID <span className="text-xs text-muted-foreground">(选填)</span></Label>
              <Input placeholder="留空自动生成" {...register('code')} />
            </div>
          </div>
          <div className="space-y-2">
            <Label>节点标签</Label>
            <div className="flex flex-wrap gap-1.5 mb-2">
              {tags.map((tag, i) => (
                <Badge key={i} variant="secondary" className="gap-1">
                  {tag}
                  <button type="button" className="ml-1 hover:text-destructive" onClick={() => {
                    const newTags = tags.filter((_, idx) => idx !== i)
                    setTags(newTags)
                    setValue('tags', newTags.join(','))
                  }}>×</button>
                </Badge>
              ))}
            </div>
            <div className="flex gap-2">
              <Input
                placeholder="输入标签后回车添加"
                value={tagInput}
                onChange={(e) => setTagInput(e.target.value)}
                onKeyDown={(e) => {
                  if (e.key === 'Enter' && tagInput.trim()) {
                    e.preventDefault()
                    const newTags = [...tags, tagInput.trim()]
                    setTags(newTags)
                    setValue('tags', newTags.join(','))
                    setTagInput('')
                  }
                }}
              />
            </div>
            <input type="hidden" {...register('tags')} />
          </div>
          <div className="flex items-center gap-6">
            <div className="flex items-center gap-2">
              <Switch checked={watch('show') ?? true} onCheckedChange={(checked) => setValue('show', checked)} />
              <Label>显示节点</Label>
            </div>
            <div className="flex items-center gap-2">
              <Switch checked={watch('banned') ?? false} onCheckedChange={(checked) => setValue('banned', checked)} />
              <Label>封禁节点</Label>
            </div>
          </div>

          {/* ── Transport Settings ── */}
          <div className="space-y-3 rounded-lg border p-4">
            <Label className="text-sm font-medium">传输设置</Label>
            <div className="space-y-2">
              <Label className="text-xs text-muted-foreground">传输类型</Label>
              <Select value={transportType} onValueChange={setTransportType}>
                <SelectTrigger><SelectValue placeholder="选择传输类型" /></SelectTrigger>
                <SelectContent>
                  {TRANSPORT_TYPES.map((t) => <SelectItem key={t} value={t}>{t}</SelectItem>)}
                </SelectContent>
              </Select>
            </div>
            <TransportSettings transportType={transportType} value={transportSettings} onChange={setTransportSettings} />
          </div>

          {/* ── TLS Settings ── */}
          <div className="space-y-3 rounded-lg border p-4">
            <Label className="text-sm font-medium">TLS 设置</Label>
            <TlsSettings mode={tlsMode} value={tlsSettings} onChange={(v) => {
              if ('mode' in v) {
                setTlsMode(v.mode as 0 | 1 | 2)
                const { mode: _, ...rest } = v
                setTlsSettings(rest)
              } else {
                setTlsSettings(v)
              }
            }} protocol={typeVal ?? ''} />
          </div>

          {/* ── Protocol-specific Settings ── */}
          {typeVal && PROTOCOLS_WITH_SETTINGS.includes(typeVal) && (
            <div className="space-y-3 rounded-lg border p-4">
              <Label className="text-sm font-medium">协议设置 ({typeVal})</Label>
              <ProtocolSettings protocol={typeVal} value={protocolSettings} onChange={setProtocolSettings} />
            </div>
          )}

          <Button type="button" variant="outline" onClick={() => setAdvancedOpen(true)}>
            高级设置
          </Button>

          <DialogFooter>
            <Button type="button" variant="outline" onClick={() => onOpenChange(false)}>取消</Button>
            <Button type="submit">保存</Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
    <AdvancedSettings
      open={advancedOpen}
      onOpenChange={setAdvancedOpen}
      certValue={certValue}
      certOnChange={setCertValue}
      multiplexValue={multiplexValue}
      multiplexOnChange={setMultiplexValue}
      echValue={echValue}
      echOnChange={setEchValue}
    />
    </>
  )
}

function NodesTab() {
  const [page, setPage] = useState(1)
  const [dialog, setDialog] = useState<'create' | 'edit' | null>(null)
  const [editNode, setEditNode] = useState<Node | null>(null)
  const [deleteId, setDeleteId] = useState<number | null>(null)
  const [searchQuery, setSearchQuery] = useState('')
  const [filterGroupId, setFilterGroupId] = useState<number | undefined>(undefined)
  const [filterType, setFilterType] = useState('')
  const [detailNode, setDetailNode] = useState<Node | null>(null)
  const [cloneNode, setCloneNode] = useState<Node | null>(null)
  const [selectedIds, setSelectedIds] = useState<Set<number>>(new Set())
  const [batchLoading, setBatchLoading] = useState(false)

  const { data, isLoading } = useAdminNodes(page, 20, filterGroupId)
  const { data: nodeStats } = useNodeStats()
  const { data: groupsData } = useServerGroups()
  const { data: allNodes } = useAdminNodes(1, 1000)
  const { data: allServers } = useAllServers()
  const { data: routesData } = useServerRoutes(1, 1000)
  const createNode = useCreateNode()
  const updateNode = useUpdateNode()
  const deleteNode = useDeleteNode()
  const updateStatus = useUpdateNodeStatus()

  const nodes = data?.list ?? []
  const groups = groupsData?.list ?? []
  const allGroups = useAllServerGroups().data ?? []
  const groupMap = useMemo(() => Object.fromEntries(groups.map((g) => [g.id, g.name])), [groups])
  const allGroupMap = useMemo(() => Object.fromEntries(allGroups.map((g) => [g.id, g.name])), [allGroups])
  const totalPages = data ? Math.ceil(data.total / (data.page_size || 20)) : 1

  const filteredNodes = useMemo(() => {
    let result = nodes
    if (filterType) {
      result = result.filter((n) => n.type === filterType)
    }
    if (searchQuery.trim()) {
      const q = searchQuery.toLowerCase()
      result = result.filter((n) => n.name.toLowerCase().includes(q) || n.address.toLowerCase().includes(q))
    }
    return result
  }, [nodes, searchQuery, filterType])

  const toggleSelect = useCallback((id: number) => {
    setSelectedIds((prev) => {
      const next = new Set(prev)
      if (next.has(id)) next.delete(id)
      else next.add(id)
      return next
    })
  }, [])

  const toggleSelectAll = useCallback(() => {
    if (selectedIds.size === filteredNodes.length && filteredNodes.length > 0) {
      setSelectedIds(new Set())
    } else {
      setSelectedIds(new Set(filteredNodes.map((n) => n.id)))
    }
  }, [filteredNodes, selectedIds.size])

  const handleBatchAction = useCallback(async (action: 'enable' | 'disable' | 'maintenance' | 'delete' | 'resetTraffic') => {
    const ids = Array.from(selectedIds)
    if (ids.length === 0) return
    setBatchLoading(true)
    try {
      if (action === 'resetTraffic') {
        await api.post('/admin/nodes/batch-reset-traffic', { ids })
        toast.success(`成功重置 ${ids.length} 个节点的流量`)
      } else {
        await api.post('/admin/nodes/batch', { ids, action })
        toast.success(`批量操作成功：${ids.length} 个节点`)
      }
      setSelectedIds(new Set())
      window.location.reload()
    } catch {
      toast.error('批量操作失败')
    } finally {
      setBatchLoading(false)
    }
  }, [selectedIds])

  const handleCreateSubmit = (d: NodeForm) => {
    const payload = { ...d }
    if (payload.server_info) {
      try { JSON.parse(payload.server_info) } catch { toast.error('server_info JSON 格式不正确'); return }
    }
    createNode.mutate(payload, { onSuccess: () => { toast.success('节点创建成功'); setDialog(null) } })
  }

  const handleUpdateSubmit = (d: NodeForm) => {
    if (!editNode) return
    const payload = { ...d }
    if (payload.server_info) {
      try { JSON.parse(payload.server_info) } catch { toast.error('server_info JSON 格式不正确'); return }
    }
    updateNode.mutate({ id: editNode.id, ...payload }, { onSuccess: () => { toast.success('节点更新成功'); setDialog(null); setEditNode(null) } })
  }

  const handleCloneSubmit = (d: NodeForm) => {
    const payload = { ...d }
    if (payload.server_info) {
      try { JSON.parse(payload.server_info) } catch { toast.error('server_info JSON 格式不正确'); return }
    }
    createNode.mutate(payload, { onSuccess: () => { toast.success('节点克隆成功'); setCloneNode(null) } })
  }

  return (
    <div className="space-y-4">
      <div className="grid gap-4 grid-cols-2 lg:grid-cols-4">
        <StatCard title="总节点" value={nodeStats?.total ?? data?.total ?? 0} icon={Server} color="bg-primary/10 text-primary" />
        <StatCard title="在线" value={nodeStats?.online ?? 0} icon={Wifi} color="bg-emerald-500/10 text-emerald-500" />
        <StatCard title="离线" value={nodeStats?.offline ?? 0} icon={WifiOff} color="bg-red-500/10 text-red-500" />
        <StatCard title="维护中" value={nodeStats?.maintenance ?? 0} icon={Wrench} color="bg-amber-500/10 text-amber-500" />
      </div>

      <div className="flex items-center gap-3 flex-wrap">
        <div className="relative flex-1 min-w-[200px] max-w-sm">
          <Search className="absolute left-2.5 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
          <Input placeholder="搜索节点名称或地址..." value={searchQuery} onChange={(e) => setSearchQuery(e.target.value)} className="pl-9" />
        </div>
        <Select value={filterType || 'all'} onValueChange={(v) => { setFilterType(v === 'all' ? '' : v); setPage(1) }}>
          <SelectTrigger className="w-[120px]"><SelectValue placeholder="全部类型" /></SelectTrigger>
          <SelectContent>
            <SelectItem value="all">全部类型</SelectItem>
            {nodeTypes.map((t) => <SelectItem key={t.value} value={t.value}>{t.label}</SelectItem>)}
          </SelectContent>
        </Select>
        <Select value={String(filterGroupId ?? 0)} onValueChange={(v) => { setFilterGroupId(v === '0' ? undefined : Number(v)); setPage(1) }}>
          <SelectTrigger className="w-[160px]"><SelectValue placeholder="全部分组" /></SelectTrigger>
          <SelectContent>
            <SelectItem value="0">全部分组</SelectItem>
            {allGroups.map((g) => <SelectItem key={g.id} value={String(g.id)}>{g.name}</SelectItem>)}
          </SelectContent>
        </Select>
        <Button onClick={() => { setCloneNode(null); setDialog('create') }}><Plus className="mr-2 h-4 w-4" />新建节点</Button>
      </div>

      {selectedIds.size > 0 && (
        <div className="flex items-center gap-3 rounded-lg border border-primary/20 bg-primary/5 px-4 py-2.5">
          <span className="text-sm font-medium">已选择 {selectedIds.size} 项</span>
          <div className="flex items-center gap-2 ml-auto">
            <Button size="sm" variant="outline" disabled={batchLoading} onClick={() => handleBatchAction('enable')}>
              <Wifi className="mr-1.5 h-3.5 w-3.5" />批量启用
            </Button>
            <Button size="sm" variant="outline" disabled={batchLoading} onClick={() => handleBatchAction('disable')}>
              <WifiOff className="mr-1.5 h-3.5 w-3.5" />批量禁用
            </Button>
            <Button size="sm" variant="outline" disabled={batchLoading} onClick={() => handleBatchAction('maintenance')}>
              <Wrench className="mr-1.5 h-3.5 w-3.5" />批量维护
            </Button>
            <Button size="sm" variant="outline" disabled={batchLoading} onClick={() => handleBatchAction('resetTraffic')}>
              <RotateCcw className="mr-1.5 h-3.5 w-3.5" />批量重置流量
            </Button>
            <Button size="sm" variant="destructive" disabled={batchLoading} onClick={() => handleBatchAction('delete')}>
              <Trash2 className="mr-1.5 h-3.5 w-3.5" />批量删除
            </Button>
          </div>
        </div>
      )}

      {isLoading ? (
        <div className="space-y-3">{Array.from({ length: 6 }).map((_, i) => <Skeleton key={i} className="h-12" />)}</div>
      ) : filteredNodes.length === 0 ? (
        <div className="flex flex-col items-center justify-center py-16 text-muted-foreground">
          <Server className="mb-3 h-10 w-10" />
          <p className="text-sm">{searchQuery || filterGroupId || filterType ? '没有匹配的节点' : '暂无节点'}</p>
        </div>
      ) : (
        <>
          <div className="rounded-md border">
            <Table>
              <TableHeader>
                <TableRow className="bg-muted/50">
                  <TableHead className="w-10 text-center">
                    <Checkbox
                      checked={filteredNodes.length > 0 && selectedIds.size === filteredNodes.length}
                      onCheckedChange={toggleSelectAll}
                    />
                  </TableHead>
                  <TableHead className="text-[11px] font-semibold uppercase tracking-wider text-muted-foreground">节点ID</TableHead>
                  <TableHead className="text-[11px] font-semibold uppercase tracking-wider text-muted-foreground w-14">显隐</TableHead>
                  <TableHead className="text-[11px] font-semibold uppercase tracking-wider text-muted-foreground">节点</TableHead>
                  <TableHead className="text-[11px] font-semibold uppercase tracking-wider text-muted-foreground">类型</TableHead>
                  <TableHead className="text-[11px] font-semibold uppercase tracking-wider text-muted-foreground">地址</TableHead>
                  <TableHead className="text-[11px] font-semibold uppercase tracking-wider text-muted-foreground">权限组</TableHead>
                  <TableHead className="text-[11px] font-semibold uppercase tracking-wider text-muted-foreground text-center">倍率</TableHead>
                  <TableHead className="text-[11px] font-semibold uppercase tracking-wider text-muted-foreground">流量使用</TableHead>
                  <TableHead className="text-[11px] font-semibold uppercase tracking-wider text-muted-foreground">部署</TableHead>
                  <TableHead className="text-[11px] font-semibold uppercase tracking-wider text-muted-foreground">负载</TableHead>
                  <TableHead className="text-[11px] font-semibold uppercase tracking-wider text-muted-foreground">状态</TableHead>
                  <TableHead className="text-[11px] font-semibold uppercase tracking-wider text-muted-foreground text-center">在线</TableHead>
                  <TableHead className="text-[11px] font-semibold uppercase tracking-wider text-muted-foreground text-right">操作</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {filteredNodes.map((n) => {
                  const info = (() => { try { return JSON.parse(n.server_info || '{}') } catch { return {} } })()
                  const isShow = info.show !== 0
                  const machineId = info.machine_id
                  const machine = machineId ? (allServers ?? []).find((s) => s.id === machineId) : null
                  const trafficLimit = info.transfer_enable_gb ?? 0
                  const metrics = info.metrics
                  const cpu = metrics?.cpu ?? 0
                  const hb = timeSince(n.last_online)
                  return (
                    <TableRow key={n.id} className={`group ${n.status === 0 ? 'opacity-50' : ''}`}>
                      {/* 选择 */}
                      <TableCell className="text-center">
                        <Checkbox checked={selectedIds.has(n.id)} onCheckedChange={() => toggleSelect(n.id)} />
                      </TableCell>
                      {/* 节点ID */}
                      <TableCell className="font-mono text-[12px] text-foreground/80">
                        {info.code || n.id}
                      </TableCell>
                      {/* 显隐 */}
                      <TableCell>
                        <Switch
                          checked={isShow}
                          onCheckedChange={(v) => {
                            info.show = v ? 1 : 0
                            updateNode.mutate({ id: n.id, server_info: JSON.stringify(info) })
                          }}
                          className="h-4 w-7 data-[state=checked]:bg-emerald-500"
                        />
                      </TableCell>
                      {/* 节点 */}
                      <TableCell>
                        <button className="text-sm font-medium hover:text-primary hover:underline cursor-pointer text-left truncate max-w-[160px] block" onClick={() => setDetailNode(n)}>
                          {n.name}
                        </button>
                        {n.tags && n.tags.split(',').filter(Boolean).length > 0 && (
                          <div className="flex gap-1 mt-0.5">
                            {n.tags.split(',').filter(Boolean).map((tag, i) => (
                              <span key={i} className="inline-block text-[10px] px-1 py-0 rounded bg-muted text-muted-foreground">{tag}</span>
                            ))}
                          </div>
                        )}
                      </TableCell>
                      {/* 类型 */}
                      <TableCell><NodeTypeBadge type={n.type} /></TableCell>
                      {/* 地址 */}
                      <TableCell>
                        <div className="flex items-center gap-1">
                          <span className="font-mono text-[12px] text-foreground/80">{n.address}:{n.port}</span>
                          <Button variant="ghost" size="icon" className="h-5 w-5 opacity-0 group-hover:opacity-100 transition-opacity"
                            onClick={() => { navigator.clipboard.writeText(`${n.address}:${n.port}`); toast.success('已复制') }}>
                            <Copy className="h-3 w-3" />
                          </Button>
                        </div>
                        {info.server_port && info.server_port !== n.port && (
                          <span className="text-[10px] text-muted-foreground">内部: {info.server_port}</span>
                        )}
                      </TableCell>
                      {/* 权限组 */}
                      <TableCell className="text-[12px] text-foreground/80">
                        {n.group_id ? (groupMap[n.group_id] || `#${n.group_id}`) : <span className="text-muted-foreground">--</span>}
                      </TableCell>
                      {/* 倍率 */}
                      <TableCell className="text-center">
                        <span className={`font-mono text-[12px] ${n.rate > 1 ? 'text-amber-600 dark:text-amber-400 font-semibold' : 'text-foreground/80'}`}>{n.rate}x</span>
                      </TableCell>
                      {/* 流量使用 */}
                      <TableCell>
                        {trafficLimit > 0 ? (
                          <Tooltip>
                            <TooltipTrigger>
                              <div className="w-20">
                                <div className="flex justify-between text-[10px] text-muted-foreground mb-0.5">
                                  <span>0G</span>
                                  <span>{trafficLimit}G</span>
                                </div>
                                <div className="h-1.5 w-full rounded-full bg-muted overflow-hidden">
                                  <div className="h-full rounded-full bg-primary" style={{ width: '0%' }} />
                                </div>
                              </div>
                            </TooltipTrigger>
                            <TooltipContent>
                              <div className="text-xs space-y-0.5">
                                <p>已用: 0 GB</p>
                                <p>限制: {trafficLimit} GB</p>
                                <p>使用率: 0%</p>
                              </div>
                            </TooltipContent>
                          </Tooltip>
                        ) : (
                          <span className="text-[12px] text-muted-foreground">不限</span>
                        )}
                      </TableCell>
                      {/* 部署 */}
                      <TableCell>
                        {machine ? (
                          <Tooltip>
                            <TooltipTrigger>
                              <Badge variant="secondary" className="text-[11px] font-normal cursor-default">
                                <HardDrive className="h-3 w-3 mr-1" />
                                {machine.name}
                              </Badge>
                            </TooltipTrigger>
                            <TooltipContent>
                              <p className="text-xs">服务器: {machine.name}</p>
                              <p className="text-xs">地址: {machine.address}</p>
                            </TooltipContent>
                          </Tooltip>
                        ) : (
                          <Badge variant="outline" className="text-[11px] font-normal">独立部署</Badge>
                        )}
                      </TableCell>
                      {/* 负载 */}
                      <TableCell>
                        {metrics ? (
                          <Tooltip>
                            <TooltipTrigger>
                              <div className="flex items-center gap-1.5">
                                <div className="w-10 h-1.5 rounded-full bg-muted overflow-hidden">
                                  <div className={`h-full rounded-full ${cpu > 80 ? 'bg-red-500' : cpu > 50 ? 'bg-amber-500' : 'bg-emerald-500'}`} style={{ width: `${Math.min(cpu, 100)}%` }} />
                                </div>
                                <span className="font-mono text-[11px] text-muted-foreground">{cpu.toFixed(0)}%</span>
                              </div>
                            </TooltipTrigger>
                            <TooltipContent>
                              <div className="text-xs space-y-1 min-w-[180px]">
                                <p className="font-medium mb-1">系统负载详情</p>
                                <div className="flex justify-between"><span>CPU 使用率:</span><span className="font-mono">{cpu.toFixed(1)}%</span></div>
                                <div className="flex justify-between"><span>内存使用:</span><span className="font-mono">{((metrics.mem_used ?? 0) / 1073741824).toFixed(1)}G / {((metrics.mem_total ?? 0) / 1073741824).toFixed(1)}G</span></div>
                                <div className="flex justify-between"><span>磁盘使用:</span><span className="font-mono">{((metrics.disk_used ?? 0) / 1073741824).toFixed(1)}G / {((metrics.disk_total ?? 0) / 1073741824).toFixed(1)}G</span></div>
                                {metrics.swap_total > 0 && <div className="flex justify-between"><span>交换区:</span><span className="font-mono">{((metrics.swap_used ?? 0) / 1073741824).toFixed(1)}G / {((metrics.swap_total ?? 0) / 1073741824).toFixed(1)}G</span></div>}
                                <div className="border-t pt-1 mt-1">
                                  <p className="font-medium mb-1">运行指标</p>
                                  {metrics.uptime && <div className="flex justify-between"><span>运行时长:</span><span className="font-mono">{formatUptime(metrics.uptime)}</span></div>}
                                  {metrics.goroutines !== undefined && <div className="flex justify-between"><span>并发协程:</span><span className="font-mono">{metrics.goroutines}</span></div>}
                                  {metrics.connections !== undefined && <div className="flex justify-between"><span>实时连接:</span><span className="font-mono">{metrics.connections}</span></div>}
                                  {metrics.users !== undefined && <div className="flex justify-between"><span>在线用户:</span><span className="font-mono">{metrics.users}</span></div>}
                                </div>
                              </div>
                            </TooltipContent>
                          </Tooltip>
                        ) : (
                          <span className="text-[12px] text-muted-foreground">-</span>
                        )}
                      </TableCell>
                      {/* 状态 */}
                      <TableCell>
                        <Select value={String(n.status)} onValueChange={(v) => updateStatus.mutate({ id: n.id, status: Number(v) })}>
                          <SelectTrigger className="h-6 w-20 border-0 bg-transparent p-0 focus:ring-0 text-[12px]">
                            <StatusDot status={n.status} />
                          </SelectTrigger>
                          <SelectContent>
                            <SelectItem value="1"><span className="flex items-center gap-1.5 text-xs"><span className="h-2 w-2 rounded-full bg-emerald-500" />运行正常</span></SelectItem>
                            <SelectItem value="0"><span className="flex items-center gap-1.5 text-xs"><span className="h-2 w-2 rounded-full bg-red-500" />未运行</span></SelectItem>
                            <SelectItem value="2"><span className="flex items-center gap-1.5 text-xs"><span className="h-2 w-2 rounded-full bg-amber-500" />维护中</span></SelectItem>
                          </SelectContent>
                        </Select>
                      </TableCell>
                      {/* 在线 */}
                      <TableCell className="text-center">
                        <span className="inline-flex items-center gap-1 font-mono text-[12px]">
                          <Users className="h-3 w-3 text-muted-foreground" />
                          {n.online_users ?? 0}
                        </span>
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
                            <DropdownMenuItem onClick={() => { setEditNode(n); setDialog('edit') }}>
                              <Pencil className="mr-2 h-4 w-4" />编辑
                            </DropdownMenuItem>
                            <DropdownMenuItem onClick={() => { setCloneNode(n); setDialog('create') }}>
                              <Copy className="mr-2 h-4 w-4" />复制
                            </DropdownMenuItem>
                            <DropdownMenuItem onClick={() => {
                              api.post(`/admin/nodes/${n.id}/reset-traffic`).then(() => {
                                toast.success('流量重置成功')
                                window.location.reload()
                              }).catch(() => toast.error('重置失败'))
                            }}>
                              <RotateCcw className="mr-2 h-4 w-4" />重置流量
                            </DropdownMenuItem>
                            <DropdownMenuSeparator />
                            <DropdownMenuItem className="text-destructive focus:text-destructive" onClick={() => setDeleteId(n.id)}>
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
          <Pagination page={page} totalPages={totalPages} total={data?.total ?? 0} onPageChange={setPage} />
        </>
      )}

      <NodeDetailDialog open={detailNode !== null} onOpenChange={() => setDetailNode(null)} node={detailNode} groupName={detailNode ? (allGroupMap[detailNode.group_id] || '') : ''} />
      <NodeFormDialog
        open={dialog === 'create'} onOpenChange={() => { setDialog(null); setCloneNode(null) }}
        onSubmit={cloneNode ? handleCloneSubmit : handleCreateSubmit}
        defaultValues={cloneNode ? { name: `${cloneNode.name}(副本)`, type: cloneNode.type, address: cloneNode.address, port: cloneNode.port, server_info: cloneNode.server_info || '{}', group_id: cloneNode.group_id, rate: cloneNode.rate, parent_id: cloneNode.parent_id, tags: cloneNode.tags } : {}}
        title={cloneNode ? '克隆节点' : '新建节点'} groups={allGroups} nodes={allNodes?.list ?? []} servers={allServers ?? []} routes={routesData?.list ?? []}
      />
      <NodeFormDialog
        open={dialog === 'edit' && !!editNode} onOpenChange={(v) => { if (!v) { setDialog(null); setEditNode(null) } }}
        onSubmit={handleUpdateSubmit}
        defaultValues={editNode ? { name: editNode.name, type: editNode.type, address: editNode.address, port: editNode.port, server_info: editNode.server_info || '{}', group_id: editNode.group_id, rate: editNode.rate, parent_id: editNode.parent_id, tags: editNode.tags } : {}}
        title="编辑节点" groups={allGroups} nodes={allNodes?.list?.filter((n) => n.id !== editNode?.id) ?? []} servers={allServers ?? []} routes={routesData?.list ?? []}
      />
      <ConfirmDialog open={deleteId !== null} onOpenChange={() => setDeleteId(null)} title="确认删除" description="删除后节点配置将无法恢复。" onConfirm={() => deleteId !== null && deleteNode.mutate(deleteId, { onSuccess: () => { setDeleteId(null); toast.success('节点已删除') } })} loading={deleteNode.isPending} />
    </div>
  )
}

// ─── Tab 2: Server Groups ─────────────────────────────
const groupSchema = z.object({
  name: z.string().min(1, '请输入名称'),
  description: z.string().optional(),
  plan_ids: z.string().optional(),
  sort: z.coerce.number().optional(),
  status: z.number().optional(),
})
type GroupForm = z.infer<typeof groupSchema>

function GroupFormDialog({ open, onOpenChange, onSubmit, defaultValues, title }: {
  open: boolean; onOpenChange: (v: boolean) => void; onSubmit: (data: GroupForm) => void;
  defaultValues?: Partial<GroupForm>; title: string
}) {
  const { register, handleSubmit, reset, setValue, watch, formState: { errors } } = useForm<GroupForm>({
    resolver: zodResolver(groupSchema), defaultValues: { sort: 0, status: 1, plan_ids: '', ...defaultValues },
  })
  const statusVal = watch('status')

  useEffect(() => {
    if (open && defaultValues) {
      reset({ sort: 0, status: 1, plan_ids: '', ...defaultValues })
    }
  }, [open, defaultValues, reset])

  return (
    <Dialog open={open} onOpenChange={(v) => { onOpenChange(v); if (!v) reset() }}>
      <DialogContent>
        <DialogHeader><DialogTitle>{title}</DialogTitle></DialogHeader>
        <form onSubmit={handleSubmit((d) => { onSubmit(d); reset() })} className="space-y-4">
          <div className="space-y-2">
            <Label>分组名称 <span className="text-destructive">*</span></Label>
            <Input placeholder="如：香港节点组" {...register('name')} />
            {errors.name && <p className="text-xs text-destructive">{errors.name.message}</p>}
          </div>
          <div className="space-y-2">
            <Label>描述</Label>
            <Textarea placeholder="分组描述" rows={2} {...register('description')} />
          </div>
          <div className="space-y-2">
            <Label>关联套餐ID</Label>
            <Input placeholder="多个用逗号分隔，如：1,2,3" {...register('plan_ids')} />
          </div>
          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-2">
              <Label>排序</Label>
              <Input type="number" placeholder="0" {...register('sort')} />
            </div>
            <div className="space-y-2">
              <Label>状态</Label>
              <div className="flex items-center gap-2 pt-1">
                <Switch checked={statusVal === 1} onCheckedChange={(checked) => setValue('status', checked ? 1 : 0)} />
                <span className="text-sm text-muted-foreground">{statusVal === 1 ? '启用' : '禁用'}</span>
              </div>
            </div>
          </div>
          <DialogFooter>
            <Button type="button" variant="outline" onClick={() => onOpenChange(false)}>取消</Button>
            <Button type="submit">保存</Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}

function GroupsTab() {
  const [page, setPage] = useState(1)
  const [dialog, setDialog] = useState<'create' | 'edit' | null>(null)
  const [editGroup, setEditGroup] = useState<ServerGroup | null>(null)
  const [deleteId, setDeleteId] = useState<number | null>(null)

  const { data, isLoading } = useServerGroups(page)
  const createGroup = useCreateServerGroup()
  const updateGroup = useUpdateServerGroup()
  const deleteGroup = useDeleteServerGroup()

  const groups = data?.list ?? []
  const totalPages = data ? Math.ceil(data.total / (data.page_size || 50)) : 1

  return (
    <div className="space-y-4">
      <div className="flex justify-end">
        <Button onClick={() => setDialog('create')}><Plus className="mr-2 h-4 w-4" />新建分组</Button>
      </div>
      {isLoading ? (
        <div className="space-y-3">{Array.from({ length: 4 }).map((_, i) => <Skeleton key={i} className="h-14" />)}</div>
      ) : groups.length === 0 ? (
        <div className="flex flex-col items-center justify-center py-16 text-muted-foreground">
          <Network className="mb-3 h-10 w-10" /><p className="text-sm">暂无分组</p>
        </div>
      ) : (
        <>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>名称</TableHead>
                <TableHead>描述</TableHead>
                <TableHead>排序</TableHead>
                <TableHead>状态</TableHead>
                <TableHead className="text-right">操作</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {groups.map((g) => (
                <TableRow key={g.id}>
                  <TableCell className="font-medium">{g.name}</TableCell>
                  <TableCell className="text-sm text-muted-foreground max-w-[300px] truncate">{g.description || '-'}</TableCell>
                  <TableCell>{g.sort}</TableCell>
                  <TableCell>
                    <Switch checked={g.status === 1} onCheckedChange={() => updateGroup.mutate({ id: g.id, status: g.status === 1 ? 0 : 1 })} />
                  </TableCell>
                  <TableCell className="text-right">
                    <div className="flex items-center justify-end gap-1">
                      <Button variant="ghost" size="icon" className="h-8 w-8" onClick={() => { setEditGroup(g); setDialog('edit') }}><Pencil className="h-3.5 w-3.5" /></Button>
                      <Button variant="ghost" size="icon" className="h-8 w-8 hover:text-destructive" onClick={() => setDeleteId(g.id)}><Trash2 className="h-3.5 w-3.5" /></Button>
                    </div>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
          <Pagination page={page} totalPages={totalPages} total={data?.total ?? 0} onPageChange={setPage} />
        </>
      )}
      <GroupFormDialog open={dialog === 'create'} onOpenChange={() => setDialog(null)} onSubmit={(d) => createGroup.mutate(d, { onSuccess: () => { toast.success('分组创建成功'); setDialog(null) } })} title="新建分组" />
      <GroupFormDialog open={dialog === 'edit' && !!editGroup} onOpenChange={(v) => { if (!v) { setDialog(null); setEditGroup(null) } }} onSubmit={(d) => editGroup && updateGroup.mutate({ id: editGroup.id, ...d }, { onSuccess: () => { toast.success('分组更新成功'); setDialog(null); setEditGroup(null) } })} defaultValues={editGroup ? { name: editGroup.name, description: editGroup.description, plan_ids: editGroup.plan_ids, sort: editGroup.sort, status: editGroup.status } : {}} title="编辑分组" />
      <ConfirmDialog open={deleteId !== null} onOpenChange={() => setDeleteId(null)} title="确认删除" description="删除分组后，关联的节点将变为无分组状态。" onConfirm={() => deleteId !== null && deleteGroup.mutate(deleteId, { onSuccess: () => { setDeleteId(null); toast.success('分组已删除') } })} loading={deleteGroup.isPending} />
    </div>
  )
}

// ─── Tab 3: Server Routes ─────────────────────────────
const routeSchema = z.object({
  name: z.string().min(1, '请输入名称'),
  group_id: z.coerce.number().min(1, '请选择分组'),
  match: z.string().min(1, '请输入匹配规则'),
  action: z.string().min(1, '请选择动作'),
  target: z.string().optional(),
  sort: z.coerce.number().optional(),
  status: z.number().optional(),
})
type RouteForm = z.infer<typeof routeSchema>

function RouteFormDialog({ open, onOpenChange, onSubmit, defaultValues, title, groups }: {
  open: boolean; onOpenChange: (v: boolean) => void; onSubmit: (data: RouteForm) => void;
  defaultValues?: Partial<RouteForm>; title: string; groups: ServerGroup[]
}) {
  const { register, handleSubmit, reset, setValue, watch, formState: { errors } } = useForm<RouteForm>({
    resolver: zodResolver(routeSchema), defaultValues: { sort: 0, action: 'proxy', status: 1, ...defaultValues },
  })
  const actionVal = watch('action')
  const groupVal = watch('group_id')
  const statusVal = watch('status')

  useEffect(() => {
    if (open && defaultValues) {
      reset({ sort: 0, action: 'proxy', status: 1, ...defaultValues })
    }
  }, [open, defaultValues, reset])

  return (
    <Dialog open={open} onOpenChange={(v) => { onOpenChange(v); if (!v) reset() }}>
      <DialogContent className="max-w-lg">
        <DialogHeader><DialogTitle>{title}</DialogTitle></DialogHeader>
        <form onSubmit={handleSubmit((d) => { onSubmit(d); reset() })} className="space-y-4">
          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-2">
              <Label>路由名称 <span className="text-destructive">*</span></Label>
              <Input placeholder="如：直连国内" {...register('name')} />
            </div>
            <div className="space-y-2">
              <Label>所属分组 <span className="text-destructive">*</span></Label>
              <Select value={String(groupVal ?? '')} onValueChange={(v) => setValue('group_id', Number(v))}>
                <SelectTrigger><SelectValue placeholder="选择分组" /></SelectTrigger>
                <SelectContent>
                  {groups.map((g) => <SelectItem key={g.id} value={String(g.id)}>{g.name}</SelectItem>)}
                </SelectContent>
              </Select>
            </div>
          </div>
          <div className="grid grid-cols-3 gap-4">
            <div className="space-y-2">
              <Label>动作 <span className="text-destructive">*</span></Label>
              <Select value={actionVal} onValueChange={(v) => setValue('action', v)}>
                <SelectTrigger><SelectValue /></SelectTrigger>
                <SelectContent>
                  <SelectItem value="proxy"><Badge variant="default" className="mr-2">proxy</Badge>代理</SelectItem>
                  <SelectItem value="direct"><Badge variant="success" className="mr-2">direct</Badge>直连</SelectItem>
                  <SelectItem value="reject"><Badge variant="destructive" className="mr-2">reject</Badge>拒绝</SelectItem>
                </SelectContent>
              </Select>
            </div>
            <div className="space-y-2">
              <Label>排序</Label>
              <Input type="number" placeholder="0" {...register('sort')} />
            </div>
            <div className="space-y-2">
              <Label>状态</Label>
              <div className="flex items-center gap-2 pt-1">
                <Switch checked={statusVal === 1} onCheckedChange={(checked) => setValue('status', checked ? 1 : 0)} />
                <span className="text-sm text-muted-foreground">{statusVal === 1 ? '启用' : '禁用'}</span>
              </div>
            </div>
          </div>
          <div className="space-y-2">
            <Label>匹配规则 <span className="text-destructive">*</span></Label>
            <Textarea placeholder='JSON 格式，如：{"domain_suffix":[".cn"]}' rows={3} {...register('match')} />
          </div>
          <div className="space-y-2">
            <Label>目标地址</Label>
            <Input placeholder="目标地址（direct/reject 时可留空）" {...register('target')} />
          </div>
          <DialogFooter>
            <Button type="button" variant="outline" onClick={() => onOpenChange(false)}>取消</Button>
            <Button type="submit">保存</Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}

function RoutesTab() {
  const [page, setPage] = useState(1)
  const [dialog, setDialog] = useState<'create' | 'edit' | null>(null)
  const [editRoute, setEditRoute] = useState<ServerRoute | null>(null)
  const [deleteId, setDeleteId] = useState<number | null>(null)

  const { data, isLoading } = useServerRoutes(page)
  const { data: groupsData } = useServerGroups()
  const createRoute = useCreateServerRoute()
  const updateRoute = useUpdateServerRoute()
  const deleteRoute = useDeleteServerRoute()

  const routes = data?.list ?? []
  const groups = groupsData?.list ?? []
  const groupMap = useMemo(() => Object.fromEntries(groups.map((g) => [g.id, g.name])), [groups])
  const totalPages = data ? Math.ceil(data.total / (data.page_size || 50)) : 1

  const actionBadge = (action: string) => {
    const map: Record<string, 'default' | 'success' | 'destructive'> = { proxy: 'default', direct: 'success', reject: 'destructive' }
    return <Badge variant={map[action] ?? 'secondary'}>{action}</Badge>
  }

  return (
    <div className="space-y-4">
      <div className="flex justify-end">
        <Button onClick={() => setDialog('create')}><Plus className="mr-2 h-4 w-4" />新建路由</Button>
      </div>
      {isLoading ? (
        <div className="space-y-3">{Array.from({ length: 4 }).map((_, i) => <Skeleton key={i} className="h-14" />)}</div>
      ) : routes.length === 0 ? (
        <div className="flex flex-col items-center justify-center py-16 text-muted-foreground">
          <Route className="mb-3 h-10 w-10" /><p className="text-sm">暂无路由规则</p>
        </div>
      ) : (
        <>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>名称</TableHead>
                <TableHead>分组</TableHead>
                <TableHead>动作</TableHead>
                <TableHead>匹配规则</TableHead>
                <TableHead>目标</TableHead>
                <TableHead>排序</TableHead>
                <TableHead>状态</TableHead>
                <TableHead className="text-right">操作</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {routes.map((r) => (
                <TableRow key={r.id}>
                  <TableCell className="font-medium">{r.name}</TableCell>
                  <TableCell className="text-sm">{groupMap[r.group_id] || `#${r.group_id}`}</TableCell>
                  <TableCell>{actionBadge(r.action)}</TableCell>
                  <TableCell className="font-mono text-xs max-w-[200px] truncate">{r.match}</TableCell>
                  <TableCell className="text-xs text-muted-foreground">{r.target || '-'}</TableCell>
                  <TableCell>{r.sort}</TableCell>
                  <TableCell>
                    <Switch checked={r.status === 1} onCheckedChange={() => updateRoute.mutate({ id: r.id, status: r.status === 1 ? 0 : 1 })} />
                  </TableCell>
                  <TableCell className="text-right">
                    <div className="flex items-center justify-end gap-1">
                      <Button variant="ghost" size="icon" className="h-8 w-8" onClick={() => { setEditRoute(r); setDialog('edit') }}><Pencil className="h-3.5 w-3.5" /></Button>
                      <Button variant="ghost" size="icon" className="h-8 w-8 hover:text-destructive" onClick={() => setDeleteId(r.id)}><Trash2 className="h-3.5 w-3.5" /></Button>
                    </div>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
          <Pagination page={page} totalPages={totalPages} total={data?.total ?? 0} onPageChange={setPage} />
        </>
      )}
      <RouteFormDialog open={dialog === 'create'} onOpenChange={() => setDialog(null)} onSubmit={(d) => createRoute.mutate(d, { onSuccess: () => { toast.success('路由创建成功'); setDialog(null) } })} title="新建路由" groups={groups} />
      <RouteFormDialog open={dialog === 'edit' && !!editRoute} onOpenChange={(v) => { if (!v) { setDialog(null); setEditRoute(null) } }} onSubmit={(d) => editRoute && updateRoute.mutate({ id: editRoute.id, ...d }, { onSuccess: () => { toast.success('路由更新成功'); setDialog(null); setEditRoute(null) } })} defaultValues={editRoute ? { name: editRoute.name, group_id: editRoute.group_id, match: editRoute.match, action: editRoute.action, target: editRoute.target, sort: editRoute.sort, status: editRoute.status } : {}} title="编辑路由" groups={groups} />
      <ConfirmDialog open={deleteId !== null} onOpenChange={() => setDeleteId(null)} title="确认删除" description="删除后路由规则将无法恢复。" onConfirm={() => deleteId !== null && deleteRoute.mutate(deleteId, { onSuccess: () => { setDeleteId(null); toast.success('路由已删除') } })} loading={deleteRoute.isPending} />
    </div>
  )
}

// ─── Tab 4: Server Machines ────────────────────────────
const serverSchema = z.object({
  name: z.string().min(1, '请输入名称'),
  host: z.string().min(1, '请输入地址'),
  port: z.coerce.number().min(1),
  protocol: z.string().optional(),
})
type ServerForm = z.infer<typeof serverSchema>

function ServerFormDialog({ open, onOpenChange, onSubmit, defaultValues, title, showToken }: {
  open: boolean; onOpenChange: (v: boolean) => void; onSubmit: (data: ServerForm) => void;
  defaultValues?: Partial<ServerForm>; title: string; showToken?: string
}) {
  const { register, handleSubmit, reset, formState: { errors } } = useForm<ServerForm>({
    resolver: zodResolver(serverSchema), defaultValues: { port: 443, ...defaultValues },
  })
  return (
    <Dialog open={open} onOpenChange={(v) => { onOpenChange(v); if (!v) reset() }}>
      <DialogContent>
        <DialogHeader><DialogTitle>{title}</DialogTitle></DialogHeader>
        <form onSubmit={handleSubmit((d) => { onSubmit(d); reset() })} className="space-y-4">
          <div className="space-y-2">
            <Label>名称 <span className="text-destructive">*</span></Label>
            <Input placeholder="如：香港-01" {...register('name')} />
          </div>
          <div className="grid grid-cols-3 gap-4">
            <div className="col-span-2 space-y-2">
              <Label>地址 <span className="text-destructive">*</span></Label>
              <Input placeholder="192.168.1.1" {...register('host')} />
            </div>
            <div className="space-y-2">
              <Label>端口 <span className="text-destructive">*</span></Label>
              <Input type="number" {...register('port')} />
            </div>
          </div>
          <div className="space-y-2">
            <Label>协议</Label>
            <Input placeholder="如：vmess" {...register('protocol')} />
          </div>
          {showToken && (
            <div className="space-y-2">
              <Label>Token</Label>
              <div className="flex items-center gap-2">
                <Input readOnly value={showToken} className="font-mono text-xs" />
                <Button type="button" variant="outline" size="icon" className="shrink-0" onClick={() => { navigator.clipboard.writeText(showToken); toast.success('Token 已复制') }}>
                  <Copy className="h-4 w-4" />
                </Button>
              </div>
            </div>
          )}
          <DialogFooter>
            <Button type="button" variant="outline" onClick={() => onOpenChange(false)}>取消</Button>
            <Button type="submit">保存</Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}

function MachinesTab() {
  const [page, setPage] = useState(1)
  const [dialog, setDialog] = useState<'create' | 'edit' | null>(null)
  const [editServer, setEditServer] = useState<ServerMachine | null>(null)
  const [deleteId, setDeleteId] = useState<number | null>(null)
  const [resetTokenId, setResetTokenId] = useState<number | null>(null)
  const [installCmd, setInstallCmd] = useState<{ server: ServerMachine; cmd: string } | null>(null)
  const [installCmdLoading, setInstallCmdLoading] = useState(false)

  const { data, isLoading } = useServers(page)
  const createServer = useCreateServer()
  const updateServer = useUpdateServer()
  const updateStatus = useUpdateServerStatus()
  const deleteServer = useDeleteServer()
  const resetToken = useResetServerToken()

  const servers = data?.list ?? []
  const totalPages = data ? Math.ceil(data.total / (data.page_size || 50)) : 1

  const maskedToken = (token?: string) => {
    if (!token) return '-'
    if (token.length <= 12) return 'xboard-****'
    return `xboard-${token.slice(8, 12)}...${token.slice(-4)}`
  }

  const handleShowInstallCmd = async (server: ServerMachine) => {
    setInstallCmdLoading(true)
    try {
      const cmd = await getInstallCommand(server.id)
      setInstallCmd({ server, cmd })
    } catch {
      // Fallback: construct from server data
      const cmd = `curl -fsSL https://get.xboard.dev/install.sh | bash -s -- --panel-url http://${server.host}:${server.port} --api-key ${server.token || '<token>'}`
      setInstallCmd({ server, cmd })
    } finally {
      setInstallCmdLoading(false)
    }
  }

  return (
    <div className="space-y-4">
      <div className="flex justify-end">
        <Button onClick={() => setDialog('create')}><Plus className="mr-2 h-4 w-4" />新建服务器</Button>
      </div>
      {isLoading ? (
        <div className="space-y-3">{Array.from({ length: 4 }).map((_, i) => <Skeleton key={i} className="h-14" />)}</div>
      ) : servers.length === 0 ? (
        <div className="flex flex-col items-center justify-center py-16 text-muted-foreground">
          <Monitor className="mb-3 h-10 w-10" /><p className="text-sm">暂无服务器</p>
        </div>
      ) : (
        <>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>名称</TableHead>
                <TableHead>地址</TableHead>
                <TableHead>端口</TableHead>
                <TableHead>协议</TableHead>
                <TableHead>Token</TableHead>
                <TableHead>状态</TableHead>
                <TableHead className="w-[120px]">CPU</TableHead>
                <TableHead className="w-[120px]">内存</TableHead>
                <TableHead className="w-[120px]">磁盘</TableHead>
                <TableHead>运行时间</TableHead>
                <TableHead className="text-right">操作</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {servers.map((s) => (
                <TableRow key={s.id}>
                  <TableCell className="font-medium">{s.name}</TableCell>
                  <TableCell className="font-mono text-xs">{s.host}</TableCell>
                  <TableCell>{s.port}</TableCell>
                  <TableCell className="text-xs">{s.protocol || '-'}</TableCell>
                  <TableCell>
                    <div className="flex items-center gap-1.5">
                      <Key className="h-3 w-3 text-muted-foreground" />
                      <span className="font-mono text-xs">{maskedToken(s.token)}</span>
                      {s.token && (
                        <Tooltip>
                          <TooltipTrigger asChild>
                            <Button variant="ghost" size="icon" className="h-5 w-5" onClick={() => { navigator.clipboard.writeText(s.token!); toast.success('Token 已复制') }}>
                              <Copy className="h-3 w-3" />
                            </Button>
                          </TooltipTrigger>
                          <TooltipContent>复制完整 Token</TooltipContent>
                        </Tooltip>
                      )}
                    </div>
                  </TableCell>
                  <TableCell>
                    <Badge variant={s.status === 1 ? 'success' : 'destructive'}>{s.status === 1 ? '在线' : '离线'}</Badge>
                  </TableCell>
                  <TableCell>
                    <div className="flex items-center gap-2">
                      <Progress value={s.cpu} className="h-2 flex-1" />
                      <span className="text-xs text-muted-foreground w-10 text-right">{s.cpu?.toFixed(1) ?? 0}%</span>
                    </div>
                  </TableCell>
                  <TableCell>
                    <div className="flex items-center gap-2">
                      <Progress value={s.memory} className="h-2 flex-1" />
                      <span className="text-xs text-muted-foreground w-10 text-right">{s.memory?.toFixed(1) ?? 0}%</span>
                    </div>
                  </TableCell>
                  <TableCell>
                    <div className="flex items-center gap-2">
                      <Progress value={s.disk} className="h-2 flex-1" />
                      <span className="text-xs text-muted-foreground w-10 text-right">{s.disk?.toFixed(1) ?? 0}%</span>
                    </div>
                  </TableCell>
                  <TableCell className="text-xs text-muted-foreground">
                    <span className="inline-flex items-center gap-1"><Clock className="h-3 w-3" />{formatUptime(s.uptime)}</span>
                  </TableCell>
                  <TableCell className="text-right">
                    <div className="flex items-center justify-end gap-1">
                      <Tooltip>
                        <TooltipTrigger asChild>
                          <Button variant="ghost" size="icon" className="h-8 w-8" onClick={() => handleShowInstallCmd(s)} disabled={installCmdLoading}>
                            <Terminal className="h-3.5 w-3.5" />
                          </Button>
                        </TooltipTrigger>
                        <TooltipContent>安装命令</TooltipContent>
                      </Tooltip>
                      <Tooltip>
                        <TooltipTrigger asChild>
                          <Button variant="ghost" size="icon" className="h-8 w-8" onClick={() => setResetTokenId(s.id)}>
                            <RotateCcw className="h-3.5 w-3.5" />
                          </Button>
                        </TooltipTrigger>
                        <TooltipContent>重置 Token</TooltipContent>
                      </Tooltip>
                      <Switch checked={s.status === 1} onCheckedChange={() => updateStatus.mutate({ id: s.id, status: s.status === 1 ? 0 : 1 })} />
                      <Button variant="ghost" size="icon" className="h-8 w-8" onClick={() => { setEditServer(s); setDialog('edit') }}><Pencil className="h-3.5 w-3.5" /></Button>
                      <Button variant="ghost" size="icon" className="h-8 w-8 hover:text-destructive" onClick={() => setDeleteId(s.id)}><Trash2 className="h-3.5 w-3.5" /></Button>
                    </div>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
          <Pagination page={page} totalPages={totalPages} total={data?.total ?? 0} onPageChange={setPage} />
        </>
      )}
      <ServerFormDialog open={dialog === 'create'} onOpenChange={() => setDialog(null)} onSubmit={(d) => createServer.mutate(d, { onSuccess: () => { toast.success('服务器创建成功'); setDialog(null) } })} title="新建服务器" />
      <ServerFormDialog open={dialog === 'edit' && !!editServer} onOpenChange={(v) => { if (!v) { setDialog(null); setEditServer(null) } }} onSubmit={(d) => editServer && updateServer.mutate({ id: editServer.id, ...d }, { onSuccess: () => { toast.success('服务器更新成功'); setDialog(null); setEditServer(null) } })} defaultValues={editServer ? { name: editServer.name, host: editServer.host, port: editServer.port, protocol: editServer.protocol } : {}} title="编辑服务器" showToken={editServer?.token} />
      <ConfirmDialog open={deleteId !== null} onOpenChange={() => setDeleteId(null)} title="确认删除" description="删除后服务器信息将无法恢复。" onConfirm={() => deleteId !== null && deleteServer.mutate(deleteId, { onSuccess: () => { setDeleteId(null); toast.success('服务器已删除') } })} loading={deleteServer.isPending} />
      <ConfirmDialog open={resetTokenId !== null} onOpenChange={() => setResetTokenId(null)} title="重置 Token" description="重置后旧 Token 将立即失效，服务器需要重新配置。确认继续？" onConfirm={() => resetTokenId !== null && resetToken.mutate(resetTokenId, { onSuccess: () => { setResetTokenId(null) } })} loading={resetToken.isPending} />
      {/* Install Command Dialog */}
      <Dialog open={installCmd !== null} onOpenChange={(v) => { if (!v) setInstallCmd(null) }}>
        <DialogContent className="max-w-2xl">
          <DialogHeader>
            <DialogTitle className="flex items-center gap-2">
              <Terminal className="h-4 w-4" />安装命令 — {installCmd?.server.name}
            </DialogTitle>
            <DialogDescription>在目标服务器上执行以下命令完成安装</DialogDescription>
          </DialogHeader>
          <div className="relative rounded-md border bg-muted/50 p-4 max-h-[300px] overflow-auto">
            <pre className="text-xs font-mono whitespace-pre-wrap break-all">{installCmd?.cmd}</pre>
            <Button variant="outline" size="sm" className="absolute top-2 right-2" onClick={() => { if (installCmd?.cmd) { navigator.clipboard.writeText(installCmd.cmd); toast.success('命令已复制') } }}>
              <Copy className="mr-1.5 h-3.5 w-3.5" />复制
            </Button>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setInstallCmd(null)}>关闭</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}

// ─── Tab 5: Monitor ────────────────────────────────────
function MonitorTab() {
  const [selectedNodeId, setSelectedNodeId] = useState<number | null>(null)
  const [metrics, setMetrics] = useState<Record<string, unknown> | null>(null)
  const [loading, setLoading] = useState(false)

  const { data: allNodesData } = useAdminNodes(1, 1000)
  const allNodes = allNodesData?.list ?? []

  const fetchMetrics = useCallback(async (nodeId: number) => {
    try {
      const res = await api.get(`/admin/nodes/${nodeId}/metrics`)
      setMetrics(res as unknown as Record<string, unknown>)
    } catch {
      // error handled by api interceptor
    }
  }, [])

  useEffect(() => {
    if (!selectedNodeId) return
    setLoading(true)
    fetchMetrics(selectedNodeId).finally(() => setLoading(false))

    const interval = setInterval(() => fetchMetrics(selectedNodeId), 10000)
    return () => clearInterval(interval)
  }, [selectedNodeId, fetchMetrics])

  const cpu = (metrics?.cpu as number) ?? 0
  const memory = (metrics?.memory as number) ?? 0
  const disk = (metrics?.disk as number) ?? 0
  const memoryUsed = (metrics?.memory_used as number) ?? 0
  const memoryTotal = (metrics?.memory_total as number) ?? 0
  const diskUsed = (metrics?.disk_used as number) ?? 0
  const diskTotal = (metrics?.disk_total as number) ?? 0
  const connections = (metrics?.connections as number) ?? 0
  const uploadTotal = (metrics?.upload_total as number) ?? 0
  const downloadTotal = (metrics?.download_total as number) ?? 0
  const onlineUsers = (metrics?.online_users as number) ?? 0
  const uptime = (metrics?.uptime as number) ?? 0

  return (
    <div className="space-y-4">
      <div className="flex items-center gap-3">
        <Label className="whitespace-nowrap">选择节点</Label>
        <Select value={selectedNodeId ? String(selectedNodeId) : ''} onValueChange={(v) => setSelectedNodeId(Number(v))}>
          <SelectTrigger className="w-[280px]"><SelectValue placeholder="请选择要监控的节点" /></SelectTrigger>
          <SelectContent>
            {allNodes.map((n) => <SelectItem key={n.id} value={String(n.id)}>{n.name} ({n.address})</SelectItem>)}
          </SelectContent>
        </Select>
        {selectedNodeId && (
          <Badge variant="outline" className="gap-1">
            <Activity className="h-3 w-3 animate-pulse" /> 实时监控 · 10s 刷新
          </Badge>
        )}
      </div>

      {!selectedNodeId ? (
        <div className="flex flex-col items-center justify-center py-20 text-muted-foreground">
          <Activity className="mb-3 h-10 w-10" />
          <p className="text-sm">请先选择一个节点以查看监控数据</p>
        </div>
      ) : loading && !metrics ? (
        <div className="grid gap-4 grid-cols-1 md:grid-cols-2 lg:grid-cols-3">
          {Array.from({ length: 6 }).map((_, i) => <Skeleton key={i} className="h-32" />)}
        </div>
      ) : (
        <>
          <div className="grid gap-4 grid-cols-1 md:grid-cols-3">
            <Card>
              <CardHeader className="pb-2"><CardTitle className="text-sm font-medium text-muted-foreground">CPU 使用率</CardTitle></CardHeader>
              <CardContent>
                <div className="flex items-end justify-between mb-2">
                  <span className="text-3xl font-bold">{cpu.toFixed(1)}%</span>
                </div>
                <Progress value={cpu} className="h-2" />
              </CardContent>
            </Card>
            <Card>
              <CardHeader className="pb-2"><CardTitle className="text-sm font-medium text-muted-foreground">内存使用率</CardTitle></CardHeader>
              <CardContent>
                <div className="flex items-end justify-between mb-2">
                  <span className="text-3xl font-bold">{memory.toFixed(1)}%</span>
                  <span className="text-xs text-muted-foreground">{formatBytes(memoryUsed)} / {formatBytes(memoryTotal)}</span>
                </div>
                <Progress value={memory} className="h-2" />
              </CardContent>
            </Card>
            <Card>
              <CardHeader className="pb-2"><CardTitle className="text-sm font-medium text-muted-foreground">磁盘使用率</CardTitle></CardHeader>
              <CardContent>
                <div className="flex items-end justify-between mb-2">
                  <span className="text-3xl font-bold">{disk.toFixed(1)}%</span>
                  <span className="text-xs text-muted-foreground">{formatBytes(diskUsed)} / {formatBytes(diskTotal)}</span>
                </div>
                <Progress value={disk} className="h-2" />
              </CardContent>
            </Card>
          </div>

          <div className="grid gap-4 grid-cols-2 md:grid-cols-4">
            <Card>
              <CardContent className="pt-6">
                <div className="flex items-center gap-3">
                  <div className="rounded-lg p-2 bg-blue-500/10 text-blue-500"><Zap className="h-5 w-5" /></div>
                  <div>
                    <p className="text-2xl font-bold">{connections}</p>
                    <p className="text-xs text-muted-foreground">活跃连接数</p>
                  </div>
                </div>
              </CardContent>
            </Card>
            <Card>
              <CardContent className="pt-6">
                <div className="flex items-center gap-3">
                  <div className="rounded-lg p-2 bg-emerald-500/10 text-emerald-500"><Globe className="h-5 w-5" /></div>
                  <div>
                    <p className="text-2xl font-bold">{onlineUsers}</p>
                    <p className="text-xs text-muted-foreground">在线用户</p>
                  </div>
                </div>
              </CardContent>
            </Card>
            <Card>
              <CardContent className="pt-6">
                <div className="flex items-center gap-3">
                  <div className="rounded-lg p-2 bg-violet-500/10 text-violet-500"><Wifi className="h-5 w-5" /></div>
                  <div>
                    <p className="text-2xl font-bold">{formatBytes(uploadTotal)}</p>
                    <p className="text-xs text-muted-foreground">总上传流量</p>
                  </div>
                </div>
              </CardContent>
            </Card>
            <Card>
              <CardContent className="pt-6">
                <div className="flex items-center gap-3">
                  <div className="rounded-lg p-2 bg-amber-500/10 text-amber-500"><WifiOff className="h-5 w-5" /></div>
                  <div>
                    <p className="text-2xl font-bold">{formatBytes(downloadTotal)}</p>
                    <p className="text-xs text-muted-foreground">总下载流量</p>
                  </div>
                </div>
              </CardContent>
            </Card>
          </div>

          {uptime > 0 && (
            <Card>
              <CardContent className="pt-6">
                <div className="flex items-center gap-3">
                  <div className="rounded-lg p-2 bg-primary/10 text-primary"><Clock className="h-5 w-5" /></div>
                  <div>
                    <p className="text-lg font-bold">{formatUptime(uptime)}</p>
                    <p className="text-xs text-muted-foreground">节点运行时间</p>
                  </div>
                </div>
              </CardContent>
            </Card>
          )}
        </>
      )}
    </div>
  )
}

// ─── Tab 6: Templates ──────────────────────────────────
const templateSchema = z.object({
  name: z.string().min(1, '请输入名称'),
  type: z.string().min(1, '请选择类型'),
  server_info: z.string().optional(),
  description: z.string().optional(),
})
type TemplateForm = z.infer<typeof templateSchema>

function TemplateFormDialog({ open, onOpenChange, onSubmit, defaultValues, title }: {
  open: boolean; onOpenChange: (v: boolean) => void; onSubmit: (data: TemplateForm) => void;
  defaultValues?: Partial<TemplateForm>; title: string
}) {
  const { register, handleSubmit, reset, setValue, watch, formState: { errors } } = useForm<TemplateForm>({
    resolver: zodResolver(templateSchema), defaultValues: { server_info: '{}', ...defaultValues },
  })
  const typeVal = watch('type')

  useEffect(() => {
    if (open && defaultValues) {
      reset({ server_info: '{}', ...defaultValues })
    }
  }, [open, defaultValues, reset])

  return (
    <Dialog open={open} onOpenChange={(v) => { onOpenChange(v); if (!v) reset() }}>
      <DialogContent className="max-w-lg">
        <DialogHeader><DialogTitle>{title}</DialogTitle></DialogHeader>
        <form onSubmit={handleSubmit((d) => { onSubmit(d); reset() })} className="space-y-4">
          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-2">
              <Label>模板名称 <span className="text-destructive">*</span></Label>
              <Input placeholder="如：香港标准模板" {...register('name')} />
              {errors.name && <p className="text-xs text-destructive">{errors.name.message}</p>}
            </div>
            <div className="space-y-2">
              <Label>协议类型 <span className="text-destructive">*</span></Label>
              <Select value={typeVal} onValueChange={(v) => setValue('type', v, { shouldValidate: true })}>
                <SelectTrigger><SelectValue placeholder="选择协议" /></SelectTrigger>
                <SelectContent>
                  {nodeTypes.map((t) => <SelectItem key={t.value} value={t.value}>{t.label}</SelectItem>)}
                </SelectContent>
              </Select>
              {errors.type && <p className="text-xs text-destructive">{errors.type.message}</p>}
            </div>
          </div>
          <div className="space-y-2">
            <Label>节点配置 (server_info)</Label>
            <Textarea placeholder='JSON 格式，如：{"network":"ws","path":"/ws"}' rows={4} {...register('server_info')} />
          </div>
          <div className="space-y-2">
            <Label>描述</Label>
            <Textarea placeholder="模板说明" rows={2} {...register('description')} />
          </div>
          <DialogFooter>
            <Button type="button" variant="outline" onClick={() => onOpenChange(false)}>取消</Button>
            <Button type="submit">保存</Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}

function TemplatesTab() {
  const [page, setPage] = useState(1)
  const [dialog, setDialog] = useState<'create' | 'edit' | null>(null)
  const [editTemplate, setEditTemplate] = useState<NodeTemplate | null>(null)
  const [deleteId, setDeleteId] = useState<number | null>(null)

  const { data, isLoading } = useNodeTemplates(page)
  const createTemplate = useCreateNodeTemplate()
  const updateTemplate = useUpdateNodeTemplate()
  const deleteTemplate = useDeleteNodeTemplate()

  const templates = data?.list ?? []
  const totalPages = data ? Math.ceil(data.total / (data.page_size || 50)) : 1

  return (
    <div className="space-y-4">
      <div className="flex justify-end">
        <Button onClick={() => setDialog('create')}><Plus className="mr-2 h-4 w-4" />新建模板</Button>
      </div>
      {isLoading ? (
        <div className="space-y-3">{Array.from({ length: 4 }).map((_, i) => <Skeleton key={i} className="h-14" />)}</div>
      ) : templates.length === 0 ? (
        <div className="flex flex-col items-center justify-center py-16 text-muted-foreground">
          <FileCode className="mb-3 h-10 w-10" /><p className="text-sm">暂无节点模板</p>
        </div>
      ) : (
        <>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>名称</TableHead>
                <TableHead>类型</TableHead>
                <TableHead>描述</TableHead>
                <TableHead>配置预览</TableHead>
                <TableHead>创建时间</TableHead>
                <TableHead className="text-right">操作</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {templates.map((t) => {
                let infoPreview = '-'
                try {
                  const parsed = JSON.parse(t.server_info || '{}')
                  infoPreview = JSON.stringify(parsed).slice(0, 60)
                  if (infoPreview.length >= 60) infoPreview += '...'
                } catch { /* ignore */ }
                return (
                  <TableRow key={t.id}>
                    <TableCell className="font-medium">{t.name}</TableCell>
                    <TableCell><NodeTypeBadge type={t.type} /></TableCell>
                    <TableCell className="text-sm text-muted-foreground max-w-[200px] truncate">{t.description || '-'}</TableCell>
                    <TableCell className="font-mono text-xs max-w-[250px] truncate">{infoPreview}</TableCell>
                    <TableCell className="text-xs text-muted-foreground">{formatDate(t.created_at)}</TableCell>
                    <TableCell className="text-right">
                      <div className="flex items-center justify-end gap-1">
                        <Button variant="ghost" size="icon" className="h-8 w-8" onClick={() => { setEditTemplate(t); setDialog('edit') }}><Pencil className="h-3.5 w-3.5" /></Button>
                        <Button variant="ghost" size="icon" className="h-8 w-8 hover:text-destructive" onClick={() => setDeleteId(t.id)}><Trash2 className="h-3.5 w-3.5" /></Button>
                      </div>
                    </TableCell>
                  </TableRow>
                )
              })}
            </TableBody>
          </Table>
          <Pagination page={page} totalPages={totalPages} total={data?.total ?? 0} onPageChange={setPage} />
        </>
      )}
      <TemplateFormDialog
        open={dialog === 'create'} onOpenChange={() => setDialog(null)}
        onSubmit={(d) => {
          if (d.server_info) { try { JSON.parse(d.server_info) } catch { toast.error('server_info JSON 格式不正确'); return } }
          createTemplate.mutate(d, { onSuccess: () => setDialog(null) })
        }}
        title="新建模板"
      />
      <TemplateFormDialog
        open={dialog === 'edit' && !!editTemplate} onOpenChange={(v) => { if (!v) { setDialog(null); setEditTemplate(null) } }}
        onSubmit={(d) => {
          if (!editTemplate) return
          if (d.server_info) { try { JSON.parse(d.server_info) } catch { toast.error('server_info JSON 格式不正确'); return } }
          updateTemplate.mutate({ id: editTemplate.id, ...d }, { onSuccess: () => { setDialog(null); setEditTemplate(null) } })
        }}
        defaultValues={editTemplate ? { name: editTemplate.name, type: editTemplate.type, server_info: editTemplate.server_info || '{}', description: editTemplate.description } : {}}
        title="编辑模板"
      />
      <ConfirmDialog open={deleteId !== null} onOpenChange={() => setDeleteId(null)} title="确认删除" description="删除后模板将无法恢复。" onConfirm={() => deleteId !== null && deleteTemplate.mutate(deleteId, { onSuccess: () => setDeleteId(null) })} loading={deleteTemplate.isPending} />
    </div>
  )
}

// ─── Main Page ─────────────────────────────────────────
export default function NodesPage() {
  return (
    <TooltipProvider>
      <div className="space-y-6">
        <div>
          <h1 className="text-2xl font-bold tracking-tight">节点中心</h1>
          <p className="text-muted-foreground">管理代理节点、分组、路由和服务器</p>
        </div>

        <Tabs defaultValue="nodes" className="space-y-4">
          <TabsList>
            <TabsTrigger value="nodes" className="gap-1.5"><Server className="h-4 w-4" />节点列表</TabsTrigger>
            <TabsTrigger value="groups" className="gap-1.5"><Network className="h-4 w-4" />服务端分组</TabsTrigger>
            <TabsTrigger value="routes" className="gap-1.5"><Route className="h-4 w-4" />路由规则</TabsTrigger>
            <TabsTrigger value="machines" className="gap-1.5"><Monitor className="h-4 w-4" />服务器机器</TabsTrigger>
            <TabsTrigger value="monitor" className="gap-1.5"><Activity className="h-4 w-4" />实时监控</TabsTrigger>
            <TabsTrigger value="templates" className="gap-1.5"><FileCode className="h-4 w-4" />节点模板</TabsTrigger>
          </TabsList>

          <TabsContent value="nodes"><NodesTab /></TabsContent>
          <TabsContent value="groups"><GroupsTab /></TabsContent>
          <TabsContent value="routes"><RoutesTab /></TabsContent>
          <TabsContent value="machines"><MachinesTab /></TabsContent>
          <TabsContent value="monitor"><MonitorTab /></TabsContent>
          <TabsContent value="templates"><TemplatesTab /></TabsContent>
        </Tabs>
      </div>
    </TooltipProvider>
  )
}
