import { useState, useCallback } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { toast } from 'sonner'
import { useAdminPlans, useCreatePlan, useUpdatePlan, useDeletePlan } from '@/hooks/usePlans'
import { Card, CardContent, CardHeader } from '@/components/ui/card'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import { Skeleton } from '@/components/ui/skeleton'
import { Switch } from '@/components/ui/switch'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { formatCurrency } from '@/lib/utils'
import { Plus, Pencil, Trash2, X } from 'lucide-react'
import type { Plan } from '@/types'

const planSchema = z.object({
  name: z.string().min(1, '请输入名称'),
  description: z.string().optional(),
  content: z.string().optional(),
  price: z.coerce.number().min(0, '价格不能为负'),
  duration_days: z.coerce.number().min(1, '天数至少为1'),
  traffic: z.coerce.number().min(0),
  device_limit: z.coerce.number().min(0),
  speed_limit: z.coerce.number().min(0).optional(),
  node_group: z.string().optional(),
  show: z.coerce.number().optional().default(1),
  sell: z.coerce.number().optional().default(1),
  renew: z.coerce.number().optional().default(1),
  capacity_limit: z.coerce.number().min(0).optional(),
  tags: z.string().optional(),
  reset_traffic_method: z.coerce.number().optional(),
})

type PlanForm = z.infer<typeof planSchema>

function PlanFormDialog({
  open, onOpenChange, onSubmit, defaultValues, title,
}: {
  open: boolean
  onOpenChange: (v: boolean) => void
  onSubmit: (data: PlanForm) => void
  defaultValues?: Partial<PlanForm>
  title: string
}) {
  const { register, handleSubmit, reset, setValue, watch, formState: { errors } } = useForm<PlanForm>({
    resolver: zodResolver(planSchema),
    defaultValues,
  })

  const tagsValue = watch('tags') ?? ''
  const resetTrafficMethod = watch('reset_traffic_method')

  const [tagInput, setTagInput] = useState('')

  const handleAddTag = useCallback((e: React.KeyboardEvent<HTMLInputElement>) => {
    if (e.key === 'Enter') {
      e.preventDefault()
      const val = tagInput.trim()
      if (!val) return
      const current = (watch('tags') ?? '').split(',').filter(Boolean)
      if (!current.includes(val)) {
        setValue('tags', [...current, val].join(','))
      }
      setTagInput('')
    }
  }, [tagInput, setValue, watch])

  const handleRemoveTag = useCallback((tag: string) => {
    const current = (watch('tags') ?? '').split(',').filter(Boolean)
    setValue('tags', current.filter(t => t !== tag).join(','))
  }, [setValue, watch])

  const handleFormSubmit = (data: PlanForm) => {
    onSubmit(data)
    reset()
  }

  return (
    <Dialog open={open} onOpenChange={(v) => { onOpenChange(v); if (!v) reset() }}>
      <DialogContent className="max-w-lg">
        <DialogHeader>
          <DialogTitle>{title}</DialogTitle>
          <DialogDescription>填写套餐信息</DialogDescription>
        </DialogHeader>
        <form onSubmit={handleSubmit(handleFormSubmit)} className="space-y-4">
          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-2">
              <Label>名称</Label>
              <Input {...register('name')} />
              {errors.name && <p className="text-xs text-destructive">{errors.name.message}</p>}
            </div>
            <div className="space-y-2">
              <Label>价格 (分)</Label>
              <Input type="number" {...register('price')} />
              {errors.price && <p className="text-xs text-destructive">{errors.price.message}</p>}
            </div>
          </div>
          <div className="space-y-2">
            <Label>描述</Label>
            <Textarea {...register('description')} rows={2} />
          </div>
          <div className="space-y-2">
            <Label>套餐说明</Label>
            <Textarea {...register('content')} rows={3} placeholder="套餐详细说明，对用户展示" />
          </div>
          <div className="grid grid-cols-3 gap-4">
            <div className="space-y-2">
              <Label>天数</Label>
              <Input type="number" {...register('duration_days')} />
            </div>
            <div className="space-y-2">
              <Label>流量 (字节)</Label>
              <Input type="number" {...register('traffic')} />
            </div>
            <div className="space-y-2">
              <Label>设备限制</Label>
              <Input type="number" {...register('device_limit')} />
            </div>
          </div>
          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-2">
              <Label>节点组</Label>
              <Input {...register('node_group')} />
            </div>
            <div className="space-y-2">
              <Label>限速 (Mbps)</Label>
              <Input type="number" {...register('speed_limit')} placeholder="0 表示不限" />
            </div>
          </div>
          <div className="space-y-2">
            <Label>容量限制</Label>
            <Input type="number" {...register('capacity_limit')} placeholder="0 = 无限制" />
          </div>
          <div className="space-y-2">
            <Label>流量重置方式</Label>
            <Select value={resetTrafficMethod != null ? String(resetTrafficMethod) : '0'} onValueChange={(v) => setValue('reset_traffic_method', Number(v))}>
              <SelectTrigger>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="0">跟随系统</SelectItem>
                <SelectItem value="1">每月1号</SelectItem>
                <SelectItem value="2">按月重置（按购买日）</SelectItem>
                <SelectItem value="3">不重置</SelectItem>
                <SelectItem value="4">每年1月1号</SelectItem>
                <SelectItem value="5">按年重置（按购买日）</SelectItem>
              </SelectContent>
            </Select>
          </div>
          <div className="space-y-2">
            <Label>标签</Label>
            <div className="flex flex-wrap gap-1.5 mb-1">
              {tagsValue.split(',').filter(Boolean).map((tag) => (
                <Badge key={tag} variant="secondary" className="gap-1 pr-1">
                  {tag}
                  <button type="button" onClick={() => handleRemoveTag(tag)} className="ml-0.5 rounded-full hover:bg-muted-foreground/20">
                    <X className="h-3 w-3" />
                  </button>
                </Badge>
              ))}
            </div>
            <Input
              value={tagInput}
              onChange={(e) => setTagInput(e.target.value)}
              onKeyDown={handleAddTag}
              placeholder="输入标签后回车添加"
            />
          </div>
          <DialogFooter>
            <Button type="button" variant="outline" onClick={() => onOpenChange(false)}>取消</Button>
            <Button type="submit">保存</Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}

export default function PlansPage() {
  const [page, setPage] = useState(1)
  const [dialog, setDialog] = useState<'create' | 'edit' | null>(null)
  const [editPlan, setEditPlan] = useState<Plan | null>(null)
  const [deleteId, setDeleteId] = useState<number | null>(null)

  const { data, isLoading } = useAdminPlans(page)
  const createPlan = useCreatePlan()
  const updatePlan = useUpdatePlan()
  const deletePlan = useDeletePlan()

  const handleCreate = (data: PlanForm) => {
    createPlan.mutate(data, { onSuccess: () => { toast.success('套餐创建成功'); setDialog(null) } })
  }

  const handleEdit = (data: PlanForm) => {
    if (!editPlan) return
    updatePlan.mutate({ id: editPlan.id, ...data }, { onSuccess: () => { toast.success('套餐更新成功'); setDialog(null); setEditPlan(null) } })
  }

  const handleDelete = () => {
    if (deleteId !== null) {
      deletePlan.mutate(deleteId, { onSuccess: () => setDeleteId(null) })
    }
  }

  const totalPages = data ? Math.ceil(data.total / data.page_size) : 1

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold">套餐管理</h1>
        <Button onClick={() => setDialog('create')}><Plus className="mr-2 h-4 w-4" />新建套餐</Button>
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
                    <TableHead>价格</TableHead>
                    <TableHead>天数</TableHead>
                    <TableHead>流量</TableHead>
                    <TableHead className="text-center">显示</TableHead>
                    <TableHead className="text-center">新购</TableHead>
                    <TableHead className="text-center">续费</TableHead>
                    <TableHead>状态</TableHead>
                    <TableHead className="text-right">操作</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {data?.list.map((p) => (
                    <TableRow key={p.id}>
                      <TableCell className="font-medium">{p.name}</TableCell>
                      <TableCell>{formatCurrency(p.price)}</TableCell>
                      <TableCell>{p.duration_days} 天</TableCell>
                      <TableCell>{p.traffic ? `${(p.traffic / 1073741824).toFixed(0)} GB` : '不限'}</TableCell>
                      <TableCell className="text-center">
                        <Switch
                          checked={p.show === 1}
                          onCheckedChange={(v) => updatePlan.mutate({ id: p.id, show: v ? 1 : 0 })}
                          aria-label="显示"
                        />
                      </TableCell>
                      <TableCell className="text-center">
                        <Switch
                          checked={p.sell === 1}
                          onCheckedChange={(v) => updatePlan.mutate({ id: p.id, sell: v ? 1 : 0 })}
                          aria-label="新购"
                        />
                      </TableCell>
                      <TableCell className="text-center">
                        <Switch
                          checked={p.renew === 1}
                          onCheckedChange={(v) => updatePlan.mutate({ id: p.id, renew: v ? 1 : 0 })}
                          aria-label="续费"
                        />
                      </TableCell>
                      <TableCell>
                        <Badge variant={p.status === 1 ? 'success' : 'secondary'}>{p.status === 1 ? '启用' : '禁用'}</Badge>
                      </TableCell>
                      <TableCell className="text-right space-x-1">
                        <Button size="icon" variant="ghost" onClick={() => { setEditPlan(p); setDialog('edit') }}>
                          <Pencil className="h-4 w-4" />
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

      <PlanFormDialog open={dialog === 'create'} onOpenChange={() => setDialog(null)} onSubmit={handleCreate} title="新建套餐" />

      <PlanFormDialog
        open={dialog === 'edit' && !!editPlan}
        onOpenChange={(v) => { if (!v) { setDialog(null); setEditPlan(null) } }}
        onSubmit={handleEdit}
        defaultValues={editPlan ?? {}}
        title="编辑套餐"
      />

      <Dialog open={deleteId !== null} onOpenChange={() => setDeleteId(null)}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>确认删除</DialogTitle>
            <DialogDescription>确定要删除该套餐吗？此操作不可撤销。</DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button variant="outline" onClick={() => setDeleteId(null)}>取消</Button>
            <Button variant="destructive" onClick={handleDelete} disabled={deletePlan.isPending}>确认删除</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}
