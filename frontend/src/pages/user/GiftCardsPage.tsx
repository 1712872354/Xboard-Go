import { useState } from 'react'
import { useMutation } from '@tanstack/react-query'
import { toast } from 'sonner'
import { Loader2, Gift, Info } from 'lucide-react'
import api from '@/lib/api'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Separator } from '@/components/ui/separator'
import { Badge } from '@/components/ui/badge'

interface RedeemResult {
  balance?: number
  traffic?: number
  duration?: number
  plan_name?: string
}

export default function GiftCardsPage() {
  const [code, setCode] = useState('')
  const [result, setResult] = useState<RedeemResult | null>(null)

  const redeem = useMutation({
    mutationFn: async (giftCode: string) =>
      (await api.post('/gift-cards/use', { code: giftCode })) as unknown as RedeemResult,
    onSuccess: (data) => {
      setResult(data)
      toast.success('礼品卡兑换成功')
      setCode('')
    },
    onError: (err: any) => toast.error(err?.message || '兑换失败'),
  })

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-bold">礼品卡</h1>

      {/* Redeem Card */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Gift className="h-5 w-5" />
            兑换礼品卡
          </CardTitle>
          <CardDescription>输入礼品卡代码来兑换奖励</CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="flex gap-2 max-w-md">
            <Input
              placeholder="请输入礼品卡代码"
              value={code}
              onChange={(e) => setCode(e.target.value.toUpperCase())}
            />
            <Button
              onClick={() => redeem.mutate(code)}
              disabled={!code.trim() || redeem.isPending}
            >
              {redeem.isPending && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
              兑换
            </Button>
          </div>

          {result && (
            <div className="rounded-lg border border-green-200 bg-green-50 p-4 dark:border-green-800 dark:bg-green-950">
              <p className="font-medium text-green-700 dark:text-green-300 mb-2">
                兑换成功！
              </p>
              <div className="space-y-1 text-sm">
                {result.balance !== undefined && result.balance > 0 && (
                  <p>余额增加: ¥{(result.balance / 100).toFixed(2)}</p>
                )}
                {result.traffic !== undefined && result.traffic > 0 && (
                  <p>流量增加: {(result.traffic / (1024 * 1024 * 1024)).toFixed(1)} GB</p>
                )}
                {result.duration !== undefined && result.duration > 0 && (
                  <p>时长增加: {result.duration} 天</p>
                )}
                {result.plan_name && (
                  <p>套餐: {result.plan_name}</p>
                )}
              </div>
            </div>
          )}
        </CardContent>
      </Card>

      {/* Usage Instructions */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Info className="h-5 w-5" />
            使用说明
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-3 text-sm text-muted-foreground">
          <div className="flex gap-2">
            <Badge variant="outline" className="shrink-0">1</Badge>
            <p>购买或获取礼品卡代码</p>
          </div>
          <Separator />
          <div className="flex gap-2">
            <Badge variant="outline" className="shrink-0">2</Badge>
            <p>在上方输入框中输入礼品卡代码</p>
          </div>
          <Separator />
          <div className="flex gap-2">
            <Badge variant="outline" className="shrink-0">3</Badge>
            <p>点击兑换按钮，系统将自动为您的账户添加对应权益</p>
          </div>
          <Separator />
          <div className="flex gap-2">
            <Badge variant="outline" className="shrink-0">4</Badge>
            <p>礼品卡可能包含余额、流量、时长或套餐订阅</p>
          </div>
          <Separator />
          <p className="text-xs">
            注意：每个礼品卡代码仅可使用一次，兑换后无法撤销。
          </p>
        </CardContent>
      </Card>
    </div>
  )
}
