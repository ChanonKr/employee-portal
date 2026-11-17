import { expect, signInThroughUi, test } from './fixtures'

test('admin changes the history window and the employee sees it immediately', async ({ page, admin, employee, browser }, testInfo) => {
  test.skip(testInfo.project.name !== 'desktop', 'one viewport is enough for a global setting')
  const before = (await (await admin.get('/api/v1/admin/settings')).json()).historyDays as number

  try {
    await page.goto('/login?next=/admin/settings')
    await page.getByLabel('Employee code').fill('ADMIN-001')
    await page.getByLabel('Password', { exact: true }).fill('demo1234')
    await page.getByRole('button', { name: 'Sign in' }).click()
    await expect(page.getByRole('heading', { name: 'Settings', level: 1 })).toBeVisible()

    const next = before === 30 ? 14 : 30
    await page.getByRole('button', { name: `${next} days` }).click()
    await page.getByRole('button', { name: 'Save' }).click()
    await expect(page.getByText(`Employees now see the last ${next} days of changes`)).toBeVisible()

    const ctx = await browser.newContext()
    const empPage = await ctx.newPage()
    await signInThroughUi(empPage, employee.code, 'e2e-pass-123')
    await expect(empPage.getByText(`last ${next} days`).first()).toBeVisible()
    await ctx.close()
  } finally {
    await admin.put('/api/v1/admin/settings/history-days', { data: { value: before } })
  }
})

test('admin renames an employee; the employee sees it marked as an admin change', async ({ page, admin, employee }, testInfo) => {
  test.skip(testInfo.project.name === 'ipad', 'covered on desktop and phone')
  const list = await (await admin.get(`/api/v1/admin/employees?q=${employee.code}`)).json()
  const id = list.items[0].id as number

  await page.goto(`/login?next=/admin/employees/${id}`)
  await page.getByLabel('Employee code').fill('ADMIN-001')
  await page.getByLabel('Password', { exact: true }).fill('demo1234')
  await page.getByRole('button', { name: 'Sign in' }).click()
  await expect(page.getByRole('heading', { name: 'Legal name' })).toBeVisible()

  await page.getByLabel('Last name').fill('Renamed')
  await page.getByRole('button', { name: 'Save name' }).click()
  await expect(page.getByText('Name updated')).toBeVisible()
  await expect(page.getByRole('heading', { level: 1, name: 'Test Renamed' })).toBeVisible()
  await expect(page.getByText('Changed by an admin')).toHaveCount(0) // it's the admin's own view

  await page.getByRole('button', { name: 'Sign out' }).click()
  await signInThroughUi(page, employee.code, 'e2e-pass-123')
  await expect(page.getByRole('heading', { name: 'Test Renamed' })).toBeVisible()
  await expect(page.getByText('Changed by an admin')).toBeVisible()
})

test('employees cannot open admin pages', async ({ signedIn: page }) => {
  await expect(page.getByRole('link', { name: 'Admin' })).toHaveCount(0)
  await page.goto('/admin/settings')
  await page.waitForURL((u) => u.pathname === '/profile')
  const res = await page.request.get('/api/v1/admin/settings')
  expect(res.status()).toBe(403)
})

test('admins land on the admin dashboard and have no Profile menu', async ({ page }) => {
  await page.goto('/login')
  await page.getByLabel('Employee code').fill('ADMIN-001')
  await page.getByLabel('Password', { exact: true }).fill('demo1234')
  await page.getByRole('button', { name: 'Sign in' }).click()
  await page.waitForURL((u) => u.pathname === '/admin')
  await expect(page.getByRole('link', { name: 'Profile' })).toHaveCount(0)
  await page.goto('/profile')
  await page.waitForURL((u) => u.pathname === '/admin')
})
