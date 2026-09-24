import { create } from 'zustand'

type Theme = 'light' | 'dark'

interface UIState {
  sidebarOpen: boolean
  theme: Theme
  toggleSidebar: () => void
  setSidebar: (open: boolean) => void
  toggleTheme: () => void
  setTheme: (theme: Theme) => void
}

/**
 * UI store — global client/UI state only.
 *
 * Sidebar open state, theme, and other transient UI preferences live here.
 * Never duplicate server data into this store.
 */
export const useUIStore = create<UIState>((set, get) => ({
  sidebarOpen: true,
  theme: 'light',
  toggleSidebar: () => set((s) => ({ sidebarOpen: !s.sidebarOpen })),
  setSidebar: (open) => set({ sidebarOpen: open }),
  toggleTheme: () => {
    const next = get().theme === 'light' ? 'dark' : 'light'
    set({ theme: next })
    applyTheme(next)
  },
  setTheme: (theme) => {
    set({ theme })
    applyTheme(theme)
  },
}))

function applyTheme(theme: Theme): void {
  if (typeof document === 'undefined') return
  document.documentElement.classList.toggle('dark', theme === 'dark')
}
