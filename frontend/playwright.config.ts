import { defineConfig, devices } from '@playwright/test'

/**
 * E2E critical paths (T085/T100/T139). Requires:
 *   - backend on :8080 with seeded trainee1@ojt.local / Trainee123!
 *   - `npm run dev` serving the app on :5173 (proxies /api to :8080)
 * Camera is Chromium's fake media stream; geolocation is stubbed inside the
 * seeded site's radius.
 */
export default defineConfig({
  testDir: './e2e',
  timeout: 60_000,
  retries: 0,
  workers: 1, // attendance dates are unique per trainee per day — serialize
  use: {
    baseURL: 'http://localhost:5173',
    geolocation: { latitude: 6.629, longitude: 124.605, accuracy: 10 },
    permissions: ['geolocation', 'camera'],
    launchOptions: {
      args: [
        '--use-fake-device-for-media-stream',
        '--use-fake-ui-for-media-stream',
      ],
    },
    trace: 'retain-on-failure',
  },
  projects: [
    { name: 'chromium', use: { ...devices['Desktop Chrome'] } },
  ],
})
