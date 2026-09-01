import { useState, useMemo } from 'react'
import { toast } from 'sonner'
import { useServerGroups, useCreateServerGroup, useUpdateServerGroup, useDeleteServerGroup } from '@/hooks/useServerGroups'
import type { ServerGroup } from '@/types'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import {
  Dialog, DialogContent, DialogDescription, DialogFooter,
  DialogHeader, DialogTitle, DialogTrigger,
} from '@/components/ui/dialog'
import {
  AlertDialog, AlertDialogAction, AlertDialogCancel, AlertDialogContent,
  AlertDialogDescription, AlertDialogFooter, AlertDialogHeader, AlertDialogTitle,
} from '@/components/ui/alert-dialog'
import { Skeleton } from '@/components/ui/skeleton'
import { Plus, Pencil, Trash2, Users, Server, Search, AlertTriangle } from 'lucide-react'

const NAME_REGEX = /^[a-zA-Z0-9\u4e00-\u9fa5_-]+$/

function GroupFormDialog({ group, onSave, children }: {
  group?: ServerGroup
  onSave: (data: Partial<ServerGroup>) => void
  children: React.ReactNode
}) {
  const [open, setOpen] = useState(false)
  const [name, setName] = useState(group?.name || '')
  const [error, setError] = useState('')

  const handleSave = () => {
    if (!name.trim()) {
      setError('请输入权限组名称')
      return
    }
    if (name.length > 40) {
      setError('名称不能超过40个字符')
      return
    }
    if (!NAME_REGEX.test(name)) {
      setError('名称只能包含字母、数字、中文、下划线和连字符')
      return
    }
    setError('')
    onSave({ name: name.trim() })
    setOpen(false)
    setName('')
  }

  return (
    <Dialog open={open} onOpenChange={(v) => { setOpen(v); if (!v) { setName(group?.name || ''); setError('') } }}>
      <DialogTrigger asChild>{children}</DialogTrigger>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{group ? '编辑权限组' : '创建权限组'}</DialogTitle>
          <DialogDescription>权限组用于控制用户对节点的访问权限</DialogDescription>
        </DialogHeader>
        <div className="space-y-4 py-4">
          <div className="space-y-2">
            <Label>权限组名称</Label>
            <Input
              value={name}
              onChange={(e) => { setName(e.target.value); setError('') }}
              placeholder="输入权限组名称"
              maxLength={40}
            />
            {error && <p className="text-xs text-destructive">{error}</p>}
            <p className="text-xs text-muted-foreground">支持字母、数字、中文、下划线和连字符，最多40个字符</p>
          </div>
        </div>
        <DialogFooter>
          <Button variant="outline" onClick={() => setOpen(false)}>取消</Button>
          <Button onClick={handleSave}>保存</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}

export default function ServerGroupsPage() {
  const { data, isLoading } = useServerGroups()
  const createGroup = useCreateServerGroup()
  const updateGroup = useUpdateServerGroup()
  const deleteGroup = useDeleteServerGroup()
  const [deleteId, setDeleteId] = useState<number | null>(null)
  const [search, setSearch] = useState('')

  const groups = data?.list || []

  const filteredGroups = useMemo(() => {
    if (!search.trim()) return groups
    const q = search.toLowerCase()
    return groups.filter((g) => g.name.toLowerCase().includes(q))
  }, [groups, search])

  const deleteTarget = deleteId !== null ? groups.find((g) => g.id === deleteId) : null
  const canDelete = deleteTarget ? (deleteTarget.users_count ?? 0) === 0 && (deleteTarget.server_count ?? 0) === 0 : false
  const deleteBlockedReason = deleteTarget
    ? (deleteTarget.server_count ?? 0) > 0 ? `该权限组正在被 ${deleteTarget.server_count} 个节点使用`
      : (deleteTarget.users_count ?? 0) > 0 ? `该权限组正在被 ${deleteTarget.users_count} 个用户使用`
      : ''
    : ''

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold tracking-tight">权限组管理</h1>
          <p className="text-muted-foreground">管理所有权限组，控制用户对节点的访问权限</p>
        </div>
        <GroupFormDialog onSave={(data) => createGroup.mutate(data, { onSuccess: () => toast.success('创建成功') })}>
          <Button><Plus className="mr-2 h-4 w-4" />添加权限组</Button>
        </GroupFormDialog>
      </div>

      {/* 搜索 */}
      <div className="relative max-w-sm">
        <Search className="absolute left-2.5 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
        <Input placeholder="搜索权限组名称..." value={search} onChange={(e) => setSearch(e.target.value)} className="pl-9" />
      </div>

      {isLoading ? (
        <div className="space-y-2">{Array.from({ length: 4 }).map((_, i) => <Skeleton key={i} className="h-12 w-full" />)}</div>
      ) : filteredGroups.length === 0 ? (
        <div className="flex flex-col items-center justify-center py-16 text-muted-foreground">
          <Server className="mb-3 h-10 w-10" />
          <p className="text-sm">{search ? '没有匹配的权限组' : '暂无权限组'}</p>
        </div>
      ) : (
        <div className="rounded-md border">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead className="w-16">ID</TableHead>
                <TableHead>名称</TableHead>
                <TableHead className="w-24 text-center">用户数</TableHead>
                <TableHead className="w-24 text-center">节点数</TableHead>
                <TableHead className="w-24 text-right">操作</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {filteredGroups.map((g) => (
                <TableRow key={g.id}>
                  <TableCell className="font-mono text-muted-foreground">{g.id}</TableCell>
                  <TableCell className="font-medium max-w-[300px] truncate">{g.name}</TableCell>
                  <TableCell className="text-center">
                    <span className="inline-flex items-center gap-1 text-sm">
                      <Users className="h-3.5 w-3.5 text-muted-foreground" />
                      {g.users_count ?? 0}
                    </span>
                  </TableCell>
                  <TableCell className="text-center">
                    <span className="inline-flex items-center gap-1 text-sm">
                      <Server className="h-3.5 w-3.5 text-muted-foreground" />
                      {g.server_count ?? 0}
                    </span>
                  </TableCell>
                  <TableCell className="text-right">
                    <div className="flex items-center justify-end gap-1">
                      <GroupFormDialog group={g} onSave={(data) => updateGroup.mutate({ id: g.id, ...data }, { onSuccess: () => toast.success('更新成功') })}>
                        <Button variant="ghost" size="icon" className="h-8 w-8">
                          <Pencil className="h-3.5 w-3.5" />
                        </Button>
                      </GroupFormDialog>
                      <Button variant="ghost" size="icon" className="h-8 w-8 text-destructive hover:text-destructive" onClick={() => setDeleteId(g.id)}>
                        <Trash2 className="h-3.5 w-3.5" />
                      </Button>
                    </div>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </div>
      )}

      <AlertDialog open={deleteId !== null} onOpenChange={() => setDeleteId(null)}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>确认删除</AlertDialogTitle>
            <AlertDialogDescription>
              {canDelete
                ? `此操作将永久删除权限组"${deleteTarget?.name}"，删除后无法恢复。确定要继续吗？`
                : (
                  <span className="flex items-start gap-2">
                    <AlertTriangle className="h-4 w-4 text-destructive shrink-0 mt-0.5" />
                    <span>{deleteBlockedReason}，无法删除。请先解除关联后再试。</span>
                  </span>
                )
              }
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>{canDelete ? '取消' : '关闭'}</AlertDialogCancel>
            {canDelete && (
              <AlertDialogAction onClick={() => { if (deleteId !== null) deleteGroup.mutate(deleteId, { onSuccess: () => { toast.success('删除成功'); setDeleteId(null) } }) }}>删除</AlertDialogAction>
            )}
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  )
}
