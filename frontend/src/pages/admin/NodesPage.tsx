import { useState, useEffect, useMemo, useCallback, useRef } from 'react'
import { useSearchParams } from 'react-router-dom'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { toast } from 'sonner'
import { useAdminNodes, useNodeStats, useCreateNode, useUpdateNode, useDeleteNode, useUpdateNodeStatus, useBatchUpdateNodes, useSortNodes } from '@/hooks/useNodes'
import { useServerGroups, useAllServerGroups } from '@/hooks/useServerGroups'
import { useServerRoutes } from '@/hooks/useServerRoutes'
import { useAllServers } from '@/hooks/useServers'
import { Card, CardContent } from '@/components/ui/card'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Switch } from '@/components/ui/switch'
import { Skeleton } from '@/components/ui/skeleton'
import { Checkbox } from '@/components/ui/checkbox'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from '@/components/ui/tooltip'
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuSeparator, DropdownMenuTrigger } from '@/components/ui/dropdown-menu'
import {
  Plus, Pencil, Trash2, Server, Wifi, WifiOff, Wrench,
  Copy, Search, Info, Clock,
  Activity, RotateCcw, MoreHorizontal, Users, HardDrive,
  ChevronRight, Filter, X,
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

// ─── 常量 ──────────────────────────────────────────────
const PROTOCOL_COLORS: Record<string, string> = {
  shadowsocks: '#489851',
  vmess: '#CB3180',
  trojan: '#EBB749',
  hysteria: '#5684e6',
  vless: '#1a1a1a',
  tuic: '#00C853',
  socks: '#2196F3',
  naive: '#9C27B0',
  http: '#FF5722',
  mieru: '#4CAF50',
  anytls: '#7E57C2',
  hysteria2: '#5684e6',
}

const nodeTypes = [
  { value: 'vmess', label: 'VMess' },
  { value: 'vless', label: 'VLESS' },
  { value: 'trojan', label: 'Trojan' },
  { value: 'shadowsocks', label: 'Shadowsocks' },
  { value: 'hysteria2', label: 'Hysteria2' },
  { value: 'hysteria', label: 'Hysteria' },
  { value: 'tuic', label: 'TUIC' },
  { value: 'anytls', label: 'AnyTLS' },
  { value: 'naive', label: 'Naive' },
  { value: 'socks', label: 'SOCKS' },
  { value: 'http', label: 'HTTP' },
  { value: 'mieru', label: 'Mieru' },
]
const nodeTypeMap = Object.fromEntries(nodeTypes.map((t) => [t.value, t]))

const TRANSPORT_TYPES = ['tcp', 'ws', 'grpc', 'h2', 'httpupgrade', 'xhttp']
const PROTOCOLS_WITH_SETTINGS = ['vless', 'shadowsocks', 'hysteria', 'hysteria2', 'tuic', 'anytls', 'mieru']

// ─── 工具函数 ──────────────────────────────────────────
function NodeTypeBadge({ type }: { type: string }) {
  const info = nodeTypeMap[type]
  const color = PROTOCOL_COLORS[type] ?? '#6b7280'
  return (
    <span
      className="inline-flex items-center rounded-md border-2 px-2 py-0.5 text-xs font-medium"
      style={{ borderColor: color, color }}
    >
      {info?.label ?? type}
    </span>
  )
}

function StatusDot({ status }: { status: number }) {
  // 原版: 0=离线(红) 1=警告(黄) 2=在线(绿)
  const map: Record<number, { dot: string; shadow: string; label: string }> = {
    0: { dot: 'bg-destructive/80', shadow: 'shadow-sm shadow-destructive/50', label: '离线' },
    1: { dot: 'bg-yellow-500/80', shadow: 'shadow-sm shadow-yellow-500/50', label: '警告' },
    2: { dot: 'bg-emerald-500/80', shadow: 'shadow-sm shadow-emerald-500/50', label: '在线' },
  }
  const s = map[status] ?? map[0]
  return (
    <Tooltip>
      <TooltipTrigger>
        <span className={`inline-block h-2.5 w-2.5 rounded-full ${s.dot} ${s.shadow}`} />
      </TooltipTrigger>
      <TooltipContent side="top"><p className="text-xs">{s.label}</p></TooltipContent>
    </Tooltip>
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

function parseServerInfo(raw: string): Record<string, any> {
  try { return JSON.parse(raw || '{}') } catch { return {} }
}

function getMachineStatus(machine: ServerMachine): { dot: string; label: string } {
  if (machine.is_active === false) return { dot: 'bg-slate-400', label: '已禁用' }
  if (!machine.last_seen_at) return { dot: 'bg-slate-400', label: '从未上报' }
  const diff = Date.now() / 1000 - new Date(machine.last_seen_at).getTime() / 1000
  if (diff < 300) return { dot: 'bg-emerald-500', label: '在线' }
  return { dot: 'bg-red-500', label: '离线' }
}

// ─── 节点详情弹窗 ─────────────────────────────────────
function NodeDetailDialog({ open, onOpenChange, node, groupNames }: {
  open: boolean; onOpenChange: (v: boolean) => void; node: Node | null; groupNames: string[]
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
          <DialogDescription>ID: {node.id} | 类型: {node.type} | 地址: {node.host}:{node.port}</DialogDescription>
        </DialogHeader>
        <div className="space-y-4 text-sm">
          <div className="grid grid-cols-2 gap-3">
            <div><span className="text-muted-foreground">分组:</span> {groupNames.length > 0 ? groupNames.join(', ') : '-'}</div>
            <div><span className="text-muted-foreground">流量倍率:</span> {node.rate}x</div>
            <div><span className="text-muted-foreground">在线用户:</span> {node.online_user_count ?? 0}</div>
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

// ─── 动态倍率规则 ─────────────────────────────────────
interface RateTimeRange {
  start: string
  end: string
  rate: string
}

// ─── 节点表单 ─────────────────────────────────────────
const rateRangeSchema = z.object({
  start: z.string().min(1),
  end: z.string().min(1),
  rate: z.string().min(0),
})

const nodeSchema = z.object({
  name: z.string().min(1, '请输入名称'),
  type: z.string().min(1, '请选择类型'),
  host: z.string().min(1, '请输入地址'),
  port: z.coerce.number().min(1).max(65535),
  server_info: z.string().optional(),
  group_ids: z.string().optional(),
  rate: z.coerce.number().min(0).max(100).optional(),
  rate_time_enable: z.boolean().optional().default(false),
  rate_time_ranges: z.array(rateRangeSchema).optional().default([]),
  parent_id: z.coerce.number().optional(),
  machine_id: z.coerce.number().optional(),
  listen_address: z.string().optional(),
  server_port: z.coerce.number().optional(),
  show: z.boolean().optional().default(true),
  banned: z.boolean().optional().default(false),
  traffic_limit: z.coerce.number().optional(),
  transfer_enable: z.coerce.number().optional(),
  tags: z.string().optional(),
  code: z.string().optional(),
  route_id: z.coerce.number().optional(),
})
type NodeForm = z.infer<typeof nodeSchema>

function NodeFormDialog({ open, onOpenChange, onSubmit, defaultValues, title, groups, nodes, servers, routes }: {
  open: boolean; onOpenChange: (v: boolean) => void; onSubmit: (data: NodeForm) => void;
  defaultValues?: Partial<NodeForm>; title: string;
  groups: ServerGroup[]; nodes: Node[]; servers: ServerMachine[]; routes?: ServerRoute[]
}) {
  const { register, handleSubmit, reset, setValue, watch, formState: { errors } } = useForm<NodeForm>({
    resolver: zodResolver(nodeSchema),
    defaultValues: { rate: 1, group_ids: '', parent_id: 0, server_info: '{}', ...defaultValues },
  })
  const typeVal = watch('type')
  const groupIdsVal = watch('group_ids')
  const [groupDropdownOpen, setGroupDropdownOpen] = useState(false)
  const groupDropdownRef = useRef<HTMLDivElement>(null)
  const [tags, setTags] = useState<string[]>(defaultValues?.tags ? defaultValues.tags.split(',').filter(Boolean) : [])
  const [tagInput, setTagInput] = useState('')

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
    const EXCLUDED_KEYS = ['network', 'network_settings', 'tls', 'tls_settings', 'cert_config', 'multiplex', 'ech', 'custom_outbounds', 'custom_routes', 'listen_ip', 'server_port', 'show', 'banned', 'transfer_enable_gb']
    const proto: Record<string, any> = {}
    for (const [k, v] of Object.entries(info)) {
      if (!EXCLUDED_KEYS.includes(k)) {
        proto[k] = v
      }
    }
    setProtocolSettings(proto)
  }, [])

  // Click-outside handler for group dropdown
  useEffect(() => {
    if (!groupDropdownOpen) return
    const handleClickOutside = (e: MouseEvent) => {
      if (groupDropdownRef.current && !groupDropdownRef.current.contains(e.target as Node)) {
        setGroupDropdownOpen(false)
      }
    }
    document.addEventListener('mousedown', handleClickOutside)
    return () => document.removeEventListener('mousedown', handleClickOutside)
  }, [groupDropdownOpen])

  useEffect(() => {
    if (open && defaultValues) {
      const merged = { rate: 1, group_ids: '', parent_id: 0, server_info: '{}', show: true, banned: false, rate_time_enable: false, rate_time_ranges: [], ...defaultValues }
      reset(merged)
      syncFromJson(merged.server_info ?? '{}')
      const info = parseServerInfo(merged.server_info ?? '{}')
      if (info.listen_ip) setValue('listen_address', String(info.listen_ip))
      if (info.server_port) setValue('server_port', Number(info.server_port))
      if (info.show !== undefined) setValue('show', Number(info.show) !== 0)
      if (info.banned !== undefined) setValue('banned', Number(info.banned) === 1)
      if (info.transfer_enable_gb) setValue('traffic_limit', Number(info.transfer_enable_gb))
      if (merged.transfer_enable) setValue('traffic_limit', Number(merged.transfer_enable) / 1073741824)
      if (info.machine_id) setValue('machine_id', Number(info.machine_id))
      // 动态倍率
      if (merged.rate_time_enable) setValue('rate_time_enable', true)
      if (merged.rate_time_ranges) {
        try {
          const ranges = typeof merged.rate_time_ranges === 'string' ? JSON.parse(merged.rate_time_ranges) : merged.rate_time_ranges
          setValue('rate_time_ranges', Array.isArray(ranges) ? ranges : [])
        } catch { setValue('rate_time_ranges', []) }
      }
    }
  }, [open, defaultValues, reset, syncFromJson, setValue])

  const handleFormSubmit = (d: NodeForm) => {
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
    Object.assign(serverInfo, protocolSettings)
    if (d.listen_address) serverInfo.listen_ip = d.listen_address
    if (certValue && Object.keys(certValue).length > 0) serverInfo.cert_config = certValue
    if (multiplexValue?.enabled) serverInfo.multiplex = multiplexValue
    if (echValue && Object.keys(echValue).filter(k => echValue[k as keyof EchSettingsValue]).length > 0) serverInfo.ech = echValue
    if (customOutbounds && customOutbounds !== '[]') try { serverInfo.custom_outbounds = JSON.parse(customOutbounds) } catch { /* ignore */ }
    if (customRoutes && customRoutes !== '[]') try { serverInfo.custom_routes = JSON.parse(customRoutes) } catch { /* ignore */ }
    if (d.server_port) serverInfo.server_port = d.server_port
    if (d.machine_id) serverInfo.machine_id = d.machine_id
    if (!d.show) serverInfo.show = 0
    if (d.banned) serverInfo.banned = 1
    if (d.tags) serverInfo.tags = d.tags.split(',').filter(Boolean)
    if (d.route_id) serverInfo.route_id = d.route_id
    if (d.code) serverInfo.code = d.code

    const payload = { ...d }
    payload.server_info = JSON.stringify(serverInfo)
    ;(payload as any).transfer_enable = d.traffic_limit ? Math.round(d.traffic_limit * 1073741824) : 0

    // 动态倍率
    ;(payload as any).rate_time_enable = d.rate_time_enable ? 1 : 0
    ;(payload as any).rate_time_ranges = d.rate_time_enable ? JSON.stringify(d.rate_time_ranges || []) : '[]'

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
      <DialogContent className="max-w-xl gap-0 overflow-hidden p-0 sm:rounded-2xl">
        <div className="border-b bg-muted/20 px-6 pb-4 pt-6">
          <div className="flex items-center justify-between pr-8">
            <div className="flex items-center gap-3">
              <DialogTitle className="font-mono text-lg tracking-tight">{title}</DialogTitle>
              {typeVal && (
                <span
                  className="rounded px-2 py-0.5 font-mono text-xs text-white"
                  style={{ background: PROTOCOL_COLORS[typeVal] ?? '#6b7280' }}
                >
                  {nodeTypeMap[typeVal]?.label ?? typeVal}
                </span>
              )}
            </div>
            <Select value={typeVal} onValueChange={(v) => setValue('type', v, { shouldValidate: true })}>
              <SelectTrigger className="h-8 w-[150px] border-2 font-mono text-xs">
                <SelectValue placeholder="选择协议" />
              </SelectTrigger>
              <SelectContent>
                {nodeTypes.map((t) => (
                  <SelectItem key={t.value} value={t.value} className="font-mono text-xs">
                    <span className="flex items-center gap-2">
                      <span className="h-2.5 w-2.5 rounded-full" style={{ background: PROTOCOL_COLORS[t.value] }} />
                      {t.label}
                    </span>
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
          <DialogDescription className="mt-2 font-mono text-xs opacity-70">
            配置代理节点的连接信息
          </DialogDescription>
        </div>
        <form onSubmit={handleSubmit(handleFormSubmit)} className="flex h-[75vh] min-h-[500px] flex-col">
          <div className="flex-1 space-y-8 overflow-y-auto px-6 py-6">
          {/* 名称 + 倍率 */}
          <div className="flex gap-4">
            <div className="flex-[2] space-y-2">
              <Label className="font-mono text-[12px] text-foreground/80">节点名称 <span className="text-destructive">*</span></Label>
              <Input placeholder="如：香港 IPLC-01" className="h-9 font-mono text-xs" {...register('name')} />
              {errors.name && <p className="text-xs text-destructive">{errors.name.message}</p>}
            </div>
            <div className="flex-[1] space-y-2">
              <Label className="font-mono text-[12px] text-foreground/80">流量倍率</Label>
              <div className="relative">
                <Input type="number" step="0.1" min="0" placeholder="1.0" className="h-9 pr-8 font-mono text-xs" {...register('rate')} disabled={!!watch('parent_id')} />
                <span className="absolute right-2.5 top-1/2 -translate-y-1/2 font-mono text-[10px] text-muted-foreground">x</span>
              </div>
              {!!watch('parent_id') && (
                <p className="flex items-center gap-1 text-xs text-muted-foreground">
                  <Info className="h-3 w-3" />子节点倍率继承自父节点
                </p>
              )}
            </div>
          </div>

          {/* 流量限制 + 自定义ID */}
          <div className="flex gap-3">
            <div className="flex-1 space-y-1">
              <Label className="font-mono text-[11px] text-muted-foreground">流量限制 <span className="ml-1 text-[9px]">(GB)</span></Label>
              <Input type="number" min="0" step="1" placeholder="不限" className="h-8 font-mono text-xs" {...register('traffic_limit')} />
            </div>
            <div className="flex-1 space-y-1">
              <Label className="font-mono text-[11px] text-muted-foreground">自定义节点ID <span className="ml-1 text-[9px]">(选填)</span></Label>
              <Input placeholder="留空自动生成" className="h-8 font-mono text-xs" {...register('code')} />
            </div>
          </div>

          {/* 标签 */}
          <div className="space-y-2">
            <Label className="font-mono text-[12px] text-foreground/80">节点标签</Label>
            <div className="flex flex-wrap gap-1.5 mb-2">
              {tags.map((tag, i) => (
                <Badge key={i} variant="secondary" className="gap-1 font-mono text-xs">
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
                className="h-9 min-h-9 font-mono text-xs"
                value={tagInput}
                onChange={(e) => setTagInput(e.target.value)}
                onKeyDown={(e) => {
                  if (e.key === 'Enter' && tagInput.trim()) {
                    e.preventDefault()
                    const newTags = [...tags, ...tagInput.split(',').map(t => t.trim()).filter(Boolean)]
                    setTags(newTags)
                    setValue('tags', newTags.join(','))
                    setTagInput('')
                  }
                }}
                onBlur={() => {
                  if (tagInput.trim()) {
                    const newTags = [...tags, ...tagInput.split(',').map(t => t.trim()).filter(Boolean)]
                    setTags(newTags)
                    setValue('tags', newTags.join(','))
                    setTagInput('')
                  }
                }}
              />
            </div>
            <input type="hidden" {...register('tags')} />
          </div>

          {/* 权限组 */}
          <div className="space-y-2">
            <Label className="font-mono text-[12px] text-foreground/80">权限组</Label>
            <div className="relative" ref={groupDropdownRef}>
              <button
                type="button"
                className="flex h-9 w-full items-center justify-between rounded-md border border-input bg-background px-3 py-2 text-xs font-mono ring-offset-background placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2"
                onClick={() => setGroupDropdownOpen(!groupDropdownOpen)}
              >
                <span className="truncate">
                  {(() => {
                    const selectedIds = (groupIdsVal || '').split(',').filter(Boolean).map(Number)
                    if (selectedIds.length === 0) return <span className="text-muted-foreground">选择权限组</span>
                    return selectedIds.map(id => groupMap[id] || `#${id}`).join(', ')
                  })()}
                </span>
                <ChevronRight className={`h-4 w-4 shrink-0 opacity-50 transition-transform ${groupDropdownOpen ? 'rotate-90' : ''}`} />
              </button>
              {groupDropdownOpen && (
                <div className="absolute z-50 mt-1 max-h-60 w-full overflow-auto rounded-md border bg-popover p-1 shadow-md">
                  {groups.map((g) => {
                    const selectedIds = (groupIdsVal || '').split(',').filter(Boolean).map(Number)
                    const checked = selectedIds.includes(g.id)
                    return (
                      <label
                        key={g.id}
                        className="flex cursor-pointer items-center gap-2 rounded-sm px-2 py-1.5 text-xs font-mono hover:bg-accent hover:text-accent-foreground"
                      >
                        <Checkbox
                          checked={checked}
                          onCheckedChange={(v) => {
                            const current = (groupIdsVal || '').split(',').filter(Boolean).map(Number)
                            let next: number[]
                            if (v) {
                              next = [...current, g.id]
                            } else {
                              next = current.filter(id => id !== g.id)
                            }
                            setValue('group_ids', next.join(','))
                          }}
                        />
                        <span>{g.name}</span>
                      </label>
                    )
                  })}
                  {groups.length === 0 && (
                    <p className="px-2 py-1.5 text-xs text-muted-foreground">暂无权限组</p>
                  )}
                </div>
              )}
            </div>
          </div>

          {/* 节点地址 + 端口 */}
          <div className="flex space-x-2">
            <div className="flex-1 space-y-2">
              <Label className="font-mono text-[12px] text-foreground/80">节点地址 <span className="text-destructive">*</span></Label>
              <Input placeholder="hk01.example.com" className="h-9 font-mono text-xs" {...register('host')} />
            </div>
            <div className="flex-1 space-y-2">
              <Label className="flex items-center gap-1.5 font-mono text-[12px] text-foreground/80">
                连接端口 <span className="text-destructive">*</span>
              </Label>
              <div className="flex items-center gap-1">
                <Input type="number" placeholder="443" className="h-9 font-mono text-xs" {...register('port')} />
                <Button
                  type="button"
                  variant="ghost"
                  size="icon"
                  className="h-6 w-6 shrink-0 text-muted-foreground/50 hover:text-muted-foreground"
                  onClick={() => {
                    const portValue = watch('port')
                    if (portValue) setValue('server_port', portValue)
                  }}
                >
                  <ChevronRight className="h-3 w-3" />
                </Button>
              </div>
            </div>
            <div className="flex-1 space-y-2">
              <Label className="font-mono text-[12px] text-foreground/80">服务端口</Label>
              <Input type="number" placeholder="与连接端口相同时留空" className="h-9 font-mono text-xs" {...register('server_port')} />
            </div>
          </div>

          {/* 监听地址 */}
          <div className="space-y-2">
            <Label className="font-mono text-[12px] text-foreground/80">监听地址</Label>
            <Input placeholder="留空使用默认 (0.0.0.0)" className="h-9 font-mono text-xs" {...register('listen_address')} />
          </div>

          {/* 动态倍率 */}
          <div className="space-y-3 rounded-xl border bg-muted/5 p-4">
            <div className="flex items-center gap-2">
              <Switch
                checked={watch('rate_time_enable') ?? false}
                onCheckedChange={(v) => setValue('rate_time_enable', v)}
                className="scale-90"
              />
              <div>
                <Label className="font-mono text-[12px] text-foreground/80">启用动态倍率</Label>
                <p className="font-mono text-[11px] opacity-70">根据时间段设置不同的倍率乘数</p>
              </div>
            </div>
            {watch('rate_time_enable') && (
              <div className="space-y-3">
                <div className="flex items-center justify-between">
                  <Label className="font-mono text-[12px] text-foreground/80">时间段规则</Label>
                  <Button
                    type="button"
                    variant="outline"
                    size="sm"
                    className="h-7 px-2 font-mono text-[10px]"
                    onClick={() => {
                      const current = watch('rate_time_ranges') || []
                      setValue('rate_time_ranges', [...current, { start: '00:00', end: '23:59', rate: '1' }])
                    }}
                  >
                    <Plus className="mr-1 h-3 w-3" />添加规则
                  </Button>
                </div>
                {(watch('rate_time_ranges') || []).map((_: RateTimeRange, i: number) => (
                  <div key={i} className="space-y-3 rounded-lg border bg-background p-3">
                    <div className="flex items-center justify-between">
                      <span className="font-mono text-[11px] font-bold">规则 {i + 1}</span>
                      <Button
                        type="button"
                        variant="ghost"
                        size="icon"
                        className="h-7 w-7 text-muted-foreground hover:text-destructive"
                        onClick={() => {
                          const current = watch('rate_time_ranges') || []
                          setValue('rate_time_ranges', current.filter((_: RateTimeRange, idx: number) => idx !== i))
                        }}
                      >
                        <Trash2 className="h-3.5 w-3.5" />
                      </Button>
                    </div>
                    <div className="grid grid-cols-3 gap-3">
                      <div className="space-y-1">
                        <Label className="font-mono text-[11px] text-foreground/80">开始时间</Label>
                        <Input
                          type="time"
                          className="h-8 px-2 font-mono text-xs"
                          {...register(`rate_time_ranges.${i}.start`)}
                        />
                      </div>
                      <div className="space-y-1">
                        <Label className="font-mono text-[11px] text-foreground/80">结束时间</Label>
                        <Input
                          type="time"
                          className="h-8 px-2 font-mono text-xs"
                          {...register(`rate_time_ranges.${i}.end`)}
                        />
                      </div>
                      <div className="space-y-1">
                        <Label className="font-mono text-[11px] text-foreground/80">倍率乘数</Label>
                        <Input
                          type="number"
                          step="0.1"
                          min="0"
                          className="h-8 px-2 font-mono text-xs"
                          placeholder="1.0"
                          {...register(`rate_time_ranges.${i}.rate`)}
                        />
                      </div>
                    </div>
                  </div>
                ))}
                {(watch('rate_time_ranges') || []).length === 0 && (
                  <p className="py-4 text-center font-mono text-[10px] italic text-muted-foreground">暂无规则，点击上方按钮添加</p>
                )}
              </div>
            )}
          </div>

          {/* 父节点 + 路由 */}
          <div className="flex gap-4">
            <div className="flex-1 space-y-2">
              <Label className="font-mono text-[12px] text-foreground/80">父节点</Label>
              <Select value={String(watch('parent_id') ?? 0)} onValueChange={(v) => setValue('parent_id', Number(v))}>
                <SelectTrigger className="h-9 font-mono text-xs"><SelectValue placeholder="无" /></SelectTrigger>
                <SelectContent>
                  <SelectItem value="0">无（独立节点）</SelectItem>
                  {nodes.map((n) => <SelectItem key={n.id} value={String(n.id)}>{n.name}</SelectItem>)}
                </SelectContent>
              </Select>
            </div>
            <div className="flex-1 space-y-2">
              <Label className="font-mono text-[12px] text-foreground/80">路由组</Label>
              <Select value={String(watch('route_id') ?? 0)} onValueChange={(v) => setValue('route_id', Number(v))}>
                <SelectTrigger className="h-9 font-mono text-xs"><SelectValue placeholder="无" /></SelectTrigger>
                <SelectContent>
                  <SelectItem value="0">无</SelectItem>
                  {(routes ?? []).map((r) => <SelectItem key={r.id} value={String(r.id)}>{r.name}</SelectItem>)}
                </SelectContent>
              </Select>
            </div>
          </div>

          {/* 服务器选择器 */}
          <div className="space-y-2">
            <Label className="font-mono text-[12px] text-foreground/80">服务器机器</Label>
            <Select
              value={String(watch('machine_id') ?? 0)}
              onValueChange={(v) => {
                const val = v === '0' ? 0 : Number(v)
                setValue('machine_id', val)
                if (val && watch('show') === false) setValue('show', true)
              }}
            >
              <SelectTrigger className="h-9 font-mono text-xs"><SelectValue placeholder="独立部署" /></SelectTrigger>
              <SelectContent>
                <SelectItem value="0">独立部署</SelectItem>
                {servers.filter((s) => s.is_active !== false).map((s) => (
                  <SelectItem key={s.id} value={String(s.id)} className="cursor-pointer text-xs">
                    <span className="flex items-center gap-2">
                      <span className={`h-2 w-2 shrink-0 rounded-full ${getMachineStatus(s).dot}`} />
                      {s.name}
                      <span className="text-muted-foreground">SID:{s.id}</span>
                    </span>
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
            {watch('machine_id') ? (() => {
              const m = servers.find((s) => s.id === watch('machine_id'))
              if (!m) return null
              return (
                <div className="flex items-center justify-between rounded-md border bg-muted/20 px-3 py-2">
                  <div>
                    <div className="flex items-center gap-1.5 font-mono text-xs font-medium">
                      <span className={`h-2 w-2 shrink-0 rounded-full ${getMachineStatus(m).dot}`} />
                      {m.name}
                    </div>
                    <p className="font-mono text-[11px] text-muted-foreground">选择是否由此服务器管理该节点</p>
                  </div>
                  <Switch
                    checked={watch('show') ?? true}
                    onCheckedChange={(v) => setValue('show', v)}
                  />
                </div>
              )
            })() : null}
          </div>

          {/* 显示/封禁开关 */}
          <div className="flex items-center gap-6">
            <div className="flex items-center gap-2">
              <Switch checked={watch('show') ?? true} onCheckedChange={(checked) => setValue('show', checked)} className="scale-90" />
              <Label className="font-mono text-[12px] text-foreground/80">显示节点</Label>
            </div>
            <div className="flex items-center gap-2">
              <Switch checked={watch('banned') ?? false} onCheckedChange={(checked) => setValue('banned', checked)} className="scale-90" />
              <Label className="font-mono text-[12px] text-foreground/80">封禁节点</Label>
            </div>
          </div>
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

          {typeVal && PROTOCOLS_WITH_SETTINGS.includes(typeVal) && (
            <div className="space-y-3 rounded-lg border p-4">
              <Label className="text-sm font-medium">协议设置 ({typeVal})</Label>
              <ProtocolSettings protocol={typeVal} value={protocolSettings} onChange={setProtocolSettings} />
            </div>
          )}

          </div>
          <div className="flex items-center justify-between border-t bg-muted/20 px-6 py-4">
            <Button type="button" variant="secondary" size="sm" className="h-7 gap-2 rounded-md border border-border/50 bg-muted/50 px-2.5 font-mono text-[11px]" onClick={() => setAdvancedOpen(true)}>
              高级设置
              {(certValue && Object.keys(certValue).length > 0) && <span className="h-1 w-1 rounded-full bg-primary" />}
              {(multiplexValue?.enabled) && <span className="h-1 w-1 rounded-full bg-emerald-500" />}
              {(customOutbounds && customOutbounds !== '[]') && <span className="h-1 w-1 rounded-full bg-blue-500" />}
              {(customRoutes && customRoutes !== '[]') && <span className="h-1 w-1 rounded-full bg-violet-500" />}
            </Button>
            <div className="flex items-center gap-3">
              <Button type="button" variant="ghost" className="h-8 px-4 font-mono text-xs font-bold" onClick={() => onOpenChange(false)}>取消</Button>
              <Button type="submit" className="h-8 bg-primary px-8 font-mono text-xs font-bold text-primary-foreground hover:bg-primary/90">保存</Button>
            </div>
          </div>
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

// ─── 主页面 ───────────────────────────────────────────
export default function NodesPage() {
  const [searchParams, setSearchParams] = useSearchParams()
  const [page, setPage] = useState(1)
  const [dialog, setDialog] = useState<'create' | 'edit' | null>(null)
  const [editNode, setEditNode] = useState<Node | null>(null)
  const [deleteId, setDeleteId] = useState<number | null>(null)
  const [searchQuery, setSearchQuery] = useState('')
  const [filterGroupId, setFilterGroupId] = useState<number | undefined>(undefined)
  const [filterType, setFilterType] = useState('')
  const [filterMachineId, setFilterMachineId] = useState<number | undefined>(undefined)
  const [detailNode, setDetailNode] = useState<Node | null>(null)
  const [cloneNode, setCloneNode] = useState<Node | null>(null)
  const [selectedIds, setSelectedIds] = useState<Set<number>>(new Set())
  const [batchLoading, setBatchLoading] = useState(false)
  const [batchUpdateOpen, setBatchUpdateOpen] = useState(false)
  const [batchUpdateField, setBatchUpdateField] = useState<'show' | 'enabled' | 'machine_id'>('show')
  const [batchUpdateValue, setBatchUpdateValue] = useState<string>('')
  const [sortMode, setSortMode] = useState(false)
  const [sortItems, setSortItems] = useState<{ id: number; order: number }[]>([])
  const sortNodes = useSortNodes()

  // URL 参数联动: ?machine_id=xxx 自动筛选，?machine_id=xxx&open_create=1 自动打开创建
  useEffect(() => {
    const machineId = searchParams.get('machine_id')
    if (machineId) {
      setFilterMachineId(Number(machineId))
    }
    const openCreate = searchParams.get('open_create')
    if (openCreate === '1' && machineId) {
      setDialog('create')
    }
    // 清除 URL 参数
    if (machineId || openCreate) {
      setSearchParams({}, { replace: true })
    }
  }, [searchParams, setSearchParams])

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
  const batchUpdateNodes = useBatchUpdateNodes()

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
    if (filterMachineId) {
      result = result.filter((n) => {
        try {
          const info = JSON.parse(n.server_info || '{}')
          return info.machine_id === filterMachineId
        } catch { return false }
      })
    }
    if (searchQuery.trim()) {
      const q = searchQuery.toLowerCase()
      result = result.filter((n) => n.name.toLowerCase().includes(q) || n.host.toLowerCase().includes(q))
    }
    return result
  }, [nodes, searchQuery, filterType, filterMachineId])

  // 拖拽排序
  const handleSortSave = useCallback(() => {
    if (sortItems.length === 0) return
    sortNodes.mutate(sortItems, {
      onSuccess: () => { toast.success('排序保存成功'); setSortMode(false); setSortItems([]) },
      onError: () => toast.error('排序保存失败'),
    })
  }, [sortItems, sortNodes])

  const handleDragEnd = useCallback((fromIndex: number, toIndex: number) => {
    if (fromIndex === toIndex) return
    const items = [...filteredNodes]
    const [moved] = items.splice(fromIndex, 1)
    items.splice(toIndex, 0, moved)
    setSortItems(items.map((n, i) => ({ id: n.id, order: i })))
  }, [filteredNodes])

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

  const handleBatchUpdate = useCallback(() => {
    const ids = Array.from(selectedIds)
    if (ids.length === 0) return
    const payload: Record<string, any> = { ids }
    if (batchUpdateField === 'show') {
      payload.show = Number(batchUpdateValue)
    } else if (batchUpdateField === 'enabled') {
      payload.enabled = batchUpdateValue === 'true'
    } else if (batchUpdateField === 'machine_id') {
      payload.machine_id = Number(batchUpdateValue) || 0
    }
    batchUpdateNodes.mutate(payload as any, {
      onSuccess: () => {
        toast.success(`批量更新成功：${ids.length} 个节点`)
        setBatchUpdateOpen(false)
        setSelectedIds(new Set())
      },
      onError: () => toast.error('批量更新失败'),
    })
  }, [selectedIds, batchUpdateField, batchUpdateValue, batchUpdateNodes])

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
    <TooltipProvider>
      <div className="space-y-6">
        <div>
          <h1 className="text-2xl font-bold tracking-tight">节点管理</h1>
          <p className="text-muted-foreground">管理代理节点，配置协议和连接参数</p>
        </div>

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
          <Select value={String(filterMachineId ?? 0)} onValueChange={(v) => { setFilterMachineId(v === '0' ? undefined : Number(v)); setPage(1) }}>
            <SelectTrigger className="w-[160px]"><SelectValue placeholder="全部服务器" /></SelectTrigger>
            <SelectContent>
              <SelectItem value="0">全部服务器</SelectItem>
              <SelectItem value="-1">独立部署</SelectItem>
              {(allServers ?? []).map((s) => <SelectItem key={s.id} value={String(s.id)}>{s.name}</SelectItem>)}
            </SelectContent>
          </Select>
          <div className="ml-auto flex items-center gap-2">
            {(searchQuery || filterType || filterGroupId || filterMachineId) && (
              <Button variant="ghost" size="sm" className="h-8 gap-1 text-muted-foreground" onClick={() => { setSearchQuery(''); setFilterType(''); setFilterGroupId(undefined); setFilterMachineId(undefined); setPage(1) }}>
                <X className="h-3.5 w-3.5" />重置筛选
              </Button>
            )}
            {sortMode ? (
              <>
                <Button variant="outline" onClick={() => { setSortMode(false); setSortItems([]) }}>取消排序</Button>
                <Button onClick={handleSortSave} disabled={sortItems.length === 0 || sortNodes.isPending}>
                  {sortNodes.isPending ? '保存中...' : '保存排序'}
                </Button>
              </>
            ) : (
              <>
                <Button variant="outline" onClick={() => setSortMode(true)}>排序模式</Button>
                <Button onClick={() => { setCloneNode(null); setDialog('create') }}><Plus className="mr-2 h-4 w-4" />新建节点</Button>
              </>
            )}
          </div>
        </div>

        {/* 服务器筛选提示条 */}
        {filterMachineId && filterMachineId !== -1 && (() => {
          const m = (allServers ?? []).find((s) => s.id === filterMachineId)
          if (!m) return null
          return (
            <div className="flex items-center gap-3 rounded-lg border border-primary/20 bg-primary/5 px-4 py-2.5">
              <Filter className="h-4 w-4 text-primary" />
              <span className="text-sm">
                当前正在查看服务器 <strong>{m.name}</strong> <span className="text-muted-foreground font-mono">(SID:{m.id})</span> 下的节点
              </span>
              <div className="ml-auto flex items-center gap-2">
                <Button size="sm" variant="outline" className="h-7 gap-1 text-xs" onClick={() => { setCloneNode(null); setDialog('create') }}>
                  <Plus className="h-3 w-3" />新增节点到此服务器
                </Button>
                <Button size="sm" variant="ghost" className="h-7 gap-1 text-xs text-muted-foreground" onClick={() => setFilterMachineId(undefined)}>
                  <X className="h-3 w-3" />清除筛选
                </Button>
              </div>
            </div>
          )
        })()}

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
              <Button size="sm" variant="outline" disabled={batchLoading} onClick={() => setBatchUpdateOpen(true)}>
                <Pencil className="mr-1.5 h-3.5 w-3.5" />批量更新
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
                    const trafficUsed = (n.u ?? 0) + (n.d ?? 0)
                    const trafficLimit = n.transfer_enable ?? 0
                    const metrics = info.metrics
                    const cpu = metrics?.cpu ?? 0
                    const hb = timeSince(n.last_online)
                    return (
                      <TableRow key={n.id} className={`group ${n.status === 0 ? 'opacity-50' : ''}`}>
                        <TableCell className="text-center">
                          <Checkbox checked={selectedIds.has(n.id)} onCheckedChange={() => toggleSelect(n.id)} />
                        </TableCell>
                        <TableCell>
                          <Badge
                            variant="outline"
                            className="border-2 font-mono text-xs gap-1"
                            style={{ borderColor: PROTOCOL_COLORS[n.type] ?? '#6b7280' }}
                          >
                            <Server className="h-3 w-3" />
                            {info.code || n.id}
                          </Badge>
                        </TableCell>
                        <TableCell>
                          <Switch
                            checked={isShow}
                            onCheckedChange={(v) => {
                              info.show = v ? 1 : 0
                              updateNode.mutate({ id: n.id, server_info: JSON.stringify(info) })
                            }}
                            className="h-4 w-7"
                            style={{ backgroundColor: isShow ? (PROTOCOL_COLORS[n.type] ?? undefined) : undefined }}
                          />
                        </TableCell>
                        <TableCell>
                          <div className="flex items-center gap-2.5">
                            <StatusDot status={n.status} />
                            <div className="min-w-0">
                              <button className="text-sm font-medium hover:text-primary cursor-pointer text-left truncate max-w-[160px] block" onClick={() => setDetailNode(n)}>
                                {n.name}
                              </button>
                              {n.tags && n.tags.split(',').filter(Boolean).length > 0 && (
                                <div className="flex gap-1 mt-0.5">
                                  {n.tags.split(',').filter(Boolean).map((tag, i) => (
                                    <span key={i} className="inline-block text-[10px] px-1 py-0 rounded bg-muted text-muted-foreground">{tag}</span>
                                  ))}
                                </div>
                              )}
                            </div>
                          </div>
                        </TableCell>
                        <TableCell><NodeTypeBadge type={n.type} /></TableCell>
                        <TableCell>
                          <div className="flex items-center gap-1">
                            <span className="font-mono text-[12px] text-foreground/80">{n.host}:{n.port}</span>
                            <Button variant="ghost" size="icon" className="h-5 w-5 opacity-0 group-hover:opacity-100 transition-opacity"
                              onClick={() => { navigator.clipboard.writeText(`${n.host}:${n.port}`); toast.success('已复制') }}>
                              <Copy className="h-3 w-3" />
                            </Button>
                          </div>
                          {info.server_port && info.server_port !== n.port && (
                            <span className="text-[10px] text-muted-foreground">内部: {info.server_port}</span>
                          )}
                        </TableCell>
                        <TableCell className="text-[12px] text-foreground/80">
                          {(() => {
                            const ids = n.group_ids?.split(',').filter(Boolean).map(Number) || []
                            if (ids.length === 0) return <span className="text-muted-foreground">--</span>
                            return (
                              <div className="flex flex-wrap gap-1">
                                {ids.map(id => (
                                  <Badge key={id} variant="secondary" className="text-[10px] font-mono">
                                    {groupMap[id] || `#${id}`}
                                  </Badge>
                                ))}
                              </div>
                            )
                          })()}
                        </TableCell>
                        <TableCell className="text-center">
                          <span className={`font-mono text-[12px] ${n.rate > 1 ? 'text-amber-600 dark:text-amber-400 font-semibold' : 'text-foreground/80'}`}>{n.rate}x</span>
                        </TableCell>
                        <TableCell>
                          {trafficLimit > 0 ? (
                            <Tooltip>
                              <TooltipTrigger>
                                <div className="flex items-center gap-2">
                                  <div className="h-1.5 w-12 rounded-full bg-secondary">
                                    <div
                                      className={`h-full rounded-full transition-all ${info.banned || (trafficUsed / trafficLimit * 100) > 90 ? 'bg-destructive' : 'bg-primary'}`}
                                      style={{ width: `${Math.min((trafficUsed / trafficLimit) * 100, 100)}%` }}
                                    />
                                  </div>
                                  <span className={`text-xs tabular-nums text-muted-foreground ${info.banned ? 'text-destructive' : ''}`}>
                                    {formatBytes(trafficUsed)}
                                  </span>
                                </div>
                              </TooltipTrigger>
                              <TooltipContent>
                                <div className="text-xs space-y-0.5">
                                  <p>已用: {formatBytes(trafficUsed)}</p>
                                  <p>限制: {formatBytes(trafficLimit)}</p>
                                  <p>使用率: {((trafficUsed / trafficLimit) * 100).toFixed(1)}%</p>
                                </div>
                              </TooltipContent>
                            </Tooltip>
                          ) : (
                            <span className="text-xs tabular-nums text-muted-foreground">{formatBytes(trafficUsed)}</span>
                          )}
                        </TableCell>
                        <TableCell>
                          <DropdownMenu>
                            <DropdownMenuTrigger asChild>
                              <button className="flex items-center gap-1.5 hover:bg-muted/50 rounded px-1.5 py-0.5 transition-colors cursor-pointer w-full text-left">
                                {machine ? (
                                  <>
                                    <span className={`h-2 w-2 rounded-full ${getMachineStatus(machine).dot}`} />
                                    <span className="truncate text-xs font-medium">{machine.name}</span>
                                    <Badge variant="outline" className={`text-[10px] font-mono ${getMachineStatus(machine).dot === 'bg-emerald-500' ? 'border-emerald-500/25 bg-emerald-500/10 text-emerald-700' : 'border-rose-500/25 bg-rose-500/10 text-rose-700'}`}>
                                      {getMachineStatus(machine).label}
                                    </Badge>
                                  </>
                                ) : (
                                  <>
                                    <Server className="h-3.5 w-3.5 text-muted-foreground" />
                                    <span className="text-xs text-muted-foreground">独立部署</span>
                                  </>
                                )}
                              </button>
                            </DropdownMenuTrigger>
                            <DropdownMenuContent align="start" className="w-48">
                              <DropdownMenuItem onClick={() => {
                                const newInfo = { ...info }
                                delete newInfo.machine_id
                                updateNode.mutate({ id: n.id, server_info: JSON.stringify(newInfo) })
                              }}>
                                <Server className="mr-2 h-3.5 w-3.5 text-muted-foreground" />
                                <span className="text-xs">独立部署</span>
                                {!machine && <span className="ml-auto text-[10px] text-muted-foreground">当前</span>}
                              </DropdownMenuItem>
                              {(allServers ?? []).filter((s) => s.is_active !== false).map((s) => (
                                <DropdownMenuItem key={s.id} onClick={() => {
                                  const newInfo = { ...info, machine_id: s.id }
                                  updateNode.mutate({ id: n.id, server_info: JSON.stringify(newInfo) })
                                }}>
                                  <span className={`mr-2 h-2 w-2 rounded-full ${getMachineStatus(s).dot}`} />
                                  <span className="text-xs truncate flex-1">{s.name}</span>
                                  <span className="text-[10px] text-muted-foreground font-mono">SID:{s.id}</span>
                                  {machine?.id === s.id && <span className="ml-1 text-[10px] text-muted-foreground">当前</span>}
                                </DropdownMenuItem>
                              ))}
                            </DropdownMenuContent>
                          </DropdownMenu>
                        </TableCell>
                        <TableCell>
                          {metrics ? (
                            <Tooltip>
                              <TooltipTrigger>
                                <div className="flex items-center gap-1.5">
                                  <div className="w-10 h-1.5 rounded-full bg-muted overflow-hidden">
                                    <div className={`h-full rounded-full ${cpu >= 85 ? 'bg-destructive' : cpu >= 70 ? 'bg-yellow-500' : 'bg-emerald-500'}`} style={{ width: `${Math.min(cpu, 100)}%` }} />
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
                        <TableCell className="text-center">
                          <span className="inline-flex items-center gap-1 font-mono text-[12px]">
                            <Users className="h-3 w-3 text-muted-foreground" />
                            {n.online_user_count ?? 0}
                          </span>
                        </TableCell>
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

        <NodeDetailDialog open={detailNode !== null} onOpenChange={() => setDetailNode(null)} node={detailNode} groupNames={detailNode ? (detailNode.group_ids?.split(',').filter(Boolean).map(Number).map(id => allGroupMap[id] || `#${id}`) || []) : []} />
        <NodeFormDialog
          open={dialog === 'create'} onOpenChange={() => { setDialog(null); setCloneNode(null) }}
          onSubmit={cloneNode ? handleCloneSubmit : handleCreateSubmit}
          defaultValues={cloneNode ? { name: `${cloneNode.name}(副本)`, type: cloneNode.type, host: cloneNode.host, port: cloneNode.port, server_info: cloneNode.server_info || '{}', group_ids: cloneNode.group_ids, rate: cloneNode.rate, parent_id: cloneNode.parent_id, tags: cloneNode.tags, transfer_enable: cloneNode.transfer_enable } : {}}
          title={cloneNode ? '克隆节点' : '新建节点'} groups={allGroups} nodes={allNodes?.list ?? []} servers={allServers ?? []} routes={routesData?.list ?? []}
        />
        <NodeFormDialog
          open={dialog === 'edit' && !!editNode} onOpenChange={(v) => { if (!v) { setDialog(null); setEditNode(null) } }}
          onSubmit={handleUpdateSubmit}
          defaultValues={editNode ? { name: editNode.name, type: editNode.type, host: editNode.host, port: editNode.port, server_info: editNode.server_info || '{}', group_ids: editNode.group_ids, rate: editNode.rate, parent_id: editNode.parent_id, tags: editNode.tags, transfer_enable: editNode.transfer_enable } : {}}
          title="编辑节点" groups={allGroups} nodes={allNodes?.list?.filter((n) => n.id !== editNode?.id) ?? []} servers={allServers ?? []} routes={routesData?.list ?? []}
        />
        <ConfirmDialog open={deleteId !== null} onOpenChange={() => setDeleteId(null)} title="确认删除" description="删除后节点配置将无法恢复。" onConfirm={() => deleteId !== null && deleteNode.mutate(deleteId, { onSuccess: () => { setDeleteId(null); toast.success('节点已删除') } })} loading={deleteNode.isPending} />
        <Dialog open={batchUpdateOpen} onOpenChange={setBatchUpdateOpen}>
          <DialogContent>
            <DialogHeader>
              <DialogTitle>批量更新</DialogTitle>
              <DialogDescription>对已选择的 {selectedIds.size} 个节点进行批量字段更新</DialogDescription>
            </DialogHeader>
            <div className="space-y-4">
              <div className="space-y-2">
                <Label>更新字段</Label>
                <Select value={batchUpdateField} onValueChange={(v) => { setBatchUpdateField(v as any); setBatchUpdateValue('') }}>
                  <SelectTrigger><SelectValue /></SelectTrigger>
                  <SelectContent>
                    <SelectItem value="show">显示 (show)</SelectItem>
                    <SelectItem value="enabled">启用 (enabled)</SelectItem>
                    <SelectItem value="machine_id">关联机器 (machine_id)</SelectItem>
                  </SelectContent>
                </Select>
              </div>
              <div className="space-y-2">
                <Label>更新值</Label>
                {batchUpdateField === 'show' ? (
                  <Select value={batchUpdateValue} onValueChange={setBatchUpdateValue}>
                    <SelectTrigger><SelectValue placeholder="选择" /></SelectTrigger>
                    <SelectContent>
                      <SelectItem value="1">显示</SelectItem>
                      <SelectItem value="0">隐藏</SelectItem>
                    </SelectContent>
                  </Select>
                ) : batchUpdateField === 'enabled' ? (
                  <Select value={batchUpdateValue} onValueChange={setBatchUpdateValue}>
                    <SelectTrigger><SelectValue placeholder="选择" /></SelectTrigger>
                    <SelectContent>
                      <SelectItem value="true">启用</SelectItem>
                      <SelectItem value="false">禁用</SelectItem>
                    </SelectContent>
                  </Select>
                ) : (
                  <Select value={batchUpdateValue} onValueChange={setBatchUpdateValue}>
                    <SelectTrigger><SelectValue placeholder="选择机器" /></SelectTrigger>
                    <SelectContent>
                      <SelectItem value="0">无 (取消关联)</SelectItem>
                      {(allServers ?? []).map((s) => <SelectItem key={s.id} value={String(s.id)}>{s.name}</SelectItem>)}
                    </SelectContent>
                  </Select>
                )}
              </div>
            </div>
            <DialogFooter>
              <Button variant="outline" onClick={() => setBatchUpdateOpen(false)}>取消</Button>
              <Button onClick={handleBatchUpdate} disabled={!batchUpdateValue || batchUpdateNodes.isPending}>
                {batchUpdateNodes.isPending ? '更新中...' : '确认更新'}
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      </div>
    </TooltipProvider>
  )
}
