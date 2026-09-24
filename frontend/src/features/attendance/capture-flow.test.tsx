import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { QueryClientProvider, QueryClient } from '@tanstack/react-query'
import { AttendanceCaptureFlow } from './AttendanceCaptureFlow'
import { ToastProvider } from '@/components/feedback/toast'

// ---- Browser API mocks -----------------------------------------------------

function mockGeoSuccess(lat = 14.5995, lon = 120.9842, acc = 12) {
  const watch = vi.fn((ok: PositionCallback) => {
    setTimeout(() => ok({ coords: { latitude: lat, longitude: lon, accuracy: acc } } as GeolocationPosition), 0)
    return 1
  })
  const clear = vi.fn()
  vi.stubGlobal('navigator', {
    ...navigator,
    geolocation: { watchPosition: watch, clearWatch: clear, getCurrentPosition: vi.fn() },
    mediaDevices: { getUserMedia: vi.fn() },
  })
  return { watch, clear }
}

function mockGeoDenied() {
  vi.stubGlobal('navigator', {
    ...navigator,
    geolocation: {
      watchPosition: vi.fn((_ok: PositionCallback, err: PositionErrorCallback) => {
        const e = { code: 1, message: 'denied', PERMISSION_DENIED: 1, POSITION_UNAVAILABLE: 2, TIMEOUT: 3 }
        setTimeout(() => err(e as GeolocationPositionError), 0)
        return 2
      }),
      clearWatch: vi.fn(),
      getCurrentPosition: vi.fn(),
    },
    mediaDevices: { getUserMedia: vi.fn() },
  })
}

function mockCameraOK() {
  const track = { stop: vi.fn() }
  const stream = { getTracks: () => [track] } as unknown as MediaStream
  navigator.mediaDevices.getUserMedia = vi.fn(async () => stream)
  return track
}

// jsdom has no canvas/toBlob — stub the bits LiveCameraCapture needs.
function mockCanvasCapture() {
  const video = window.HTMLVideoElement.prototype
  Object.defineProperty(video, 'videoWidth', { get: () => 640, configurable: true })
  Object.defineProperty(video, 'videoHeight', { get: () => 480, configurable: true })
  video.play = vi.fn(async () => {})
  window.HTMLCanvasElement.prototype.getContext = vi.fn(() => ({
    drawImage: vi.fn(),
  })) as never
  window.HTMLCanvasElement.prototype.toBlob = vi.fn(function (this: HTMLCanvasElement, cb: BlobCallback) {
    cb(new Blob(['jpeg'], { type: 'image/jpeg' }))
  })
}

let fetchCalls: { url: string; body?: FormData }[]
function mockApi(handler: (fd: FormData) => { status: number; body: unknown }) {
  fetchCalls = []
  vi.stubGlobal('fetch', async (input: RequestInfo | URL, init?: RequestInit) => {
    const url = typeof input === 'string' ? input : input instanceof URL ? input.href : input.url
    const path = new URL(url, 'http://localhost').pathname
    const fd = init?.body instanceof FormData ? init.body : undefined
    fetchCalls.push({ url: path, body: fd })
    if (fd) {
      const r = handler(fd)
      return new Response(JSON.stringify(r.body), {
        status: r.status,
        headers: { 'Content-Type': 'application/json' },
      })
    }
    return new Response('{"error":{"code":"X","message":"unmocked"}}', { status: 404 })
  })
}

const qc = () =>
  new QueryClient({ defaultOptions: { queries: { retry: false }, mutations: { retry: 0 } } })

function renderFlow(props?: Partial<Parameters<typeof AttendanceCaptureFlow>[0]>) {
  const onDone = vi.fn()
  render(
    <QueryClientProvider client={qc()}>
      <ToastProvider>
        <AttendanceCaptureFlow action="time_in" onDone={onDone} onCancel={vi.fn()} {...props} />
      </ToastProvider>
    </QueryClientProvider>,
  )
  return { onDone }
}

const okTimeIn = {
  status: 201,
  body: {
    data: {
      attendance: {
        id: 'a1', date: '2026-09-25', status: 'open',
        official_time_in_at: '2026-09-25T00:05:00Z', location_status: 'verified', flags: [],
      },
    },
  },
}

beforeEach(() => {
  mockCanvasCapture()
})

afterEach(() => {
  vi.unstubAllGlobals()
})

describe('AttendanceCaptureFlow', () => {
  it('walks location → camera → preview → submit → done', async () => {
    mockGeoSuccess()
    mockCameraOK()
    mockApi(() => okTimeIn)
    const { onDone } = renderFlow()

    // location
    await userEvent.click(screen.getByRole('button', { name: /get location/i }))
    await screen.findByText(/location captured/i)
    await userEvent.click(screen.getByRole('button', { name: /continue to camera/i }))

    // camera — capture button appears when 'live'
    const captureBtn = await screen.findByRole('button', { name: /capture photo/i })
    await userEvent.click(captureBtn)

    // preview + confirm
    await userEvent.click(await screen.findByRole('button', { name: /confirm time in/i }))

    await screen.findByText(/timed in at/i)
    expect(fetchCalls).toHaveLength(1)
    expect(fetchCalls[0]!.url).toBe('/api/v1/attendance/time-in')
    const fd = fetchCalls[0]!.body!
    expect(fd.get('client_action_id')).toBeTruthy()
    expect(fd.get('latitude')).toBe('14.5995')
    expect(fd.get('photo')).toBeTruthy()

    await userEvent.click(screen.getByRole('button', { name: 'Done' }))
    expect(onDone).toHaveBeenCalled()
  })

  it('requires a reason when location is unavailable', async () => {
    mockGeoDenied()
    mockApi(() => okTimeIn)
    renderFlow()

    await userEvent.click(screen.getByRole('button', { name: /get location/i }))
    await screen.findByText(/location permission denied/i)

    // Cannot continue without a reason.
    expect(screen.queryByRole('button', { name: /continue to camera/i })).not.toBeInTheDocument()

    await userEvent.type(screen.getByLabelText(/why is location unavailable/i), 'GPS denied')
    await userEvent.click(screen.getByRole('button', { name: /continue to camera/i }))

    // With no camera mock it still mounts; capture is not possible — but the
    // flow reached the camera step, which is what we're asserting.
    await waitFor(() =>
      expect(screen.getByRole('button', { name: 'Back' })).toBeInTheDocument(),
    )
  })

  it('reuses the same client_action_id on retry after failure', async () => {
    mockGeoSuccess()
    mockCameraOK()
    let attempt = 0
    mockApi(() => {
      attempt++
      return attempt === 1
        ? { status: 500, body: { error: { code: 'STORAGE_UNAVAILABLE', message: 'storage down' } } }
        : okTimeIn
    })
    renderFlow()

    await userEvent.click(screen.getByRole('button', { name: /get location/i }))
    await screen.findByText(/location captured/i)
    await userEvent.click(screen.getByRole('button', { name: /continue to camera/i }))
    await userEvent.click(await screen.findByRole('button', { name: /capture photo/i }))

    await userEvent.click(await screen.findByRole('button', { name: /confirm time in/i }))
    await screen.findByText(/storage down/i)

    // Retry — same action id must be resent.
    await userEvent.click(screen.getByRole('button', { name: /confirm time in/i }))
    await screen.findByText(/timed in at/i)

    expect(fetchCalls).toHaveLength(2)
    expect(fetchCalls[0]!.body!.get('client_action_id')).toBe(fetchCalls[1]!.body!.get('client_action_id'))
  })

  it('disables the submit button while uploading (no double submit)', async () => {
    mockGeoSuccess()
    mockCameraOK()
    let resolve!: (v: Response) => void
    vi.stubGlobal('fetch', async () => new Promise<Response>((r) => (resolve = r)))
    renderFlow()

    await userEvent.click(screen.getByRole('button', { name: /get location/i }))
    await screen.findByText(/location captured/i)
    await userEvent.click(screen.getByRole('button', { name: /continue to camera/i }))
    await userEvent.click(await screen.findByRole('button', { name: /capture photo/i }))

    const btn = await screen.findByRole('button', { name: /confirm time in/i })
    await userEvent.click(btn)
    expect(screen.getByRole('button', { name: /uploading/i })).toBeDisabled()

    resolve(new Response(JSON.stringify(okTimeIn.body), { status: 201 }))
    await screen.findByText(/timed in at/i)
  })
})
