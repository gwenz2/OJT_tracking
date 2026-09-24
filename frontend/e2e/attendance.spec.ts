/**
 * T085/T100/T139 — Playwright critical path:
 *   login → Time In (GPS+camera) → Time Out → journal auto-created →
 *   submit → staff needs-revision → resubmit → reviewed.
 * Camera frames come from Chromium's fake media stream; geolocation is
 * stubbed inside the seeded site's radius (see playwright.config.ts).
 */
import { test, expect, type Page } from '@playwright/test'

const TRAINEE = { email: 'trainee1@ojt.local', password: 'Trainee123!' }
const ADMIN = { email: 'admin@ojt.local', password: 'Admin123!' }

async function login(page: Page, creds = TRAINEE) {
  await page.goto('/login')
  await page.getByLabel('Email').fill(creds.email)
  await page.getByLabel('Password').fill(creds.password)
  await page.getByRole('button', { name: /sign in/i }).click()
  await expect(page).not.toHaveURL(/login/)
}

async function captureFlow(page: Page, actionLabel: RegExp) {
  await page.getByRole('button', { name: actionLabel }).click()
  await page.getByRole('button', { name: /get location/i }).click()
  await page.getByRole('button', { name: /continue to camera/i }).click()
  await page.getByRole('button', { name: /capture photo/i }).click()
  await page.getByRole('button', { name: /confirm (time in|time out)/i }).click()
  await expect(page.getByText(/timed (in|out) at/i)).toBeVisible()
  await page.getByRole('button', { name: /done/i }).click()
}

test('trainee attendance → journal → staff review lifecycle', async ({ page }) => {
  test.setTimeout(120_000)

  await login(page)
  await expect(page).toHaveURL(/today/)

  // Time In.
  await captureFlow(page, /^time in$/i)

  // Time Out — auto-creates the day's journal.
  await captureFlow(page, /^time out$/i)

  // Time Out auto-navigates straight to the freshly created journal.
  await expect(page).toHaveURL(/\/journals\//)
  await page.getByLabel('Journal narrative').fill(
    'Deployed fixes to the attendance kiosk and shadowed the IT officer during morning rounds.',
  )
  // action bar clears the fixed tab bar (bottom-16)
  await page.getByRole('button', { name: /save draft/i }).click()
  await page.getByRole('button', { name: /submit journal/i }).click()
  await expect(page.getByText(/^submitted$/i).first()).toBeVisible()

  // Staff side: admin opens the queue, requests revision, then approves.
  // clearCookies + reload so the in-memory session state also resets.
  await page.context().clearCookies()
  await page.goto('/login')
  await page.reload()

  await login(page, ADMIN)
  await page.goto('/staff/journals')
  await page.getByRole('link', { name: /trainee one/i }).first().click()

  await expect(page).toHaveURL(/\/staff\/journals\//)
  await page.getByLabel('Review comment').fill('Please add more detail about the kiosk work.')
  await page.getByRole('button', { name: /request revision/i }).click()
  await expect(page.getByText(/needs revision/i).first()).toBeVisible()

  // Trainee sees the revision request, edits, resubmits.
  await page.context().clearCookies()
  await page.goto('/login')
  await page.reload()
  await login(page)
  await page.getByRole('button', { name: /revise journal/i }).click()
  await page.getByLabel('Journal narrative').fill(
    'Deployed fixes to the attendance kiosk, replaced a failing barcode scanner, ' +
      'and shadowed the IT officer during morning rounds and the afternoon asset audit.',
  )
  await page.getByRole('button', { name: /resubmit/i }).click()
  await expect(page.getByText(/^submitted$/i).first()).toBeVisible()

  // Admin approves.
  await page.context().clearCookies()
  await page.goto('/login')
  await page.reload()
  await login(page, ADMIN)
  await page.goto('/staff/journals')
  await page.getByRole('link', { name: /trainee one/i }).first().click()
  await page.getByRole('button', { name: /^approve$/i }).click()
  await expect(page.getByText(/reviewed/i).first()).toBeVisible()
})
