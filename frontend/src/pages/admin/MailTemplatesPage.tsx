import { useState, useRef, useEffect } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useForm, useWatch } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { toast } from 'sonner'
import api from '@/lib/api'
import type { MailTemplate, PaginatedResponse } from '@/types'
import { Card, CardContent, CardHeader } from '@/components/ui/card'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Skeleton } from '@/components/ui/skeleton'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/tooltip'
import { formatDate } from '@/lib/utils'
import { Plus, Pencil, Trash2, Copy, Loader2 } from 'lucide-react'

const useAdminMailTemplates = (page = 1, pageSize = 20) =>
  useQuery({
    queryKey: ['admin', 'mail-templates', page, pageSize],
    queryFn: async () => (await api.get('/admin/mail-templates', { params: { page, page_size: pageSize } })) as unknown as PaginatedResponse<MailTemplate>,
  })

const useCreateMailTemplate = () => {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (data: Partial<MailTemplate>) => await api.post('/admin/mail-templates', data),
    onSuccess: () => { qc.invalidateQueries({ queryKey: ['admin', 'mail-templates'] }); toast.success('模板创建成功') },
  })
}

const useUpdateMailTemplate = () => {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async ({ id, ...data }: Partial<MailTemplate> & { id: number }) => await api.put(`/admin/mail-templates/${id}`, data),
    onSuccess: () => { qc.invalidateQueries({ queryKey: ['admin', 'mail-templates'] }); toast.success('模板更新成功') },
  })
}

const useDeleteMailTemplate = () => {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (id: number) => await api.delete(`/admin/mail-templates/${id}`),
    onSuccess: () => { qc.invalidateQueries({ queryKey: ['admin', 'mail-templates'] }); toast.success('模板删除成功') },
  })
}

const useResetMailTemplate = () => {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (id: number) => await api.put(`/admin/mail-templates/${id}/reset`),
    onSuccess: () => { qc.invalidateQueries({ queryKey: ['admin', 'mail-templates'] }); toast.success('已恢复默认模板') },
  })
}

const useSendTestMail = () => {
  return useMutation({
    mutationFn: async ({ id, email }: { id: number; email?: string }) => await api.post(`/admin/mail-templates/${id}/test`, { email }),
    onSuccess: () => { toast.success('测试邮件已发送') },
  })
}

const templateSchema = z.object({
  name: z.string().min(1, '请输入名称'),
  subject: z.string().min(1, '请输入主题'),
  body: z.string().min(1, '请输入内容'),
  language: z.string().min(1, '请选择语言'),
  remark: z.string().optional(),
})

type TemplateForm = z.infer<typeof templateSchema>

const languages = ['zh-CN', 'en-US']
const templateVars = ['{{user.email}}', '{{user.name}}', '{{plan.name}}', '{{order.trade_no}}', '{{order.amount}}', '{{site.name}}', '{{site.url}}', '{{subscribe.url}}']

export default function MailTemplatesPage() {
  const [page, setPage] = useState(1)
  const [dialog, setDialog] = useState<'create' | 'edit' | null>(null)
  const [editItem, setEditItem] = useState<MailTemplate | null>(null)
  const [deleteId, setDeleteId] = useState<number | null>(null)
  const [resetId, setResetId] = useState<number | null>(null)
  const [testEmailDialog, setTestEmailDialog] = useState(false)
  const [testEmail, setTestEmail] = useState('')
  const [langValue, setLangValue] = useState('zh-CN')
  const previewRef = useRef<HTMLIFrameElement>(null)

  const { data, isLoading } = useAdminMailTemplates(page)
  const createTmpl = useCreateMailTemplate()
  const updateTmpl = useUpdateMailTemplate()
  const deleteTmpl = useDeleteMailTemplate()
  const resetTmpl = useResetMailTemplate()
  const sendTestMail = useSendTestMail()

  const { register, handleSubmit, reset, control, formState: { errors } } = useForm<TemplateForm>({
    resolver: zodResolver(templateSchema),
  })

  const bodyValue = useWatch({ control, name: 'body' })

  useEffect(() => {
    if (previewRef.current) {
      const doc = previewRef.current.contentDocument
      if (doc) {
        doc.open()
        doc.write(bodyValue || '')
        doc.close()
      }
    }
  }, [bodyValue])

  const openCreate = () => {
    reset({ name: '', subject: '', body: '', language: 'zh-CN', remark: '' })
    setLangValue('zh-CN')
    setDialog('create')
  }

  const openEdit = (t: MailTemplate) => {
    reset({ name: t.name, subject: t.subject, body: t.body, language: t.language, remark: t.remark })
    setLangValue(t.language)
    setEditItem(t)
    setDialog('edit')
  }

  const handleCreate = (data: TemplateForm) => {
    createTmpl.mutate({ ...data, language: langValue }, { onSuccess: () => setDialog(null) })
  }

  const handleEdit = (data: TemplateForm) => {
    if (!editItem) return
    updateTmpl.mutate({ id: editItem.id, ...data, language: langValue }, { onSuccess: () => { setDialog(null); setEditItem(null) } })
  }

  const handleDelete = () => {
    if (deleteId !== null) deleteTmpl.mutate(deleteId, { onSuccess: () => setDeleteId(null) })
  }

  const handleReset = () => {
    if (resetId !== null) resetTmpl.mutate(resetId, { onSuccess: () => { setResetId(null); setDialog(null); setEditItem(null) } })
  }

  const copyVar = (v: string) => {
    navigator.clipboard.writeText(v)
    toast.success(`已复制 ${v}`)
  }

  const handleSendTest = () => {
    if (!editItem) return
    sendTestMail.mutate(
      { id: editItem.id, email: testEmail || undefined },
      { onSuccess: () => { setTestEmailDialog(false); setTestEmail('') } }
    )
  }

  const totalPages = data ? Math.ceil(data.total / data.page_size) : 1

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold">邮件模板</h1>
        <Button onClick={openCreate}><Plus className="mr-2 h-4 w-4" />新建模板</Button>
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
                    <TableHead>名称</TableHead>
                    <TableHead>主题</TableHead>
                    <TableHead>语言</TableHead>
                    <TableHead>备注</TableHead>
                    <TableHead>创建时间</TableHead>
                    <TableHead className="text-right">操作</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {data?.list.map((t) => (
                    <TableRow key={t.id}>
                      <TableCell className="font-medium">{t.name}</TableCell>
                      <TableCell>{t.subject}</TableCell>
                      <TableCell><Badge variant="outline">{t.language}</Badge></TableCell>
                      <TableCell className="text-muted-foreground">{t.remark || '-'}</TableCell>
                      <TableCell className="text-xs">{formatDate(t.created_at)}</TableCell>
                      <TableCell className="text-right space-x-1">
                        <Button size="icon" variant="ghost" onClick={() => openEdit(t)}>
                          <Pencil className="h-4 w-4" />
                        </Button>
                        <Button size="icon" variant="ghost" onClick={() => setDeleteId(t.id)}>
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
        <DialogContent className="max-w-4xl">
          <DialogHeader>
            <DialogTitle>{dialog === 'create' ? '新建模板' : '编辑模板'}</DialogTitle>
          </DialogHeader>
          <form onSubmit={handleSubmit(dialog === 'create' ? handleCreate : handleEdit)} className="space-y-4">
            <div className="grid grid-cols-2 gap-4">
              <div className="space-y-2">
                <Label>名称</Label>
                <Input {...register('name')} />
                {errors.name && <p className="text-xs text-destructive">{errors.name.message}</p>}
              </div>
              <div className="space-y-2">
                <Label>语言</Label>
                <Select value={langValue} onValueChange={setLangValue}>
                  <SelectTrigger><SelectValue /></SelectTrigger>
                  <SelectContent>{languages.map((l) => <SelectItem key={l} value={l}>{l}</SelectItem>)}</SelectContent>
                </Select>
              </div>
            </div>
            <div className="space-y-2">
              <Label>主题</Label>
              <Input {...register('subject')} />
              {errors.subject && <p className="text-xs text-destructive">{errors.subject.message}</p>}
            </div>
            <div className="space-y-2">
              <Label>模板变量</Label>
              <div className="flex flex-wrap gap-1">
                {templateVars.map((v) => (
                  <Tooltip key={v}>
                    <TooltipTrigger asChild>
                      <Badge variant="outline" className="cursor-pointer text-xs" onClick={() => copyVar(v)}>
                        <Copy className="mr-1 h-2.5 w-2.5" />{v}
                      </Badge>
                    </TooltipTrigger>
                    <TooltipContent>点击复制</TooltipContent>
                  </Tooltip>
                ))}
              </div>
            </div>
            <div className="space-y-2">
              <Label>内容 (HTML)</Label>
              <div className="grid grid-cols-2 gap-4">
                <Textarea {...register('body')} rows={14} className="font-mono text-xs" />
                <div className="rounded-md border bg-white">
                  <div className="border-b px-3 py-1.5 text-xs font-medium text-muted-foreground">实时预览</div>
                  <iframe
                    ref={previewRef}
                    sandbox=""
                    className="h-[340px] w-full"
                    title="邮件预览"
                  />
                </div>
              </div>
              {errors.body && <p className="text-xs text-destructive">{errors.body.message}</p>}
            </div>
            <div className="space-y-2">
              <Label>备注</Label>
              <Input {...register('remark')} />
            </div>
            <DialogFooter className="flex-row justify-between">
              {dialog === 'edit' && editItem && (
                <div className="flex gap-2">
                  <Button type="button" variant="destructive" size="sm" onClick={() => setResetId(editItem.id)}>
                    恢复默认
                  </Button>
                  <Button type="button" variant="outline" size="sm" onClick={() => setTestEmailDialog(true)}>
                    发送测试
                  </Button>
                </div>
              )}
              <div className="flex gap-2 ml-auto">
                <Button type="button" variant="outline" onClick={() => setDialog(null)}>取消</Button>
                <Button type="submit" disabled={createTmpl.isPending || updateTmpl.isPending}>
                  {(createTmpl.isPending || updateTmpl.isPending) && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
                  保存
                </Button>
              </div>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>

      <Dialog open={deleteId !== null} onOpenChange={() => setDeleteId(null)}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>确认删除</DialogTitle>
            <DialogDescription>确定要删除该模板吗？此操作不可撤销。</DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button variant="outline" onClick={() => setDeleteId(null)}>取消</Button>
            <Button variant="destructive" onClick={handleDelete} disabled={deleteTmpl.isPending}>确认删除</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <Dialog open={resetId !== null} onOpenChange={() => setResetId(null)}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>恢复默认模板</DialogTitle>
            <DialogDescription>确定要将该模板恢复为系统默认内容吗？当前修改将被覆盖。</DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button variant="outline" onClick={() => setResetId(null)}>取消</Button>
            <Button variant="destructive" onClick={handleReset} disabled={resetTmpl.isPending}>确认恢复</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <Dialog open={testEmailDialog} onOpenChange={() => { setTestEmailDialog(false); setTestEmail('') }}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>发送测试邮件</DialogTitle>
            <DialogDescription>输入测试邮箱地址，留空将发送到管理员邮箱。</DialogDescription>
          </DialogHeader>
          <div className="space-y-2">
            <Label>邮箱地址</Label>
            <Input
              type="email"
              placeholder="留空则发送到管理员邮箱"
              value={testEmail}
              onChange={(e) => setTestEmail(e.target.value)}
            />
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => { setTestEmailDialog(false); setTestEmail('') }}>取消</Button>
            <Button onClick={handleSendTest} disabled={sendTestMail.isPending}>发送</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}
