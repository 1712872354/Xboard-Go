import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { toast } from 'sonner'
import api from '@/lib/api'
import type { Coupon, PaginatedResponse } from '@/types'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Skeleton } from '@/components/ui/skeleton'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { formatDate, formatCurrency } from '@/lib/utils'
import { Plus, Pencil, Trash2, Copy, Ticket, CheckCircle, Tag } from 'lucide-react'

const useAdminCoupons = (page = 1, pageSize = 20) =>
  useQuery({
    queryKey: ['admin', 'coupons', page, pageSize],
    queryFn: async () => (await api.get('/admin/coupons', { params: { page, page_size: pageSize } })) as unknown as PaginatedResponse<Coupon>,
  })

const useCreateCoupon = () => {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (data: Partial<Coupon>) => await api.post('/admin/coupons', data),
    onSuccess: () => { qc.invalidateQueries({ queryKey: ['admin', 'coupons'] }); toast.success('优惠券创建成功') },
  })
}

const useUpdateCoupon = () => {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async ({ id, ...data }: Partial<Coupon> & { id: number }) => await api.put(`/admin/coupons/${id}`, data),
    onSuccess: () => { qc.invalidateQueries({ queryKey: ['admin', 'coupons'] }); toast.success('优惠券更新成功') },
  })
}

const useDeleteCoupon = () => {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (id: number) => await api.delete(`/admin/coupons/${id}`),
    onSuccess: () => { qc.invalidateQueries({ queryKey: ['admin', 'coupons'] }); toast.success('优惠券删除成功') },
  })
}

const couponSchema = z.object({
  code: z.string().min(1, '请输入优惠码'),
  name: z.string().min(1, '请输入名称'),
  type: z.coerce.number().min(0),
  value: z.coerce.number().min(0),
  min_amount: z.coerce.number().min(0).optional(),
  max_discount: z.coerce.number().min(0).optional(),
  limit_count: z.coerce.number().min(0).optional(),
  start_date: z.string().optional(),
  end_date: z.string().optional(),
})

type CouponForm = z.infer<typeof couponSchema>

export default function CouponsPage() {
  const [page, setPage] = useState(1)
  const [dialog, setDialog] = useState<'create' | 'edit' | null>(null)
  const [editItem, setEditItem] = useState<Coupon | null>(null)
  const [deleteId, setDeleteId] = useState<number | null>(null)
  const [typeValue, setTypeValue] = useState('0')

  const { data, isLoading } = useAdminCoupons(page)
  const createCoupon = useCreateCoupon()
  const updateCoupon = useUpdateCoupon()
  const deleteCoupon = useDeleteCoupon()

  const coupons = data?.list ?? []
  const totalCoupons = data?.total ?? 0
  const usedCount = coupons.reduce((sum, c) => sum + c.used_count, 0)
  const availableCount = coupons.filter((c) => c.status === 1).length

  const { register, handleSubmit, reset, formState: { errors } } = useForm<CouponForm>({
    resolver: zodResolver(couponSchema),
  })

  const openCreate = () => {
    reset({ code: '', name: '', type: 0, value: 0, min_amount: 0, max_discount: 0, limit_count: 1, start_date: '', end_date: '' })
    setTypeValue('0')
    setDialog('create')
  }

  const openEdit = (c: Coupon) => {
    reset({ code: c.code, name: c.name, type: c.type, value: c.value, min_amount: c.min_amount, max_discount: c.max_discount, limit_count: c.limit_count, start_date: c.start_date, end_date: c.end_date })
    setTypeValue(String(c.type))
    setEditItem(c)
    setDialog('edit')
  }

  const handleCreate = (data: CouponForm) => {
    createCoupon.mutate({ ...data, type: Number(typeValue) }, { onSuccess: () => setDialog(null) })
  }

  const handleEdit = (data: CouponForm) => {
    if (!editItem) return
    updateCoupon.mutate({ id: editItem.id, ...data, type: Number(typeValue) }, { onSuccess: () => { setDialog(null); setEditItem(null) } })
  }

  const handleDelete = () => {
    if (deleteId !== null) deleteCoupon.mutate(deleteId, { onSuccess: () => setDeleteId(null) })
  }

  const copyCode = (code: string) => {
    navigator.clipboard.writeText(code)
    toast.success('已复制优惠码')
  }

  const totalPages = data ? Math.ceil(data.total / data.page_size) : 1

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold">优惠券管理</h1>
        <Button onClick={openCreate}><Plus className="mr-2 h-4 w-4" />创建优惠券</Button>
      </div>

      <div className="grid gap-4 sm:grid-cols-3">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">总计</CardTitle>
            <Ticket className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent><div className="text-2xl font-bold">{totalCoupons}</div></CardContent>
        </Card>
        <Card>
          <CardHeader className="flex flex-row items-center justify-between pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">已使用</CardTitle>
            <CheckCircle className="h-4 w-4 text-emerald-500" />
          </CardHeader>
          <CardContent><div className="text-2xl font-bold">{usedCount}</div></CardContent>
        </Card>
        <Card>
          <CardHeader className="flex flex-row items-center justify-between pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">可用</CardTitle>
            <Tag className="h-4 w-4 text-blue-500" />
          </CardHeader>
          <CardContent><div className="text-2xl font-bold">{availableCount}</div></CardContent>
        </Card>
      </div>

      <Card>
        <CardHeader />
        <CardContent>
          {isLoading ? (
            <div className="space-y-2">{Array.from({ length: 8 }).map((_, i) => <Skeleton key={i} className="h-12" />)}</div>
          ) : (
            <>
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>优惠码</TableHead>
                    <TableHead>名称</TableHead>
                    <TableHead>类型</TableHead>
                    <TableHead>值</TableHead>
                    <TableHead>使用/限制</TableHead>
                    <TableHead>有效期</TableHead>
                    <TableHead>状态</TableHead>
                    <TableHead className="text-right">操作</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {coupons.map((c) => (
                    <TableRow key={c.id}>
                      <TableCell>
                        <div className="flex items-center gap-1">
                          <span className="font-mono text-xs">{c.code}</span>
                          <Button size="icon" variant="ghost" className="h-5 w-5" onClick={() => copyCode(c.code)}>
                            <Copy className="h-3 w-3" />
                          </Button>
                        </div>
                      </TableCell>
                      <TableCell>{c.name}</TableCell>
                      <TableCell>
                        <Badge variant={c.type === 0 ? 'default' : 'secondary'}>{c.type === 0 ? '固定金额' : '百分比'}</Badge>
                      </TableCell>
                      <TableCell>{c.type === 0 ? formatCurrency(c.value) : `${c.value}%`}</TableCell>
                      <TableCell>{c.used_count} / {c.limit_count || '不限'}</TableCell>
                      <TableCell className="text-xs">
                        {c.start_date ? `${c.start_date} ~ ${c.end_date}` : '永久'}
                      </TableCell>
                      <TableCell>
                        <Badge variant={c.status === 1 ? 'success' : 'secondary'}>{c.status === 1 ? '可用' : '禁用'}</Badge>
                      </TableCell>
                      <TableCell className="text-right space-x-1">
                        <Button size="icon" variant="ghost" onClick={() => openEdit(c)}>
                          <Pencil className="h-4 w-4" />
                        </Button>
                        <Button size="icon" variant="ghost" onClick={() => setDeleteId(c.id)}>
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
            <DialogTitle>{dialog === 'create' ? '创建优惠券' : '编辑优惠券'}</DialogTitle>
          </DialogHeader>
          <form onSubmit={handleSubmit(dialog === 'create' ? handleCreate : handleEdit)} className="space-y-4">
            <div className="grid grid-cols-2 gap-4">
              <div className="space-y-2">
                <Label>优惠码</Label>
                <Input {...register('code')} />
                {errors.code && <p className="text-xs text-destructive">{errors.code.message}</p>}
              </div>
              <div className="space-y-2">
                <Label>名称</Label>
                <Input {...register('name')} />
                {errors.name && <p className="text-xs text-destructive">{errors.name.message}</p>}
              </div>
            </div>
            <div className="grid grid-cols-2 gap-4">
              <div className="space-y-2">
                <Label>类型</Label>
                <Select value={typeValue} onValueChange={setTypeValue}>
                  <SelectTrigger><SelectValue /></SelectTrigger>
                  <SelectContent>
                    <SelectItem value="0">固定金额</SelectItem>
                    <SelectItem value="1">百分比</SelectItem>
                  </SelectContent>
                </Select>
              </div>
              <div className="space-y-2">
                <Label>{typeValue === '0' ? '金额 (分)' : '百分比 (%)'}</Label>
                <Input type="number" {...register('value')} />
              </div>
            </div>
            <div className="grid grid-cols-3 gap-4">
              <div className="space-y-2">
                <Label>最低消费</Label>
                <Input type="number" {...register('min_amount')} />
              </div>
              <div className="space-y-2">
                <Label>最大折扣</Label>
                <Input type="number" {...register('max_discount')} />
              </div>
              <div className="space-y-2">
                <Label>使用限制</Label>
                <Input type="number" {...register('limit_count')} />
              </div>
            </div>
            <div className="grid grid-cols-2 gap-4">
              <div className="space-y-2">
                <Label>开始日期</Label>
                <Input type="date" {...register('start_date')} />
              </div>
              <div className="space-y-2">
                <Label>结束日期</Label>
                <Input type="date" {...register('end_date')} />
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
            <DialogDescription>确定要删除该优惠券吗？此操作不可撤销。</DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button variant="outline" onClick={() => setDeleteId(null)}>取消</Button>
            <Button variant="destructive" onClick={handleDelete} disabled={deleteCoupon.isPending}>确认删除</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}
