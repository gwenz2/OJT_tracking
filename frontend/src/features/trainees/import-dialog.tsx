import { useState, useRef } from 'react'
import { api, ApiError } from '@/api/client'
import { Button } from '@/components/ui/button'
import { Modal } from '@/components/ui/modal'
import { Badge } from '@/components/ui/badge'

interface ImportRowResult {
  row: number
  status: 'valid' | 'invalid'
  normalized?: { email: string; display_name: string; student_number: string }
  errors?: Record<string, string>
}

interface ImportPreview {
  valid_rows: number
  invalid_rows: number
  rows: ImportRowResult[]
  import_token: string
}

interface CommitResult {
  created: number
  credentials: { email: string; temporary_password: string }[]
}

/**
 * Two-step CSV import: upload for server-side validation (returns a staged
 * import_token), review row errors, then commit atomically.
 */
export function ImportDialog({ onClose, onDone }: { onClose: () => void; onDone: (created: number) => void }) {
  const fileRef = useRef<HTMLInputElement>(null)
  const [preview, setPreview] = useState<ImportPreview | null>(null)
  const [result, setResult] = useState<CommitResult | null>(null)
  const [error, setError] = useState('')
  const [busy, setBusy] = useState(false)

  async function validate() {
    const file = fileRef.current?.files?.[0]
    if (!file) {
      setError('Choose a .csv file first.')
      return
    }
    setError('')
    setBusy(true)
    try {
      const fd = new FormData()
      fd.append('file', file)
      setPreview(await api.post<ImportPreview>('/staff/trainees/import/validate', fd))
    } catch (e) {
      setError(e instanceof ApiError ? e.message : 'Upload failed')
    } finally {
      setBusy(false)
    }
  }

  async function commit() {
    if (!preview) return
    setBusy(true)
    try {
      const res = await api.post<CommitResult>('/staff/trainees/import/commit', {
        import_token: preview.import_token,
      })
      setResult(res)
    } catch (e) {
      setError(e instanceof ApiError ? e.message : 'Import failed')
    } finally {
      setBusy(false)
    }
  }

  if (result) {
    return (
      <Modal open onClose={() => onDone(result.created)} title="Import complete">
        <p className="text-sm text-[var(--color-text-muted)]">
          {result.created} account{result.created === 1 ? '' : 's'} created. Share each temporary
          password with the trainee — they are shown only once.
        </p>
        <div className="mt-3 max-h-64 space-y-1 overflow-y-auto rounded-[var(--radius-md)] border border-[var(--color-border)] p-3 text-xs">
          {result.credentials.map((cred) => (
            <div key={cred.email} className="flex justify-between gap-4">
              <span className="truncate">{cred.email}</span>
              <code className="font-mono">{cred.temporary_password}</code>
            </div>
          ))}
        </div>
        <div className="mt-4 flex justify-end">
          <Button onClick={() => onDone(result.created)}>Done</Button>
        </div>
      </Modal>
    )
  }

  return (
    <Modal open onClose={onClose} title="Import trainees (CSV)">
      {!preview ? (
        <div className="space-y-3">
          <p className="text-sm text-[var(--color-text-muted)]">
            CSV columns: <code>email, student_number, display_name, program, year_level, contact_number</code>.
            An optional <code>password</code> column sets a specific temporary password.
          </p>
          <input ref={fileRef} type="file" accept=".csv" className="text-sm" aria-label="CSV file" />
          {error && <p role="alert" className="text-sm text-[var(--color-danger)]">{error}</p>}
          <div className="flex justify-end gap-2">
            <Button variant="ghost" onClick={onClose}>Cancel</Button>
            <Button onClick={validate} loading={busy}>Validate</Button>
          </div>
        </div>
      ) : (
        <div className="space-y-3">
          <p className="text-sm">
            <Badge variant="success">{preview.valid_rows} valid</Badge>{' '}
            {preview.invalid_rows > 0 && <Badge variant="danger">{preview.invalid_rows} invalid</Badge>}
          </p>
          <div className="max-h-64 overflow-y-auto rounded-[var(--radius-md)] border border-[var(--color-border)]">
            <table className="w-full text-xs">
              <thead>
                <tr className="border-b border-[var(--color-border)] text-left text-[var(--color-text-muted)]">
                  <th className="px-2 py-2">Row</th>
                  <th className="px-2 py-2">Email</th>
                  <th className="px-2 py-2">Result</th>
                </tr>
              </thead>
              <tbody>
                {preview.rows.map((r) => (
                  <tr key={r.row} className="border-b border-[var(--color-border)] last:border-0">
                    <td className="px-2 py-2">{r.row}</td>
                    <td className="px-2 py-2">{r.normalized?.email ?? '—'}</td>
                    <td className="px-2 py-2">
                      {r.status === 'valid' ? (
                        <Badge variant="success">ok</Badge>
                      ) : (
                        <span className="text-[var(--color-danger)]">
                          {Object.entries(r.errors ?? {}).map(([k, v]) => `${k}: ${v}`).join('; ')}
                        </span>
                      )}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
          {preview.invalid_rows > 0 && (
            <p className="text-xs text-[var(--color-text-muted)]">
              Rows with errors are skipped; only valid rows are committed.
            </p>
          )}
          {error && <p role="alert" className="text-sm text-[var(--color-danger)]">{error}</p>}
          <div className="flex justify-end gap-2">
            <Button variant="ghost" onClick={onClose}>Cancel</Button>
            <Button onClick={commit} loading={busy} disabled={preview.valid_rows === 0}>
              Import {preview.valid_rows} row{preview.valid_rows === 1 ? '' : 's'}
            </Button>
          </div>
        </div>
      )}
    </Modal>
  )
}
