import { EmptyState } from '@/components/feedback/states'

/** Placeholder for routes delivered in a later phase. */
export function StubPage({ title, phase }: { title: string; phase: number }) {
  return (
    <div className="p-4">
      <h1 className="mb-4 text-lg font-semibold">{title}</h1>
      <EmptyState title="Coming soon" description={`This workflow is delivered in Phase ${phase}.`} />
    </div>
  )
}
