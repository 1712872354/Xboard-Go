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
import { Switch } from '@/components/ui/switch'
import { Skeleton } from '@/components/ui/skeleton'
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from '@/components/ui/dialog'
import { Plus, Pencil, Trash2, Loader2 } from 'lucide-react'

interface ServerGroup {
  id: number
  name: string
  description: string
  status: number
  created_at: string
}

function GroupForm({ group, onClose, onSave }: { group?: ServerGroup; onClose: () => void; onSave: (data: Partial<ServerGroup>) => void }) {
  const [name, setName] = useState(group?.name ?? '')
  const [description, setDescription] = useState(group?.description ?? '')
  const [status, setStatus] = useState(group?.status ?? 1)
  const [saving, setSaving] = useState(false)

  const handleSubmit = async () => {
    if (!name.trim()) {
      toast.error('请输入分组名称')
      return
    }
    setSaving(true)
    await onSave({ name, description, status })
    setSaving(false)
  }

  return (
    <div className="space-y-4">
      <div className="space-y-2">
        <Label>分组名称</Label>
        <Input value={name} onChange={(e) => setName(e.target.value)} placeholder="请输入分组名称" />
      </div>
      <div className="space-y-2">
        <Label>描述</Label>
        <Input value={description} onChange={(e) => setDescription(e.target.value)} placeholder="请输入描述" />
      </div>
      <div className="flex items-center justify-between">
        <Label>启用状态</Label>
        <Switch checked={status === 1} onCheckedChange={(v) => setStatus(v ? 1 : 0)} />
      </div>
      <DialogFooter>
        <Button variant="outline" onClick={onClose}>取消</Button>
        <Button onClick={handleSubmit} disabled={saving}>
          {saving && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
          {group ? '更新' : '创建'}
        </Button>
      </DialogFooter>
    </div>
  )
}

export default function ServerGroupsPage() {
  const qc = useQueryClient()
  const [dialogOpen, setDialogOpen] = useState(false)
  const [editingGroup, setEditingGroup] = useState<ServerGroup | undefined>()

  const { data, isLoading } = useQuery({
    queryKey: ['admin', 'server-groups'],
    queryFn: async () => await api.get('/admin/server-groups') as unknown as { list: ServerGroup[]; total: number },
  })

  const createMutation = useMutation({
    mutationFn: async (data: Partial<ServerGroup>) => await api.post('/admin/server-groups', data),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['admin', 'server-groups'] })
      setDialogOpen(false)
      toast.success('创建成功')
    },
    onError: () => toast.error('创建失败'),
  })

  const updateMutation = useMutation({
    mutationFn: async ({ id, ...data }: Partial<ServerGroup> & { id: number }) =>
      await api.put(`/admin/server-groups/${id}`, data),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['admin', 'server-groups'] })
      setDialogOpen(false)
      toast.success('更新成功')
    },
    onError: () => toast.error('更新失败'),
  })

  const deleteMutation = useMutation({
    mutationFn: async (id: number) => await api.delete(`/admin/server-groups/${id}`),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['admin', 'server-groups'] })
      toast.success('删除成功')
    },
    onError: () => toast.error('删除失败'),
  })

  const handleSave = async (formData: Partial<ServerGroup>) => {
    if (editingGroup) {
      await updateMutation.mutateAsync({ id: editingGroup.id, ...formData })
    } else {
      await createMutation.mutateAsync(formData)
    }
  }

  const handleEdit = (group: ServerGroup) => {
    setEditingGroup(group)
    setDialogOpen(true)
  }

  const handleAdd = () => {
    setEditingGroup(undefined)
    setDialogOpen(true)
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold">权限组管理</h1>
          <p className="text-muted-foreground">管理服务器权限分组</p>
        </div>
        <Button onClick={handleAdd}>
          <Plus className="mr-2 h-4 w-4" />
          添加分组
        </Button>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>分组列表</CardTitle>
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
                  <TableHead>描述</TableHead>
                  <TableHead>状态</TableHead>
                  <TableHead>创建时间</TableHead>
                  <TableHead>操作</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {data?.list?.length === 0 ? (
                  <TableRow><TableCell colSpan={6} className="h-24 text-center text-muted-foreground">暂无数据</TableCell></TableRow>
                ) : data?.list?.map((g) => (
                  <TableRow key={g.id}>
                    <TableCell className="font-mono text-xs">{g.id}</TableCell>
                    <TableCell className="font-medium">{g.name}</TableCell>
                    <TableCell className="text-sm text-muted-foreground">{g.description || '-'}</TableCell>
                    <TableCell>
                      <Badge variant={g.status === 1 ? 'success' : 'secondary'}>
                        {g.status === 1 ? '启用' : '禁用'}
                      </Badge>
                    </TableCell>
                    <TableCell className="text-xs text-muted-foreground">{g.created_at?.slice(0, 16)}</TableCell>
                    <TableCell>
                      <div className="flex gap-1">
                        <Button variant="ghost" size="sm" onClick={() => handleEdit(g)}>
                          <Pencil className="h-4 w-4" />
                        </Button>
                        <Button variant="ghost" size="sm" onClick={() => { if (confirm('确定删除？')) deleteMutation.mutate(g.id) }}>
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
            <DialogTitle>{editingGroup ? '编辑分组' : '添加分组'}</DialogTitle>
          </DialogHeader>
          <GroupForm group={editingGroup} onClose={() => setDialogOpen(false)} onSave={handleSave} />
        </DialogContent>
      </Dialog>
    </div>
  )
}
