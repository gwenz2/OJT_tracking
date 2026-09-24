/**
 * Thin, typed wrapper around localStorage. Falls back to no-op when
 * storage is unavailable (e.g. SSR, privacy mode).
 */
const memoryStore = new Map<string, string>()

function getStorage(): Storage | null {
  try {
    if (typeof window === 'undefined') return null
    const test = '__storage_test__'
    window.localStorage.setItem(test, '1')
    window.localStorage.removeItem(test)
    return window.localStorage
  } catch {
    return null
  }
}

export const storage = {
  get(key: string): string | null {
    const s = getStorage()
    if (s) return s.getItem(key)
    return memoryStore.get(key) ?? null
  },

  set(key: string, value: string): void {
    const s = getStorage()
    if (s) s.setItem(key, value)
    else memoryStore.set(key, value)
  },

  remove(key: string): void {
    const s = getStorage()
    if (s) s.removeItem(key)
    else memoryStore.delete(key)
  },

  /** Get and JSON-parse a value. Returns fallback on parse failure. */
  getJSON<T>(key: string, fallback: T): T {
    const raw = storage.get(key)
    if (!raw) return fallback
    try {
      return JSON.parse(raw) as T
    } catch {
      return fallback
    }
  },

  /** JSON-stringify and set a value. */
  setJSON(key: string, value: unknown): void {
    storage.set(key, JSON.stringify(value))
  },
}

/** Keys used by the app. Centralized to avoid typos. */
export const storageKeys = {
  accessToken: 'fst.access_token',
  refreshToken: 'fst.refresh_token',
  theme: 'fst.theme',
} as const
