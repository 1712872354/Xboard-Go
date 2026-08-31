import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { toast } from 'sonner'
import api from '@/lib/api'
import type { GiftCardTemplate, GiftCardCode, PaginatedResponse } from '@/types'
import { Card, CardContent, CardHeader } from '@/components/ui/card'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Skeleton } from '@/components/ui/skeleton'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { formatCurrency, formatDate } from '@/lib/utils'
import { Plus, Pencil, Trash2, Copy, Download, Gift } from 'lucide-react'

const useAdminGiftTemplates = (page = 1, pageSize = 20) =>
  useQuery({
    queryKey: ['admin', 'gift-templates', page, pageSize],
    queryFn: async () => (await api.get('/admin/gift-card-templates', { params: { page, page_size: pageSize } })) as unknown as PaginatedResponse<GiftCardTemplate>,
  })

const useAdminGiftCodes = (page = 1, pageSize = 20) =>
  useQuery({
    queryKey: ['admin', 'gift-codes', page, pageSize],
    queryFn: async () => (await api.get('/admin/gift-card-codes', { params: { page, page_size: pageSize } })) as unknown as PaginatedResponse<GiftCardCode>,
  })

const useCreateGiftTemplate = () => {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (data: Partial<GiftCardTemplate>) => await api.post('/admin/gift-card-templates', data),
    onSuccess: () => { qc.invalidateQueries({ queryKey: ['admin', 'gift-templates'] }); toast.success('模板创建成功') },
  })
}

const useUpdateGiftTemplate = () => {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async ({ id, ...data }: Partial<GiftCardTemplate> & { id: number }) => await api.put(`/admin/gift-card-templates/${id}`, data),
    onSuccess: () => { qc.invalidateQueries({ queryKey: ['admin', 'gift-templates'] }); toast.success('模板更新成功') },
  })
}

const useDeleteGiftTemplate = () => {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (id: number) => await api.delete(`/admin/gift-card-templates/${id}`),
    onSuccess: () => { qc.invalidateQueries({ queryKey: ['admin', 'gift-templates'] }); toast.success('模板删除成功') },
  })
}

const useGenerateCodes = () => {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (data: { template_id: number; count: number }) => await api.post('/admin/gift-card-codes/generate', data),
    onSuccess: () => { qc.invalidateQueries({ queryKey: ['admin', 'gift-codes'] }); toast.success('卡密生成成功') },
  })
}

const useDeleteGiftCode = () => {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (id: number) => await api.delete(`/admin/gift-card-codes/${id}`),
    onSuccess: () => { qc.invalidateQueries({ queryKey: ['admin', 'gift-codes'] }); toast.success('卡密删除成功') },
  })
}

const templateSchema = z.object({
  name: z.string().min(1, '请输入名称'),
  type: z.coerce.number().min(0),
  value: z.coerce.number().min(0),
  price: z.coerce.number().min(0),
  traffic: z.coerce.number().min(0).optional(),
  duration: z.coerce.number().min(0).optional(),
  plan_id: z.coerce.number().optional(),
})

type TemplateForm = z.infer<typeof templateSchema>

const typeMap: Record<number, string> = { 0: '余额', 1: '流量', 2: '套餐时长' }

export default function GiftCardsPage() {
  const [tmplPage, setTmplPage] = useState(1)
  const [codePage, setCodePage] = useState(1)
  const [dialog, setDialog] = useState<'create' | 'edit' | 'generate' | null>(null)
  const [editItem, setEditItem] = useState<GiftCardTemplate | null>(null)
  const [deleteId, setDeleteId] = useState<number | null>(null)
  const [genCount, setGenCount] = useState(10)
  const [genTemplateId, setGenTemplateId] = useState<string>('')
  const [typeValue, setTypeValue] = useState('0')

  const { data: templates, isLoading: loadingTmpl } = useAdminGiftTemplates(tmplPage)
  const { data: codes, isLoading: loadingCodes } = useAdminGiftCodes(codePage)
  const createTmpl = useCreateGiftTemplate()
  const updateTmpl = useUpdateGiftTemplate()
  const deleteTmpl = useDeleteGiftTemplate()
  const generateCodes = useGenerateCodes()
  const deleteCode = useDeleteGiftCode()

  const { register, handleSubmit, reset, formState: { errors } } = useForm<TemplateForm>({
    resolver: zodResolver(templateSchema),
  })

  const openCreate = () => {
    reset({ name: '', type: 0, value: 0, price: 0 })
    setTypeValue('0')
    setDialog('create')
  }

  const openEdit = (t: GiftCardTemplate) => {
    reset({ name: t.name, type: t.type, value: t.value, price: t.price, traffic: t.traffic, duration: t.duration, plan_id: t.plan_id })
    setTypeValue(String(t.type))
    setEditItem(t)
    setDialog('edit')
  }

  const handleCreate = (data: TemplateForm) => {
    createTmpl.mutate({ ...data, type: Number(typeValue) }, { onSuccess: () => setDialog(null) })
  }

  const handleEdit = (data: TemplateForm) => {
    if (!editItem) return
    updateTmpl.mutate({ id: editItem.id, ...data, type: Number(typeValue) }, { onSuccess: () => { setDialog(null); setEditItem(null) } })
  }

  const handleGenerate = () => {
    if (!genTemplateId) return
    generateCodes.mutate({ template_id: Number(genTemplateId), count: genCount }, { onSuccess: () => setDialog(null) })
  }

  const handleDelete = () => {
    if (deleteId !== null) deleteTmpl.mutate(deleteId, { onSuccess: () => setDeleteId(null) })
  }

  const copyCode = (code: string) => {
    navigator.clipboard.writeText(code)
    toast.success('已复制卡密')
  }

  const tmplTotalPages = templates ? Math.ceil(templates.total / templates.page_size) : 1
  const codeTotalPages = codes ? Math.ceil(codes.total / codes.page_size) : 1

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold">礼品卡管理</h1>
      </div>

      <Tabs defaultValue="templates">
        <TabsList>
          <TabsTrigger value="templates">模板</TabsTrigger>
          <TabsTrigger value="codes">卡密</TabsTrigger>
        </TabsList>

        <TabsContent value="templates">
          <Card>
            <CardHeader>
              <div className="flex justify-end">
                <Button onClick={openCreate}><Plus className="mr-2 h-4 w-4" />新建模板</Button>
              </div>
            </CardHeader>
            <CardContent>
              {loadingTmpl ? (
                <div className="space-y-2">{Array.from({ length: 5 }).map((_, i) => <Skeleton key={i} className="h-12" />)}</div>
              ) : (
                <>
                  <Table>
                    <TableHeader>
                      <TableRow>
                        <TableHead>名称</TableHead>
                        <TableHead>类型</TableHead>
                        <TableHead>值</TableHead>
                        <TableHead>价格</TableHead>
                        <TableHead>状态</TableHead>
                        <TableHead className="text-right">操作</TableHead>
                      </TableRow>
                    </TableHeader>
                    <TableBody>
                      {templates?.list.map((t) => (
                        <TableRow key={t.id}>
                          <TableCell className="font-medium">{t.name}</TableCell>
                          <TableCell><Badge variant="outline">{typeMap[t.type] ?? '未知'}</Badge></TableCell>
                          <TableCell>{formatCurrency(t.value)}</TableCell>
                          <TableCell>{formatCurrency(t.price)}</TableCell>
                          <TableCell>
                            <Badge variant={t.status === 1 ? 'success' : 'secondary'}>{t.status === 1 ? '启用' : '禁用'}</Badge>
                          </TableCell>
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
                    <p className="text-sm text-muted-foreground">共 {templates?.total ?? 0} 条</p>
                    <div className="flex gap-2">
                      <Button size="sm" variant="outline" disabled={tmplPage <= 1} onClick={() => setTmplPage(tmplPage - 1)}>上一页</Button>
                      <span className="flex items-center text-sm">{tmplPage} / {tmplTotalPages}</span>
                      <Button size="sm" variant="outline" disabled={tmplPage >= tmplTotalPages} onClick={() => setTmplPage(tmplPage + 1)}>下一页</Button>
                    </div>
                  </div>
                </>
              )}
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="codes">
          <Card>
            <CardHeader>
              <div className="flex justify-end gap-2">
                <Button variant="outline" onClick={() => setDialog('generate')}>
                  <Gift className="mr-2 h-4 w-4" />生成卡密
                </Button>
              </div>
            </CardHeader>
            <CardContent>
              {loadingCodes ? (
                <div className="space-y-2">{Array.from({ length: 8 }).map((_, i) => <Skeleton key={i} className="h-12" />)}</div>
              ) : (
                <>
                  <Table>
                    <TableHeader>
                      <TableRow>
                        <TableHead>卡密</TableHead>
                        <TableHead>状态</TableHead>
                        <TableHead>使用者</TableHead>
                        <TableHead>创建时间</TableHead>
                        <TableHead className="text-right">操作</TableHead>
                      </TableRow>
                    </TableHeader>
                    <TableBody>
                      {codes?.list.map((c) => (
                        <TableRow key={c.id}>
                          <TableCell>
                            <div className="flex items-center gap-1">
                              <span className="font-mono text-xs">{c.code}</span>
                              <Button size="icon" variant="ghost" className="h-5 w-5" onClick={() => copyCode(c.code)}>
                                <Copy className="h-3 w-3" />
                              </Button>
                            </div>
                          </TableCell>
                          <TableCell>
                            <Badge variant={c.status === 1 ? 'success' : c.status === 2 ? 'destructive' : 'secondary'}>
                              {c.status === 0 ? '未使用' : c.status === 1 ? '已使用' : '已禁用'}
                            </Badge>
                          </TableCell>
                          <TableCell>{c.user_id || '-'}</TableCell>
                          <TableCell className="text-xs">{formatDate(c.created_at)}</TableCell>
                          <TableCell className="text-right">
                            <Button size="icon" variant="ghost" onClick={() => deleteCode.mutate(c.id)}>
                              <Trash2 className="h-4 w-4 text-destructive" />
                            </Button>
                          </TableCell>
                        </TableRow>
                      ))}
                    </TableBody>
                  </Table>
                  <div className="mt-4 flex items-center justify-between">
                    <p className="text-sm text-muted-foreground">共 {codes?.total ?? 0} 条</p>
                    <div className="flex gap-2">
                      <Button size="sm" variant="outline" disabled={codePage <= 1} onClick={() => setCodePage(codePage - 1)}>上一页</Button>
                      <span className="flex items-center text-sm">{codePage} / {codeTotalPages}</span>
                      <Button size="sm" variant="outline" disabled={codePage >= codeTotalPages} onClick={() => setCodePage(codePage + 1)}>下一页</Button>
                    </div>
                  </div>
                </>
              )}
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>

      {/* Create / Edit template dialog */}
      <Dialog open={dialog === 'create' || dialog === 'edit'} onOpenChange={() => { setDialog(null); setEditItem(null) }}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{dialog === 'create' ? '新建模板' : '编辑模板'}</DialogTitle>
          </DialogHeader>
          <form onSubmit={handleSubmit(dialog === 'create' ? handleCreate : handleEdit)} className="space-y-4">
            <div className="space-y-2">
              <Label>名称</Label>
              <Input {...register('name')} />
              {errors.name && <p className="text-xs text-destructive">{errors.name.message}</p>}
            </div>
            <div className="grid grid-cols-3 gap-4">
              <div className="space-y-2">
                <Label>类型</Label>
                <Select value={typeValue} onValueChange={setTypeValue}>
                  <SelectTrigger><SelectValue /></SelectTrigger>
                  <SelectContent>
                    <SelectItem value="0">余额</SelectItem>
                    <SelectItem value="1">流量</SelectItem>
                    <SelectItem value="2">套餐时长</SelectItem>
                  </SelectContent>
                </Select>
              </div>
              <div className="space-y-2">
                <Label>值</Label>
                <Input type="number" {...register('value')} />
              </div>
              <div className="space-y-2">
                <Label>价格</Label>
                <Input type="number" {...register('price')} />
              </div>
            </div>
            <DialogFooter>
              <Button type="button" variant="outline" onClick={() => setDialog(null)}>取消</Button>
              <Button type="submit">保存</Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>

      {/* Generate codes dialog */}
      <Dialog open={dialog === 'generate'} onOpenChange={() => setDialog(null)}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>生成卡密</DialogTitle>
            <DialogDescription>选择模板并输入生成数量</DialogDescription>
          </DialogHeader>
          <div className="space-y-4">
            <div className="space-y-2">
              <Label>模板</Label>
              <Select value={genTemplateId} onValueChange={setGenTemplateId}>
                <SelectTrigger><SelectValue placeholder="选择模板" /></SelectTrigger>
                <SelectContent>
                  {templates?.list.map((t) => <SelectItem key={t.id} value={String(t.id)}>{t.name}</SelectItem>)}
                </SelectContent>
              </Select>
            </div>
            <div className="space-y-2">
              <Label>数量</Label>
              <Input type="number" value={genCount} onChange={(e) => setGenCount(Number(e.target.value))} min={1} max={1000} />
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setDialog(null)}>取消</Button>
            <Button onClick={handleGenerate} disabled={generateCodes.isPending || !genTemplateId}>生成</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Delete dialog */}
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
    </div>
  )
}
