import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Switch } from '@/components/ui/switch'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'

export interface MultiplexSettingsValue {
  enabled?: boolean
  protocol?: string
  max_connections?: number
  min_streams?: number
  max_streams?: number
  padding?: boolean
  brutal?: {
    enabled?: boolean
    up_mbps?: number
    down_mbps?: number
  }
}

export interface MultiplexSettingsProps {
  value: MultiplexSettingsValue
  onChange: (val: MultiplexSettingsValue) => void
}

const MULTIPLEX_PROTOCOLS = [
  { value: 'smux', label: 'smux' },
  { value: 'yamux', label: 'yamux' },
  { value: 'h2mux', label: 'h2mux' },
]

export function MultiplexSettings({ value, onChange }: MultiplexSettingsProps) {
  const update = (key: string, v: any) => {
    onChange({ ...value, [key]: v })
  }

  const updateBrutal = (key: string, v: any) => {
    onChange({
      ...value,
      brutal: { ...(value.brutal ?? {}), [key]: v },
    })
  }

  return (
    <div className="space-y-4">
      <div className="flex items-center gap-3">
        <Switch checked={value.enabled ?? false} onCheckedChange={(checked) => update('enabled', checked)} />
        <Label>启用多路复用</Label>
      </div>

      {value.enabled && (
        <>
          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-2">
              <Label>协议</Label>
              <Select value={value.protocol ?? ''} onValueChange={(v) => update('protocol', v)}>
                <SelectTrigger><SelectValue placeholder="选择协议" /></SelectTrigger>
                <SelectContent>
                  {MULTIPLEX_PROTOCOLS.map((p) => (
                    <SelectItem key={p.value} value={p.value}>{p.label}</SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
            <div className="space-y-2">
              <Label>最大连接数</Label>
              <Input type="number" placeholder="如：4" value={value.max_connections ?? ''} onChange={(e) => update('max_connections', Number(e.target.value))} />
            </div>
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-2">
              <Label>最小流数</Label>
              <Input type="number" placeholder="如：1" value={value.min_streams ?? ''} onChange={(e) => update('min_streams', Number(e.target.value))} />
            </div>
            <div className="space-y-2">
              <Label>最大流数</Label>
              <Input type="number" placeholder="如：8" value={value.max_streams ?? ''} onChange={(e) => update('max_streams', Number(e.target.value))} />
            </div>
          </div>

          <div className="flex items-center gap-3">
            <Switch checked={value.padding ?? false} onCheckedChange={(checked) => update('padding', checked)} />
            <Label>启用填充 (Padding)</Label>
          </div>

          {/* TCP Brutal */}
          <div className="rounded-lg border p-4 space-y-3">
            <div className="flex items-center gap-3">
              <Switch
                checked={value.brutal?.enabled ?? false}
                onCheckedChange={(checked) => updateBrutal('enabled', checked)}
              />
              <Label>TCP Brutal</Label>
            </div>

            {value.brutal?.enabled && (
              <div className="grid grid-cols-2 gap-4">
                <div className="space-y-2">
                  <Label>上行带宽 (Mbps)</Label>
                  <Input type="number" placeholder="100" value={value.brutal?.up_mbps ?? ''} onChange={(e) => updateBrutal('up_mbps', Number(e.target.value))} />
                </div>
                <div className="space-y-2">
                  <Label>下行带宽 (Mbps)</Label>
                  <Input type="number" placeholder="100" value={value.brutal?.down_mbps ?? ''} onChange={(e) => updateBrutal('down_mbps', Number(e.target.value))} />
                </div>
              </div>
            )}
          </div>
        </>
      )}
    </div>
  )
}
