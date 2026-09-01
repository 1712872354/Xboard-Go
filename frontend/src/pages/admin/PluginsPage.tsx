import { useState, useRef } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { toast } from 'sonner'
import api from '@/lib/api'
import type { Plugin, PaginatedResponse } from '@/types'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import { Switch } from '@/components/ui/switch'
import { Skeleton } from '@/components/ui/skeleton'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { Plus, Trash2, Settings, Plug, Zap, Upload } from 'lucide-react'
import { Tabs, TabsList, TabsTrigger } from '@/components/ui/tabs'

const useAdminPlugins = (page = 1, pageSize = 20) =>
  useQuery({
    queryKey: ['admin', 'plugins', page, pageSize],
    queryFn: async () => (await api.get('/admin/plugins', { params: { page, page_size: pageSize } })) as unknown as PaginatedResponse<Plugin>,
  })

const useInstallPlugin = () => {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (data: Partial<Plugin>) => await api.post('/admin/plugins', data),
    onSuccess: () => { qc.invalidateQueries({ queryKey: ['admin', 'plugins'] }); toast.success('插件安装成功') },
  })
}

const useTogglePlugin = () => {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async ({ id, status }: { id: number; status: number }) => await api.put(`/admin/plugins/${id}/status`, { status }),
    onSuccess: () => { qc.invalidateQueries({ queryKey: ['admin', 'plugins'] }); toast.success('插件状态更新成功') },
  })
}

const useConfigurePlugin = () => {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async ({ id, config }: { id: number; config: string }) => await api.put(`/admin/plugins/${id}`, { config }),
    onSuccess: () => { qc.invalidateQueries({ queryKey: ['admin', 'plugins'] }); toast.success('插件配置更新成功') },
  })
}

const useDeletePlugin = () => {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (id: number) => await api.delete(`/admin/plugins/${id}`),
    onSuccess: () => { qc.invalidateQueries({ queryKey: ['admin', 'plugins'] }); toast.success('插件删除成功') },
  })
}

const useUploadPlugin = () => {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (file: File) => {
      const formData = new FormData()
      formData.append('file', file)
      return await api.post('/admin/plugins/upload', formData, {
        headers: { 'Content-Type': 'multipart/form-data' },
      })
    },
    onSuccess: () => { qc.invalidateQueries({ queryKey: ['admin', 'plugins'] }); toast.success('插件上传成功') },
  })
}

const pluginSchema = z.object({
  name: z.string().min(1, '请输入名称'),
  title: z.string().min(1, '请输入标题'),
  description: z.string().optional(),
  version: z.string().optional(),
  author: z.string().optional(),
  homepage: z.string().optional(),
})

type PluginForm = z.infer<typeof pluginSchema>

export default function PluginsPage() {
  const [page, setPage] = useState(1)
  const [activeTab, setActiveTab] = useState('all')
  const [installDialog, setInstallDialog] = useState(false)
  const [configDialog, setConfigDialog] = useState<Plugin | null>(null)
  const [configValue, setConfigValue] = useState('')
  const [deleteId, setDeleteId] = useState<number | null>(null)
  const fileInputRef = useRef<HTMLInputElement>(null)

  const { data, isLoading } = useAdminPlugins(page)
  const installPlugin = useInstallPlugin()
  const togglePlugin = useTogglePlugin()
  const configurePlugin = useConfigurePlugin()
  const deletePlugin = useDeletePlugin()
  const uploadPlugin = useUploadPlugin()

  const plugins = data?.list ?? []
  const filteredPlugins = activeTab === 'all' ? plugins : plugins.filter((p) => p.type === activeTab)
  const totalPlugins = data?.total ?? 0
  const enabledCount = plugins.filter((p) => p.status === 1).length
  const disabledCount = plugins.filter((p) => p.status !== 1).length

  const handleUpload = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    if (!file) return
    if (!file.name.endsWith('.zip')) {
      toast.error('请上传 .zip 格式的插件包')
      return
    }
    uploadPlugin.mutate(file, { onSettled: () => { if (fileInputRef.current) fileInputRef.current.value = '' } })
  }

  const { register, handleSubmit, reset, formState: { errors } } = useForm<PluginForm>({
    resolver: zodResolver(pluginSchema),
  })

  const handleInstall = (data: PluginForm) => {
    installPlugin.mutate(data, { onSuccess: () => { setInstallDialog(false); reset() } })
  }

  const handleConfigure = () => {
    if (!configDialog) return
    configurePlugin.mutate({ id: configDialog.id, config: configValue }, { onSuccess: () => setConfigDialog(null) })
  }

  const handleDelete = () => {
    if (deleteId !== null) deletePlugin.mutate(deleteId, { onSuccess: () => setDeleteId(null) })
  }

  const totalPages = data ? Math.ceil(data.total / data.page_size) : 1

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold">插件管理</h1>
        <div className="flex gap-2">
          <input type="file" ref={fileInputRef} accept=".zip" className="hidden" onChange={handleUpload} />
          <Button variant="outline" onClick={() => fileInputRef.current?.click()} disabled={uploadPlugin.isPending}>
            <Upload className="mr-2 h-4 w-4" />
            {uploadPlugin.isPending ? '上传中...' : '上传插件'}
          </Button>
          <Button onClick={() => setInstallDialog(true)}><Plus className="mr-2 h-4 w-4" />安装插件</Button>
        </div>
      </div>

      <div className="grid gap-4 sm:grid-cols-3">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">总计</CardTitle>
            <Plug className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent><div className="text-2xl font-bold">{totalPlugins}</div></CardContent>
        </Card>
        <Card>
          <CardHeader className="flex flex-row items-center justify-between pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">已启用</CardTitle>
            <Zap className="h-4 w-4 text-emerald-500" />
          </CardHeader>
          <CardContent><div className="text-2xl font-bold text-emerald-600">{enabledCount}</div></CardContent>
        </Card>
        <Card>
          <CardHeader className="flex flex-row items-center justify-between pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">已禁用</CardTitle>
            <Plug className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent><div className="text-2xl font-bold text-muted-foreground">{disabledCount}</div></CardContent>
        </Card>
      </div>

      <Tabs value={activeTab} onValueChange={setActiveTab}>
        <TabsList>
          <TabsTrigger value="all">全部插件</TabsTrigger>
          <TabsTrigger value="feature">功能插件</TabsTrigger>
          <TabsTrigger value="payment">支付插件</TabsTrigger>
        </TabsList>
      </Tabs>

      <Card>
        <CardHeader />
        <CardContent>
          {isLoading ? (
            <div className="space-y-2">{Array.from({ length: 5 }).map((_, i) => <Skeleton key={i} className="h-12" />)}</div>
          ) : (
            <>
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>标题</TableHead>
                    <TableHead>名称</TableHead>
                    <TableHead>类型</TableHead>
                    <TableHead>描述</TableHead>
                    <TableHead>版本</TableHead>
                    <TableHead>作者</TableHead>
                    <TableHead>状态</TableHead>
                    <TableHead className="text-right">操作</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {filteredPlugins.map((p) => (
                    <TableRow key={p.id}>
                      <TableCell className="font-medium">{p.title}</TableCell>
                      <TableCell className="font-mono text-xs">{p.name}</TableCell>
                      <TableCell>
                        <Badge
                          variant="outline"
                          className={p.type === 'feature' ? 'bg-blue-500/10 text-blue-600 border-blue-500/25' : p.type === 'payment' ? 'bg-emerald-500/10 text-emerald-600 border-emerald-500/25' : undefined}
                        >
                          {p.type === 'payment' ? '支付' : p.type === 'feature' ? '功能' : p.type || '-'}
                        </Badge>
                      </TableCell>
                      <TableCell className="max-w-48 truncate text-muted-foreground">{p.description || '-'}</TableCell>
                      <TableCell>{p.version || '-'}</TableCell>
                      <TableCell>{p.author || '-'}</TableCell>
                      <TableCell>
                        <Badge variant={p.status === 1 ? 'success' : 'secondary'}>{p.status === 1 ? '启用' : '禁用'}</Badge>
                      </TableCell>
                      <TableCell className="text-right space-x-1">
                        <Switch checked={p.status === 1} onCheckedChange={() => togglePlugin.mutate({ id: p.id, status: p.status === 1 ? 0 : 1 })} aria-label="启用" />
                        <Button size="icon" variant="ghost" onClick={() => { setConfigDialog(p); setConfigValue(p.config || '{}') }} title="配置">
                          <Settings className="h-4 w-4" />
                        </Button>
                        <Button size="icon" variant="ghost" onClick={() => setDeleteId(p.id)}>
                          <Trash2 className="h-4 w-4 text-destructive" />
                        </Button>
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
              <div className="mt-4 flex items-center justify-between">
                <p className="text-sm text-muted-foreground">共 {data?.total ?? 0} 条</p>
                <div className="flex gap-2">
                  <Button size="sm" variant="outline" disabled={page <= 1} onClick={() => setPage(page - 1)}>上一页</Button>
                  <span className="flex items-center text-sm">{page} / {totalPages}</span>
                  <Button size="sm" variant="outline" disabled={page >= totalPages} onClick={() => setPage(page + 1)}>下一页</Button>
                </div>
              </div>
            </>
          )}
        </CardContent>
      </Card>

      {/* Install dialog */}
      <Dialog open={installDialog} onOpenChange={(v) => { setInstallDialog(v); if (!v) reset() }}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>安装插件</DialogTitle>
            <DialogDescription>填写插件信息</DialogDescription>
          </DialogHeader>
          <form onSubmit={handleSubmit(handleInstall)} className="space-y-4">
            <div className="grid grid-cols-2 gap-4">
              <div className="space-y-2">
                <Label>插件名称</Label>
                <Input {...register('name')} placeholder="my_plugin" />
                {errors.name && <p className="text-xs text-destructive">{errors.name.message}</p>}
              </div>
              <div className="space-y-2">
                <Label>显示标题</Label>
                <Input {...register('title')} />
                {errors.title && <p className="text-xs text-destructive">{errors.title.message}</p>}
              </div>
            </div>
            <div className="space-y-2">
              <Label>描述</Label>
              <Input {...register('description')} />
            </div>
            <div className="grid grid-cols-2 gap-4">
              <div className="space-y-2">
                <Label>版本</Label>
                <Input {...register('version')} placeholder="1.0.0" />
              </div>
              <div className="space-y-2">
                <Label>作者</Label>
                <Input {...register('author')} />
              </div>
            </div>
            <div className="space-y-2">
              <Label>主页</Label>
              <Input {...register('homepage')} placeholder="https://..." />
            </div>
            <DialogFooter>
              <Button type="button" variant="outline" onClick={() => setInstallDialog(false)}>取消</Button>
              <Button type="submit">安装</Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>

      {/* Configure dialog */}
      <Dialog open={!!configDialog} onOpenChange={() => setConfigDialog(null)}>
        <DialogContent className="max-w-lg">
          <DialogHeader>
            <DialogTitle>插件配置 - {configDialog?.title}</DialogTitle>
            <DialogDescription>编辑 JSON 配置</DialogDescription>
          </DialogHeader>
          <div className="space-y-4">
            <Textarea
              value={configValue}
              onChange={(e) => setConfigValue(e.target.value)}
              rows={12}
              className="font-mono text-xs"
              placeholder="{}"
            />
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setConfigDialog(null)}>取消</Button>
            <Button onClick={handleConfigure} disabled={configurePlugin.isPending}>保存</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Delete dialog */}
      <Dialog open={deleteId !== null} onOpenChange={() => setDeleteId(null)}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>确认删除</DialogTitle>
            <DialogDescription>确定要删除该插件吗？此操作不可撤销。</DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button variant="outline" onClick={() => setDeleteId(null)}>取消</Button>
            <Button variant="destructive" onClick={handleDelete} disabled={deletePlugin.isPending}>确认删除</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}
