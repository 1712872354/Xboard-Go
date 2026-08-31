import { create } from 'zustand'
import { persist } from 'zustand/middleware'

type Locale = 'zh-CN' | 'en-US' | 'ja-JP'

interface LocaleState {
  locale: Locale
  setLocale: (locale: Locale) => void
}

export const useLocaleStore = create<LocaleState>()(
  persist(
    (set) => ({
      locale: 'zh-CN',

      setLocale: (locale) => {
        document.documentElement.lang = locale
        set({ locale })
      },
    }),
    { name: 'locale-storage' },
  ),
)

export function getBrowserLocale(): Locale {
  const lang = navigator.language
  if (lang.startsWith('zh')) return 'zh-CN'
  if (lang.startsWith('ja')) return 'ja-JP'
  return 'en-US'
}
