import { test, expect } from '@playwright/test'

test.describe('Landing page', () => {
  test('loads with correct title', async ({ page }) => {
    await page.goto('/')
    await expect(page).toHaveTitle('Cereb-Ray')
  })

  test('shows sign-in button', async ({ page }) => {
    await page.goto('/')
    await expect(page.getByRole('button', { name: 'Sign in to get started' })).toBeVisible()
  })

  test('shows feature descriptions', async ({ page }) => {
    await page.goto('/')
    await expect(page.getByText('Learn through conversation')).toBeVisible()
    await expect(page.getByText('AI-structured notes')).toBeVisible()
    await expect(page.getByText('Grow your knowledge base')).toBeVisible()
    await expect(page.getByText('Connect everything')).toBeVisible()
  })

  test('shows hero tagline', async ({ page }) => {
    await page.goto('/')
    await expect(
      page.getByText('A personal Zettelkasten powered by AI conversations')
    ).toBeVisible()
  })
})

test.describe('Auth redirect', () => {
  test('unauthenticated visit to /dashboard redirects to login', async ({ page }) => {
    // AppLayout wraps children in ProtectedRoute, which calls login()
    // when no user is found. login() navigates to /auth/login.
    // Without a backend, the navigation will fail, but we can verify
    // the page does not render dashboard content.
    await page.goto('/')
    await expect(page.getByRole('button', { name: 'Sign in to get started' })).toBeVisible()

    // Going to /dashboard without auth should not show dashboard content
    await page.goto('/dashboard')
    // ProtectedRoute shows "Loading..." then redirects via login()
    // We just verify the dashboard heading is NOT visible
    await expect(page.getByRole('heading', { name: 'Dashboard' })).not.toBeVisible()
  })
})

test.describe('Authenticated flow (requires backend)', () => {
  // These tests only run when the backend is reachable in local auth mode
  test.beforeAll(async ({ request }) => {
    try {
      const resp = await request.get('http://localhost:8080/health')
      if (resp.status() !== 200) {
        test.skip(true, 'Backend not reachable - skipping authenticated tests')
      }
    } catch {
      test.skip(true, 'Backend not reachable - skipping authenticated tests')
    }
  })

  test('login and navigate app pages', async ({ page }) => {
    // Local auth mode: GET /auth/login auto-creates user and sets cookie
    await page.goto('/auth/login')
    await page.waitForURL('**/dashboard', { timeout: 10000 })

    // Dashboard loads
    await expect(page.getByRole('heading', { name: /dashboard/i })).toBeVisible()

    // Navigate to Inbox
    await page.goto('/inbox')
    await page.waitForLoadState('networkidle')
    await expect(page.locator('body')).not.toBeEmpty()

    // Navigate to Echoes
    await page.goto('/echoes')
    await page.waitForLoadState('networkidle')
    await expect(page.locator('body')).not.toBeEmpty()

    // Navigate to Codex
    await page.goto('/codex')
    await page.waitForLoadState('networkidle')
    await expect(page.locator('body')).not.toBeEmpty()
  })
})
