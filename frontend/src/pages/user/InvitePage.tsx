import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'
import { Copy, Users, DollarSign, ArrowUpRight, Loader2, Wallet } from 'lucide-react'
import api from '@/lib/api'
import { useUserProfile } from '@/hooks/useAuth'
import { formatDate, formatCurrency } from '@/lib/utils'
import type { InviteCode, CommissionLog } from '@/types'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Badge } from '@/components/ui/badge'
import { Skeleton } from '@/components/ui/skeleton'
import { Separator } from '@/components/ui/separator'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import {
  Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle,
} from '@/components/ui/dialog'

export default function InvitePage() {
  const queryClient = useQueryClient()
  const { data: user } = useUserProfile()
  const [showWithdraw, setShowWithdraw] = useState(false)

  const { data: inviteCode, isLoading: codeLoading } = useQuery({
    queryKey: ['user', 'invite', 'code'],
    queryFn: async () => (await api.get('/user/invite/code')) as unknown as InviteCode,
  })

  const { data: commissionStats } = useQuery({
    queryKey: ['user', 'invite', 'commission-stats'],
    queryFn: async () =>
      (await api.get('/invite/commission/stats')) as unknown as { total: number; pending: number },
  })

  const { data: commissionLogs, isLoading: logsLoading } = useQuery({
    queryKey: ['user', 'invite', 'commissions'],
    queryFn: async () =>
      (await api.get('/invite/commission/logs')) as unknown as { list: CommissionLog[]; total: number },
  })

  // Withdraw commission to balance
  const withdrawMutation = useMutation({
    mutationFn: async () => await api.post('/invite/commission/withdraw'),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['user', 'profile'] })
      queryClient.invalidateQueries({ queryKey: ['user', 'invite'] })
      toast.success('佣金已转入余额')
      setShowWithdraw(false)
    },
    onError: (err: any) => toast.error(err?.message || '提现失败'),
  })

  const inviteLink = inviteCode
    ? `${window.location.origin}/user/register?code=${inviteCode.code}`
    : ''

  const copyToClipboard = (text: string) => {
    navigator.clipboard.writeText(text)
    toast.success('已复制到剪贴板')
  }

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-bold">邀请</h1>

      {/* Invite Code & Link */}
      <div className="grid gap-4 md:grid-cols-2">
        <Card>
          <CardHeader>
            <CardTitle>邀请码</CardTitle>
            <CardDescription>分享邀请码给好友</CardDescription>
          </CardHeader>
          <CardContent>
            {codeLoading ? (
              <Skeleton className="h-10 w-full" />
            ) : (
              <div className="flex gap-2">
                <Input readOnly value={inviteCode?.code || '暂无邀请码'} className="font-mono" />
                <Button
                  variant="outline"
                  size="icon"
                  onClick={() => inviteCode?.code && copyToClipboard(inviteCode.code)}
                >
                  <Copy className="h-4 w-4" />
                </Button>
              </div>
            )}
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>邀请链接</CardTitle>
            <CardDescription>直接分享链接注册</CardDescription>
          </CardHeader>
          <CardContent>
            {codeLoading ? (
              <Skeleton className="h-10 w-full" />
            ) : (
              <div className="flex gap-2">
                <Input readOnly value={inviteLink} className="font-mono text-xs" />
                <Button
                  variant="outline"
                  size="icon"
                  onClick={() => inviteLink && copyToClipboard(inviteLink)}
                >
                  <Copy className="h-4 w-4" />
                </Button>
              </div>
            )}
          </CardContent>
        </Card>
      </div>

      {/* Commission Stats */}
      <div className="grid gap-4 md:grid-cols-3">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between pb-2">
            <CardTitle className="text-sm font-medium">累计佣金</CardTitle>
            <DollarSign className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{formatCurrency(commissionStats?.total || 0)}</div>
            <p className="text-xs text-muted-foreground">邀请好友注册购买可获得佣金</p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between pb-2">
            <CardTitle className="text-sm font-medium">待结算佣金</CardTitle>
            <ArrowUpRight className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{formatCurrency(commissionStats?.pending || 0)}</div>
            <p className="text-xs text-muted-foreground">等待结算的佣金金额</p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between pb-2">
            <CardTitle className="text-sm font-medium">邀请人数</CardTitle>
            <Users className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{inviteCode?.used_count || 0}</div>
            <p className="text-xs text-muted-foreground">
              限制: {inviteCode?.limit_count || '不限'}
            </p>
          </CardContent>
        </Card>
      </div>

      {/* Withdraw Button */}
      <Card>
        <CardContent className="pt-6">
          <div className="flex items-center justify-between">
            <div>
              <p className="font-medium">佣金提现</p>
              <p className="text-sm text-muted-foreground">
                将佣金转入账户余额，可用于购买套餐
              </p>
            </div>
            <Button
              onClick={() => setShowWithdraw(true)}
              disabled={!commissionStats?.pending || commissionStats.pending <= 0}
            >
              <Wallet className="mr-2 h-4 w-4" />
              提现到余额
            </Button>
          </div>
        </CardContent>
      </Card>

      {/* Commission Logs */}
      <Card>
        <CardHeader>
          <CardTitle>佣金记录</CardTitle>
          <CardDescription>详细的佣金获得记录</CardDescription>
        </CardHeader>
        <CardContent>
          {logsLoading ? (
            <div className="space-y-3">
              {Array.from({ length: 5 }).map((_, i) => (
                <Skeleton key={i} className="h-12 w-full" />
              ))}
            </div>
          ) : (
            <div className="overflow-x-auto">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>来源用户</TableHead>
                    <TableHead>订单金额</TableHead>
                    <TableHead>佣金</TableHead>
                    <TableHead>状态</TableHead>
                    <TableHead>时间</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {commissionLogs?.list?.length ? (
                    commissionLogs.list.map((log) => (
                      <TableRow key={log.id}>
                        <TableCell>{log.from_user_email || `用户 #${log.from_user_id}`}</TableCell>
                        <TableCell>{formatCurrency(log.order_amount)}</TableCell>
                        <TableCell className="font-medium text-green-600">
                          +{formatCurrency(log.commission)}
                        </TableCell>
                        <TableCell>
                          <Badge variant={log.status === 1 ? 'default' : 'secondary'}>
                            {log.status === 1 ? '已到账' : '待结算'}
                          </Badge>
                        </TableCell>
                        <TableCell className="text-sm text-muted-foreground">
                          {formatDate(log.created_at)}
                        </TableCell>
                      </TableRow>
                    ))
                  ) : (
                    <TableRow>
                      <TableCell colSpan={5} className="text-center py-8 text-muted-foreground">
                        暂无佣金记录
                      </TableCell>
                    </TableRow>
                  )}
                </TableBody>
              </Table>
            </div>
          )}
        </CardContent>
      </Card>

      {/* Withdraw Dialog */}
      <Dialog open={showWithdraw} onOpenChange={setShowWithdraw}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>佣金提现</DialogTitle>
            <DialogDescription>
              将待结算佣金转入账户余额
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-4">
            <div className="rounded-lg border p-4 space-y-2">
              <div className="flex justify-between text-sm">
                <span className="text-muted-foreground">待结算佣金</span>
                <span className="font-medium">{formatCurrency(commissionStats?.pending || 0)}</span>
              </div>
              <Separator />
              <div className="flex justify-between">
                <span className="font-medium">转入余额</span>
                <span className="text-lg font-bold text-primary">{formatCurrency(commissionStats?.pending || 0)}</span>
              </div>
            </div>
            <p className="text-sm text-muted-foreground">
              提现后佣金将转入您的账户余额，可用于购买套餐。
            </p>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setShowWithdraw(false)}>
              取消
            </Button>
            <Button
              onClick={() => withdrawMutation.mutate()}
              disabled={withdrawMutation.isPending}
            >
              {withdrawMutation.isPending && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
              确认提现
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}
