import { useState } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { toast } from 'sonner'
import {
  useAdminUsers, useUpdateUserStatus, useDeleteUser, useAdminUpdateUser, useGenerateUsers, useExportUsers,
} from '@/hooks/useAdminUsers'
import { usePlans } from '@/hooks/usePlans'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Skeleton } from '@/components/ui/skeleton'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { AlertDialog, AlertDialogAction, AlertDialogCancel, AlertDialogContent, AlertDialogDescription, AlertDialogFooter, AlertDialogHeader, AlertDialogTitle } from '@/components/ui/alert-dialog'
import { Checkbox } from '@/components/ui/checkbox'
import { formatBytes, formatDate, formatCurrency } from '@/lib/utils'
import { Textarea } from '@/components/ui/textarea'
import { Search, Trash2, Ban, CheckCircle, Pencil, Download, Users, Loader2, Eye, SlidersHorizontal, RotateCcw } from 'lucide-react'

const editSchema = z.object({
  email: z.string().email('请输入有效的邮箱').optional().or(z.literal('')),
  password: z.string().min(6, '密码至少6位').optional().or(z.literal('')),
  traffic_limit: z.number().min(0).optional(),
  expired_at: z.string().optional(),
  plan_id: z.number().optional(),
  balance: z.number().optional(),
  commission: z.number().optional(),
  phone: z.string().optional(),
  remarks: z.string().optional(),
  device_limit: z.number().min(0).optional(),
  speed_limit: z.number().min(0).optional(),
  discount: z.number().min(0).max(100).optional(),
})

type EditFormData = z.infer<typeof editSchema>

const generateSchema = z.object({
  count: z.number().min(1).max(1000),
  prefix: z.string().optional(),
  password: z.string().min(6).optional().or(z.literal('')),
  plan_id: z.number().optional(),
  expired_at: z.string().optional(),
})

type GenerateFormData = z.infer<typeof generateSchema>

export default function UsersPage() {
  const [page, setPage] = useState(1)
  const [keyword, setKeyword] = useState('')
  const [search, setSearch] = useState('')
  const [deleteId, setDeleteId] = useState<number | null>(null)
  const [editUser, setEditUser] = useState<any>(null)
  const [detailUser, setDetailUser] = useState<any>(null)
  const [showGenerate, setShowGenerate] = useState(false)
  const [selectedUsers, setSelectedUsers] = useState<Set<number>>(new Set())
  const [showBatchBan, setShowBatchBan] = useState(false)
  const [showFilters, setShowFilters] = useState(false)
  const [filters, setFilters] = useState<{ status?: number; role?: string; plan_id?: number; expired?: boolean }>({})

  const { data, isLoading } = useAdminUsers(page, 20, search, Object.keys(filters).length > 0 ? filters : undefined)
  const { data: plans } = usePlans()
  const updateStatus = useUpdateUserStatus()
  const updateUser = useAdminUpdateUser()
  const deleteUser = useDeleteUser()
  const generateUsers = useGenerateUsers()
  const exportUsers = useExportUsers()

  const editForm = useForm<EditFormData>({
    resolver: zodResolver(editSchema),
    defaultValues: {
      email: '',
      password: '',
      traffic_limit: 0,
      expired_at: '',
      plan_id: 0,
      balance: 0,
      commission: 0,
      phone: '',
      remarks: '',
      device_limit: 0,
      speed_limit: 0,
      discount: 0,
    },
  })

  const generateForm = useForm<GenerateFormData>({
    resolver: zodResolver(generateSchema),
    defaultValues: {
      count: 10,
      prefix: 'user',
      password: '123456',
      plan_id: 0,
      expired_at: '',
    },
  })

  const handleSearch = () => setSearch(keyword)

  const activeFilterCount = Object.values(filters).filter((v) => v !== undefined && v !== '').length

  const updateFilter = (key: keyof typeof filters, value: string | number | boolean | undefined) => {
    setFilters((prev) => {
      const next = { ...prev }
      if (value === undefined || value === '' || value === -1) {
        delete next[key]
      } else {
        next[key] = value as any
      }
      return next
    })
    setPage(1)
  }

  const resetFilters = () => {
    setFilters({})
    setPage(1)
  }

  const toggleUser = (id: number) => {
    setSelectedUsers((prev) => {
      const next = new Set(prev)
      next.has(id) ? next.delete(id) : next.add(id)
      return next
    })
  }

  const handleBatchBan = async () => {
    const targets = data?.list.filter((u) => selectedUsers.has(u.id) && u.status === 1) ?? []
    let ok = 0, fail = 0
    for (const u of targets) {
      try { await new Promise<void>((res, rej) => updateStatus.mutate({ id: u.id, status: 0 }, { onSuccess: () => res(), onError: rej })); ok++ } catch { fail++ }
    }
    setSelectedUsers(new Set())
    setShowBatchBan(false)
    toast.success(`批量封禁完成：成功 ${ok}，失败 ${fail}`)
  }

  const handleToggleStatus = (id: number, currentStatus: number) => {
    updateStatus.mutate({ id, status: currentStatus === 1 ? 0 : 1 })
  }

  const handleDelete = () => {
    if (deleteId !== null) {
      deleteUser.mutate(deleteId, { onSuccess: () => setDeleteId(null) })
    }
  }

  const handleEdit = (user: any) => {
    setEditUser(user)
    editForm.reset({
      email: user.email,
      password: '',
      traffic_limit: user.traffic_limit || 0,
      expired_at: user.expired_at ? user.expired_at.slice(0, 16).replace('T', ' ') : '',
      plan_id: user.plan_id || 0,
      balance: user.balance || 0,
      commission: user.commission || 0,
      phone: user.phone || '',
      remarks: user.remarks || '',
      device_limit: user.device_limit || 0,
      speed_limit: user.speed_limit || 0,
      discount: user.discount || 0,
    })
  }

  const handleEditSubmit = (data: EditFormData) => {
    if (!editUser) return

    const payload: any = {}
    if (data.email && data.email !== editUser.email) payload.email = data.email
    if (data.password) payload.password = data.password
    if (data.traffic_limit !== undefined) payload.traffic_limit = data.traffic_limit
    if (data.expired_at !== undefined) payload.expired_at = data.expired_at || ''
    if (data.plan_id !== undefined) payload.plan_id = data.plan_id || null
    if (data.balance !== undefined) payload.balance = data.balance
    if (data.commission !== undefined) payload.commission = data.commission
    if (data.phone !== undefined) payload.phone = data.phone
    if (data.remarks !== undefined) payload.remarks = data.remarks
    if (data.device_limit !== undefined) payload.device_limit = data.device_limit
    if (data.speed_limit !== undefined) payload.speed_limit = data.speed_limit
    if (data.discount !== undefined) payload.discount = data.discount

    updateUser.mutate(
      { id: editUser.id, data: payload },
      { onSuccess: () => setEditUser(null) },
    )
  }

  const handleGenerate = (data: GenerateFormData) => {
    generateUsers.mutate(data, { onSuccess: () => setShowGenerate(false) })
  }

  const handleExport = () => {
    exportUsers.mutate(
      { keyword: search },
      {
        onSuccess: (res: any) => {
          const url = window.URL.createObjectURL(new Blob([res]))
          const a = document.createElement('a')
          a.href = url
          a.download = `users_${new Date().toISOString().slice(0, 10)}.csv`
          a.click()
          window.URL.revokeObjectURL(url)
        },
      },
    )
  }

  const totalPages = data ? Math.ceil(data.total / data.page_size) : 1

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold">用户管理</h1>
        <div className="flex gap-2">
          <Button variant="outline" onClick={handleExport}>
            <Download className="mr-2 h-4 w-4" />
            导出CSV
          </Button>
          <Button onClick={() => setShowGenerate(true)}>
            <Users className="mr-2 h-4 w-4" />
            批量生成
          </Button>
        </div>
      </div>

      <Card>
        <CardHeader>
          <div className="flex items-center gap-2">
            <Input
              placeholder="搜索邮箱..."
              value={keyword}
              onChange={(e) => setKeyword(e.target.value)}
              onKeyDown={(e) => e.key === 'Enter' && handleSearch()}
              className="max-w-sm"
            />
            <Button size="sm" onClick={handleSearch}>
              <Search className="h-4 w-4" />
            </Button>
            <Button
              size="sm"
              variant={showFilters ? 'default' : 'outline'}
              onClick={() => setShowFilters(!showFilters)}
            >
              <SlidersHorizontal className="h-4 w-4" />
              {activeFilterCount > 0 && (
                <Badge variant="secondary" className="ml-1 h-5 min-w-5 px-1">{activeFilterCount}</Badge>
              )}
            </Button>
          </div>
          {showFilters && (
            <div className="mt-3 flex flex-wrap items-center gap-3 rounded-md border p-3">
              <div className="flex items-center gap-2">
                <Label className="text-xs text-muted-foreground whitespace-nowrap">状态</Label>
                <Select
                  value={filters.status !== undefined ? String(filters.status) : ''}
                  onValueChange={(v) => updateFilter('status', v === '' ? undefined : Number(v))}
                >
                  <SelectTrigger className="h-8 w-[110px]">
                    <SelectValue placeholder="全部" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="">全部</SelectItem>
                    <SelectItem value="1">正常</SelectItem>
                    <SelectItem value="0">禁用</SelectItem>
                  </SelectContent>
                </Select>
              </div>
              <div className="flex items-center gap-2">
                <Label className="text-xs text-muted-foreground whitespace-nowrap">角色</Label>
                <Select
                  value={filters.role || ''}
                  onValueChange={(v) => updateFilter('role', v || undefined)}
                >
                  <SelectTrigger className="h-8 w-[110px]">
                    <SelectValue placeholder="全部" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="">全部</SelectItem>
                    <SelectItem value="admin">管理员</SelectItem>
                    <SelectItem value="user">普通用户</SelectItem>
                  </SelectContent>
                </Select>
              </div>
              <div className="flex items-center gap-2">
                <Label className="text-xs text-muted-foreground whitespace-nowrap">套餐</Label>
                <Select
                  value={filters.plan_id !== undefined ? String(filters.plan_id) : ''}
                  onValueChange={(v) => updateFilter('plan_id', v === '' ? undefined : Number(v))}
                >
                  <SelectTrigger className="h-8 w-[130px]">
                    <SelectValue placeholder="全部" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="">全部</SelectItem>
                    {plans?.map((p) => (
                      <SelectItem key={p.id} value={String(p.id)}>{p.name}</SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
              <div className="flex items-center gap-2">
                <Label className="text-xs text-muted-foreground whitespace-nowrap">到期</Label>
                <Select
                  value={filters.expired !== undefined ? String(filters.expired) : ''}
                  onValueChange={(v) => updateFilter('expired', v === '' ? undefined : v === 'true')}
                >
                  <SelectTrigger className="h-8 w-[110px]">
                    <SelectValue placeholder="全部" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="">全部</SelectItem>
                    <SelectItem value="true">已过期</SelectItem>
                    <SelectItem value="false">未过期</SelectItem>
                  </SelectContent>
                </Select>
              </div>
              {activeFilterCount > 0 && (
                <Button size="sm" variant="ghost" onClick={resetFilters}>
                  <RotateCcw className="mr-1 h-3 w-3" />
                  重置
                </Button>
              )}
            </div>
          )}
        </CardHeader>
        <CardContent>
          {selectedUsers.size > 0 && (
            <div className="mb-4 flex items-center gap-2">
              <span className="text-sm text-muted-foreground">已选 {selectedUsers.size} 个用户</span>
              <Button size="sm" variant="destructive" onClick={() => setShowBatchBan(true)}>
                <Ban className="mr-1 h-4 w-4" />批量封禁
              </Button>
              <Button size="sm" variant="ghost" onClick={() => setSelectedUsers(new Set())}>取消选择</Button>
            </div>
          )}
          {isLoading ? (
            <div className="space-y-2">{Array.from({ length: 10 }).map((_, i) => <Skeleton key={i} className="h-12" />)}</div>
          ) : (
            <>
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead className="w-10">
                      <Checkbox
                        checked={data?.list.length > 0 && data.list.every((u) => selectedUsers.has(u.id))}
                        onCheckedChange={(v) => {
                          if (v) setSelectedUsers(new Set(data?.list.map((u) => u.id)))
                          else setSelectedUsers(new Set())
                        }}
                      />
                    </TableHead>
                    <TableHead>邮箱</TableHead>
                    <TableHead>角色</TableHead>
                    <TableHead>在线设备</TableHead>
                    <TableHead>状态</TableHead>
                    <TableHead>套餐</TableHead>
                    <TableHead>流量</TableHead>
                    <TableHead>余额</TableHead>
                    <TableHead>佣金</TableHead>
                    <TableHead>到期时间</TableHead>
                    <TableHead>注册时间</TableHead>
                    <TableHead className="text-right">操作</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {data?.list.map((u) => (
                    <TableRow key={u.id} data-state={selectedUsers.has(u.id) ? 'selected' : undefined}>
                      <TableCell>
                        <Checkbox checked={selectedUsers.has(u.id)} onCheckedChange={() => toggleUser(u.id)} />
                      </TableCell>
                      <TableCell>{u.email}</TableCell>
                      <TableCell>
                        <Badge variant={u.role === 'admin' ? 'default' : 'secondary'}>{u.role}</Badge>
                      </TableCell>
                      <TableCell>{u.online_count ?? 0}</TableCell>
                      <TableCell>
                        <Badge variant={u.status === 1 ? 'success' : 'destructive'}>{u.status === 1 ? '正常' : '禁用'}</Badge>
                      </TableCell>
                      <TableCell>{u.plan_id ? (plans?.find((p) => p.id === u.plan_id)?.name || `#${u.plan_id}`) : '-'}</TableCell>
                      <TableCell>{formatBytes(u.used_traffic)} / {u.traffic_limit ? formatBytes(u.traffic_limit) : '不限'}</TableCell>
                      <TableCell>{formatCurrency(u.balance || 0)}</TableCell>
                      <TableCell>{formatCurrency(u.commission || 0)}</TableCell>
                      <TableCell className="text-xs">{u.expired_at ? formatDate(u.expired_at) : '永不过期'}</TableCell>
                      <TableCell className="text-xs">{formatDate(u.created_at)}</TableCell>
                      <TableCell className="text-right space-x-1">
                        <Button size="icon" variant="ghost" onClick={() => setDetailUser(u)} title="查看详情">
                          <Eye className="h-4 w-4" />
                        </Button>
                        <Button size="icon" variant="ghost" onClick={() => handleEdit(u)} title="编辑">
                          <Pencil className="h-4 w-4" />
                        </Button>
                        <Button size="icon" variant="ghost" onClick={() => handleToggleStatus(u.id, u.status)} title={u.status === 1 ? '禁用' : '启用'}>
                          {u.status === 1 ? <Ban className="h-4 w-4" /> : <CheckCircle className="h-4 w-4" />}
                        </Button>
                        <Button size="icon" variant="ghost" onClick={() => setDeleteId(u.id)} title="删除">
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

      {/* Edit User Dialog */}
      <Dialog open={!!editUser} onOpenChange={() => setEditUser(null)}>
        <DialogContent className="sm:max-w-lg">
          <DialogHeader>
            <DialogTitle>编辑用户</DialogTitle>
            <DialogDescription>修改用户 {editUser?.email} 的信息</DialogDescription>
          </DialogHeader>
          <form onSubmit={editForm.handleSubmit(handleEditSubmit)} className="space-y-4">
            <div className="space-y-2">
              <Label>邮箱</Label>
              <Input {...editForm.register('email')} placeholder="留空则不修改" />
            </div>
            <div className="space-y-2">
              <Label>新密码</Label>
              <Input type="password" {...editForm.register('password')} placeholder="留空则不修改" />
            </div>
            <div className="grid grid-cols-2 gap-4">
              <div className="space-y-2">
                <Label>流量限制 (字节)</Label>
                <Input
                  type="number"
                  {...editForm.register('traffic_limit', { valueAsNumber: true })}
                />
              </div>
              <div className="space-y-2">
                <Label>套餐</Label>
                <Select
                  value={String(editForm.watch('plan_id') || 0)}
                  onValueChange={(v) => editForm.setValue('plan_id', Number(v))}
                >
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="0">无套餐</SelectItem>
                    {plans?.map((p) => (
                      <SelectItem key={p.id} value={String(p.id)}>{p.name}</SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
            </div>
            <div className="space-y-2">
              <Label>到期时间</Label>
              <Input
                type="datetime-local"
                {...editForm.register('expired_at')}
              />
              <p className="text-xs text-muted-foreground">留空表示永不过期</p>
            </div>
            <div className="grid grid-cols-2 gap-4">
              <div className="space-y-2">
                <Label>账户余额</Label>
                <Input
                  type="number"
                  step="0.01"
                  {...editForm.register('balance', { valueAsNumber: true })}
                />
              </div>
              <div className="space-y-2">
                <Label>佣金余额</Label>
                <Input
                  type="number"
                  step="0.01"
                  {...editForm.register('commission', { valueAsNumber: true })}
                />
              </div>
            </div>
            <div className="grid grid-cols-2 gap-4">
              <div className="space-y-2">
                <Label>手机号</Label>
                <Input {...editForm.register('phone')} placeholder="留空则不修改" />
              </div>
              <div className="space-y-2">
                <Label>设备限制</Label>
                <Input
                  type="number"
                  {...editForm.register('device_limit', { valueAsNumber: true })}
                  placeholder="0 表示不限"
                />
              </div>
            </div>
            <div className="grid grid-cols-2 gap-4">
              <div className="space-y-2">
                <Label>限速 (Mbps)</Label>
                <Input
                  type="number"
                  {...editForm.register('speed_limit', { valueAsNumber: true })}
                  placeholder="0 表示不限"
                />
              </div>
              <div className="space-y-2">
                <Label>专享折扣 (%)</Label>
                <Input
                  type="number"
                  {...editForm.register('discount', { valueAsNumber: true })}
                  placeholder="0-100"
                />
              </div>
            </div>
            <div className="space-y-2">
              <Label>备注</Label>
              <Textarea {...editForm.register('remarks')} placeholder="留空则不修改" rows={3} />
            </div>
            <DialogFooter>
              <Button type="button" variant="outline" onClick={() => setEditUser(null)}>取消</Button>
              <Button type="submit" disabled={updateUser.isPending}>
                {updateUser.isPending && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
                保存
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>

      {/* User Detail Dialog */}
      <Dialog open={!!detailUser} onOpenChange={() => setDetailUser(null)}>
        <DialogContent className="sm:max-w-lg">
          <DialogHeader>
            <DialogTitle>用户详情</DialogTitle>
            <DialogDescription>{detailUser?.email}</DialogDescription>
          </DialogHeader>
          {detailUser && (
            <div className="space-y-4">
              <div>
                <h4 className="text-sm font-medium text-muted-foreground mb-2">基本信息</h4>
                <div className="grid grid-cols-2 gap-2 text-sm">
                  <div><span className="text-muted-foreground">ID：</span>{detailUser.id}</div>
                  <div><span className="text-muted-foreground">邮箱：</span>{detailUser.email}</div>
                  <div>
                    <span className="text-muted-foreground">角色：</span>
                    <Badge variant={detailUser.role === 'admin' ? 'default' : 'secondary'} className="ml-1">{detailUser.role}</Badge>
                  </div>
                  <div>
                    <span className="text-muted-foreground">状态：</span>
                    <Badge variant={detailUser.status === 1 ? 'success' : 'destructive'} className="ml-1">
                      {detailUser.status === 1 ? '正常' : '禁用'}
                    </Badge>
                  </div>
                  {detailUser.phone && <div><span className="text-muted-foreground">手机号：</span>{detailUser.phone}</div>}
                  {detailUser.remarks && <div className="col-span-2"><span className="text-muted-foreground">备注：</span>{detailUser.remarks}</div>}
                </div>
              </div>
              <div>
                <h4 className="text-sm font-medium text-muted-foreground mb-2">订阅信息</h4>
                <div className="grid grid-cols-2 gap-2 text-sm">
                  <div>
                    <span className="text-muted-foreground">套餐：</span>
                    {detailUser.plan_id ? (plans?.find((p) => p.id === detailUser.plan_id)?.name || `#${detailUser.plan_id}`) : '无'}
                  </div>
                  <div><span className="text-muted-foreground">到期时间：</span>{detailUser.expired_at ? formatDate(detailUser.expired_at) : '永不过期'}</div>
                  <div><span className="text-muted-foreground">设备限制：</span>{detailUser.device_limit || '不限'}</div>
                  <div><span className="text-muted-foreground">在线设备：</span>{detailUser.online_count ?? 0}</div>
                </div>
              </div>
              <div>
                <h4 className="text-sm font-medium text-muted-foreground mb-2">流量信息</h4>
                <div className="grid grid-cols-2 gap-2 text-sm">
                  <div><span className="text-muted-foreground">已用流量：</span>{formatBytes(detailUser.used_traffic)}</div>
                  <div><span className="text-muted-foreground">总流量：</span>{detailUser.traffic_limit ? formatBytes(detailUser.traffic_limit) : '不限'}</div>
                  {detailUser.u != null && <div><span className="text-muted-foreground">上传：</span>{formatBytes(detailUser.u)}</div>}
                  {detailUser.d != null && <div><span className="text-muted-foreground">下载：</span>{formatBytes(detailUser.d)}</div>}
                </div>
              </div>
              <div>
                <h4 className="text-sm font-medium text-muted-foreground mb-2">财务信息</h4>
                <div className="grid grid-cols-2 gap-2 text-sm">
                  <div><span className="text-muted-foreground">余额：</span>{formatCurrency(detailUser.balance || 0)}</div>
                  <div><span className="text-muted-foreground">佣金余额：</span>{formatCurrency(detailUser.commission || 0)}</div>
                </div>
              </div>
              <div>
                <h4 className="text-sm font-medium text-muted-foreground mb-2">时间信息</h4>
                <div className="grid grid-cols-2 gap-2 text-sm">
                  <div><span className="text-muted-foreground">注册时间：</span>{formatDate(detailUser.created_at)}</div>
                  <div><span className="text-muted-foreground">最后登录：</span>{detailUser.last_login_at ? formatDate(detailUser.last_login_at) : '-'}</div>
                </div>
              </div>
            </div>
          )}
          <DialogFooter>
            <Button variant="outline" onClick={() => setDetailUser(null)}>关闭</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Generate Users Dialog */}
      <Dialog open={showGenerate} onOpenChange={setShowGenerate}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>批量生成用户</DialogTitle>
            <DialogDescription>批量生成用户账号</DialogDescription>
          </DialogHeader>
          <form onSubmit={generateForm.handleSubmit(handleGenerate)} className="space-y-4">
            <div className="grid grid-cols-2 gap-4">
              <div className="space-y-2">
                <Label>数量</Label>
                <Input
                  type="number"
                  {...generateForm.register('count', { valueAsNumber: true })}
                />
              </div>
              <div className="space-y-2">
                <Label>前缀</Label>
                <Input {...generateForm.register('prefix')} placeholder="user" />
              </div>
            </div>
            <div className="space-y-2">
              <Label>密码</Label>
              <Input {...generateForm.register('password')} placeholder="123456" />
            </div>
            <div className="space-y-2">
              <Label>套餐</Label>
              <Select
                value={String(generateForm.watch('plan_id') || 0)}
                onValueChange={(v) => generateForm.setValue('plan_id', Number(v))}
              >
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="0">无套餐</SelectItem>
                  {plans?.map((p) => (
                    <SelectItem key={p.id} value={String(p.id)}>{p.name}</SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
            <div className="space-y-2">
              <Label>到期时间</Label>
              <Input
                type="datetime-local"
                {...generateForm.register('expired_at')}
              />
            </div>
            <DialogFooter>
              <Button type="button" variant="outline" onClick={() => setShowGenerate(false)}>取消</Button>
              <Button type="submit" disabled={generateUsers.isPending}>
                {generateUsers.isPending && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
                生成
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>

      {/* Batch Ban Dialog */}
      <AlertDialog open={showBatchBan} onOpenChange={setShowBatchBan}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>批量封禁</AlertDialogTitle>
            <AlertDialogDescription>
              确定要封禁选中的 {selectedUsers.size} 个用户吗？被封禁的用户将无法使用服务。
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>取消</AlertDialogCancel>
            <AlertDialogAction onClick={handleBatchBan} className="bg-destructive text-destructive-foreground hover:bg-destructive/90">
              确认封禁
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>

      {/* Delete Dialog */}
      <Dialog open={deleteId !== null} onOpenChange={() => setDeleteId(null)}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>确认删除</DialogTitle>
            <DialogDescription>确定要删除该用户吗？此操作不可撤销。</DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button variant="outline" onClick={() => setDeleteId(null)}>取消</Button>
            <Button variant="destructive" onClick={handleDelete} disabled={deleteUser.isPending}>确认删除</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}
