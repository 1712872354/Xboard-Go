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
import { Textarea } from '@/components/ui/textarea'
import { Switch } from '@/components/ui/switch'
import { Skeleton } from '@/components/ui/skeleton'
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from '@/components/ui/dialog'
import { Plus, Pencil, Trash2, ArrowUp, ArrowDown, Loader2, ExternalLink, Search } from 'lucide-react'
import { AlertDialog, AlertDialogAction, AlertDialogCancel, AlertDialogContent, AlertDialogDescription, AlertDialogFooter, AlertDialogHeader, AlertDialogTitle } from '@/components/ui/alert-dialog'

interface PaymentGateway {
  id: number
  name: string
  payment: string
  config: string
  icon: string
  notify_domain: string
  handling_fee_percent: number
  handling_fee_fixed: number
  notify_url: string
  sort: number
  status: number
  created_at: string
}

const PAYMENT_NAME_MAP: Record<string, string> = {
  alipay_f2f: '支付宝当面付',
  wechat_pay: '微信支付',
  epay: '易支付',
  btcpay: 'BTCPay',
  stripe: 'Stripe',
}

function GatewayForm({ gateway, onClose, onSave }: { gateway?: PaymentGateway; onClose: () => void; onSave: (data: Partial<PaymentGateway>) => void }) {
  const [name, setName] = useState(gateway?.name ?? '')
  const [payment, setPayment] = useState(gateway?.payment ?? '')
  const [config, setConfig] = useState(gateway?.config ?? '{}')
  const [icon, setIcon] = useState(gateway?.icon ?? '')
  const [notifyDomain, setNotifyDomain] = useState(gateway?.notify_domain ?? '')
  const [feePercent, setFeePercent] = useState(gateway?.handling_fee_percent?.toString() ?? '0')
  const [feeFixed, setFeeFixed] = useState(gateway?.handling_fee_fixed?.toString() ?? '0')
  const [sort, setSort] = useState(gateway?.sort?.toString() ?? '0')
  const [status, setStatus] = useState(gateway?.status ?? 1)
  const [saving, setSaving] = useState(false)

  const handleSubmit = async () => {
    if (!name.trim() || !payment.trim()) {
      toast.error('请填写必填项')
      return
    }
    setSaving(true)
    await onSave({
      name, payment, config, icon, notify_domain: notifyDomain,
      handling_fee_percent: parseFloat(feePercent) || 0,
      handling_fee_fixed: parseFloat(feeFixed) || 0,
      sort: parseInt(sort), status,
    })
    setSaving(false)
  }

  return (
    <div className="space-y-4">
      <div className="grid grid-cols-2 gap-4">
        <div className="space-y-2">
          <Label>显示名称 <span className="text-destructive">*</span></Label>
          <Input value={name} onChange={(e) => setName(e.target.value)} placeholder="如：支付宝" />
        </div>
        <div className="space-y-2">
          <Label>支付接口 <span className="text-destructive">*</span></Label>
          <select className="w-full rounded-md border border-input bg-background px-3 py-2 text-sm" value={payment} onChange={(e) => setPayment(e.target.value)}>
            <option value="">请选择</option>
            <option value="alipay_f2f">支付宝当面付</option>
            <option value="wechat_pay">微信支付</option>
            <option value="epay">易支付</option>
            <option value="btcpay">BTCPay</option>
            <option value="stripe">Stripe</option>
          </select>
        </div>
      </div>
      <div className="space-y-2">
        <Label>图标 URL</Label>
        <Input value={icon} onChange={(e) => setIcon(e.target.value)} placeholder="支付方式图标地址" />
      </div>
      <div className="space-y-2">
        <Label>通知域名</Label>
        <Input value={notifyDomain} onChange={(e) => setNotifyDomain(e.target.value)} placeholder="留空使用默认站点域名，如 https://example.com" />
        <p className="text-xs text-muted-foreground">自定义支付回调通知域名，需确保防火墙放行该域名</p>
      </div>
      <div className="grid grid-cols-2 gap-4">
        <div className="space-y-2">
          <Label>百分比手续费 (%)</Label>
          <Input type="number" min="0" max="100" step="0.1" value={feePercent} onChange={(e) => setFeePercent(e.target.value)} placeholder="0" />
          <p className="text-xs text-muted-foreground">订单金额的百分比，0 表示不收取</p>
        </div>
        <div className="space-y-2">
          <Label>固定手续费 (分)</Label>
          <Input type="number" min="0" step="1" value={feeFixed} onChange={(e) => setFeeFixed(e.target.value)} placeholder="0" />
          <p className="text-xs text-muted-foreground">每笔订单固定收取的手续费，单位分</p>
        </div>
      </div>
      <div className="space-y-2">
        <Label>配置 (JSON)</Label>
        <Textarea value={config} onChange={(e) => setConfig(e.target.value)} placeholder='{"app_id":"...","private_key":"..."}' rows={6} className="font-mono text-xs" />
      </div>
      <div className="grid grid-cols-2 gap-4">
        <div className="space-y-2">
          <Label>排序</Label>
          <Input type="number" value={sort} onChange={(e) => setSort(e.target.value)} />
        </div>
        <div className="flex items-end justify-between pb-2">
          <Label>启用状态</Label>
          <Switch checked={status === 1} onCheckedChange={(v) => setStatus(v ? 1 : 0)} />
        </div>
      </div>
      <DialogFooter>
        <Button variant="outline" onClick={onClose}>取消</Button>
        <Button onClick={handleSubmit} disabled={saving}>
          {saving && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
          {gateway ? '更新' : '创建'}
        </Button>
      </DialogFooter>
    </div>
  )
}

export default function PaymentGatewaysPage() {
  const qc = useQueryClient()
  const [dialogOpen, setDialogOpen] = useState(false)
  const [editingGateway, setEditingGateway] = useState<PaymentGateway | undefined>()
  const [deleteId, setDeleteId] = useState<number | null>(null)
  const [gatewaySearch, setGatewaySearch] = useState('')

  const { data, isLoading } = useQuery({
    queryKey: ['admin', 'payment-gateways'],
    queryFn: async () => await api.get('/admin/payment/gateways') as unknown as PaymentGateway[],
  })

  const createMutation = useMutation({
    mutationFn: async (data: Partial<PaymentGateway>) => await api.post('/admin/payment/gateways', data),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['admin', 'payment-gateways'] })
      setDialogOpen(false)
      toast.success('创建成功')
    },
    onError: () => toast.error('创建失败'),
  })

  const updateMutation = useMutation({
    mutationFn: async ({ id, ...data }: Partial<PaymentGateway> & { id: number }) =>
      await api.put(`/admin/payment/gateways/${id}`, data),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['admin', 'payment-gateways'] })
      setDialogOpen(false)
      toast.success('更新成功')
    },
    onError: () => toast.error('更新失败'),
  })

  const deleteMutation = useMutation({
    mutationFn: async (id: number) => await api.delete(`/admin/payment/gateways/${id}`),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['admin', 'payment-gateways'] })
      toast.success('删除成功')
    },
    onError: () => toast.error('删除失败'),
  })

  const updateSortMutation = useMutation({
    mutationFn: async ({ id, sort }: { id: number; sort: number }) =>
      await api.put(`/admin/payment/gateways/${id}/sort`, { sort }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['admin', 'payment-gateways'] }),
  })

  const handleSave = async (formData: Partial<PaymentGateway>) => {
    if (editingGateway) {
      await updateMutation.mutateAsync({ id: editingGateway.id, ...formData })
    } else {
      await createMutation.mutateAsync(formData)
    }
  }

  const gateways = Array.isArray(data) ? data : []
  const filteredGateways = gatewaySearch
    ? gateways.filter((g) => g.name.toLowerCase().includes(gatewaySearch.toLowerCase()))
    : gateways

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold">支付配置</h1>
          <p className="text-muted-foreground">管理支付接口和支付方式</p>
        </div>
        <Button onClick={() => { setEditingGateway(undefined); setDialogOpen(true) }}>
          <Plus className="mr-2 h-4 w-4" />
          添加支付方式
        </Button>
      </div>

      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <CardTitle>支付方式列表</CardTitle>
            <div className="relative w-64">
              <Search className="absolute left-2.5 top-2.5 h-4 w-4 text-muted-foreground" />
              <Input placeholder="搜索名称..." className="pl-8" value={gatewaySearch} onChange={(e) => setGatewaySearch(e.target.value)} />
            </div>
          </div>
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
                  <TableHead>接口</TableHead>
                  <TableHead>通知地址</TableHead>
                  <TableHead>排序</TableHead>
                  <TableHead>状态</TableHead>
                  <TableHead className="text-right">操作</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {filteredGateways.length === 0 ? (
                  <TableRow><TableCell colSpan={7} className="h-24 text-center text-muted-foreground">暂无支付方式</TableCell></TableRow>
                ) : filteredGateways.map((g) => (
                  <TableRow key={g.id}>
                    <TableCell className="font-mono text-xs">{g.id}</TableCell>
                    <TableCell className="font-medium">{g.name}</TableCell>
                    <TableCell><Badge variant="outline">{PAYMENT_NAME_MAP[g.payment] ?? g.payment}</Badge></TableCell>
                    <TableCell>
                      {g.notify_url ? (
                        <div className="flex items-center gap-1 max-w-[200px]">
                          <span className="text-xs text-muted-foreground truncate">{g.notify_url}</span>
                          <Button variant="ghost" size="icon" className="h-5 w-5 shrink-0" onClick={() => { navigator.clipboard.writeText(g.notify_url); toast.success('已复制') }}>
                            <ExternalLink className="h-3 w-3" />
                          </Button>
                        </div>
                      ) : <span className="text-xs text-muted-foreground">-</span>}
                    </TableCell>
                    <TableCell>
                      <div className="flex items-center gap-1">
                        <span className="text-sm">{g.sort}</span>
                        <Button variant="ghost" size="sm" className="h-6 w-6 p-0" onClick={() => updateSortMutation.mutate({ id: g.id, sort: g.sort - 1 })}>
                          <ArrowUp className="h-3 w-3" />
                        </Button>
                        <Button variant="ghost" size="sm" className="h-6 w-6 p-0" onClick={() => updateSortMutation.mutate({ id: g.id, sort: g.sort + 1 })}>
                          <ArrowDown className="h-3 w-3" />
                        </Button>
                      </div>
                    </TableCell>
                    <TableCell>
                      <Switch
                        checked={g.status === 1}
                        onCheckedChange={(v) => updateMutation.mutate({ id: g.id, status: v ? 1 : 0 })}
                      />
                    </TableCell>
                    <TableCell className="text-right">
                      <div className="flex justify-end gap-1">
                        <Button variant="ghost" size="icon" className="h-8 w-8" onClick={() => { setEditingGateway(g); setDialogOpen(true) }}>
                          <Pencil className="h-4 w-4" />
                        </Button>
                        <Button variant="ghost" size="icon" className="h-8 w-8" onClick={() => setDeleteId(g.id)}>
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
        <DialogContent className="max-w-lg">
          <DialogHeader>
            <DialogTitle>{editingGateway ? '编辑支付方式' : '添加支付方式'}</DialogTitle>
          </DialogHeader>
          <GatewayForm gateway={editingGateway} onClose={() => setDialogOpen(false)} onSave={handleSave} />
        </DialogContent>
      </Dialog>

      <AlertDialog open={deleteId !== null} onOpenChange={() => setDeleteId(null)}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>确认删除</AlertDialogTitle>
            <AlertDialogDescription>确定要删除该支付方式吗？此操作不可撤销。</AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>取消</AlertDialogCancel>
            <AlertDialogAction onClick={() => { if (deleteId !== null) { deleteMutation.mutate(deleteId); setDeleteId(null) } }}>确认删除</AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  )
}
