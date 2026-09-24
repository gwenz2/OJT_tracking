import { useCallback, useRef, useState } from 'react'

export type GeoState =
  | { phase: 'idle' }
  | { phase: 'acquiring' }
  | {
      phase: 'acquired'
      latitude: number
      longitude: number
      accuracyM: number
    }
  | { phase: 'denied' }
  | { phase: 'unavailable'; reason: 'timeout' | 'position_unavailable' | 'unsupported' }

const TIMEOUT_MS = 15_000
const MAX_AGE_MS = 10_000

/**
 * One-shot geolocation capture with explicit states. The result is evidence
 * the server evaluates — this hook never decides validity, it only reports
 * what the device could produce.
 */
export function useGeolocationCapture() {
  const [state, setState] = useState<GeoState>({ phase: 'idle' })
  const watchId = useRef<number | null>(null)

  const cancel = useCallback(() => {
    if (watchId.current != null && 'geolocation' in navigator) {
      navigator.geolocation.clearWatch(watchId.current)
      watchId.current = null
    }
  }, [])

  const acquire = useCallback(() => {
    cancel()
    if (!('geolocation' in navigator)) {
      setState({ phase: 'unavailable', reason: 'unsupported' })
      return
    }
    setState({ phase: 'acquiring' })

    let done = false
    const finish = (s: GeoState) => {
      if (done) return
      done = true
      cancel()
      setState(s)
    }

    const timer = setTimeout(() => finish({ phase: 'unavailable', reason: 'timeout' }), TIMEOUT_MS)

    watchId.current = navigator.geolocation.watchPosition(
      (pos) => {
        clearTimeout(timer)
        finish({
          phase: 'acquired',
          latitude: pos.coords.latitude,
          longitude: pos.coords.longitude,
          accuracyM: pos.coords.accuracy,
        })
      },
      (err) => {
        clearTimeout(timer)
        if (err.code === err.PERMISSION_DENIED) {
          finish({ phase: 'denied' })
        } else {
          finish({ phase: 'unavailable', reason: 'position_unavailable' })
        }
      },
      { enableHighAccuracy: true, timeout: TIMEOUT_MS, maximumAge: MAX_AGE_MS },
    )
  }, [cancel])

  return { state, acquire, cancel }
}
