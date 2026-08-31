import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { toast } from 'sonner'
import api from '@/lib/api'
import type { Knowledge, PaginatedResponse } from '@/types'
import { Card, CardContent, CardHeader } from '@/components/ui/card'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Switch } from '@/components/ui/switch'
import { Skeleton } from '@/components/ui/skeleton'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { formatDate } from '@/lib/utils'
import { Plus, Pencil, Trash2 } from 'lucide-react'

const useAdminKnowledges = (page = 1, pageSize = 20) =>
  useQuery({
    queryKey: ['admin', 'knowledges', page, pageSize],
    queryFn: async () => (await api.get('/admin/knowledges', { params: { page, page_size: pageSize } })) as unknown as PaginatedResponse<Knowledge>,
  })

const useCreateKnowledge = () => {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (data: Partial<Knowledge>) => await api.post('/admin/knowledges', data),
    onSuccess: () => { qc.invalidateQueries({ queryKey: ['admin', 'knowledges'] }); toast.success('知识库创建成功') },
  })
}

const useUpdateKnowledge = () => {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async ({ id, ...data }: Partial<Knowledge> & { id: number }) => await api.put(`/admin/knowledges/${id}`, data),
    onSuccess: () => { qc.invalidateQueries({ queryKey: ['admin', 'knowledges'] }); toast.success('知识库更新成功') },
  })
}

const useDeleteKnowledge = () => {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (id: number) => await api.delete(`/admin/knowledges/${id}`),
    onSuccess: () => { qc.invalidateQueries({ queryKey: ['admin', 'knowledges'] }); toast.success('知识库删除成功') },
  })
}

const knowledgeSchema = z.object({
  title: z.string().min(1, '请输入标题'),
  category: z.string().min(1, '请输入分类'),
  content: z.string().min(1, '请输入内容'),
  language: z.string().min(1, '请选择语言'),
  show: z.number().optional(),
  sort: z.coerce.number().optional(),
})

type KnowledgeForm = z.infer<typeof knowledgeSchema>

const languages = ['zh-CN', 'en-US', 'ja-JP', 'ko-KR']

export default function KnowledgesPage() {
  const [page, setPage] = useState(1)
  const [dialog, setDialog] = useState<'create' | 'edit' | null>(null)
  const [editItem, setEditItem] = useState<Knowledge | null>(null)
  const [deleteId, setDeleteId] = useState<number | null>(null)
  const [showSwitch, setShowSwitch] = useState(true)
  const [langValue, setLangValue] = useState('zh-CN')

  const { data, isLoading } = useAdminKnowledges(page)
  const createKb = useCreateKnowledge()
  const updateKb = useUpdateKnowledge()
  const deleteKb = useDeleteKnowledge()

  const { register, handleSubmit, reset, formState: { errors } } = useForm<KnowledgeForm>({
    resolver: zodResolver(knowledgeSchema),
  })

  const openCreate = () => {
    reset({ title: '', category: '', content: '', language: 'zh-CN', sort: 0 })
    setShowSwitch(true)
    setLangValue('zh-CN')
    setDialog('create')
  }

  const openEdit = (k: Knowledge) => {
    reset({ title: k.title, category: k.category, content: k.content, language: k.language, sort: k.sort })
    setShowSwitch(k.show === 1)
    setLangValue(k.language)
    setEditItem(k)
    setDialog('edit')
  }

  const handleCreate = (data: KnowledgeForm) => {
    createKb.mutate({ ...data, language: langValue, show: showSwitch ? 1 : 0 }, { onSuccess: () => setDialog(null) })
  }

  const handleEdit = (data: KnowledgeForm) => {
    if (!editItem) return
    updateKb.mutate({ id: editItem.id, ...data, language: langValue, show: showSwitch ? 1 : 0 }, { onSuccess: () => { setDialog(null); setEditItem(null) } })
  }

  const handleDelete = () => {
    if (deleteId !== null) deleteKb.mutate(deleteId, { onSuccess: () => setDeleteId(null) })
  }

  const totalPages = data ? Math.ceil(data.total / data.page_size) : 1

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold">知识库管理</h1>
        <Button onClick={openCreate}><Plus className="mr-2 h-4 w-4" />新建知识</Button>
      </div>

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
                    <TableHead>分类</TableHead>
                    <TableHead>语言</TableHead>
                    <TableHead>显示</TableHead>
                    <TableHead>排序</TableHead>
                    <TableHead>创建时间</TableHead>
                    <TableHead className="text-right">操作</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {data?.list.map((k) => (
                    <TableRow key={k.id}>
                      <TableCell className="font-medium">{k.title}</TableCell>
                      <TableCell>{k.category}</TableCell>
                      <TableCell><Badge variant="outline">{k.language}</Badge></TableCell>
                      <TableCell>
                        <Badge variant={k.show === 1 ? 'success' : 'secondary'}>{k.show === 1 ? '显示' : '隐藏'}</Badge>
                      </TableCell>
                      <TableCell>{k.sort}</TableCell>
                      <TableCell className="text-xs">{formatDate(k.created_at)}</TableCell>
                      <TableCell className="text-right space-x-1">
                        <Button size="icon" variant="ghost" onClick={() => openEdit(k)}>
                          <Pencil className="h-4 w-4" />
                        </Button>
                        <Button size="icon" variant="ghost" onClick={() => setDeleteId(k.id)}>
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

      <Dialog open={dialog !== null} onOpenChange={() => { setDialog(null); setEditItem(null) }}>
        <DialogContent className="max-w-lg">
          <DialogHeader>
            <DialogTitle>{dialog === 'create' ? '新建知识' : '编辑知识'}</DialogTitle>
          </DialogHeader>
          <form onSubmit={handleSubmit(dialog === 'create' ? handleCreate : handleEdit)} className="space-y-4">
            <div className="grid grid-cols-2 gap-4">
              <div className="space-y-2">
                <Label>标题</Label>
                <Input {...register('title')} />
                {errors.title && <p className="text-xs text-destructive">{errors.title.message}</p>}
              </div>
              <div className="space-y-2">
                <Label>分类</Label>
                <Input {...register('category')} />
                {errors.category && <p className="text-xs text-destructive">{errors.category.message}</p>}
              </div>
            </div>
            <div className="space-y-2">
              <Label>内容</Label>
              <Textarea {...register('content')} rows={6} />
              {errors.content && <p className="text-xs text-destructive">{errors.content.message}</p>}
            </div>
            <div className="grid grid-cols-3 gap-4">
              <div className="space-y-2">
                <Label>语言</Label>
                <Select value={langValue} onValueChange={setLangValue}>
                  <SelectTrigger><SelectValue /></SelectTrigger>
                  <SelectContent>{languages.map((l) => <SelectItem key={l} value={l}>{l}</SelectItem>)}</SelectContent>
                </Select>
              </div>
              <div className="space-y-2">
                <Label>排序</Label>
                <Input type="number" {...register('sort')} />
              </div>
              <div className="flex items-end">
                <div className="flex items-center gap-2">
                  <Label>显示</Label>
                  <Switch checked={showSwitch} onCheckedChange={setShowSwitch} />
                </div>
              </div>
            </div>
            <DialogFooter>
              <Button type="button" variant="outline" onClick={() => setDialog(null)}>取消</Button>
              <Button type="submit">保存</Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>

      <Dialog open={deleteId !== null} onOpenChange={() => setDeleteId(null)}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>确认删除</DialogTitle>
            <DialogDescription>确定要删除该知识条目吗？此操作不可撤销。</DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button variant="outline" onClick={() => setDeleteId(null)}>取消</Button>
            <Button variant="destructive" onClick={handleDelete} disabled={deleteKb.isPending}>确认删除</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}
