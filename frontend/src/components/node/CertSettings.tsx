import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'

export interface CertSettingsValue {
  cert_mode?: string
  domain?: string
  email?: string
  http_port?: number
  dns_provider?: string
  dns_env?: string
  cert_content?: string
  key_content?: string
}

export interface CertSettingsProps {
  value: CertSettingsValue
  onChange: (val: CertSettingsValue) => void
}

const CERT_MODES = [
  { value: 'none', label: '无' },
  { value: 'http-01', label: 'HTTP-01' },
  { value: 'dns-01', label: 'DNS-01' },
  { value: 'self-signed', label: '自签名' },
  { value: 'content', label: '手动上传' },
]

export function CertSettings({ value, onChange }: CertSettingsProps) {
  const update = (key: string, v: any) => {
    onChange({ ...value, [key]: v })
  }

  const mode = value.cert_mode ?? 'none'

  return (
    <div className="space-y-3">
      <div className="space-y-2">
        <Label>证书模式</Label>
        <Select value={mode} onValueChange={(v) => update('cert_mode', v)}>
          <SelectTrigger><SelectValue placeholder="选择证书模式" /></SelectTrigger>
          <SelectContent>
            {CERT_MODES.map((m) => (
              <SelectItem key={m.value} value={m.value}>{m.label}</SelectItem>
            ))}
          </SelectContent>
        </Select>
      </div>

      {mode === 'http-01' && (
        <div className="grid grid-cols-2 gap-4">
          <div className="space-y-2">
            <Label>域名</Label>
            <Input placeholder="example.com" value={value.domain ?? ''} onChange={(e) => update('domain', e.target.value)} />
          </div>
          <div className="space-y-2">
            <Label>邮箱</Label>
            <Input placeholder="admin@example.com" value={value.email ?? ''} onChange={(e) => update('email', e.target.value)} />
          </div>
          <div className="space-y-2">
            <Label>HTTP 端口</Label>
            <Input type="number" placeholder="80" value={value.http_port ?? 80} onChange={(e) => update('http_port', Number(e.target.value))} />
          </div>
        </div>
      )}

      {mode === 'dns-01' && (
        <div className="space-y-3">
          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-2">
              <Label>域名</Label>
              <Input placeholder="example.com" value={value.domain ?? ''} onChange={(e) => update('domain', e.target.value)} />
            </div>
            <div className="space-y-2">
              <Label>邮箱</Label>
              <Input placeholder="admin@example.com" value={value.email ?? ''} onChange={(e) => update('email', e.target.value)} />
            </div>
          </div>
          <div className="space-y-2">
            <Label>DNS 提供商</Label>
            <Input placeholder="如：cloudflare, alidns" value={value.dns_provider ?? ''} onChange={(e) => update('dns_provider', e.target.value)} />
          </div>
          <div className="space-y-2">
            <Label>DNS 环境变量</Label>
            <Textarea
              placeholder="每行一个 KEY=VALUE，如：&#10;CF_API_TOKEN=your_token&#10;CF_ZONE_ID=your_zone_id"
              rows={4}
              value={value.dns_env ?? ''}
              onChange={(e) => update('dns_env', e.target.value)}
            />
            <p className="text-xs text-muted-foreground">每行一个 KEY=VALUE 格式的环境变量</p>
          </div>
        </div>
      )}

      {mode === 'self-signed' && (
        <div className="space-y-2">
          <Label>域名</Label>
          <Input placeholder="example.com" value={value.domain ?? ''} onChange={(e) => update('domain', e.target.value)} />
        </div>
      )}

      {mode === 'content' && (
        <div className="space-y-3">
          <div className="space-y-2">
            <Label>证书内容 (PEM)</Label>
            <Textarea
              placeholder="-----BEGIN CERTIFICATE-----&#10;...&#10;-----END CERTIFICATE-----"
              rows={6}
              className="font-mono text-xs"
              value={value.cert_content ?? ''}
              onChange={(e) => update('cert_content', e.target.value)}
            />
          </div>
          <div className="space-y-2">
            <Label>私钥内容 (PEM)</Label>
            <Textarea
              placeholder="-----BEGIN PRIVATE KEY-----&#10;...&#10;-----END PRIVATE KEY-----"
              rows={6}
              className="font-mono text-xs"
              value={value.key_content ?? ''}
              onChange={(e) => update('key_content', e.target.value)}
            />
          </div>
        </div>
      )}
    </div>
  )
}
