import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'
import api from '@/lib/api'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import { Switch } from '@/components/ui/switch'
import { Skeleton } from '@/components/ui/skeleton'
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from '@/components/ui/dialog'
import { Plus, Pencil, Trash2, Loader2 } from 'lucide-react'

interface ServerRoute {
  id: number
  name: string
  remarks: string
  match: string[]
  action: string
  action_value: string
  status: number
  created_at: string
}

function RouteForm({ route, onClose, onSave }: { route?: ServerRoute; onClose: () => void; onSave: (data: Partial<ServerRoute>) => void }) {
  const [name, setName] = useState(route?.name ?? '')
  const [remarks, setRemarks] = useState(route?.remarks ?? '')
  const [match, setMatch] = useState(route?.match?.join('\n') ?? '')
  const [action, setAction] = useState(route?.action ?? 'dns')
  const [actionValue, setActionValue] = useState(route?.action_value ?? '')
  const [status, setStatus] = useState(route?.status ?? 1)
  const [saving, setSaving] = useState(false)

  const handleSubmit = async () => {
    if (!name.trim()) {
      toast.error('请输入路由名称')
      return
    }
    setSaving(true)
    await onSave({
      name,
      remarks,
      match: match.split('\n').filter((s) => s.trim()),
      action,
      action_value: actionValue,
      status,
    })
    setSaving(false)
  }

  return (
    <div className="space-y-4">
      <div className="space-y-2">
        <Label>路由名称</Label>
        <Input value={name} onChange={(e) => setName(e.target.value)} placeholder="请输入路由名称" />
      </div>
      <div className="space-y-2">
        <Label>备注</Label>
        <Input value={remarks} onChange={(e) => setRemarks(e.target.value)} placeholder="备注信息" />
      </div>
      <div className="space-y-2">
        <Label>匹配规则（每行一个）</Label>
        <Textarea value={match} onChange={(e) => setMatch(e.target.value)} placeholder="域名或IP，每行一个&#10;example.com&#10;192.168.0.0/16" rows={4} />
      </div>
      <div className="space-y-2">
        <Label>动作</Label>
        <select className="w-full rounded-md border border-input bg-background px-3 py-2 text-sm" value={action} onChange={(e) => setAction(e.target.value)}>
          <option value="dns">DNS 解析</option>
          <option value="proxy">代理</option>
          <option value="direct">直连</option>
          <option value="block">阻断</option>
        </select>
      </div>
      <div className="space-y-2">
        <Label>动作值</Label>
        <Input value={actionValue} onChange={(e) => setActionValue(e.target.value)} placeholder="如 DNS 服务器地址" />
      </div>
      <div className="flex items-center justify-between">
        <Label>启用状态</Label>
        <Switch checked={status === 1} onCheckedChange={(v) => setStatus(v ? 1 : 0)} />
      </div>
      <DialogFooter>
        <Button variant="outline" onClick={onClose}>取消</Button>
        <Button onClick={handleSubmit} disabled={saving}>
          {saving && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
          {route ? '更新' : '创建'}
        </Button>
      </DialogFooter>
    </div>
  )
}

export default function ServerRoutesPage() {
  const qc = useQueryClient()
  const [dialogOpen, setDialogOpen] = useState(false)
  const [editingRoute, setEditingRoute] = useState<ServerRoute | undefined>()

  const { data, isLoading } = useQuery({
    queryKey: ['admin', 'server-routes'],
    queryFn: async () => await api.get('/admin/server-routes') as unknown as { list: ServerRoute[]; total: number },
  })

  const createMutation = useMutation({
    mutationFn: async (data: Partial<ServerRoute>) => await api.post('/admin/server-routes', data),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['admin', 'server-routes'] })
      setDialogOpen(false)
      toast.success('创建成功')
    },
    onError: () => toast.error('创建失败'),
  })

  const updateMutation = useMutation({
    mutationFn: async ({ id, ...data }: Partial<ServerRoute> & { id: number }) =>
      await api.put(`/admin/server-routes/${id}`, data),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['admin', 'server-routes'] })
      setDialogOpen(false)
      toast.success('更新成功')
    },
    onError: () => toast.error('更新失败'),
  })

  const deleteMutation = useMutation({
    mutationFn: async (id: number) => await api.delete(`/admin/server-routes/${id}`),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['admin', 'server-routes'] })
      toast.success('删除成功')
    },
    onError: () => toast.error('删除失败'),
  })

  const handleSave = async (formData: Partial<ServerRoute>) => {
    if (editingRoute) {
      await updateMutation.mutateAsync({ id: editingRoute.id, ...formData })
    } else {
      await createMutation.mutateAsync(formData)
    }
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold">路由管理</h1>
          <p className="text-muted-foreground">管理服务器路由规则</p>
        </div>
        <Button onClick={() => { setEditingRoute(undefined); setDialogOpen(true) }}>
          <Plus className="mr-2 h-4 w-4" />
          添加路由
        </Button>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>路由列表</CardTitle>
        </CardHeader>
        <CardContent>
          {isLoading ? (
            <div className="space-y-3">{Array.from({ length: 3 }).map((_, i) => <Skeleton key={i} className="h-12" />)}</div>
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>ID</TableHead>
                  <TableHead>名称</TableHead>
                  <TableHead>匹配规则</TableHead>
                  <TableHead>动作</TableHead>
                  <TableHead>状态</TableHead>
                  <TableHead>操作</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {data?.list?.length === 0 ? (
                  <TableRow><TableCell colSpan={6} className="h-24 text-center text-muted-foreground">暂无数据</TableCell></TableRow>
                ) : data?.list?.map((r) => (
                  <TableRow key={r.id}>
                    <TableCell className="font-mono text-xs">{r.id}</TableCell>
                    <TableCell className="font-medium">{r.name}</TableCell>
                    <TableCell>
                      <div className="flex flex-wrap gap-1">
                        {r.match?.slice(0, 3).map((m, i) => (
                          <Badge key={i} variant="outline" className="text-xs">{m}</Badge>
                        ))}
                        {(r.match?.length ?? 0) > 3 && <Badge variant="outline" className="text-xs">+{r.match!.length - 3}</Badge>}
                      </div>
                    </TableCell>
                    <TableCell><Badge>{r.action}</Badge></TableCell>
                    <TableCell>
                      <Badge variant={r.status === 1 ? 'success' : 'secondary'}>{r.status === 1 ? '启用' : '禁用'}</Badge>
                    </TableCell>
                    <TableCell>
                      <div className="flex gap-1">
                        <Button variant="ghost" size="sm" onClick={() => { setEditingRoute(r); setDialogOpen(true) }}>
                          <Pencil className="h-4 w-4" />
                        </Button>
                        <Button variant="ghost" size="sm" onClick={() => { if (confirm('确定删除？')) deleteMutation.mutate(r.id) }}>
                          <Trash2 className="h-4 w-4 text-destructive" />
                        </Button>
                      </div>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          )}
        </CardContent>
      </Card>

      <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{editingRoute ? '编辑路由' : '添加路由'}</DialogTitle>
          </DialogHeader>
          <RouteForm route={editingRoute} onClose={() => setDialogOpen(false)} onSave={handleSave} />
        </DialogContent>
      </Dialog>
    </div>
  )
}
