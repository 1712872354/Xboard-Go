import { useState } from 'react'
import { toast } from 'sonner'
import { Copy, RefreshCw, Loader2, Download, Check, Link2, QrCode } from 'lucide-react'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useUserProfile } from '@/hooks/useAuth'
import api from '@/lib/api'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Badge } from '@/components/ui/badge'
import { Skeleton } from '@/components/ui/skeleton'
import { Separator } from '@/components/ui/separator'
import {
  Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle,
} from '@/components/ui/dialog'

interface SubscribeFormat {
  name: string
  format: string
  description: string
  icon: string
  clients: string[]
  color: string
}

const subscribeFormats: SubscribeFormat[] = [
  {
    name: 'Clash',
    format: 'clash',
    description: 'Clash for Windows / ClashX / Clash Verge',
    icon: '⚡',
    clients: ['Clash for Windows', 'ClashX', 'Clash Verge', 'Clash Verge Rev'],
    color: 'bg-blue-500',
  },
  {
    name: 'ClashMeta',
    format: 'clashmeta',
    description: 'Clash Meta / mihomo 内核客户端',
    icon: '🔧',
    clients: ['Clash Meta', 'mihomo', 'Clash Meta for Android'],
    color: 'bg-purple-500',
  },
  {
    name: 'V2Ray',
    format: 'v2ray',
    description: 'V2RayN / V2RayNG / V2RayU',
    icon: '🚀',
    clients: ['V2RayN (Windows)', 'V2RayNG (Android)', 'V2RayU (macOS)'],
    color: 'bg-green-500',
  },
  {
    name: 'Shadowrocket',
    format: 'shadowrocket',
    description: 'iOS 小火箭',
    icon: '🚀',
    clients: ['Shadowrocket (iOS)'],
    color: 'bg-red-500',
  },
  {
    name: 'Sing-box',
    format: 'sing-box',
    description: 'sing-box 通用客户端',
    icon: '📦',
    clients: ['sing-box (全平台)', 'SFA (Android)', 'SFM (macOS)'],
    color: 'bg-orange-500',
  },
  {
    name: 'Surge',
    format: 'surge',
    description: 'Surge 5 for Mac / iOS',
    icon: '🌊',
    clients: ['Surge 5 (macOS/iOS)'],
    color: 'bg-cyan-500',
  },
  {
    name: 'Loon',
    format: 'loon',
    description: 'Loon (iOS)',
    icon: '🎈',
    clients: ['Loon (iOS)'],
    color: 'bg-pink-500',
  },
  {
    name: 'Quantumult X',
    format: 'quantumultx',
    description: 'Quantumult X (iOS)',
    icon: '⚡',
    clients: ['Quantumult X (iOS)'],
    color: 'bg-indigo-500',
  },
  {
    name: 'Stash',
    format: 'stash',
    description: 'Stash (macOS/iOS)',
    icon: '💎',
    clients: ['Stash (macOS/iOS)'],
    color: 'bg-teal-500',
  },
  {
    name: 'Surfboard',
    format: 'surfboard',
    description: 'Surfboard (Android)',
    icon: '🏄',
    clients: ['Surfboard (Android)'],
    color: 'bg-amber-500',
  },
]

const clientDownloads = [
  { name: 'Clash Verge Rev', platform: 'Windows/macOS/Linux', url: 'https://github.com/clash-verge-rev/clash-verge-rev/releases' },
  { name: 'V2RayN', platform: 'Windows', url: 'https://github.com/2dust/v2rayN/releases' },
  { name: 'V2RayNG', platform: 'Android', url: 'https://github.com/2dust/v2rayNG/releases' },
  { name: 'Shadowrocket', platform: 'iOS', url: 'https://apps.apple.com/app/shadowrocket/id932747118' },
  { name: 'sing-box', platform: '全平台', url: 'https://github.com/SagerNet/sing-box/releases' },
  { name: 'Clash for Android', platform: 'Android', url: 'https://github.com/AkinoKaede/clash-for-android/releases' },
]

export default function SubscribePage() {
  const { data: user, isLoading } = useUserProfile()
  const queryClient = useQueryClient()
  const [showResetDialog, setShowResetDialog] = useState(false)
  const [copiedFormat, setCopiedFormat] = useState<string | null>(null)

  const baseUrl = window.location.origin
  const token = user?.subscribe_token || ''

  const getSubscribeUrl = (format: string) =>
    `${baseUrl}/api/v1/client/subscribe?token=${token}&format=${format}`

  const resetToken = useMutation({
    mutationFn: async () => await api.post('/user/subscribe/reset'),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['user', 'profile'] })
      toast.success('订阅链接已重置')
      setShowResetDialog(false)
    },
    onError: (err: any) => {
      toast.error(err?.message || '重置失败')
    },
  })

  const copyToClipboard = (text: string, format: string) => {
    navigator.clipboard.writeText(text)
    setCopiedFormat(format)
    toast.success('已复制到剪贴板')
    setTimeout(() => setCopiedFormat(null), 2000)
  }

  if (isLoading) {
    return (
      <div className="space-y-6">
        <Skeleton className="h-8 w-32" />
        <Skeleton className="h-40 w-full" />
        <Skeleton className="h-40 w-full" />
      </div>
    )
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold">订阅</h1>
          <p className="text-muted-foreground">管理您的订阅链接</p>
        </div>
        <Button
          variant="destructive"
          size="sm"
          onClick={() => setShowResetDialog(true)}
        >
          <RefreshCw className="mr-2 h-4 w-4" />
          重置订阅链接
        </Button>
      </div>

      {/* 主订阅链接 */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Link2 className="h-5 w-5" />
            通用订阅链接
          </CardTitle>
          <CardDescription>
            大部分客户端支持自动识别格式，推荐优先使用此链接
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="flex gap-2">
            <Input
              readOnly
              value={getSubscribeUrl('clash')}
              className="font-mono text-sm"
            />
            <Button
              variant="outline"
              size="icon"
              onClick={() => copyToClipboard(getSubscribeUrl('clash'), 'default')}
            >
              {copiedFormat === 'default' ? (
                <Check className="h-4 w-4 text-green-500" />
              ) : (
                <Copy className="h-4 w-4" />
              )}
            </Button>
          </div>
        </CardContent>
      </Card>

      {/* 按客户端格式展示 */}
      <Card>
        <CardHeader>
          <CardTitle>按客户端格式复制</CardTitle>
          <CardDescription>
            选择您使用的客户端，复制对应的订阅链接
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="grid gap-3 md:grid-cols-2">
            {subscribeFormats.map((fmt) => (
              <div
                key={fmt.format}
                className="flex items-center gap-3 rounded-lg border p-3 transition-colors hover:bg-muted/50"
              >
                <div className={`flex h-10 w-10 shrink-0 items-center justify-center rounded-lg ${fmt.color} text-white text-lg`}>
                  {fmt.icon}
                </div>
                <div className="flex-1 min-w-0">
                  <div className="flex items-center gap-2">
                    <p className="text-sm font-medium">{fmt.name}</p>
                    <Badge variant="outline" className="text-xs">
                      {fmt.clients[0]}
                    </Badge>
                  </div>
                  <p className="text-xs text-muted-foreground truncate">
                    {fmt.description}
                  </p>
                </div>
                <Button
                  variant="ghost"
                  size="icon"
                  className="shrink-0"
                  onClick={() => copyToClipboard(getSubscribeUrl(fmt.format), fmt.format)}
                >
                  {copiedFormat === fmt.format ? (
                    <Check className="h-4 w-4 text-green-500" />
                  ) : (
                    <Copy className="h-4 w-4" />
                  )}
                </Button>
              </div>
            ))}
          </div>
        </CardContent>
      </Card>

      {/* 客户端下载 */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Download className="h-5 w-5" />
            客户端下载
          </CardTitle>
          <CardDescription>
            推荐使用的代理客户端
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="grid gap-3 md:grid-cols-2 lg:grid-cols-3">
            {clientDownloads.map((client) => (
              <a
                key={client.name}
                href={client.url}
                target="_blank"
                rel="noopener noreferrer"
                className="flex items-center gap-3 rounded-lg border p-3 transition-colors hover:bg-muted"
              >
                <Download className="h-4 w-4 text-muted-foreground" />
                <div>
                  <p className="text-sm font-medium">{client.name}</p>
                  <p className="text-xs text-muted-foreground">{client.platform}</p>
                </div>
              </a>
            ))}
          </div>
        </CardContent>
      </Card>

      {/* 使用说明 */}
      <Card>
        <CardHeader>
          <CardTitle>使用说明</CardTitle>
        </CardHeader>
        <CardContent className="space-y-4 text-sm text-muted-foreground">
          <div className="space-y-2">
            <p className="font-medium text-foreground">1. 选择客户端</p>
            <p>根据您的操作系统和设备选择合适的代理客户端。</p>
          </div>
          <Separator />
          <div className="space-y-2">
            <p className="font-medium text-foreground">2. 复制订阅链接</p>
            <p>点击上方对应格式的复制按钮，获取订阅链接。</p>
          </div>
          <Separator />
          <div className="space-y-2">
            <p className="font-medium text-foreground">3. 导入订阅</p>
            <p>在客户端中添加订阅链接，即可自动获取节点配置。</p>
          </div>
          <Separator />
          <div className="space-y-2">
            <p className="font-medium text-foreground">4. 更新订阅</p>
            <p>客户端会定期自动更新订阅，也可以手动更新获取最新节点。</p>
          </div>
        </CardContent>
      </Card>

      {/* 重置确认对话框 */}
      <Dialog open={showResetDialog} onOpenChange={setShowResetDialog}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>重置订阅链接</DialogTitle>
            <DialogDescription>
              重置后旧的订阅链接将失效，请确保更新所有客户端的订阅地址。
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button variant="outline" onClick={() => setShowResetDialog(false)}>
              取消
            </Button>
            <Button
              variant="destructive"
              onClick={() => resetToken.mutate()}
              disabled={resetToken.isPending}
            >
              {resetToken.isPending && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
              确认重置
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}
