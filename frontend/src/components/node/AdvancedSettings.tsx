import { useState } from 'react'
import { Dialog, DialogContent, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Textarea } from '@/components/ui/textarea'
import { ScrollArea } from '@/components/ui/scroll-area'
import { Label } from '@/components/ui/label'
import { CertSettings, type CertSettingsValue } from '@/components/node/CertSettings'
import { MultiplexSettings, type MultiplexSettingsValue } from '@/components/node/MultiplexSettings'
import { EchSettings, type EchSettingsValue } from '@/components/node/EchSettings'

export interface AdvancedSettingsProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  certValue: CertSettingsValue
  certOnChange: (val: CertSettingsValue) => void
  multiplexValue: MultiplexSettingsValue
  multiplexOnChange: (val: MultiplexSettingsValue) => void
  echValue: EchSettingsValue
  echOnChange: (val: EchSettingsValue) => void
}

export function AdvancedSettings({
  open,
  onOpenChange,
  certValue,
  certOnChange,
  multiplexValue,
  multiplexOnChange,
  echValue,
  echOnChange,
}: AdvancedSettingsProps) {
  const [customOutbound, setCustomOutbound] = useState('{}')
  const [customRoute, setCustomRoute] = useState('{}')

  // Show ECH settings when TLS mode is standard TLS (cert_mode is not 'none')
  const showEch = certValue.cert_mode && certValue.cert_mode !== 'none'

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-2xl max-h-[85vh]">
        <DialogHeader>
          <DialogTitle>高级设置</DialogTitle>
        </DialogHeader>
        <ScrollArea className="max-h-[calc(85vh-120px)]">
          <div className="pr-4">
            <Tabs defaultValue="cert" className="space-y-4">
              <TabsList className="w-full justify-start">
                <TabsTrigger value="cert">证书配置</TabsTrigger>
                <TabsTrigger value="multiplex">多路复用</TabsTrigger>
                <TabsTrigger value="custom-outbound">自定义出站</TabsTrigger>
                <TabsTrigger value="custom-route">自定义路由</TabsTrigger>
              </TabsList>

              <TabsContent value="cert" className="space-y-4">
                <CertSettings value={certValue} onChange={certOnChange} />
                {showEch && (
                  <div className="rounded-lg border p-4 space-y-3">
                    <Label className="text-sm font-medium">ECH (Encrypted Client Hello)</Label>
                    <EchSettings value={echValue} onChange={echOnChange} />
                  </div>
                )}
              </TabsContent>

              <TabsContent value="multiplex">
                <MultiplexSettings value={multiplexValue} onChange={multiplexOnChange} />
              </TabsContent>

              <TabsContent value="custom-outbound">
                <div className="space-y-2">
                  <Label>自定义出站配置 (JSON)</Label>
                  <Textarea
                    placeholder={'[\n  {\n    "protocol": "freedom",\n    "tag": "direct"\n  }\n]'}
                    rows={12}
                    className="font-mono text-xs"
                    value={customOutbound}
                    onChange={(e) => setCustomOutbound(e.target.value)}
                  />
                  <p className="text-xs text-muted-foreground">自定义 sing-box 出站配置，格式为 JSON 数组</p>
                </div>
              </TabsContent>

              <TabsContent value="custom-route">
                <div className="space-y-2">
                  <Label>自定义路由规则 (JSON)</Label>
                  <Textarea
                    placeholder={'{\n  "rules": [\n    {\n      "domain_suffix": [".cn"],\n      "outbound": "direct"\n    }\n  ]\n}'}
                    rows={12}
                    className="font-mono text-xs"
                    value={customRoute}
                    onChange={(e) => setCustomRoute(e.target.value)}
                  />
                  <p className="text-xs text-muted-foreground">自定义 sing-box 路由规则，格式为 JSON 对象</p>
                </div>
              </TabsContent>
            </Tabs>
          </div>
        </ScrollArea>
      </DialogContent>
    </Dialog>
  )
}
