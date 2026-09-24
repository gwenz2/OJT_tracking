import { useCallback, useEffect, useRef, useState } from 'react'
import { Camera, RefreshCw, SwitchCamera } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { cn } from '@/lib/utils'

type CameraState =
  | { phase: 'idle' }
  | { phase: 'starting' }
  | { phase: 'live' }
  | { phase: 'denied' }
  | { phase: 'unavailable' }
  | { phase: 'insecure' }

interface Props {
  /** Called with the captured JPEG blob when the user taps capture. */
  onCapture: (blob: Blob, previewUrl: string) => void
  disabled?: boolean
}

/**
 * Live camera capture via getUserMedia. Renders <video playsInline> for iOS,
 * captures a frame to canvas → JPEG Blob. Tracks are always stopped on
 * unmount or phase change so the camera LED never stays on.
 */
export function LiveCameraCapture({ onCapture, disabled }: Props) {
  const videoRef = useRef<HTMLVideoElement>(null)
  const streamRef = useRef<MediaStream | null>(null)
  const [state, setState] = useState<CameraState>({ phase: 'idle' })
  const [facing, setFacing] = useState<'user' | 'environment'>('environment')

  const stop = useCallback(() => {
    streamRef.current?.getTracks().forEach((t) => t.stop())
    streamRef.current = null
  }, [])

  const start = useCallback(async () => {
    stop()
    if (!navigator.mediaDevices?.getUserMedia) {
      // mediaDevices is only exposed in secure contexts (HTTPS or localhost).
      setState({ phase: window.isSecureContext ? 'unavailable' : 'insecure' })
      return
    }
    setState({ phase: 'starting' })
    try {
      const stream = await navigator.mediaDevices.getUserMedia({
        video: { facingMode: facing, width: { ideal: 1280 }, height: { ideal: 720 } },
        audio: false,
      })
      streamRef.current = stream
      if (videoRef.current) {
        videoRef.current.srcObject = stream
        await videoRef.current.play().catch(() => undefined)
      }
      setState({ phase: 'live' })
    } catch (err) {
      const name = (err as DOMException).name
      setState(
        name === 'NotAllowedError' || name === 'SecurityError'
          ? { phase: 'denied' }
          : { phase: 'unavailable' },
      )
    }
  }, [facing, stop])

  useEffect(() => {
    void start()
    return stop
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [facing])

  const capture = useCallback(() => {
    const video = videoRef.current
    if (!video || state.phase !== 'live') return
    const canvas = document.createElement('canvas')
    canvas.width = video.videoWidth || 1280
    canvas.height = video.videoHeight || 720
    const ctx = canvas.getContext('2d')
    if (!ctx) return
    ctx.drawImage(video, 0, 0)
    canvas.toBlob(
      (blob) => {
        if (blob) onCapture(blob, URL.createObjectURL(blob))
      },
      'image/jpeg',
      0.9,
    )
  }, [onCapture, state.phase])

  return (
    <div className="relative overflow-hidden rounded-[var(--radius-lg)] bg-black">
      <video
        ref={videoRef}
        playsInline
        muted
        autoPlay
        className={cn(
          'aspect-[3/4] max-h-[52dvh] w-full object-cover',
          state.phase !== 'live' && 'invisible',
        )}
        aria-label="Live camera preview"
      />

      {state.phase === 'starting' && (
        <div className="absolute inset-0 flex items-center justify-center text-sm text-white/80">
          Starting camera…
        </div>
      )}

      {state.phase === 'denied' && (
        <div className="absolute inset-0 flex flex-col items-center justify-center gap-3 p-6 text-center">
          <Camera size={32} className="text-white/60" />
          <p className="text-sm text-white/90">
            Camera access was denied. Enable it in your browser settings and retry.
          </p>
          <Button variant="secondary" size="sm" onClick={() => void start()}>
            <RefreshCw size={14} /> Retry
          </Button>
        </div>
      )}

      {state.phase === 'unavailable' && (
        <div className="absolute inset-0 flex flex-col items-center justify-center gap-3 p-6 text-center">
          <Camera size={32} className="text-white/60" />
          <p className="text-sm text-white/90">No camera is available on this device.</p>
          <Button variant="secondary" size="sm" onClick={() => void start()}>
            <RefreshCw size={14} /> Retry
          </Button>
        </div>
      )}

      {state.phase === 'insecure' && (
        <div className="absolute inset-0 flex flex-col items-center justify-center gap-3 p-6 text-center">
          <Camera size={32} className="text-white/60" />
          <p className="text-sm text-white/90">
            The camera needs a secure connection. Open this app over HTTPS, or via
            localhost on the device itself.
          </p>
        </div>
      )}

      {state.phase === 'live' && (
        <div className="absolute inset-x-0 bottom-0 flex items-center justify-center gap-6 p-4">
          <button
            onClick={() => setFacing((f) => (f === 'user' ? 'environment' : 'user'))}
            className="flex h-11 w-11 items-center justify-center rounded-full bg-black/50 text-white focus-ring"
            aria-label="Switch camera"
            type="button"
          >
            <SwitchCamera size={20} />
          </button>
          <button
            onClick={capture}
            disabled={disabled}
            className="h-16 w-16 rounded-full border-4 border-white bg-white/20 focus-ring disabled:opacity-50"
            aria-label="Capture photo"
            type="button"
          >
            <span className="sr-only">Capture</span>
          </button>
          <div className="w-11" />
        </div>
      )}
    </div>
  )
}
