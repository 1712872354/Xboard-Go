import { useCallback, useMemo } from 'react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Switch } from '@/components/ui/switch'

// ---------------------------------------------------------------------------
// Types
// ---------------------------------------------------------------------------

export interface ProtocolSettingsProps {
  protocol: string
  value: Record<string, any>
  onChange: (val: Record<string, any>) => void
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

/** Shorthand to merge a single key change into the value object. */
function useUpdate(
  value: Record<string, any>,
  onChange: (val: Record<string, any>) => void,
) {
  return useCallback(
    (key: string, v: any) => {
      onChange({ ...value, [key]: v })
    },
    [value, onChange],
  )
}

/** Generate a random alphanumeric string (16 chars). */
function generateRandomString(length = 16): string {
  const chars = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789'
  const arr = new Uint8Array(length)
  crypto.getRandomValues(arr)
  return Array.from(arr, (b) => chars[b % chars.length]).join('')
}

/** Generate a random hex string (suitable for passwords / keys). */
function generateRandomHex(length = 16): string {
  const arr = new Uint8Array(length)
  crypto.getRandomValues(arr)
  return Array.from(arr, (b) => b.toString(16).padStart(2, '0')).join('')
}

// ---------------------------------------------------------------------------
// Constants
// ---------------------------------------------------------------------------

const SHADOWSOCKS_CIPHER_PRESETS = [
  'aes-256-gcm',
  'chacha20-ietf-poly1305',
  '2022-blake3-aes-256-gcm',
  '2022-blake3-chacha20-poly1305',
  'aes-128-gcm',
  'xchacha20-ietf-poly1305',
]

const SHADOWSOCKS_PLUGIN_OPTIONS = [
  'none',
  'obfs',
  'v2ray-plugin',
  'gost-plugin',
  'shadow-tls',
  'restls',
  'kcptun',
]

const SHADOWSOCKS_FINGERPRINT_PLUGINS = ['shadow-tls', 'restls']

const HYSTERIA_OBFS_TYPES = ['salamander']

const TUIC_VERSIONS = [
  { value: '4', label: 'V4' },
  { value: '5', label: 'V5' },
]

const TUIC_CONGESTION_CONTROLS = ['cubic', 'bbr', 'new_reno']

const TUIC_UDP_RELAY_MODES = ['native', 'quic']

const MIERU_TRANSPORTS = ['TCP']

// ---------------------------------------------------------------------------
// Sub-sections per protocol
// ---------------------------------------------------------------------------

// ---- Shadowsocks ---------------------------------------------------------

function ShadowsocksSettings({
  value,
  update,
}: {
  value: Record<string, any>
  update: (key: string, v: any) => void
}) {
  const currentCipher: string = value.cipher ?? ''
  const isCustomCipher = useMemo(
    () => currentCipher !== '' && !SHADOWSOCKS_CIPHER_PRESETS.includes(currentCipher),
    [currentCipher],
  )
  const cipherSelectValue = isCustomCipher ? '__custom__' : currentCipher

  const plugin: string = value.plugin ?? 'none'
  const showPluginOpts = plugin !== 'none'
  const showFingerprint = SHADOWSOCKS_FINGERPRINT_PLUGINS.includes(plugin)

  return (
    <div className="space-y-4">
      {/* Cipher */}
      <div className="grid grid-cols-2 gap-4">
        <div className="space-y-2">
          <Label>加密方式 (Cipher)</Label>
          <Select
            value={cipherSelectValue}
            onValueChange={(v) => {
              if (v === '__custom__') {
                update('cipher', '')
              } else {
                update('cipher', v)
              }
            }}
          >
            <SelectTrigger>
              <SelectValue placeholder="选择加密方式" />
            </SelectTrigger>
            <SelectContent>
              {SHADOWSOCKS_CIPHER_PRESETS.map((c) => (
                <SelectItem key={c} value={c}>
                  {c}
                </SelectItem>
              ))}
              <SelectItem value="__custom__">自定义...</SelectItem>
            </SelectContent>
          </Select>
        </div>

        {isCustomCipher && (
          <div className="space-y-2">
            <Label>自定义加密方式</Label>
            <Input
              placeholder="输入自定义加密方式"
              value={currentCipher}
              onChange={(e) => update('cipher', e.target.value)}
            />
          </div>
        )}
      </div>

      {/* Plugin */}
      <div className="grid grid-cols-2 gap-4">
        <div className="space-y-2">
          <Label>插件 (Plugin)</Label>
          <Select value={plugin} onValueChange={(v) => update('plugin', v)}>
            <SelectTrigger>
              <SelectValue placeholder="选择插件" />
            </SelectTrigger>
            <SelectContent>
              {SHADOWSOCKS_PLUGIN_OPTIONS.map((p) => (
                <SelectItem key={p} value={p}>
                  {p}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>

        {showPluginOpts && (
          <div className="space-y-2">
            <Label>插件参数 (Plugin Opts)</Label>
            <Input
              placeholder="插件参数"
              value={value.plugin_opts ?? ''}
              onChange={(e) => update('plugin_opts', e.target.value)}
            />
          </div>
        )}
      </div>

      {/* Client fingerprint (for shadow-tls / restls) */}
      {showFingerprint && (
        <div className="space-y-2">
          <Label>客户端指纹 (Client Fingerprint)</Label>
          <Input
            placeholder="如：chrome, firefox"
            value={value.client_fingerprint ?? ''}
            onChange={(e) => update('client_fingerprint', e.target.value)}
          />
        </div>
      )}
    </div>
  )
}

// ---- VMess ---------------------------------------------------------------

function VmessSettings() {
  return (
    <p className="text-sm text-muted-foreground">
      VMess 无需额外协议设置，TLS 与传输层设置请在其他区域配置。
    </p>
  )
}

// ---- VLESS ---------------------------------------------------------------

function VlessSettings({
  value,
  update,
}: {
  value: Record<string, any>
  update: (key: string, v: any) => void
}) {
  const encryptionEnabled: boolean = !!value.encryption

  return (
    <div className="space-y-4">
      {/* Flow */}
      <div className="space-y-2">
        <Label>流控 (Flow)</Label>
        <Select
          value={value.flow ?? ''}
          onValueChange={(v) => update('flow', v)}
        >
          <SelectTrigger>
            <SelectValue placeholder="无" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="">无</SelectItem>
            <SelectItem value="xtls-rprx-vision">xtls-rprx-vision</SelectItem>
          </SelectContent>
        </Select>
      </div>

      {/* Encryption toggle */}
      <div className="flex items-center gap-3">
        <Switch
          id="vless-encryption"
          checked={encryptionEnabled}
          onCheckedChange={(checked) => {
            update('encryption', checked)
            if (!checked) {
              update('decryption', '')
            }
          }}
        />
        <Label htmlFor="vless-encryption">启用加密 (Encryption)</Label>
      </div>

      {/* Decryption key */}
      {encryptionEnabled && (
        <div className="space-y-2">
          <Label>加密密钥 (Decryption)</Label>
          <Input
            placeholder="输入加密密钥"
            value={value.decryption ?? ''}
            onChange={(e) => update('decryption', e.target.value)}
          />
        </div>
      )}
    </div>
  )
}

// ---- Trojan --------------------------------------------------------------

function TrojanSettings() {
  return (
    <p className="text-sm text-muted-foreground">
      Trojan 无需额外协议设置，TLS 与传输层设置请在其他区域配置。
    </p>
  )
}

// ---- Hysteria / Hysteria2 ------------------------------------------------

function HysteriaSettings({
  value,
  update,
  protocol,
}: {
  value: Record<string, any>
  update: (key: string, v: any) => void
  protocol: string
}) {
  const version: number = value.version ?? 2
  const obfsEnabled: boolean = !!value.obfs_enabled
  const isV1 = version === 1

  return (
    <div className="space-y-4">
      {/* Version */}
      <div className="space-y-2">
        <Label>版本 (Version)</Label>
        <Select
          value={String(version)}
          onValueChange={(v) => update('version', Number(v))}
        >
          <SelectTrigger>
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="1">V1</SelectItem>
            <SelectItem value="2">V2</SelectItem>
          </SelectContent>
        </Select>
      </div>

      {/* Obfs toggle */}
      <div className="flex items-center gap-3">
        <Switch
          id="hysteria-obfs"
          checked={obfsEnabled}
          onCheckedChange={(checked) => {
            update('obfs_enabled', checked)
            if (!checked) {
              update('obfs_type', '')
              update('obfs_password', '')
            }
          }}
        />
        <Label htmlFor="hysteria-obfs">启用混淆 (Obfs)</Label>
      </div>

      {/* Obfs settings */}
      {obfsEnabled && (
        <div className="grid grid-cols-2 gap-4">
          <div className="space-y-2">
            <Label>混淆类型 (Obfs Type)</Label>
            <Select
              value={value.obfs_type ?? ''}
              onValueChange={(v) => update('obfs_type', v)}
            >
              <SelectTrigger>
                <SelectValue placeholder="选择混淆类型" />
              </SelectTrigger>
              <SelectContent>
                {HYSTERIA_OBFS_TYPES.map((t) => (
                  <SelectItem key={t} value={t}>
                    {t}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>

          <div className="space-y-2">
            <Label>混淆密码 (Obfs Password)</Label>
            <div className="flex gap-2">
              <Input
                placeholder="混淆密码"
                value={value.obfs_password ?? ''}
                onChange={(e) => update('obfs_password', e.target.value)}
              />
              <Button
                type="button"
                variant="outline"
                size="sm"
                onClick={() => update('obfs_password', generateRandomString())}
              >
                生成
              </Button>
            </div>
          </div>
        </div>
      )}

      {/* Bandwidth */}
      <div className="grid grid-cols-2 gap-4">
        <div className="space-y-2">
          <Label>上行带宽 (Mbps)</Label>
          <Input
            type="number"
            placeholder="留空则使用 BBR"
            value={value.bandwidth_up ?? ''}
            onChange={(e) =>
              update('bandwidth_up', e.target.value === '' ? undefined : Number(e.target.value))
            }
          />
        </div>
        <div className="space-y-2">
          <Label>下行带宽 (Mbps)</Label>
          <Input
            type="number"
            placeholder="下行带宽"
            value={value.bandwidth_down ?? ''}
            onChange={(e) =>
              update('bandwidth_down', e.target.value === '' ? undefined : Number(e.target.value))
            }
          />
        </div>
      </div>

      {/* Hop interval - V1 only */}
      {isV1 && (
        <div className="space-y-2">
          <Label>跳跃间隔 (Hop Interval, 秒)</Label>
          <Input
            type="number"
            placeholder="跳跃间隔（秒）"
            value={value.hop_interval ?? ''}
            onChange={(e) =>
              update('hop_interval', e.target.value === '' ? undefined : Number(e.target.value))
            }
          />
        </div>
      )}
    </div>
  )
}

// ---- TUIC ----------------------------------------------------------------

function TuicSettings({
  value,
  update,
}: {
  value: Record<string, any>
  update: (key: string, v: any) => void
}) {
  return (
    <div className="space-y-4">
      {/* Version */}
      <div className="space-y-2">
        <Label>版本 (Version)</Label>
        <Select
          value={String(value.version ?? '5')}
          onValueChange={(v) => update('version', Number(v))}
        >
          <SelectTrigger>
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            {TUIC_VERSIONS.map((v) => (
              <SelectItem key={v.value} value={v.value}>
                {v.label}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      </div>

      <div className="grid grid-cols-2 gap-4">
        {/* Congestion control */}
        <div className="space-y-2">
          <Label>拥塞控制 (Congestion Control)</Label>
          <Select
            value={value.congestion_control ?? 'cubic'}
            onValueChange={(v) => update('congestion_control', v)}
          >
            <SelectTrigger>
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              {TUIC_CONGESTION_CONTROLS.map((c) => (
                <SelectItem key={c} value={c}>
                  {c}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>

        {/* UDP relay mode */}
        <div className="space-y-2">
          <Label>UDP 中继模式 (UDP Relay Mode)</Label>
          <Select
            value={value.udp_relay_mode ?? 'native'}
            onValueChange={(v) => update('udp_relay_mode', v)}
          >
            <SelectTrigger>
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              {TUIC_UDP_RELAY_MODES.map((m) => (
                <SelectItem key={m} value={m}>
                  {m}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>
      </div>

      {/* Password */}
      <div className="space-y-2">
        <Label>密码 (Password)</Label>
        <div className="flex gap-2">
          <Input
            placeholder="密码"
            value={value.password ?? ''}
            onChange={(e) => update('password', e.target.value)}
          />
          <Button
            type="button"
            variant="outline"
            size="sm"
            onClick={() => update('password', generateRandomString())}
          >
            生成
          </Button>
        </div>
      </div>

      {/* ALPN */}
      <div className="space-y-2">
        <Label>ALPN</Label>
        <Input
          placeholder="如：h3"
          value={value.alpn ?? ''}
          onChange={(e) => update('alpn', e.target.value)}
        />
      </div>
    </div>
  )
}

// ---- SOCKS ---------------------------------------------------------------

function SocksSettings({
  value,
  update,
}: {
  value: Record<string, any>
  update: (key: string, v: any) => void
}) {
  return (
    <div className="space-y-2">
      <Label>版本 (Version)</Label>
      <Select
        value={String(value.version ?? '5')}
        onValueChange={(v) => update('version', Number(v))}
      >
        <SelectTrigger>
          <SelectValue />
        </SelectTrigger>
        <SelectContent>
          <SelectItem value="5">SOCKS5</SelectItem>
        </SelectContent>
      </Select>
    </div>
  )
}

// ---- HTTP ----------------------------------------------------------------

function HttpSettings() {
  return (
    <p className="text-sm text-muted-foreground">
      HTTP 无需额外协议设置。
    </p>
  )
}

// ---- Naive ---------------------------------------------------------------

function NaiveSettings() {
  return (
    <p className="text-sm text-muted-foreground">
      Naive 无需额外协议设置。
    </p>
  )
}

// ---- AnyTLS --------------------------------------------------------------

function AnyTlsSettings({
  value,
  update,
}: {
  value: Record<string, any>
  update: (key: string, v: any) => void
}) {
  return (
    <div className="space-y-2">
      <Label>填充方案 (Padding Scheme)</Label>
      <Input
        placeholder='输入 JSON 或 "default"'
        value={value.padding_scheme ?? ''}
        onChange={(e) => update('padding_scheme', e.target.value)}
      />
    </div>
  )
}

// ---- Mieru ---------------------------------------------------------------

function MieruSettings({
  value,
  update,
}: {
  value: Record<string, any>
  update: (key: string, v: any) => void
}) {
  return (
    <div className="grid grid-cols-2 gap-4">
      <div className="space-y-2">
        <Label>传输协议 (Transport)</Label>
        <Select
          value={value.transport ?? 'TCP'}
          onValueChange={(v) => update('transport', v)}
        >
          <SelectTrigger>
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            {MIERU_TRANSPORTS.map((t) => (
              <SelectItem key={t} value={t}>
                {t}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      </div>

      <div className="space-y-2">
        <Label>流量模式 (Traffic Pattern, Base64)</Label>
        <Input
          placeholder="Base64 编码的流量模式"
          value={value.traffic_pattern ?? ''}
          onChange={(e) => update('traffic_pattern', e.target.value)}
        />
      </div>
    </div>
  )
}

// ---------------------------------------------------------------------------
// Main component
// ---------------------------------------------------------------------------

export function ProtocolSettings({
  protocol,
  value,
  onChange,
}: ProtocolSettingsProps) {
  const update = useUpdate(value, onChange)

  switch (protocol) {
    case 'shadowsocks':
      return <ShadowsocksSettings value={value} update={update} />
    case 'vmess':
      return <VmessSettings />
    case 'vless':
      return <VlessSettings value={value} update={update} />
    case 'trojan':
      return <TrojanSettings />
    case 'hysteria':
    case 'hysteria2':
      return <HysteriaSettings value={value} update={update} protocol={protocol} />
    case 'tuic':
      return <TuicSettings value={value} update={update} />
    case 'socks':
      return <SocksSettings value={value} update={update} />
    case 'http':
      return <HttpSettings />
    case 'naive':
      return <NaiveSettings />
    case 'anytls':
      return <AnyTlsSettings value={value} update={update} />
    case 'mieru':
      return <MieruSettings value={value} update={update} />
    default:
      return null
  }
}
