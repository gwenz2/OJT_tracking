import { Link } from 'react-router-dom'
import { Button } from '@/components/ui/button'

export function NotFoundPage() {
  return (
    <div className="flex min-h-screen flex-col items-center justify-center gap-4 p-6 text-center">
      <p className="text-6xl font-bold text-[var(--color-text)]">404</p>
      <p className="text-lg text-[var(--color-text-muted)]">The page you're looking for doesn't exist.</p>
      <Link to="/">
        <Button>Back home</Button>
      </Link>
    </div>
  )
}
