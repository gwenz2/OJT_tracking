import { forwardRef, type ButtonHTMLAttributes } from 'react'
import { cn } from '@/lib/utils'
import { Spinner } from './spinner'

type Variant = 'primary' | 'secondary' | 'outline' | 'ghost' | 'danger'
type Size = 'sm' | 'md' | 'lg'

export interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: Variant
  size?: Size
  loading?: boolean
  fullWidth?: boolean
}

const variantClasses: Record<Variant, string> = {
  primary:
    'bg-[var(--color-primary)] text-[var(--color-primary-foreground)] hover:bg-[var(--color-primary-hover)]',
  secondary:
    'bg-[var(--color-surface-hover)] text-[var(--color-text)] hover:bg-[var(--color-border)]',
  outline:
    'border border-[var(--color-border-strong)] text-[var(--color-text)] hover:bg-[var(--color-surface-hover)]',
  ghost: 'text-[var(--color-text)] hover:bg-[var(--color-surface-hover)]',
  danger:
    'bg-[var(--color-danger)] text-white hover:bg-[var(--color-danger-hover)]',
}

const sizeClasses: Record<Size, string> = {
  sm: 'h-9 px-3 text-sm',
  md: 'h-11 px-4 text-sm',
  lg: 'h-12 px-6 text-base',
}

export const Button = forwardRef<HTMLButtonElement, ButtonProps>(
  (
    { className, variant = 'primary', size = 'md', loading, fullWidth, disabled, children, ...props },
    ref,
  ) => {
    return (
      <button
        ref={ref}
        disabled={disabled || loading}
        className={cn(
          'inline-flex items-center justify-center gap-2 rounded-[var(--radius-md)] font-medium transition-colors focus-ring disabled:cursor-not-allowed disabled:opacity-50',
          variantClasses[variant],
          sizeClasses[size],
          fullWidth && 'w-full',
          className,
        )}
        {...props}
      >
        {loading && <Spinner size={size === 'lg' ? 18 : 16} />}
        {children}
      </button>
    )
  },
)
Button.displayName = 'Button'
