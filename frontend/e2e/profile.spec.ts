import { deflateSync } from 'node:zlib'
import { expect, hasNoHorizontalScroll, PASSWORD, signInThroughUi, test } from './fixtures'

test('sign in, edit mobile, and see it in recent changes and history', async ({ signedIn: page }) => {
  await expect(page.getByText('Welcome, Test')).toBeVisible()

  // Same accessible button on every screen size: inline form on tablet/desktop, bottom sheet on phones.
  await page.getByRole('button', { name: 'Add mobile' }).click()
  const input = page.getByRole('textbox', { name: 'Mobile' })
  await input.fill('12ab')
  await page.getByRole('button', { name: 'Save' }).click()
  await expect(page.getByText(/Enter 9–15 digits/)).toBeVisible()

  await input.fill('081-234-5678')
  await page.getByRole('button', { name: 'Save' }).click()
  await expect(page.getByText('Mobile saved')).toBeVisible()
  const card = (title: string) => page.locator('section').filter({ has: page.getByRole('heading', { name: title }) })
  await expect(card('Contact').getByText('081 234 5678')).toBeVisible()
  await expect(page.getByRole('button', { name: 'Edit mobile' })).toBeVisible()
  await expect(card('Recent changes').getByText('081 234 5678')).toBeVisible()

  await page.getByRole('link', { name: 'View all' }).click()
  await expect(page.getByRole('heading', { name: 'Change history' })).toBeVisible()
  await expect(page.getByText(/Showing changes from the last \d+ days/)).toBeVisible()
  await expect(page.getByText('081 234 5678')).toBeVisible()
  expect(await hasNoHorizontalScroll(page)).toBe(true)
})

test('the name is read-only for employees', async ({ signedIn: page, employee }) => {
  await expect(page.getByRole('heading', { name: `${employee.firstName} ${employee.lastName}` })).toBeVisible()
  await expect(page.getByRole('textbox', { name: /first name/i })).toHaveCount(0)
  await expect(page.getByText('Contact HR to change it')).toBeVisible()
  // The API has no route for it either.
  const res = await page.request.put('/api/v1/profile/name', { data: { firstName: 'Hacker' } })
  expect(res.status()).toBe(404)
})

test('choose skills and upload a photo', async ({ signedIn: page }) => {
  await page.getByRole('button', { name: 'Add', exact: true }).click()
  await page.getByRole('button', { name: 'Go', exact: true }).click()
  await page.getByRole('button', { name: 'React', exact: true }).click()
  // Close the picker (Done on phones, Escape for the desktop popover).
  const done = page.getByRole('button', { name: 'Done' })
  if (await done.isVisible()) await done.click()
  else await page.keyboard.press('Escape')
  await page.getByRole('button', { name: 'Save skills' }).click()
  await expect(page.getByText('Skills saved')).toBeVisible()
  await expect(page.getByRole('list', { name: 'Your skills' }).getByText('React')).toBeVisible()

  await page.getByLabel('Profile photo').setInputFiles({ name: 'me.png', mimeType: 'image/png', buffer: tinyPng() })
  await expect(page.getByText('Photo updated')).toBeVisible()
  await expect(page.getByRole('button', { name: 'Change photo' })).toBeVisible()
})

test('no horizontal scroll on the main pages', async ({ signedIn: page }) => {
  expect(await hasNoHorizontalScroll(page)).toBe(true)
  await page.goto('/history')
  await expect(page.getByRole('heading', { name: 'Change history' })).toBeVisible()
  expect(await hasNoHorizontalScroll(page)).toBe(true)
})

test('a wrong password shows one generic message', async ({ page, employee }) => {
  await page.goto('/login')
  await page.getByLabel('Employee code').fill(employee.code)
  await page.getByLabel('Password', { exact: true }).fill('not-my-password1')
  await page.getByRole('button', { name: 'Sign in' }).click()
  await expect(page.getByRole('alert')).toHaveText('Employee code or password is incorrect.')
  await signInThroughUi(page, employee.code, PASSWORD)
})

/** A valid 8×8 PNG built in memory (no fixture files). */
function tinyPng(): Buffer {
  const crcTable = Array.from({ length: 256 }, (_, n) => {
    let c = n
    for (let k = 0; k < 8; k++) c = c & 1 ? 0xedb88320 ^ (c >>> 1) : c >>> 1
    return c >>> 0
  })
  const crc = (b: Buffer) => {
    let c = 0xffffffff
    for (const x of b) c = crcTable[(c ^ x) & 0xff]! ^ (c >>> 8)
    return (c ^ 0xffffffff) >>> 0
  }
  const chunk = (type: string, data: Buffer) => {
    const len = Buffer.alloc(4)
    len.writeUInt32BE(data.length)
    const body = Buffer.concat([Buffer.from(type), data])
    const sum = Buffer.alloc(4)
    sum.writeUInt32BE(crc(body))
    return Buffer.concat([len, body, sum])
  }
  const w = 8, h = 8
  const ihdr = Buffer.alloc(13)
  ihdr.writeUInt32BE(w, 0)
  ihdr.writeUInt32BE(h, 4)
  ihdr[8] = 8
  ihdr[9] = 2
  const raw = Buffer.concat(Array.from({ length: h }, () => Buffer.concat([Buffer.from([0]), Buffer.alloc(w * 3, 0x66)])))
  return Buffer.concat([Buffer.from('\x89PNG\r\n\x1a\n', 'binary'), chunk('IHDR', ihdr), chunk('IDAT', deflateSync(raw)), chunk('IEND', Buffer.alloc(0))])
}
