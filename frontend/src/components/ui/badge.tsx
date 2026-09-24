import type { HTMLAttributes } from 'react'
import { cn } from '@/lib/utils'
import { initials } from '@/lib/utils'

type BadgeVariant = 'default' | 'success' | 'warning' | 'danger'

const badgeVariants: Record<BadgeVariant, string> = {
  default: 'bg-[var(--color-surface-hover)] text-[var(--color-text-muted)]',
  success: 'bg-[var(--color-success)]/10 text-[var(--color-success)]',
  warning: 'bg-[var(--color-warning)]/10 text-[var(--color-warning)]',
  danger: 'bg-[var(--color-danger)]/10 text-[var(--color-danger)]',
}

export function Badge({
  variant = 'default',
  className,
  ...props
}: HTMLAttributes<HTMLSpanElement> & { variant?: BadgeVariant }) {
  return (
    <span
      className={cn(
        'inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium',
        badgeVariants[variant],
        className,
      )}
      {...props}
    />
  )
}

export function Avatar({ name, src, size = 40 }: { name: string; src?: string; size?: number }) {
  if (src) {
    return (
      <img
        src={src}
        alt={name}
        width={size}
        height={size}
        className="rounded-full object-cover"
        style={{ width: size, height: size }}
      />
    )
  }
  return (
    <div
      className="flex items-center justify-center rounded-full bg-[var(--color-primary)] text-[var(--color-primary-foreground)] font-medium"
      style={{ width: size, height: size, fontSize: size * 0.4 }}
    >
      {initials(name)}
    </div>
  )
}
