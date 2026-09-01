import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { toast } from 'sonner'
import api from '@/lib/api'
import type { Notice, PaginatedResponse } from '@/types'
import { Card, CardContent, CardHeader } from '@/components/ui/card'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import { Switch } from '@/components/ui/switch'
import { Skeleton } from '@/components/ui/skeleton'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { formatDate } from '@/lib/utils'
import { Plus, Pencil, Trash2, X, Search, ArrowUp, ArrowDown } from 'lucide-react'

const useAdminNotices = (page = 1, pageSize = 20) =>
  useQuery({
    queryKey: ['admin', 'notices', page, pageSize],
    queryFn: async () => (await api.get('/admin/notices', { params: { page, page_size: pageSize } })) as unknown as PaginatedResponse<Notice>,
  })

const useCreateNotice = () => {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (data: Partial<Notice>) => await api.post('/admin/notices', data),
    onSuccess: () => { qc.invalidateQueries({ queryKey: ['admin', 'notices'] }); toast.success('公告创建成功') },
  })
}

const useUpdateNotice = () => {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async ({ id, ...data }: Partial<Notice> & { id: number }) => await api.put(`/admin/notices/${id}`, data),
    onSuccess: () => { qc.invalidateQueries({ queryKey: ['admin', 'notices'] }); toast.success('公告更新成功') },
  })
}

const useDeleteNotice = () => {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (id: number) => await api.delete(`/admin/notices/${id}`),
    onSuccess: () => { qc.invalidateQueries({ queryKey: ['admin', 'notices'] }); toast.success('公告删除成功') },
  })
}

const noticeSchema = z.object({
  title: z.string().min(1, '请输入标题'),
  content: z.string().min(1, '请输入内容'),
  img_url: z.string().optional(),
  tags: z.string().optional(),
  show: z.number().optional(),
  popup: z.boolean().optional().default(false),
  sort: z.coerce.number().optional(),
})

type NoticeForm = z.infer<typeof noticeSchema>

export default function NoticesPage() {
  const [page, setPage] = useState(1)
  const [dialog, setDialog] = useState<'create' | 'edit' | null>(null)
  const [editNotice, setEditNotice] = useState<Notice | null>(null)
  const [deleteId, setDeleteId] = useState<number | null>(null)
  const [showSwitch, setShowSwitch] = useState(true)
  const [popupSwitch, setPopupSwitch] = useState(false)
  const [tags, setTags] = useState<string[]>([])
  const [tagInput, setTagInput] = useState('')
  const [noticeSearch, setNoticeSearch] = useState('')

  const { data, isLoading } = useAdminNotices(page)
  const createNotice = useCreateNotice()
  const updateNotice = useUpdateNotice()
  const deleteNotice = useDeleteNotice()

  const { register, handleSubmit, reset, setValue, formState: { errors } } = useForm<NoticeForm>({
    resolver: zodResolver(noticeSchema),
  })

  const handleAddTag = (e: React.KeyboardEvent<HTMLInputElement>) => {
    if (e.key === 'Enter') {
      e.preventDefault()
      const value = tagInput.trim()
      if (value && !tags.includes(value)) {
        const newTags = [...tags, value]
        setTags(newTags)
        setTagInput('')
        setValue('tags', newTags.join(','))
      }
    }
  }

  const handleRemoveTag = (tag: string) => {
    const newTags = tags.filter((t) => t !== tag)
    setTags(newTags)
    setValue('tags', newTags.join(','))
  }

  const openCreate = () => {
    reset({ title: '', content: '', img_url: '', tags: '', sort: 0 })
    setTags([])
    setTagInput('')
    setShowSwitch(true)
    setPopupSwitch(false)
    setDialog('create')
  }

  const openEdit = (n: Notice) => {
    reset({ title: n.title, content: n.content, img_url: n.img_url || '', tags: n.tags || '', sort: n.sort })
    setTags(n.tags ? n.tags.split(',').filter(Boolean) : [])
    setTagInput('')
    setShowSwitch(n.show === 1)
    setPopupSwitch(!!(n as any).popup)
    setEditNotice(n)
    setDialog('edit')
  }

  const handleCreate = (data: NoticeForm) => {
    createNotice.mutate({ ...data, show: showSwitch ? 1 : 0, popup: popupSwitch ? 1 : 0 }, { onSuccess: () => setDialog(null) })
  }

  const handleEdit = (data: NoticeForm) => {
    if (!editNotice) return
    updateNotice.mutate({ id: editNotice.id, ...data, show: showSwitch ? 1 : 0, popup: popupSwitch ? 1 : 0 }, { onSuccess: () => { setDialog(null); setEditNotice(null) } })
  }

  const handleDelete = () => {
    if (deleteId !== null) deleteNotice.mutate(deleteId, { onSuccess: () => setDeleteId(null) })
  }

  const totalPages = data ? Math.ceil(data.total / data.page_size) : 1
  const filteredNotices = noticeSearch
    ? (data?.list ?? []).filter((n) => n.title.toLowerCase().includes(noticeSearch.toLowerCase()))
    : data?.list ?? []

  const handleSort = (notice: Notice, delta: number) => {
    updateNotice.mutate({ id: notice.id, sort: Math.max(0, (notice.sort ?? 0) + delta) })
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold">公告管理</h1>
        <Button onClick={openCreate}><Plus className="mr-2 h-4 w-4" />新建公告</Button>
      </div>

      <Card>
        <CardHeader>
          <div className="relative max-w-xs">
            <Search className="absolute left-2.5 top-2.5 h-4 w-4 text-muted-foreground" />
            <Input placeholder="搜索标题..." className="pl-8" value={noticeSearch} onChange={(e) => setNoticeSearch(e.target.value)} />
          </div>
        </CardHeader>
        <CardContent>
          {isLoading ? (
            <div className="space-y-2">{Array.from({ length: 5 }).map((_, i) => <Skeleton key={i} className="h-12" />)}</div>
          ) : (
            <>
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>标题</TableHead>
                    <TableHead>背景图片</TableHead>
                    <TableHead>标签</TableHead>
                    <TableHead>显示</TableHead>
                    <TableHead className="w-24">排序</TableHead>
                    <TableHead>创建时间</TableHead>
                    <TableHead className="text-right">操作</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {filteredNotices.map((n) => (
                    <TableRow key={n.id}>
                      <TableCell className="font-medium">{n.title}</TableCell>
                      <TableCell>
                        {n.img_url ? (
                          <img src={n.img_url} alt="背景" className="h-10 w-16 rounded object-cover" />
                        ) : (
                          <span className="text-muted-foreground text-xs">-</span>
                        )}
                      </TableCell>
                      <TableCell>
                        {n.tags ? (
                          <div className="flex flex-wrap gap-1">
                            {n.tags.split(',').filter(Boolean).map((tag) => (
                              <Badge key={tag} variant="secondary">{tag}</Badge>
                            ))}
                          </div>
                        ) : (
                          <span className="text-muted-foreground text-xs">-</span>
                        )}
                      </TableCell>
                      <TableCell>
                        <Badge variant={n.show === 1 ? 'success' : 'secondary'}>{n.show === 1 ? '显示' : '隐藏'}</Badge>
                      </TableCell>
                      <TableCell>
                        <div className="flex items-center gap-1">
                          <span className="text-sm">{n.sort}</span>
                          <Button variant="ghost" size="sm" className="h-5 w-5 p-0" onClick={() => handleSort(n, 1)}>
                            <ArrowUp className="h-3 w-3" />
                          </Button>
                          <Button variant="ghost" size="sm" className="h-5 w-5 p-0" onClick={() => handleSort(n, -1)}>
                            <ArrowDown className="h-3 w-3" />
                          </Button>
                        </div>
                      </TableCell>
                      <TableCell className="text-xs">{formatDate(n.created_at)}</TableCell>
                      <TableCell className="text-right space-x-1">
                        <Button size="icon" variant="ghost" onClick={() => openEdit(n)}>
                          <Pencil className="h-4 w-4" />
                        </Button>
                        <Button size="icon" variant="ghost" onClick={() => setDeleteId(n.id)}>
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

      <Dialog open={dialog !== null} onOpenChange={() => { setDialog(null); setEditNotice(null) }}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{dialog === 'create' ? '新建公告' : '编辑公告'}</DialogTitle>
          </DialogHeader>
          <form onSubmit={handleSubmit(dialog === 'create' ? handleCreate : handleEdit)} className="space-y-4">
            <div className="space-y-2">
              <Label>标题</Label>
              <Input {...register('title')} />
              {errors.title && <p className="text-xs text-destructive">{errors.title.message}</p>}
            </div>
            <div className="space-y-2">
              <Label>内容</Label>
              <Textarea {...register('content')} rows={6} />
              {errors.content && <p className="text-xs text-destructive">{errors.content.message}</p>}
            </div>
            <div className="space-y-2">
              <Label>背景图片URL</Label>
              <Input {...register('img_url')} placeholder="请输入背景图片URL" />
            </div>
            <div className="space-y-2">
              <Label>标签</Label>
              <Input
                value={tagInput}
                onChange={(e) => setTagInput(e.target.value)}
                onKeyDown={handleAddTag}
                placeholder="输入标签后回车添加"
              />
              {tags.length > 0 && (
                <div className="flex flex-wrap gap-2 mt-2">
                  {tags.map((tag) => (
                    <Badge key={tag} variant="secondary" className="gap-1">
                      {tag}
                      <button type="button" onClick={() => handleRemoveTag(tag)} className="ml-1 hover:text-destructive">
                        <X className="h-3 w-3" />
                      </button>
                    </Badge>
                  ))}
                </div>
              )}
            </div>
            <div className="flex items-center justify-between">
              <Label>显示</Label>
              <Switch checked={showSwitch} onCheckedChange={setShowSwitch} />
            </div>
            <div className="flex items-center justify-between">
              <Label>弹窗显示</Label>
              <Switch checked={popupSwitch} onCheckedChange={setPopupSwitch} />
            </div>
            <div className="space-y-2">
              <Label>排序</Label>
              <Input type="number" {...register('sort')} />
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
            <DialogDescription>确定要删除该公告吗？此操作不可撤销。</DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button variant="outline" onClick={() => setDeleteId(null)}>取消</Button>
            <Button variant="destructive" onClick={handleDelete} disabled={deleteNotice.isPending}>确认删除</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}
