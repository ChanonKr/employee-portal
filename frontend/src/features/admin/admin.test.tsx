import { screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { mockFetch, renderRoute } from '@/test/render'
import { EmployeesPage } from './EmployeesPage'
import { SettingsPage } from './SettingsPage'

afterEach(() => vi.unstubAllGlobals())

describe('SettingsPage', () => {
  it('validates 1 to the retention and saves with presets', async () => {
    const fetch = mockFetch((url, init) => {
      if (init.method === 'PUT') return { status: 204 }
      if (url.endsWith('/admin/settings')) return { status: 200, body: { historyDays: 7, maxHistoryDays: 90, updatedBy: null, updatedAt: null } }
      return undefined
    })
    renderRoute(<SettingsPage />)
    const input = await screen.findByLabelText('Days shown')
    expect(input).toHaveValue(7)
    expect(screen.getByText('Using the default value.')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Save' })).toBeDisabled()

    await userEvent.clear(input)
    await userEvent.type(input, '91')
    await userEvent.click(screen.getByRole('button', { name: 'Save' }))
    expect(await screen.findByText('Enter a whole number from 1 to 90.')).toBeInTheDocument()

    await userEvent.click(screen.getByRole('button', { name: '30 days' }))
    await userEvent.click(screen.getByRole('button', { name: 'Save' }))
    await waitFor(() => expect(fetch.mock.calls.some(([, i]) => i?.method === 'PUT')).toBe(true))
    const [url, init] = fetch.mock.calls.find(([, i]) => i?.method === 'PUT')!
    expect(url).toBe('/api/v1/admin/settings/history-days')
    expect(JSON.parse(init!.body as string)).toEqual({ value: 30 })
  })
})

describe('EmployeesPage', () => {
  const list = { items: [], page: { page: 1, pageSize: 20, hasNext: false }, total: 0 }

  it('creates an account and shows the temporary password once', async () => {
    mockFetch((url, init) => {
      if (init.method === 'POST') {
        const body = JSON.parse(init.body as string)
        return body.employeeCode === 'EMP-00001'
          ? { status: 409, body: { error: { code: 'ALREADY_EXISTS', message: 'x' } } }
          : { status: 201, body: { id: 9, employeeCode: body.employeeCode, temporaryPassword: 'Ab3dEf6hJk9m' } }
      }
      if (url.startsWith('/api/v1/admin/employees')) return { status: 200, body: list }
      return undefined
    })
    renderRoute(<EmployeesPage />)
    await userEvent.click(await screen.findByRole('button', { name: 'New account' }))

    await userEvent.type(screen.getByLabelText('Employee code'), 'emp-00001')
    await userEvent.type(screen.getByLabelText('First name'), 'New')
    await userEvent.type(screen.getByLabelText('Last name'), 'Hire')
    await userEvent.click(screen.getByRole('button', { name: 'Create account' }))
    expect(await screen.findByText('This employee code is already used')).toBeInTheDocument()

    await userEvent.clear(screen.getByLabelText('Employee code'))
    await userEvent.type(screen.getByLabelText('Employee code'), 'emp-00042')
    await userEvent.click(screen.getByRole('button', { name: 'Create account' }))
    expect(await screen.findByText('Ab3dEf6hJk9m')).toBeInTheDocument()
    expect(screen.getByText('EMP-00042')).toBeInTheDocument()
    expect(screen.getByText(/shown only once/)).toBeInTheDocument()
  })

  it('shows an empty search result', async () => {
    mockFetch((url) => (url.startsWith('/api/v1/admin/employees') ? { status: 200, body: list } : undefined))
    renderRoute(<EmployeesPage />)
    await userEvent.type(await screen.findByLabelText('Search employees'), 'zzz')
    expect(await screen.findByText('No employees match “zzz”')).toBeInTheDocument()
  })
})
