import { useState, useEffect } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'
import api from '@/lib/api'
import type { Setting } from '@/types'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Button } from '@/components/ui/button'
import { Switch } from '@/components/ui/switch'
import { Textarea } from '@/components/ui/textarea'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Skeleton } from '@/components/ui/skeleton'
import { Save, Send, Loader2 } from 'lucide-react'

// --- Hooks ---
const useSettings = (group: string) =>
  useQuery({
    queryKey: ['admin', 'settings', group],
    queryFn: async () => (await api.get(`/admin/settings/group/${group}`)) as unknown as Setting[],
  })

const useUpdateSettings = (group: string) => {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (values: Record<string, string>) =>
      await api.put('/admin/settings', { settings: Object.entries(values).map(([key, value]) => ({ key, value, group })) }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['admin', 'settings'] })
      toast.success('设置保存成功')
    },
    onError: () => toast.error('设置保存失败'),
  })
}

const useTestEmail = () =>
  useMutation({
    mutationFn: async () => await api.post('/admin/settings/test-email'),
    onSuccess: () => toast.success('测试邮件已发送'),
    onError: () => toast.error('测试邮件发送失败'),
  })

const useSetTelegramWebhook = () =>
  useMutation({
    mutationFn: async () => await api.post('/admin/telegram/set-webhook', {
      webhook_url: '',
    }),
    onSuccess: () => toast.success('Webhook 设置成功'),
    onError: () => toast.error('Webhook 设置失败'),
  })

// --- Field Definition ---
interface FieldDef {
  key: string
  label: string
  description?: string
  placeholder?: string
  type?: 'text' | 'password' | 'number' | 'switch' | 'textarea' | 'select'
  options?: { value: string; label: string }[]
}

// --- Settings Tab Component ---
function SettingsTab({ group, fields, extraActions }: { group: string; fields: FieldDef[]; extraActions?: React.ReactNode }) {
  const { data: settings, isLoading } = useSettings(group)
  const updateSettings = useUpdateSettings(group)
  const [values, setValues] = useState<Record<string, string>>({})

  useEffect(() => {
    if (settings) {
      const map: Record<string, string> = {}
      settings.forEach((s) => { map[s.key] = s.value })
      setValues(map)
    }
  }, [settings])

  const handleSave = () => {
    updateSettings.mutate(values)
  }

  if (isLoading) return <div className="space-y-4">{Array.from({ length: 6 }).map((_, i) => <Skeleton key={i} className="h-16" />)}</div>

  return (
    <div className="space-y-6">
      {fields.map((f) => (
        <div key={f.key} className="space-y-2">
          {f.type === 'switch' ? (
            <div className="flex items-center justify-between rounded-lg border p-4">
              <div className="space-y-0.5">
                <Label className="text-base">{f.label}</Label>
                {f.description && <p className="text-sm text-muted-foreground">{f.description}</p>}
              </div>
              <Switch
                checked={values[f.key] === '1' || values[f.key] === 'true'}
                onCheckedChange={(v) => setValues((prev) => ({ ...prev, [f.key]: v ? '1' : '0' }))}
              />
            </div>
          ) : f.type === 'textarea' ? (
            <div className="space-y-2">
              <Label>{f.label}</Label>
              {f.description && <p className="text-sm text-muted-foreground">{f.description}</p>}
              <Textarea
                placeholder={f.placeholder}
                value={values[f.key] ?? ''}
                onChange={(e) => setValues((prev) => ({ ...prev, [f.key]: e.target.value }))}
                rows={4}
              />
            </div>
          ) : f.type === 'select' ? (
            <div className="space-y-2">
              <Label>{f.label}</Label>
              {f.description && <p className="text-sm text-muted-foreground">{f.description}</p>}
              <Select
                value={values[f.key] ?? ''}
                onValueChange={(v) => setValues((prev) => ({ ...prev, [f.key]: v }))}
              >
                <SelectTrigger>
                  <SelectValue placeholder={f.placeholder ?? '请选择'} />
                </SelectTrigger>
                <SelectContent>
                  {f.options?.map((opt) => (
                    <SelectItem key={opt.value} value={opt.value}>{opt.label}</SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
          ) : (
            <div className="space-y-2">
              <Label>{f.label}</Label>
              {f.description && <p className="text-sm text-muted-foreground">{f.description}</p>}
              <Input
                type={f.type ?? 'text'}
                placeholder={f.placeholder}
                value={values[f.key] ?? ''}
                onChange={(e) => setValues((prev) => ({ ...prev, [f.key]: e.target.value }))}
              />
            </div>
          )}
        </div>
      ))}
      <div className="flex items-center gap-2">
        <Button onClick={handleSave} disabled={updateSettings.isPending}>
          {updateSettings.isPending ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : <Save className="mr-2 h-4 w-4" />}
          保存
        </Button>
        {extraActions}
      </div>
    </div>
  )
}

// --- Field Definitions ---
const siteFields: FieldDef[] = [
  { key: 'app_name', label: '应用名称', placeholder: 'XBoard' },
  { key: 'app_url', label: '站点 URL', placeholder: 'https://example.com' },
  { key: 'app_description', label: '站点描述', placeholder: 'XBoard is best!' },
  { key: 'logo', label: '站点 Logo', placeholder: 'Logo URL' },
  { key: 'tos_url', label: '服务条款 URL', placeholder: 'https://example.com/tos' },
  { key: 'currency', label: '货币单位', placeholder: 'CNY' },
  { key: 'currency_symbol', label: '货币符号', placeholder: '¥' },
  { key: 'stop_register', label: '停止注册', type: 'switch', description: '开启后将关闭新用户注册' },
  { key: 'try_out_plan_id', label: '试用套餐 ID', placeholder: '0 表示不启用', type: 'number' },
  { key: 'try_out_hour', label: '试用时长(小时)', placeholder: '1', type: 'number' },
]

const subscribeFields: FieldDef[] = [
  { key: 'subscribe_path', label: '订阅路径', placeholder: 's', description: '自定义订阅路径前缀' },
  { key: 'subscribe_url', label: '订阅 URL', placeholder: '留空则使用站点 URL', description: '自定义订阅域名' },
  { key: 'plan_change_enable', label: '允许更换套餐', type: 'switch', description: '开启后用户可以更换订阅套餐' },
  { key: 'surplus_enable', label: '开启折抵', type: 'switch', description: '更换套餐时对原有套餐进行折抵' },
  { key: 'reset_traffic_method', label: '流量重置方式', type: 'select', options: [
    { value: '0', label: '每月1号' },
    { value: '1', label: '按月重置（按开通日）' },
    { value: '2', label: '每年1月1号' },
    { value: '3', label: '按年重置（按开通日）' },
    { value: '4', label: '不重置' },
  ]},
  { key: 'show_info_to_server_enable', label: '订阅中展示流量信息', type: 'switch', description: '在订阅节点中显示流量和到期信息' },
  { key: 'show_protocol_to_server_enable', label: '线路名显示协议', type: 'switch', description: '线路名附带协议前缀如 [Hy2]' },
  { key: 'default_remind_expire', label: '默认开启到期提醒', type: 'switch' },
  { key: 'default_remind_traffic', label: '默认开启流量提醒', type: 'switch' },
]

const serverFields: FieldDef[] = [
  { key: 'server_token', label: '通讯密钥', placeholder: '请输入通讯密钥', description: '节点与面板通讯的密钥' },
  { key: 'server_pull_interval', label: '拉取间隔(秒)', placeholder: '60', type: 'number', description: '节点拉取配置的间隔' },
  { key: 'server_push_interval', label: '推送间隔(秒)', placeholder: '60', type: 'number', description: '节点推送数据的间隔' },
  { key: 'device_limit_mode', label: '设备限制模式', type: 'select', options: [
    { value: '0', label: '不限制' },
    { value: '1', label: '宽松模式' },
    { value: '2', label: '严格模式' },
  ]},
]

const emailFields: FieldDef[] = [
  { key: 'email_host', label: 'SMTP 主机', placeholder: 'smtp.example.com' },
  { key: 'email_port', label: 'SMTP 端口', placeholder: '465', type: 'number' },
  { key: 'email_username', label: 'SMTP 用户名', placeholder: 'noreply@example.com' },
  { key: 'email_password', label: 'SMTP 密码', type: 'password' },
  { key: 'email_encryption', label: '加密方式', type: 'select', options: [
    { value: 'ssl', label: 'SSL' },
    { value: 'tls', label: 'TLS' },
    { value: 'none', label: '无' },
  ]},
  { key: 'email_from_address', label: '发件人地址', placeholder: 'noreply@example.com' },
  { key: 'remind_mail_enable', label: '启用邮件提醒', type: 'switch', description: '开启后将发送到期和流量提醒邮件' },
]

const telegramFields: FieldDef[] = [
  { key: 'telegram_bot_enable', label: '启用 Telegram Bot', type: 'switch', description: '开启后将启用 Telegram 机器人功能' },
  { key: 'telegram_bot_token', label: 'Bot Token', placeholder: '0000000000:xxxxxxxxx_xxxxxxxxxxxxxxx', description: '从 @BotFather 获取的令牌' },
  { key: 'telegram_webhook_url', label: 'Webhook Base URL', placeholder: 'https://example.com', description: '留空则使用站点 URL' },
  { key: 'telegram_discuss_link', label: '群组链接', placeholder: 'https://t.me/xxxxxx' },
]

const safeFields: FieldDef[] = [
  { key: 'email_verify', label: '强制邮箱验证', type: 'switch', description: '开启后用户必须验证邮箱才能使用' },
  { key: 'safe_mode_enable', label: '安全模式', type: 'switch', description: '开启后非站点域名访问将返回 403' },
  { key: 'secure_path', label: '后台路径', placeholder: 'admin', description: '自定义管理后台路径' },
  { key: 'email_whitelist_enable', label: '邮箱白名单', type: 'switch', description: '仅允许白名单中的邮箱后缀注册' },
  { key: 'email_whitelist_suffix', label: '邮箱后缀白名单', placeholder: '每行一个，如 @gmail.com', type: 'textarea' },
  { key: 'email_gmail_limit_enable', label: '禁止 Gmail 别名', type: 'switch', description: '禁止使用 Gmail 多别名注册' },
  { key: 'captcha_enable', label: '启用验证码', type: 'switch', description: '注册/登录时需要验证码' },
  { key: 'captcha_type', label: '验证码类型', type: 'select', options: [
    { value: 'turnstile', label: 'Cloudflare Turnstile' },
    { value: 'recaptcha', label: 'Google reCAPTCHA v2' },
    { value: 'recaptcha-v3', label: 'Google reCAPTCHA v3' },
  ]},
  { key: 'captcha_site_key', label: '验证码 Site Key', placeholder: '请输入 Site Key' },
  { key: 'captcha_secret_key', label: '验证码 Secret Key', placeholder: '请输入 Secret Key', type: 'password' },
  { key: 'captcha_min_score', label: '最低分数 (v3)', placeholder: '0.5', description: 'reCAPTCHA v3 的最低通过分数' },
  { key: 'register_limit_by_ip_enable', label: 'IP 注册限制', type: 'switch', description: '限制同一 IP 的注册次数' },
  { key: 'register_limit_count', label: '注册次数限制', placeholder: '3', type: 'number' },
  { key: 'register_limit_expire', label: '限制时长(秒)', placeholder: '3600', type: 'number' },
  { key: 'password_limit_enable', label: '密码尝试限制', type: 'switch', description: '限制密码错误次数' },
  { key: 'password_limit_count', label: '最大尝试次数', placeholder: '5', type: 'number' },
  { key: 'password_limit_expire', label: '锁定时长(秒)', placeholder: '60', type: 'number' },
]

const inviteFields: FieldDef[] = [
  { key: 'invite_force', label: '强制邀请注册', type: 'switch', description: '开启后只有被邀请的用户才能注册' },
  { key: 'invite_commission', label: '佣金比例(%)', placeholder: '10', type: 'number', description: '默认全局佣金分配比例' },
  { key: 'invite_gen_limit', label: '邀请码上限', placeholder: '5', type: 'number', description: '每个用户可创建的邀请码数量上限' },
  { key: 'invite_never_expire', label: '邀请码永不失效', type: 'switch', description: '开启后邀请码被使用后仍然有效' },
  { key: 'commission_first_time_enable', label: '佣金仅首次发放', type: 'switch', description: '仅被邀请人首次支付时产生佣金' },
  { key: 'commission_auto_check_enable', label: '佣金自动确认', type: 'switch', description: '订单完成3日后自动确认佣金' },
  { key: 'commission_withdraw_limit', label: '提现门槛(元)', placeholder: '100', type: 'number' },
  { key: 'commission_withdraw_method', label: '提现方式', placeholder: '支付宝,微信', description: '多个用逗号分隔' },
  { key: 'withdraw_close_enable', label: '关闭提现', type: 'switch', description: '关闭后佣金直接进入余额' },
]

const appFields: FieldDef[] = [
  { key: 'windows_version', label: 'Windows 版本', placeholder: '1.0.0' },
  { key: 'windows_download_url', label: 'Windows 下载地址', placeholder: 'https://...' },
  { key: 'macos_version', label: 'macOS 版本', placeholder: '1.0.0' },
  { key: 'macos_download_url', label: 'macOS 下载地址', placeholder: 'https://...' },
  { key: 'android_version', label: 'Android 版本', placeholder: '1.0.0' },
  { key: 'android_download_url', label: 'Android 下载地址', placeholder: 'https://...' },
]

// --- Tab Config ---
const tabConfigs = [
  { value: 'site', label: '站点设置', fields: siteFields },
  { value: 'subscribe', label: '订阅设置', fields: subscribeFields },
  { value: 'server', label: '节点配置', fields: serverFields },
  { value: 'email', label: '邮件设置', fields: emailFields, hasTestEmail: true },
  { value: 'telegram', label: 'Telegram', fields: telegramFields, hasWebhook: true },
  { value: 'safe', label: '安全设置', fields: safeFields },
  { value: 'invite', label: '邀请&佣金', fields: inviteFields },
  { value: 'app', label: '客户端', fields: appFields },
]

// --- Main Component ---
export default function SettingsPage() {
  const testEmail = useTestEmail()
  const setWebhook = useSetTelegramWebhook()

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold">系统设置</h1>
        <p className="text-muted-foreground">管理系统的各项配置</p>
      </div>
      <Card>
        <CardContent className="pt-6">
          <Tabs defaultValue="site">
            <TabsList className="mb-4 flex-wrap">
              {tabConfigs.map((t) => (
                <TabsTrigger key={t.value} value={t.value}>{t.label}</TabsTrigger>
              ))}
            </TabsList>
            {tabConfigs.map((t) => (
              <TabsContent key={t.value} value={t.value}>
                <SettingsTab
                  group={t.value}
                  fields={t.fields}
                  extraActions={
                    <>
                      {t.hasTestEmail && (
                        <Button variant="outline" onClick={() => testEmail.mutate()} disabled={testEmail.isPending}>
                          {testEmail.isPending ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : <Send className="mr-2 h-4 w-4" />}
                          发送测试邮件
                        </Button>
                      )}
                      {t.hasWebhook && (
                        <Button variant="outline" onClick={() => setWebhook.mutate()} disabled={setWebhook.isPending}>
                          {setWebhook.isPending ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : <Send className="mr-2 h-4 w-4" />}
                          设置 Webhook
                        </Button>
                      )}
                    </>
                  }
                />
              </TabsContent>
            ))}
          </Tabs>
        </CardContent>
      </Card>
    </div>
  )
}
