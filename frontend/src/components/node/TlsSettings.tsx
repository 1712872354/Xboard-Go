import { useState } from 'react'
import { toast } from 'sonner'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Switch } from '@/components/ui/switch'
import { Button } from '@/components/ui/button'
import { Loader2, KeyRound } from 'lucide-react'
import api from '@/lib/api'

export interface TlsSettingsProps {
  mode: 0 | 1 | 2
  value: Record<string, any>
  onChange: (val: Record<string, any>) => void
  protocol: string
}

const TLS_REQUIRED_PROTOCOLS = ['trojan', 'hysteria', 'hysteria2', 'tuic', 'anytls']

export function TlsSettings({ mode, value, onChange, protocol }: TlsSettingsProps) {
  const [echGenerating, setEchGenerating] = useState(false)

  const update = (key: string, v: any) => {
    onChange({ ...value, [key]: v })
  }

  const hideModeSelector = TLS_REQUIRED_PROTOCOLS.includes(protocol)

  const handleGenerateEchKey = async () => {
    const publicName = value.server_name || value.ech_query_server_name || ''
    if (!publicName) {
      toast.error('请先填写 Server Name 或查询服务器名称')
      return
    }
    setEchGenerating(true)
    try {
      const res = await api.get('/admin/nodes/generate-ech-key', { params: { public_name: publicName } }) as unknown as { key: string; config: string }
      onChange({
        ...value,
        ech_key: res.key,
        ech_config: res.config,
      })
      toast.success('ECH 密钥已生成')
    } catch {
      toast.error('ECH 密钥生成失败')
    } finally {
      setEchGenerating(false)
    }
  }

  return (
    <div className="space-y-3">
      {!hideModeSelector && (
        <div className="space-y-2">
          <Label>TLS 模式</Label>
          <Select value={String(mode)} onValueChange={(v) => onChange({ mode: Number(v) })}>
            <SelectTrigger><SelectValue /></SelectTrigger>
            <SelectContent>
              <SelectItem value="0">关闭</SelectItem>
              <SelectItem value="1">标准 TLS</SelectItem>
              <SelectItem value="2">Reality</SelectItem>
            </SelectContent>
          </Select>
        </div>
      )}

      {mode === 1 && (
        <div className="space-y-3">
          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-2">
              <Label>Server Name (SNI)</Label>
              <Input placeholder="example.com" value={value.server_name ?? ''} onChange={(e) => update('server_name', e.target.value)} />
            </div>
            <div className="space-y-2">
              <Label>ALPN</Label>
              <Input placeholder="h2,http/1.1（逗号分隔）" value={value.alpn ?? ''} onChange={(e) => update('alpn', e.target.value)} />
            </div>
          </div>
          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-2">
              <Label>uTLS 指纹</Label>
              <Select value={value.fingerprint ?? ''} onValueChange={(v) => update('fingerprint', v)}>
                <SelectTrigger><SelectValue placeholder="选择指纹" /></SelectTrigger>
                <SelectContent>
                  <SelectItem value="">不使用</SelectItem>
                  <SelectItem value="chrome">Chrome</SelectItem>
                  <SelectItem value="firefox">Firefox</SelectItem>
                  <SelectItem value="safari">Safari</SelectItem>
                  <SelectItem value="edge">Edge</SelectItem>
                  <SelectItem value="random">Random</SelectItem>
                </SelectContent>
              </Select>
            </div>
            <div className="flex items-end">
              <div className="flex items-center gap-2">
                <Switch checked={!!value.allow_insecure} onCheckedChange={(v) => update('allow_insecure', v)} />
                <Label>允许不安全连接</Label>
              </div>
            </div>
          </div>
          {/* ECH Settings */}
          <div className="rounded-lg border p-3 space-y-3">
            <Label className="text-sm font-medium">ECH (Encrypted Client Hello)</Label>
            <div className="grid grid-cols-2 gap-4">
              <div className="space-y-2">
                <Label className="text-xs text-muted-foreground">查询服务器名称</Label>
                <Input placeholder="example.com" value={value.ech_query_server_name ?? ''} onChange={(e) => update('ech_query_server_name', e.target.value)} />
              </div>
              <div className="flex items-end">
                <Button type="button" variant="outline" size="sm" onClick={handleGenerateEchKey} disabled={echGenerating}>
                  {echGenerating ? <Loader2 className="mr-1.5 h-3.5 w-3.5 animate-spin" /> : <KeyRound className="mr-1.5 h-3.5 w-3.5" />}
                  生成ECH密钥
                </Button>
              </div>
            </div>
            <div className="space-y-2">
              <Label className="text-xs text-muted-foreground">ECH 配置</Label>
              <Textarea
                placeholder="-----BEGIN ECH CONFIGS-----&#10;...&#10;-----END ECH CONFIGS-----"
                rows={3}
                className="font-mono text-xs"
                value={value.ech_config ?? ''}
                onChange={(e) => update('ech_config', e.target.value)}
              />
            </div>
            <div className="space-y-2">
              <Label className="text-xs text-muted-foreground">ECH 密钥</Label>
              <Input placeholder="ECH 私钥" className="font-mono text-xs" value={value.ech_key ?? ''} onChange={(e) => update('ech_key', e.target.value)} />
            </div>
          </div>
        </div>
      )}

      {mode === 2 && (
        <div className="space-y-3">
          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-2">
              <Label>Private Key</Label>
              <Input placeholder="Private key" value={value.private_key ?? ''} onChange={(e) => update('private_key', e.target.value)} />
            </div>
            <div className="space-y-2">
              <Label>Short ID</Label>
              <Input placeholder="Short ID（最多16字符）" value={value.short_id ?? ''} onChange={(e) => update('short_id', e.target.value)} />
            </div>
            <div className="space-y-2">
              <Label>Dest / Server Name</Label>
              <Input placeholder="example.com:443" value={value.server_name ?? value.dest ?? ''} onChange={(e) => update('server_name', e.target.value)} />
            </div>
            <div className="space-y-2">
              <Label>握手端口</Label>
              <Input type="number" placeholder="443" value={value.server_port ?? ''} onChange={(e) => update('server_port', Number(e.target.value) || undefined)} />
            </div>
          </div>
          <div className="flex items-center gap-2">
            <Switch checked={!!value.allow_insecure} onCheckedChange={(v) => update('allow_insecure', v)} />
            <Label>允许不安全连接</Label>
          </div>
        </div>
      )}
    </div>
  )
}
