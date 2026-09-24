import { describe, it, expect } from 'vitest'
import { cn, initials, formatDate } from '@/lib/utils'

describe('cn', () => {
  it('merges class names', () => {
    expect(cn('px-2', 'py-1')).toBe('px-2 py-1')
  })

  it('dedupes conflicting tailwind classes', () => {
    expect(cn('px-2', 'px-4')).toBe('px-4')
  })
})

describe('initials', () => {
  it('returns initials for a full name', () => {
    expect(initials('Alice Wonderland')).toBe('AW')
  })

  it('returns first two chars for a single name', () => {
    expect(initials('Alice')).toBe('AL')
  })

  it('returns ? for empty', () => {
    expect(initials('')).toBe('?')
  })
})

describe('formatDate', () => {
  it('formats an ISO date', () => {
    const out = formatDate('2024-01-15T10:00:00Z')
    expect(out).not.toBe('—')
    expect(out.length).toBeGreaterThan(0)
  })

  it('returns — for null', () => {
    expect(formatDate(null)).toBe('—')
  })
})
