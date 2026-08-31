import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { toast } from 'sonner'
import { Loader2, User as UserIcon, Mail, Lock, Package, Calendar, Wallet, Activity, ArrowUpRight } from 'lucide-react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { useUserProfile } from '@/hooks/useAuth'
import api from '@/lib/api'
import { formatDate, formatBytes, formatCurrency } from '@/lib/utils'
import type { Plan } from '@/types'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Separator } from '@/components/ui/separator'
import { Skeleton } from '@/components/ui/skeleton'
import { Progress } from '@/components/ui/progress'
import { Badge } from '@/components/ui/badge'

const emailSchema = z.object({
  email: z.string().email('请输入有效的邮箱地址'),
})

const passwordSchema = z
  .object({
    old_password: z.string().min(6, '请输入当前密码'),
    new_password: z.string().min(6, '新密码至少6位'),
    confirm_password: z.string().min(6, '请确认新密码'),
  })
  .refine((data) => data.new_password === data.confirm_password, {
    message: '两次密码不一致',
    path: ['confirm_password'],
  })

type EmailFormData = z.infer<typeof emailSchema>
type PasswordFormData = z.infer<typeof passwordSchema>

export default function ProfilePage() {
  const queryClient = useQueryClient()
  const { data: user, isLoading } = useUserProfile()

  // 获取用户当前套餐信息
  const { data: plans } = useQuery({
    queryKey: ['plans'],
    queryFn: async () => (await api.get('/plans')) as unknown as Plan[],
  })

  const currentPlan = plans?.find((p) => p.id === user?.plan_id)

  const emailForm = useForm<EmailFormData>({
    resolver: zodResolver(emailSchema),
    values: { email: user?.email || '' },
  })

  const passwordForm = useForm<PasswordFormData>({
    resolver: zodResolver(passwordSchema),
    defaultValues: { old_password: '', new_password: '', confirm_password: '' },
  })

  const updateEmail = useMutation({
    mutationFn: async (data: EmailFormData) => await api.put('/user/profile', data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['user', 'profile'] })
      toast.success('邮箱更新成功')
    },
    onError: (err: any) => toast.error(err?.message || '更新失败'),
  })

  const changePassword = useMutation({
    mutationFn: async (data: PasswordFormData) =>
      await api.put('/user/password', {
        old_password: data.old_password,
        new_password: data.new_password,
      }),
    onSuccess: () => {
      toast.success('密码修改成功')
      passwordForm.reset()
    },
    onError: (err: any) => toast.error(err?.message || '修改失败'),
  })

  if (isLoading) {
    return (
      <div className="space-y-6">
        <Skeleton className="h-8 w-32" />
        <Skeleton className="h-60 w-full" />
        <Skeleton className="h-60 w-full" />
      </div>
    )
  }

  const trafficUsed = user?.used_traffic || 0
  const trafficTotal = user?.traffic_limit || 0
  const trafficPercent = trafficTotal > 0 ? Math.min((trafficUsed / trafficTotal) * 100, 100) : 0
  const isExpired = user?.expired_at ? new Date(user.expired_at) < new Date() : false
  const daysLeft = user?.expired_at
    ? Math.max(0, Math.ceil((new Date(user.expired_at).getTime() - Date.now()) / (1000 * 60 * 60 * 24)))
    : null

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-bold">个人资料</h1>

      {/* 套餐信息卡片 */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Package className="h-5 w-5" />
            我的套餐
          </CardTitle>
          <CardDescription>当前订阅套餐信息</CardDescription>
        </CardHeader>
        <CardContent className="space-y-6">
          {/* 套餐名称和状态 */}
          <div className="flex items-center justify-between">
            <div>
              <p className="text-lg font-semibold">
                {currentPlan?.name || '未订阅套餐'}
              </p>
              {user?.expired_at && (
                <p className="text-sm text-muted-foreground">
                  {isExpired ? '已过期' : `剩余 ${daysLeft} 天`}
                  {' · '}
                  到期时间：{formatDate(user.expired_at)}
                </p>
              )}
            </div>
            <Badge variant={isExpired ? 'destructive' : 'default'}>
              {isExpired ? '已过期' : '有效'}
            </Badge>
          </div>

          {/* 流量使用进度 */}
          <div className="space-y-2">
            <div className="flex items-center justify-between text-sm">
              <span className="text-muted-foreground">流量使用</span>
              <span className="font-medium">
                {formatBytes(trafficUsed)} / {trafficTotal > 0 ? formatBytes(trafficTotal) : '无限制'}
              </span>
            </div>
            <Progress value={trafficPercent} className="h-2" />
            <p className="text-xs text-muted-foreground text-right">
              {trafficPercent.toFixed(1)}%
            </p>
          </div>

          {/* 账户余额和佣金 */}
          <div className="grid gap-4 md:grid-cols-2">
            <div className="flex items-center gap-3 rounded-lg border p-3">
              <div className="rounded-lg bg-blue-500/10 p-2">
                <Wallet className="h-5 w-5 text-blue-500" />
              </div>
              <div>
                <p className="text-sm text-muted-foreground">账户余额</p>
                <p className="text-lg font-semibold">{formatCurrency(user?.balance || 0)}</p>
              </div>
            </div>
            <div className="flex items-center gap-3 rounded-lg border p-3">
              <div className="rounded-lg bg-green-500/10 p-2">
                <ArrowUpRight className="h-5 w-5 text-green-500" />
              </div>
              <div>
                <p className="text-sm text-muted-foreground">佣金收入</p>
                <p className="text-lg font-semibold">{formatCurrency(user?.commission || 0)}</p>
              </div>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* 账户信息 */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <UserIcon className="h-5 w-5" />
            账户信息
          </CardTitle>
          <CardDescription>您的基本账户信息</CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="grid gap-4 md:grid-cols-2">
            <div>
              <Label className="text-muted-foreground">用户 ID</Label>
              <p className="text-sm font-medium mt-1">{user?.id}</p>
            </div>
            <div>
              <Label className="text-muted-foreground">邮箱</Label>
              <p className="text-sm font-medium mt-1">{user?.email}</p>
            </div>
            <div>
              <Label className="text-muted-foreground">注册时间</Label>
              <p className="text-sm font-medium mt-1">
                {user?.created_at ? formatDate(user.created_at) : '-'}
              </p>
            </div>
            <div>
              <Label className="text-muted-foreground">账户状态</Label>
              <p className="text-sm font-medium mt-1">
                {user?.status === 1 ? '正常' : '禁用'}
              </p>
            </div>
            <div>
              <Label className="text-muted-foreground">在线设备数</Label>
              <p className="text-sm font-medium mt-1">{user?.online_count || 0}</p>
            </div>
            <div>
              <Label className="text-muted-foreground">双重验证</Label>
              <p className="text-sm font-medium mt-1">
                {user?.two_factor_enabled ? '已启用' : '未启用'}
              </p>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* 修改邮箱 */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Mail className="h-5 w-5" />
            修改邮箱
          </CardTitle>
          <CardDescription>更新您的邮箱地址</CardDescription>
        </CardHeader>
        <CardContent>
          <form
            onSubmit={emailForm.handleSubmit((d) => updateEmail.mutate(d))}
            className="space-y-4"
          >
            <div className="space-y-2 max-w-sm">
              <Label htmlFor="email">邮箱</Label>
              <Input id="email" type="email" {...emailForm.register('email')} />
              {emailForm.formState.errors.email && (
                <p className="text-sm text-destructive">
                  {emailForm.formState.errors.email.message}
                </p>
              )}
            </div>
            <Button type="submit" disabled={updateEmail.isPending}>
              {updateEmail.isPending && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
              保存
            </Button>
          </form>
        </CardContent>
      </Card>

      {/* 修改密码 */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Lock className="h-5 w-5" />
            修改密码
          </CardTitle>
          <CardDescription>更新您的登录密码</CardDescription>
        </CardHeader>
        <CardContent>
          <form
            onSubmit={passwordForm.handleSubmit((d) => changePassword.mutate(d))}
            className="space-y-4"
          >
            <div className="space-y-4 max-w-sm">
              <div className="space-y-2">
                <Label htmlFor="old_password">当前密码</Label>
                <Input
                  id="old_password"
                  type="password"
                  {...passwordForm.register('old_password')}
                />
                {passwordForm.formState.errors.old_password && (
                  <p className="text-sm text-destructive">
                    {passwordForm.formState.errors.old_password.message}
                  </p>
                )}
              </div>
              <Separator />
              <div className="space-y-2">
                <Label htmlFor="new_password">新密码</Label>
                <Input
                  id="new_password"
                  type="password"
                  {...passwordForm.register('new_password')}
                />
                {passwordForm.formState.errors.new_password && (
                  <p className="text-sm text-destructive">
                    {passwordForm.formState.errors.new_password.message}
                  </p>
                )}
              </div>
              <div className="space-y-2">
                <Label htmlFor="confirm_password">确认新密码</Label>
                <Input
                  id="confirm_password"
                  type="password"
                  {...passwordForm.register('confirm_password')}
                />
                {passwordForm.formState.errors.confirm_password && (
                  <p className="text-sm text-destructive">
                    {passwordForm.formState.errors.confirm_password.message}
                  </p>
                )}
              </div>
            </div>
            <Button type="submit" disabled={changePassword.isPending}>
              {changePassword.isPending && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
              修改密码
            </Button>
          </form>
        </CardContent>
      </Card>
    </div>
  )
}
