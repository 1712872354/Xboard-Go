import { useState, useMemo } from 'react'
import { toast } from 'sonner'
import { useServerRoutes, useCreateServerRoute, useUpdateServerRoute, useDeleteServerRoute } from '@/hooks/useServerRoutes'
import type { ServerRoute } from '@/types'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import { Badge } from '@/components/ui/badge'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import {
  Dialog, DialogContent, DialogDescription, DialogFooter,
  DialogHeader, DialogTitle, DialogTrigger,
} from '@/components/ui/dialog'
import {
  AlertDialog, AlertDialogAction, AlertDialogCancel, AlertDialogContent,
  AlertDialogDescription, AlertDialogFooter, AlertDialogHeader, AlertDialogTitle,
} from '@/components/ui/alert-dialog'
import { Skeleton } from '@/components/ui/skeleton'
import { Plus, Pencil, Trash2, Route, ShieldX, Globe, ArrowRight, Wifi, Search, X } from 'lucide-react'

const ACTION_OPTIONS = [
  { value: 'block', label: '阻止访问', icon: ShieldX, variant: 'destructive' as const, dotClass: 'bg-red-500' },
  { value: 'dns', label: 'DNS解析', icon: Globe, variant: 'secondary' as const, dotClass: 'bg-blue-500' },
  { value: 'direct', label: '直连', icon: ArrowRight, variant: 'secondary' as const, dotClass: 'bg-emerald-500' },
  { value: 'proxy', label: '转发', icon: Wifi, variant: 'default' as const, dotClass: 'bg-violet-500' },
]

const actionMap = Object.fromEntries(ACTION_OPTIONS.map((a) => [a.value, a]))

function ActionBadge({ action }: { action: string }) {
  const info = actionMap[action]
  if (!info) return <Badge variant="secondary">{action}</Badge>
  const Icon = info.icon
  return (
    <Badge variant={info.variant} className="gap-1">
      <Icon className="h-3 w-3" />
      {info.label}
    </Badge>
  )
}

function RouteFormDialog({ route, onSave, children }: {
  route?: ServerRoute
  onSave: (data: Partial<ServerRoute>) => void
  children: React.ReactNode
}) {
  const [open, setOpen] = useState(false)
  const [remarks, setRemarks] = useState(route?.name || '')
  const [match, setMatch] = useState(route?.match || '')
  const [action, setAction] = useState(route?.action || 'direct')
  const [actionValue, setActionValue] = useState(route?.action_value || '')
  const [errors, setErrors] = useState<Record<string, string>>({})

  const handleSave = () => {
    const newErrors: Record<string, string> = {}
    if (!remarks.trim()) newErrors.remarks = '请输入备注'
    if (!match.trim()) newErrors.match = '请输入匹配规则'
    if (Object.keys(newErrors).length > 0) {
      setErrors(newErrors)
      return
    }
    setErrors({})
    onSave({ name: remarks.trim(), match, action, action_value: actionValue || undefined })
    setOpen(false)
    setRemarks(''); setMatch(''); setAction('direct'); setActionValue('')
  }

  return (
    <Dialog open={open} onOpenChange={(v) => { setOpen(v); if (!v) { setRemarks(route?.name || ''); setMatch(route?.match || ''); setAction(route?.action || 'direct'); setActionValue(route?.action_value || ''); setErrors({}) } }}>
      <DialogTrigger asChild>{children}</DialogTrigger>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{route ? '编辑路由' : '添加路由'}</DialogTitle>
          <DialogDescription>配置路由匹配规则和动作</DialogDescription>
        </DialogHeader>
        <div className="space-y-4 py-4">
          <div className="space-y-2">
            <Label>备注</Label>
            <Input value={remarks} onChange={(e) => { setRemarks(e.target.value); setErrors((p) => ({ ...p, remarks: '' })) }} placeholder="如：直连国内" />
            {errors.remarks && <p className="text-xs text-destructive">{errors.remarks}</p>}
          </div>
          <div className="space-y-2">
            <Label>匹配规则</Label>
            <Textarea value={match} onChange={(e) => { setMatch(e.target.value); setErrors((p) => ({ ...p, match: '' })) }} placeholder={"example.com\n*.example.com"} rows={4} />
            {errors.match && <p className="text-xs text-destructive">{errors.match}</p>}
            <p className="text-xs text-muted-foreground">每行一条规则，支持域名和通配符</p>
          </div>
          <div className="space-y-2">
            <Label>动作</Label>
            <Select value={action} onValueChange={(v) => { setAction(v); setActionValue('') }}>
              <SelectTrigger><SelectValue /></SelectTrigger>
              <SelectContent>
                {ACTION_OPTIONS.map((o) => (
                  <SelectItem key={o.value} value={o.value}>
                    <span className="flex items-center gap-2">
                      <span className={`h-2 w-2 rounded-full ${o.dotClass}`} />
                      {o.label}
                    </span>
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
          {action === 'dns' && (
            <div className="space-y-2">
              <Label>DNS 服务器</Label>
              <Input value={actionValue} onChange={(e) => setActionValue(e.target.value)} placeholder="8.8.8.8" />
            </div>
          )}
          {action === 'proxy' && (
            <div className="space-y-2">
              <Label>转发标签 (Outbound Tag)</Label>
              <Input value={actionValue} onChange={(e) => setActionValue(e.target.value)} placeholder="proxy" />
            </div>
          )}
        </div>
        <DialogFooter>
          <Button variant="outline" onClick={() => setOpen(false)}>取消</Button>
          <Button onClick={handleSave}>保存</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}

export default function ServerRoutesPage() {
  const { data, isLoading } = useServerRoutes()
  const createRoute = useCreateServerRoute()
  const updateRoute = useUpdateServerRoute()
  const deleteRoute = useDeleteServerRoute()
  const [deleteId, setDeleteId] = useState<number | null>(null)
  const [search, setSearch] = useState('')

  const routes = data?.list || []

  const filteredRoutes = useMemo(() => {
    if (!search.trim()) return routes
    const q = search.toLowerCase()
    return routes.filter((r) => r.name.toLowerCase().includes(q))
  }, [routes, search])

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold tracking-tight">路由管理</h1>
          <p className="text-muted-foreground">管理所有路由组，控制域名访问策略</p>
        </div>
        <RouteFormDialog onSave={(data) => createRoute.mutate(data, { onSuccess: () => toast.success('创建成功') })}>
          <Button><Plus className="mr-2 h-4 w-4" />添加路由</Button>
        </RouteFormDialog>
      </div>

      {/* 搜索 */}
      <div className="flex items-center gap-2 max-w-sm">
        <div className="relative flex-1">
          <Search className="absolute left-2.5 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
          <Input placeholder="搜索路由名称..." value={search} onChange={(e) => setSearch(e.target.value)} className="pl-9" />
        </div>
        {search && (
          <Button variant="ghost" size="icon" className="h-9 w-9 shrink-0" onClick={() => setSearch('')}>
            <X className="h-4 w-4" />
          </Button>
        )}
      </div>

      {isLoading ? (
        <div className="space-y-2">{Array.from({ length: 4 }).map((_, i) => <Skeleton key={i} className="h-12 w-full" />)}</div>
      ) : filteredRoutes.length === 0 ? (
        <div className="flex flex-col items-center justify-center py-16 text-muted-foreground">
          <Route className="mb-3 h-10 w-10" />
          <p className="text-sm">{search ? '没有匹配的路由' : '暂无路由规则'}</p>
        </div>
      ) : (
        <div className="rounded-md border">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead className="w-16">ID</TableHead>
                <TableHead>备注</TableHead>
                <TableHead>动作</TableHead>
                <TableHead>动作值</TableHead>
                <TableHead className="w-24 text-right">操作</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {filteredRoutes.map((r) => {
                const matchCount = r.match ? r.match.split('\n').filter(Boolean).length : 0
                return (
                  <TableRow key={r.id}>
                    <TableCell className="font-mono text-muted-foreground">{r.id}</TableCell>
                    <TableCell className="font-medium max-w-[200px] truncate">{r.name}</TableCell>
                    <TableCell><ActionBadge action={r.action} /></TableCell>
                    <TableCell>
                      <div className="flex flex-col gap-0.5">
                        {r.action_value && <span className="text-sm">{r.action_value}</span>}
                        <span className="text-xs text-muted-foreground">匹配 {matchCount} 条规则</span>
                      </div>
                    </TableCell>
                    <TableCell className="text-right">
                      <div className="flex items-center justify-end gap-1">
                        <RouteFormDialog route={r} onSave={(data) => updateRoute.mutate({ id: r.id, ...data }, { onSuccess: () => toast.success('更新成功') })}>
                          <Button variant="ghost" size="icon" className="h-8 w-8">
                            <Pencil className="h-3.5 w-3.5" />
                          </Button>
                        </RouteFormDialog>
                        <Button variant="ghost" size="icon" className="h-8 w-8 text-destructive hover:text-destructive" onClick={() => setDeleteId(r.id)}>
                          <Trash2 className="h-3.5 w-3.5" />
                        </Button>
                      </div>
                    </TableCell>
                  </TableRow>
                )
              })}
            </TableBody>
          </Table>
        </div>
      )}

      <AlertDialog open={deleteId !== null} onOpenChange={() => setDeleteId(null)}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>确认删除</AlertDialogTitle>
            <AlertDialogDescription>此操作将永久删除该路由组，删除后无法恢复。确定要继续吗？</AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>取消</AlertDialogCancel>
            <AlertDialogAction onClick={() => { if (deleteId !== null) deleteRoute.mutate(deleteId, { onSuccess: () => { toast.success('删除成功'); setDeleteId(null) } }) }}>删除</AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  )
}
