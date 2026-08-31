import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'

export interface TransportSettingsProps {
  transportType: string
  value: Record<string, any>
  onChange: (val: Record<string, any>) => void
}

export function TransportSettings({ transportType, value, onChange }: TransportSettingsProps) {
  const update = (key: string, v: any) => {
    onChange({ ...value, [key]: v })
  }

  if (transportType === 'ws') {
    return (
      <div className="grid grid-cols-2 gap-4">
        <div className="space-y-2">
          <Label>Path</Label>
          <Input placeholder="/ws" value={value.path ?? ''} onChange={(e) => update('path', e.target.value)} />
        </div>
        <div className="space-y-2">
          <Label>Host</Label>
          <Input placeholder="example.com" value={value.host ?? ''} onChange={(e) => update('host', e.target.value)} />
        </div>
        <div className="space-y-2">
          <Label>Max Early Data</Label>
          <Input type="number" placeholder="0" value={value.max_early_data ?? ''} onChange={(e) => update('max_early_data', Number(e.target.value))} />
        </div>
        <div className="space-y-2">
          <Label>Early Data Header Name</Label>
          <Input placeholder="" value={value.early_data_header_name ?? ''} onChange={(e) => update('early_data_header_name', e.target.value)} />
        </div>
      </div>
    )
  }

  if (transportType === 'grpc') {
    return (
      <div className="space-y-2">
        <Label>Service Name</Label>
        <Input placeholder="grpc-service" value={value.serviceName ?? ''} onChange={(e) => update('serviceName', e.target.value)} />
      </div>
    )
  }

  if (transportType === 'h2') {
    return (
      <div className="grid grid-cols-2 gap-4">
        <div className="space-y-2">
          <Label>Host</Label>
          <Input placeholder="example.com" value={value.host ?? ''} onChange={(e) => update('host', e.target.value)} />
        </div>
        <div className="space-y-2">
          <Label>Path</Label>
          <Input placeholder="/h2" value={value.path ?? ''} onChange={(e) => update('path', e.target.value)} />
        </div>
      </div>
    )
  }

  if (transportType === 'httpupgrade') {
    return (
      <div className="grid grid-cols-2 gap-4">
        <div className="space-y-2">
          <Label>Path</Label>
          <Input placeholder="/httpupgrade" value={value.path ?? ''} onChange={(e) => update('path', e.target.value)} />
        </div>
        <div className="space-y-2">
          <Label>Host</Label>
          <Input placeholder="example.com" value={value.host ?? ''} onChange={(e) => update('host', e.target.value)} />
        </div>
      </div>
    )
  }

  if (transportType === 'xhttp') {
    return (
      <div className="grid grid-cols-2 gap-4">
        <div className="space-y-2">
          <Label>Mode</Label>
          <Select value={value.mode ?? 'auto'} onValueChange={(v) => update('mode', v)}>
            <SelectTrigger><SelectValue placeholder="选择模式" /></SelectTrigger>
            <SelectContent>
              <SelectItem value="auto">auto</SelectItem>
              <SelectItem value="stream-one">stream-one</SelectItem>
              <SelectItem value="packet-up">packet-up</SelectItem>
            </SelectContent>
          </Select>
        </div>
        <div className="space-y-2">
          <Label>Path</Label>
          <Input placeholder="/xhttp" value={value.path ?? ''} onChange={(e) => update('path', e.target.value)} />
        </div>
        <div className="col-span-2 space-y-2">
          <Label>Host</Label>
          <Input placeholder="example.com" value={value.host ?? ''} onChange={(e) => update('host', e.target.value)} />
        </div>
      </div>
    )
  }

  // tcp / kcp: no extra fields
  return null
}
