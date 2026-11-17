import { fireEvent, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import type { HistoryPage, Profile } from '@/api/types'
import { mockFetch, renderRoute } from '@/test/render'
import { checklist } from './ProfileChecklist'
import { ProfilePage } from './ProfilePage'

afterEach(() => vi.unstubAllGlobals())

const empty: Profile = {
  id: 2, employeeCode: 'EMP-00002', firstName: 'Suda', lastName: 'Rakdee', photoUrl: null,
  contact: { mobilePhone: null, email: null, address: null }, skills: [], version: 1,
}
const noHistory: HistoryPage = { days: 7, hasOlder: false, items: [], page: { page: 1, pageSize: 3, hasNext: false } }

function server(profile: Profile, onPut?: (url: string, init: RequestInit) => { status: number; body?: unknown }) {
  return mockFetch((url, init) => {
    if (init.method === 'PUT' && onPut) return onPut(url, init)
    if (url === '/api/v1/profile') return { status: 200, body: profile }
    if (url.startsWith('/api/v1/profile/history')) return { status: 200, body: noHistory }
    if (url === '/api/v1/skills') return { status: 200, body: [{ id: 1, name: 'Go' }, { id: 2, name: 'React' }] }
    if (url === '/api/v1/auth/me') return { status: 200, body: { userId: 2, role: 'employee' } }
    return undefined
  })
}

describe('ProfilePage', () => {
  it('first-time state: checklist, empty fields and a dynamic empty history message', async () => {
    server(empty)
    renderRoute(<ProfilePage />)
    expect(await screen.findByText('Welcome, Suda')).toBeInTheDocument()
    expect(screen.getAllByText('Not added yet')).toHaveLength(3)
    expect(screen.getByRole('button', { name: 'Add photo' })).toBeInTheDocument()
    expect(await screen.findByText('No changes yet')).toBeInTheDocument()
    expect(screen.getByText(/from the last 7 days/)).toBeInTheDocument()
  })

  it('name is read-only (no edit control)', async () => {
    server(empty)
    renderRoute(<ProfilePage />)
    await screen.findByRole('heading', { name: /Suda Rakdee/ })
    expect(screen.queryByRole('button', { name: /edit name/i })).not.toBeInTheDocument()
  })

  it('saves one contact field with If-Match and shows client-side errors first', async () => {
    const fetch = server(empty, () => ({ status: 200, body: { version: 2 } }))
    renderRoute(<ProfilePage />)
    await userEvent.click(await screen.findByRole('button', { name: 'Add mobile' }))
    const input = screen.getByRole('textbox', { name: 'Mobile' })

    await userEvent.type(input, '12ab')
    await userEvent.click(screen.getByRole('button', { name: 'Save' }))
    expect(await screen.findByText(/Enter 9–15 digits/)).toBeInTheDocument()
    expect(fetch.mock.calls.some(([, i]) => i?.method === 'PUT')).toBe(false)

    await userEvent.clear(input)
    await userEvent.type(input, '081-234-5678')
    await userEvent.click(screen.getByRole('button', { name: 'Save' }))
    await waitFor(() => expect(fetch.mock.calls.some(([, i]) => i?.method === 'PUT')).toBe(true))
    const [url, init] = fetch.mock.calls.find(([, i]) => i?.method === 'PUT')!
    expect(url).toBe('/api/v1/profile/contact/mobile')
    expect((init!.headers as Record<string, string>)['If-Match']).toBe('"1"')
    expect(JSON.parse(init!.body as string)).toEqual({ value: '0812345678' })
  })

  it('shows the server message when it rejects a value', async () => {
    server(empty, () => ({
      status: 400,
      body: { error: { code: 'VALIDATION_ERROR', message: 'Some fields are invalid.', fields: { email: 'must be a valid email address' } } },
    }))
    renderRoute(<ProfilePage />)
    await userEvent.click(await screen.findByRole('button', { name: 'Add email' }))
    await userEvent.type(screen.getByRole('textbox', { name: 'Email' }), 'a@b.co')
    await userEvent.click(screen.getByRole('button', { name: 'Save' }))
    expect(await screen.findByText('Email must be a valid email address')).toBeInTheDocument()
  })

  it('skills: Save appears only after a change and sends the whole set', async () => {
    const fetch = server({ ...empty, skills: [{ id: 1, name: 'Go' }] }, () => ({ status: 200, body: { version: 2 } }))
    renderRoute(<ProfilePage />)
    const list = await screen.findByRole('list', { name: 'Your skills' })
    expect(within(list).getByText('Go')).toBeInTheDocument()
    expect(screen.queryByRole('button', { name: 'Save skills' })).not.toBeInTheDocument()

    await userEvent.click(screen.getByRole('button', { name: 'Add' }))
    await userEvent.click(await screen.findByRole('button', { name: 'React' }))
    await userEvent.click(screen.getByRole('button', { name: 'Save skills' }))
    await waitFor(() => expect(fetch.mock.calls.some(([u]) => u === '/api/v1/profile/skills')).toBe(true))
    const [, init] = fetch.mock.calls.find(([u]) => u === '/api/v1/profile/skills')!
    expect(JSON.parse(init!.body as string)).toEqual({ skillIds: [1, 2] })
  })
})

describe('checklist', () => {
  it('needs all contact fields for "Contact info"', () => {
    const half = { ...empty, contact: { mobilePhone: '0812345678', email: null, address: null } }
    expect(checklist(half).find((i) => i.label === 'Contact info')?.done).toBe(false)
    const full = { ...empty, contact: { mobilePhone: '1', email: 'a', address: 'b' } }
    expect(checklist(full).find((i) => i.label === 'Contact info')?.done).toBe(true)
  })

  it('a chosen photo is only previewed; it uploads after "Save photo" and Cancel discards it', async () => {
    // jsdom has no object URLs.
    URL.createObjectURL = () => 'blob:preview'
    URL.revokeObjectURL = () => {}
    const fetch = server(empty, () => ({ status: 200, body: { version: 2, photoUrl: '/api/v1/files/photos/x.jpg' } }))
    renderRoute(<ProfilePage />)
    await screen.findByRole('button', { name: 'Add photo' })
    const file = new File(['x'], 'me.png', { type: 'image/png' })
    const puts = () => fetch.mock.calls.filter(([, i]) => i?.method === 'PUT')

    await userEvent.upload(screen.getByLabelText('Profile photo'), file)
    expect(await screen.findByRole('button', { name: 'Save photo' })).toBeInTheDocument()
    expect(puts()).toHaveLength(0)

    await userEvent.click(screen.getByRole('button', { name: 'Cancel' }))
    expect(screen.queryByRole('button', { name: 'Save photo' })).not.toBeInTheDocument()
    expect(puts()).toHaveLength(0)

    await userEvent.upload(screen.getByLabelText('Profile photo'), file)
    await userEvent.click(await screen.findByRole('button', { name: 'Save photo' }))
    await waitFor(() => expect(puts()).toHaveLength(1))
    expect(puts()[0]?.[0]).toBe('/api/v1/profile/photo')
  })

  it('dropping a file on the picture does nothing', async () => {
    const fetch = server(empty)
    renderRoute(<ProfilePage />)
    const picture = (await screen.findByRole('button', { name: 'Add photo' })).parentElement!
    fireEvent.drop(picture, { dataTransfer: { files: [new File(['x'], 'me.png', { type: 'image/png' })] } })
    expect(screen.queryByRole('button', { name: 'Save photo' })).not.toBeInTheDocument()
    expect(fetch.mock.calls.some(([, i]) => i?.method === 'PUT')).toBe(false)
  })
})
