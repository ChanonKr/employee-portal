import { test as base, expect, request, type APIRequestContext, type Page } from '@playwright/test'

export const PASSWORD = 'e2e-pass-123'
const ADMIN = { employeeCode: 'ADMIN-001', password: 'demo1234' }

async function signedInApi(baseURL: string, employeeCode: string, password: string): Promise<APIRequestContext> {
  const api = await request.newContext({ baseURL })
  const res = await api.post('/api/v1/auth/login', { data: { employeeCode, password } })
  expect(res.status(), `sign in as ${employeeCode}`).toBe(204)
  return api
}

type Employee = { code: string; firstName: string; lastName: string }

/**
 * Every test gets its own fresh employee (created through the admin API, temporary
 * password already replaced), so tests never depend on each other's data.
 */
export const test = base.extend<{ admin: APIRequestContext; employee: Employee; signedIn: Page }>({
  admin: async ({ baseURL }, use) => {
    const api = await signedInApi(baseURL!, ADMIN.employeeCode, ADMIN.password)
    await use(api)
    await api.dispose()
  },

  employee: async ({ admin, baseURL }, use) => {
    const code = `E2E-${Date.now().toString(36).toUpperCase()}`
    const created = await admin.post('/api/v1/admin/employees', { data: { employeeCode: code, firstName: 'Test', lastName: 'Person' } })
    expect(created.status()).toBe(201)
    const { temporaryPassword } = await created.json()

    const user = await signedInApi(baseURL!, code, temporaryPassword)
    const changed = await user.post('/api/v1/auth/change-password', { data: { currentPassword: temporaryPassword, newPassword: PASSWORD } })
    expect(changed.status()).toBe(204)
    await user.dispose()

    await use({ code, firstName: 'Test', lastName: 'Person' })
  },

  /** A page already signed in as the fresh employee, on /profile. */
  signedIn: async ({ page, employee }, use) => {
    await signInThroughUi(page, employee.code, PASSWORD)
    await use(page)
  },
})

export async function signInThroughUi(page: Page, code: string, password: string) {
  await page.goto('/login')
  await page.getByLabel('Employee code').fill(code)
  await page.getByLabel('Password', { exact: true }).fill(password)
  await page.getByRole('button', { name: 'Sign in' }).click()
  await page.waitForURL((u) => u.pathname === '/profile')
  await expect(page.getByRole('heading', { name: 'Contact' })).toBeVisible()
}

/** True when nothing overflows the viewport sideways. */
export async function hasNoHorizontalScroll(page: Page) {
  return page.evaluate(() => document.documentElement.scrollWidth <= document.documentElement.clientWidth)
}

export { expect }
