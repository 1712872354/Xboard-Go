import { useState } from 'react'
import { useMutation } from '@tanstack/react-query'
import { toast } from 'sonner'
import { Loader2, Tag } from 'lucide-react'
import api from '@/lib/api'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'

export default function CouponsPage() {
  const [code, setCode] = useState('')
  const [result, setResult] = useState<any>(null)

  const validateCoupon = useMutation({
    mutationFn: async (couponCode: string) =>
      await api.post('/coupons/validate', { code: couponCode }),
    onSuccess: (data: any) => {
      setResult(data)
      toast.success('优惠券验证成功')
      setCode('')
    },
    onError: (err: any) => toast.error(err?.message || '优惠券无效'),
  })

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-bold">优惠券</h1>

      {/* Validate Coupon */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Tag className="h-5 w-5" />
            兑换优惠券
          </CardTitle>
          <CardDescription>输入优惠码来兑换优惠券</CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="flex gap-2 max-w-md">
            <Input
              placeholder="请输入优惠码"
              value={code}
              onChange={(e) => setCode(e.target.value.toUpperCase())}
            />
            <Button
              onClick={() => validateCoupon.mutate(code)}
              disabled={!code.trim() || validateCoupon.isPending}
            >
              {validateCoupon.isPending && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
              兑换
            </Button>
          </div>

          {result && (
            <div className="rounded-lg border border-green-200 bg-green-50 p-4 dark:border-green-800 dark:bg-green-950">
              <p className="font-medium text-green-700 dark:text-green-300 mb-2">
                验证成功！
              </p>
              <div className="space-y-1 text-sm">
                {result.name && <p>优惠券: {result.name}</p>}
                {result.type === 1
                  ? <p>优惠金额: ¥{(result.value / 100).toFixed(2)}</p>
                  : <p>折扣比例: {result.value}%</p>}
                {result.max_discount > 0 && (
                  <p>最高优惠: ¥{(result.max_discount / 100).toFixed(2)}</p>
                )}
              </div>
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  )
}
