import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'

export interface EchSettingsValue {
  config?: string
  config_path?: string
  query_server_name?: string
  key?: string
  key_path?: string
}

export interface EchSettingsProps {
  value: EchSettingsValue
  onChange: (val: EchSettingsValue) => void
}

export function EchSettings({ value, onChange }: EchSettingsProps) {
  const update = (key: string, v: any) => {
    onChange({ ...value, [key]: v })
  }

  return (
    <div className="space-y-3">
      <div className="space-y-2">
        <Label>ECH 配置 (PEM)</Label>
        <Textarea
          placeholder="-----BEGIN ECH CONFIGS-----&#10;...&#10;-----END ECH CONFIGS-----"
          rows={4}
          className="font-mono text-xs"
          value={value.config ?? ''}
          onChange={(e) => update('config', e.target.value)}
        />
      </div>

      <div className="space-y-2">
        <Label>ECH 配置文件路径</Label>
        <Input placeholder="/path/to/ech_config.pem" value={value.config_path ?? ''} onChange={(e) => update('config_path', e.target.value)} />
      </div>

      <div className="space-y-2">
        <Label>查询服务器名称</Label>
        <Input placeholder="example.com" value={value.query_server_name ?? ''} onChange={(e) => update('query_server_name', e.target.value)} />
      </div>

      <div className="grid grid-cols-2 gap-4">
        <div className="space-y-2">
          <Label>ECH 密钥</Label>
          <Input placeholder="ECH 私钥" value={value.key ?? ''} onChange={(e) => update('key', e.target.value)} />
        </div>
        <div className="space-y-2">
          <Label>密钥文件路径</Label>
          <Input placeholder="/path/to/ech_key.pem" value={value.key_path ?? ''} onChange={(e) => update('key_path', e.target.value)} />
        </div>
      </div>
    </div>
  )
}
