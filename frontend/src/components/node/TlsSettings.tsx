import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Switch } from '@/components/ui/switch'

export interface TlsSettingsProps {
  mode: 0 | 1 | 2
  value: Record<string, any>
  onChange: (val: Record<string, any>) => void
  protocol: string
}

const TLS_REQUIRED_PROTOCOLS = ['trojan', 'hysteria', 'hysteria2', 'tuic', 'anytls']

export function TlsSettings({ mode, value, onChange, protocol }: TlsSettingsProps) {
  const update = (key: string, v: any) => {
    onChange({ ...value, [key]: v })
  }

  const hideModeSelector = TLS_REQUIRED_PROTOCOLS.includes(protocol)

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
