import { useState } from 'react'
import { toast } from 'sonner'
import { Check, Loader2, Tag, ShoppingCart, ArrowRight } from 'lucide-react'
import { usePlans } from '@/hooks/usePlans'
import { useCreateOrder } from '@/hooks/useOrders'
import { formatBytes, formatCurrency } from '@/lib/utils'
import api from '@/lib/api'
import type { Plan } from '@/types'
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Skeleton } from '@/components/ui/skeleton'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Separator } from '@/components/ui/separator'
import {
  Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle,
} from '@/components/ui/dialog'

export default function PlansPage() {
  const { data: plans, isLoading } = usePlans()
  const createOrder = useCreateOrder()

  const [selectedPlan, setSelectedPlan] = useState<Plan | null>(null)
  const [showBuyDialog, setShowBuyDialog] = useState(false)
  const [couponCode, setCouponCode] = useState('')
  const [couponLoading, setCouponLoading] = useState(false)
  const [couponValid, setCouponValid] = useState(false)
  const [discount, setDiscount] = useState(0)
  const [couponError, setCouponError] = useState('')

  const handleSelectPlan = (plan: Plan) => {
    setSelectedPlan(plan)
    setCouponCode('')
    setCouponValid(false)
    setDiscount(0)
    setCouponError('')
    setShowBuyDialog(true)
  }

  const handleValidateCoupon = async () => {
    if (!couponCode.trim() || !selectedPlan) return

    setCouponLoading(true)
    setCouponError('')
    setCouponValid(false)
    setDiscount(0)

    try {
      const res = (await api.post('/coupons/validate', {
        code: couponCode.trim(),
        plan_id: selectedPlan.id,
        amount: selectedPlan.price,
      })) as unknown as { discount: number }

      if (res.discount > 0) {
        setCouponValid(true)
        setDiscount(res.discount)
        toast.success('优惠券有效')
      } else {
        setCouponError('优惠券不可用')
      }
    } catch (err: any) {
      setCouponError(err?.message || '优惠券无效')
    } finally {
      setCouponLoading(false)
    }
  }

  const handleBuy = () => {
    if (!selectedPlan) return

    createOrder.mutate(
      {
        plan_id: selectedPlan.id,
        coupon_code: couponValid ? couponCode.trim() : undefined,
      },
      {
        onSuccess: () => {
          toast.success('订单创建成功，请前往订单页面支付')
          setShowBuyDialog(false)
        },
        onError: (err: any) => {
          toast.error(err?.message || '创建订单失败')
        },
      },
    )
  }

  const finalPrice = selectedPlan ? selectedPlan.price - discount : 0

  if (isLoading) {
    return (
      <div className="space-y-6">
        <h1 className="text-2xl font-bold">套餐</h1>
        <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
          {Array.from({ length: 6 }).map((_, i) => (
            <Card key={i}>
              <CardHeader>
                <Skeleton className="h-6 w-32" />
                <Skeleton className="h-4 w-48" />
              </CardHeader>
              <CardContent className="space-y-3">
                <Skeleton className="h-8 w-24" />
                <Skeleton className="h-4 w-full" />
                <Skeleton className="h-4 w-full" />
                <Skeleton className="h-4 w-3/4" />
              </CardContent>
              <CardFooter>
                <Skeleton className="h-10 w-full" />
              </CardFooter>
            </Card>
          ))}
        </div>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold">套餐</h1>
        <p className="text-muted-foreground">选择适合您的套餐</p>
      </div>

      <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
        {plans?.map((plan) => (
          <Card key={plan.id} className="flex flex-col">
            <CardHeader>
              <div className="flex items-center justify-between">
                <CardTitle>{plan.name}</CardTitle>
                {plan.sort === 0 && <Badge>推荐</Badge>}
              </div>
              <CardDescription>{plan.description || '高速稳定线路'}</CardDescription>
            </CardHeader>
            <CardContent className="flex-1 space-y-4">
              <div>
                <span className="text-3xl font-bold">{formatCurrency(plan.price)}</span>
                <span className="text-muted-foreground">
                  /{plan.duration_days ? `${plan.duration_days}天` : '月'}
                </span>
              </div>
              <ul className="space-y-2 text-sm">
                <li className="flex items-center gap-2">
                  <Check className="h-4 w-4 text-primary" />
                  <span>流量: {formatBytes(plan.traffic)}</span>
                </li>
                <li className="flex items-center gap-2">
                  <Check className="h-4 w-4 text-primary" />
                  <span>时长: {plan.duration_days ? `${plan.duration_days}天` : '永久'}</span>
                </li>
                <li className="flex items-center gap-2">
                  <Check className="h-4 w-4 text-primary" />
                  <span>设备限制: {plan.device_limit || '不限'}</span>
                </li>
                {plan.node_group && (
                  <li className="flex items-center gap-2">
                    <Check className="h-4 w-4 text-primary" />
                    <span>节点组: {plan.node_group}</span>
                  </li>
                )}
              </ul>
            </CardContent>
            <CardFooter>
              <Button
                className="w-full"
                onClick={() => handleSelectPlan(plan)}
              >
                <ShoppingCart className="mr-2 h-4 w-4" />
                立即购买
              </Button>
            </CardFooter>
          </Card>
        ))}
      </div>

      {plans?.length === 0 && (
        <div className="text-center py-12 text-muted-foreground">
          暂无可用套餐
        </div>
      )}

      {/* 购买确认对话框 */}
      <Dialog open={showBuyDialog} onOpenChange={setShowBuyDialog}>
        <DialogContent className="sm:max-w-md">
          <DialogHeader>
            <DialogTitle>确认购买</DialogTitle>
            <DialogDescription>
              请确认套餐信息，可选输入优惠券码
            </DialogDescription>
          </DialogHeader>

          {selectedPlan && (
            <div className="space-y-4">
              {/* 套餐信息 */}
              <div className="rounded-lg border p-4 space-y-2">
                <div className="flex items-center justify-between">
                  <span className="font-medium">{selectedPlan.name}</span>
                  <Badge variant="outline">
                    {selectedPlan.duration_days ? `${selectedPlan.duration_days}天` : '永久'}
                  </Badge>
                </div>
                <p className="text-sm text-muted-foreground">
                  流量: {formatBytes(selectedPlan.traffic)} · 设备: {selectedPlan.device_limit || '不限'}
                </p>
              </div>

              {/* 优惠券输入 */}
              <div className="space-y-2">
                <Label>优惠券码（可选）</Label>
                <div className="flex gap-2">
                  <Input
                    placeholder="请输入优惠券码"
                    value={couponCode}
                    onChange={(e) => {
                      setCouponCode(e.target.value.toUpperCase())
                      setCouponValid(false)
                      setDiscount(0)
                      setCouponError('')
                    }}
                  />
                  <Button
                    variant="outline"
                    onClick={handleValidateCoupon}
                    disabled={!couponCode.trim() || couponLoading}
                  >
                    {couponLoading ? (
                      <Loader2 className="h-4 w-4 animate-spin" />
                    ) : (
                      '验证'
                    )}
                  </Button>
                </div>
                {couponError && (
                  <p className="text-sm text-destructive">{couponError}</p>
                )}
                {couponValid && (
                  <p className="text-sm text-green-600">✓ 优惠券有效，折扣 {formatCurrency(discount)}</p>
                )}
              </div>

              <Separator />

              {/* 价格明细 */}
              <div className="space-y-2">
                <div className="flex items-center justify-between text-sm">
                  <span className="text-muted-foreground">套餐原价</span>
                  <span>{formatCurrency(selectedPlan.price)}</span>
                </div>
                {couponValid && discount > 0 && (
                  <div className="flex items-center justify-between text-sm">
                    <span className="text-muted-foreground">优惠券折扣</span>
                    <span className="text-green-600">-{formatCurrency(discount)}</span>
                  </div>
                )}
                <Separator />
                <div className="flex items-center justify-between">
                  <span className="font-medium">实付金额</span>
                  <span className="text-xl font-bold text-primary">
                    {formatCurrency(finalPrice)}
                  </span>
                </div>
              </div>
            </div>
          )}

          <DialogFooter>
            <Button variant="outline" onClick={() => setShowBuyDialog(false)}>
              取消
            </Button>
            <Button onClick={handleBuy} disabled={createOrder.isPending}>
              {createOrder.isPending && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
              <ArrowRight className="mr-2 h-4 w-4" />
              确认购买
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}
