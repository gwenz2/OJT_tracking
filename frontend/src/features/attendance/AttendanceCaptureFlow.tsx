import { useCallback, useMemo, useRef, useState } from 'react'
import { useMutation } from '@tanstack/react-query'
import { api, ApiError } from '@/api/client'
import { queryClient } from '@/app/query-client'
import { useGeolocationCapture } from './hooks/useGeolocationCapture'
import { LiveCameraCapture } from './components/LiveCameraCapture'
import { Button } from '@/components/ui/button'
import { Textarea, Label } from '@/components/ui/input'
import { Badge } from '@/components/ui/badge'
import { useToast } from '@/components/feedback/toast'
import { MapPin, AlertTriangle, CheckCircle2 } from 'lucide-react'
import { cn } from '@/lib/utils'

type Step = 'location' | 'camera' | 'preview' | 'submitting' | 'done' | 'failed'

interface CaptureResult {
  attendance: {
    id: string
    status: string
    official_time_in_at?: string
    official_time_out_at?: string
    credited_minutes?: number
    location_status?: string
    flags: string[]
  }
  progress?: { completed_minutes: number; required_minutes: number; progress_percent: number }
  journal?: { id: string; status: string }
}

interface Props {
  action: 'time_in' | 'time_out'
  attendanceId?: string // required for time_out
  onDone: () => void
  onCancel: () => void
}

const actionLabel = { time_in: 'Time In', time_out: 'Time Out' }

/**
 * The capture flow: acquire GPS → live camera → preview → confirm upload.
 *
 * Idempotency: one client_action_id is generated per *attempt session* and
 * reused across retries of the same submission, so flaky-network retries can
 * never double-record.
 */
export function AttendanceCaptureFlow({ action, attendanceId, onDone, onCancel }: Props) {
  const geo = useGeolocationCapture()
  const toast = useToast()
  const [step, setStep] = useState<Step>('location')
  const [photo, setPhoto] = useState<{ blob: Blob; url: string } | null>(null)
  const [reason, setReason] = useState('')
  const [error, setError] = useState('')
  const [result, setResult] = useState<CaptureResult | null>(null)

  // One idempotency key per mounted flow. A network-uncertain retry reuses it;
  // starting over (cancel → re-enter) mounts a new flow and a new key.
  const clientActionId = useRef<string>(crypto.randomUUID())
  const deviceCapturedAt = useRef<string | null>(null)

  const locationUnavailable = geo.state.phase === 'denied' || geo.state.phase === 'unavailable'
  const canContinueLocation =
    geo.state.phase === 'acquired' ||
    (locationUnavailable && reason.trim().length >= 4)

  const submit = useMutation({
    mutationFn: async () => {
      if (!photo) throw new Error('No photo captured')
      const fd = new FormData()
      fd.set('client_action_id', clientActionId.current)
      fd.set('photo', photo.blob, 'capture.jpg')
      if (deviceCapturedAt.current) fd.set('device_captured_at', deviceCapturedAt.current)
      fd.set('device_timezone_offset_min', String(-new Date().getTimezoneOffset()))
      if (geo.state.phase === 'acquired') {
        fd.set('latitude', String(geo.state.latitude))
        fd.set('longitude', String(geo.state.longitude))
        fd.set('gps_accuracy_m', String(geo.state.accuracyM))
      } else {
        fd.set('location_exception_reason', reason.trim())
      }
      const path =
        action === 'time_in' ? '/attendance/time-in' : `/attendance/${attendanceId}/time-out`
      return api.post<CaptureResult>(path, fd)
    },
    onSuccess: (res) => {
      setResult(res)
      setStep('done')
      void queryClient.invalidateQueries({ queryKey: ['attendance'] })
      void queryClient.invalidateQueries({ queryKey: ['trainee', 'dashboard'] })
      toast.success(
        action === 'time_in' ? 'Timed in' : 'Timed out',
        'Recorded at ' +
          new Date(
            (action === 'time_in'
              ? res.attendance.official_time_in_at
              : res.attendance.official_time_out_at) ?? Date.now(),
          ).toLocaleTimeString(),
      )
    },
    onError: (e) => {
      setStep('failed')
      setError(e instanceof ApiError ? e.message : 'Upload failed — your photo is kept, retry safely.')
    },
  })

  const onCapture = useCallback((blob: Blob, url: string) => {
    deviceCapturedAt.current = new Date().toISOString()
    setPhoto({ blob, url })
    setStep('preview')
  }, [])

  const locationSummary = useMemo(() => {
    switch (geo.state.phase) {
      case 'acquiring':
        return { icon: MapPin, text: 'Getting your location…', tone: 'muted' }
      case 'acquired':
        return {
          icon: CheckCircle2,
          text: `Location captured (±${Math.round(geo.state.accuracyM)} m)`,
          tone: 'ok',
        }
      case 'denied':
        return { icon: AlertTriangle, text: 'Location permission denied', tone: 'warn' }
      case 'unavailable':
        return { icon: AlertTriangle, text: 'Location unavailable — explain below', tone: 'warn' }
      default:
        return { icon: MapPin, text: 'Location not captured yet', tone: 'muted' }
    }
  }, [geo.state])

  return (
    <div className="space-y-4">
      {/* Step 1: location */}
      <section aria-label="Location" className={cn('space-y-2', step !== 'location' && 'opacity-60')}>
        <div className="flex items-center gap-2 text-sm">
          <locationSummary.icon
            size={16}
            className={cn(
              locationSummary.tone === 'ok' && 'text-[var(--color-success)]',
              locationSummary.tone === 'warn' && 'text-[var(--color-warning)]',
              locationSummary.tone === 'muted' && 'text-[var(--color-text-muted)]',
            )}
          />
          <span>{locationSummary.text}</span>
        </div>

        {step === 'location' && (
          <>
            {geo.state.phase === 'idle' && (
              <Button variant="secondary" fullWidth onClick={geo.acquire}>
                <MapPin size={16} /> Get location
              </Button>
            )}
            {geo.state.phase === 'acquiring' && (
              <Button variant="secondary" fullWidth disabled>
                Locating…
              </Button>
            )}
            {locationUnavailable && (
              <div>
                <Label htmlFor="loc-reason">Why is location unavailable?</Label>
                <Textarea
                  id="loc-reason"
                  value={reason}
                  onChange={(e) => setReason(e.target.value)}
                  placeholder="e.g. GPS denied, indoors with no signal…"
                  required
                />
                <p className="mt-1 text-xs text-[var(--color-text-subtle)]">
                  Your attendance is still recorded — it will be flagged for your coordinator.
                </p>
              </div>
            )}
            {(geo.state.phase === 'acquired' || (locationUnavailable && canContinueLocation)) && (
              <Button fullWidth onClick={() => setStep('camera')}>
                Continue to camera
              </Button>
            )}
            {locationUnavailable && !canContinueLocation && (
              <Button variant="ghost" onClick={geo.acquire}>
                Retry location
              </Button>
            )}
          </>
        )}
      </section>

      {/* Step 2: live camera — only mounted after location step, so the camera
          light turns on exactly when the user reaches for it. */}
      {step === 'camera' && (
        <section aria-label="Camera">
          <LiveCameraCapture onCapture={onCapture} />
          <Button variant="ghost" className="mt-2" onClick={() => setStep('location')}>
            Back
          </Button>
        </section>
      )}

      {/* Step 3: preview + confirm */}
      {(step === 'preview' || step === 'submitting' || step === 'failed') && photo && (
        <section aria-label="Preview" className="space-y-3">
          <img
            src={photo.url}
            alt="Captured evidence preview"
            className="max-h-[38dvh] w-full rounded-[var(--radius-lg)] object-cover"
          />
          <div className="flex items-center gap-2 text-xs text-[var(--color-text-muted)]">
            {geo.state.phase === 'acquired' ? (
              <Badge variant="success">GPS ±{Math.round(geo.state.accuracyM)} m</Badge>
            ) : (
              <Badge variant="warning">No GPS — flagged</Badge>
            )}
            <span>The server stamps the official time and site name on the photo.</span>
          </div>
          {step === 'failed' && (
            <p role="alert" className="text-sm text-[var(--color-danger)]">
              {error}
            </p>
          )}
          <div className="flex gap-2">
            <Button
              variant="outline"
              fullWidth
              onClick={() => {
                URL.revokeObjectURL(photo.url)
                setPhoto(null)
                setStep('camera')
              }}
              disabled={submit.isPending}
            >
              Retake
            </Button>
            <Button fullWidth size="lg" loading={submit.isPending} onClick={() => submit.mutate()}>
              {submit.isPending ? 'Uploading…' : `Confirm ${actionLabel[action]}`}
            </Button>
          </div>
          {step === 'failed' && (
            <p className="text-xs text-[var(--color-text-subtle)]">
              Retrying is safe — this attempt reuses the same action ID and cannot double-record.
            </p>
          )}
        </section>
      )}

      {/* Done */}
      {step === 'done' && result && (
        <section aria-label="Result" className="space-y-3 text-center">
          <CheckCircle2 size={40} className="mx-auto text-[var(--color-success)]" />
          <p className="text-sm font-medium">
            {action === 'time_in' ? 'Timed in' : 'Timed out'} at{' '}
            {new Date(
              (action === 'time_in'
                ? result.attendance.official_time_in_at
                : result.attendance.official_time_out_at) ?? '',
            ).toLocaleTimeString()}
          </p>
          {result.attendance.flags.length > 0 && (
            <p className="text-xs text-[var(--color-warning)]">
              Flagged for review: {result.attendance.flags.join(', ')}
            </p>
          )}
          {result.attendance.credited_minutes != null && (
            <p className="text-sm text-[var(--color-text-muted)]">
              Credited: {Math.floor(result.attendance.credited_minutes / 60)}h{' '}
              {result.attendance.credited_minutes % 60}m
            </p>
          )}
          {result.journal && (
            <p className="text-xs text-[var(--color-text-muted)]">
              Your daily journal is ready — complete it while the day is fresh.
            </p>
          )}
          <Button fullWidth size="lg" onClick={onDone}>
            Done
          </Button>
        </section>
      )}

      {step !== 'done' && (
        <Button variant="ghost" onClick={onCancel}>
          Cancel
        </Button>
      )}
    </div>
  )
}
