import { forwardRef, type InputHTMLAttributes } from 'react'
import { cn } from '@/lib/utils'

export interface InputProps extends InputHTMLAttributes<HTMLInputElement> {
  invalid?: boolean
}

export const Input = forwardRef<HTMLInputElement, InputProps>(
  ({ className, invalid, ...props }, ref) => {
    return (
      <input
        ref={ref}
        className={cn(
          'h-11 w-full rounded-[var(--radius-md)] border border-[var(--color-border)] bg-[var(--color-surface)] px-3 text-base text-[var(--color-text)] placeholder:text-[var(--color-text-subtle)] transition-colors focus-ring disabled:cursor-not-allowed disabled:opacity-50',
          invalid && 'border-[var(--color-danger)]',
          className,
        )}
        {...props}
      />
    )
  },
)
Input.displayName = 'Input'

export interface TextareaProps extends React.TextareaHTMLAttributes<HTMLTextAreaElement> {
  invalid?: boolean
}

export const Textarea = forwardRef<HTMLTextAreaElement, TextareaProps>(
  ({ className, invalid, ...props }, ref) => {
    return (
      <textarea
        ref={ref}
        className={cn(
          'min-h-[80px] w-full rounded-[var(--radius-md)] border border-[var(--color-border)] bg-[var(--color-surface)] px-3 py-2 text-base text-[var(--color-text)] placeholder:text-[var(--color-text-subtle)] transition-colors focus-ring disabled:cursor-not-allowed disabled:opacity-50',
          invalid && 'border-[var(--color-danger)]',
          className,
        )}
        {...props}
      />
    )
  },
)
Textarea.displayName = 'Textarea'

export function Label({ children, htmlFor, className }: { children: React.ReactNode; htmlFor?: string; className?: string }) {
  return (
    <label
      htmlFor={htmlFor}
      className={cn('mb-1.5 block text-sm font-medium text-[var(--color-text)]', className)}
    >
      {children}
    </label>
  )
}

export function FieldError({ children }: { children?: React.ReactNode }) {
  if (!children) return null
  return <p className="mt-1 text-xs text-[var(--color-danger)]">{children}</p>
}
