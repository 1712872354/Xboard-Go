import { useLocaleStore } from '@/stores/locale'
import zhCN from '@/locales/zh-CN.json'
import enUS from '@/locales/en-US.json'

const translations: Record<string, Record<string, string>> = {
  'zh-CN': zhCN as Record<string, string>,
  'en-US': enUS as Record<string, string>,
}

export function useTranslation() {
  const locale = useLocaleStore((s) => s.locale)

  const t = (key: string, params?: Record<string, string | number>): string => {
    let value = translations[locale]?.[key] || translations['zh-CN']?.[key] || key
    if (params) {
      Object.entries(params).forEach(([k, v]) => {
        value = value.replace(`{{${k}}}`, String(v))
      })
    }
    return value
  }

  return { t, locale }
}
