import { screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { mockFetch, renderRoute } from '@/test/render'
import { LoginPage } from './LoginPage'

afterEach(() => vi.unstubAllGlobals())

const routes = { path: '/login', at: '/login?next=%2Fhistory', extra: [{ path: '/history', element: <p>History page</p> }] }

describe('LoginPage', () => {
  it('validates before sending anything', async () => {
    const fetch = mockFetch(() => undefined)
    renderRoute(<LoginPage />, routes)
    await userEvent.click(screen.getByRole('button', { name: 'Sign in' }))
    expect(await screen.findByText('Enter your employee code')).toBeInTheDocument()
    expect(screen.getByText('Enter your password')).toBeInTheDocument()
    expect(fetch).not.toHaveBeenCalled()
  })

  it('shows one message for any wrong credentials', async () => {
    mockFetch(() => ({ status: 401, body: { error: { code: 'UNAUTHORIZED', message: 'Please sign in.' } } }))
    renderRoute(<LoginPage />, routes)
    await userEvent.type(screen.getByLabelText('Employee code'), 'EMP-00001')
    await userEvent.type(screen.getByLabelText('Password'), 'wrong')
    await userEvent.click(screen.getByRole('button', { name: 'Sign in' }))
    expect(await screen.findByRole('alert')).toHaveTextContent('Employee code or password is incorrect.')
  })

  it('explains rate limiting', async () => {
    mockFetch(() => ({ status: 429, body: { error: { code: 'RATE_LIMITED', message: 'x' } } }))
    renderRoute(<LoginPage />, routes)
    await userEvent.type(screen.getByLabelText('Employee code'), 'EMP-00001')
    await userEvent.type(screen.getByLabelText('Password'), 'x')
    await userEvent.click(screen.getByRole('button', { name: 'Sign in' }))
    expect(await screen.findByRole('alert')).toHaveTextContent('Too many attempts')
  })

  it('goes to ?next after signing in', async () => {
    const fetch = mockFetch((url) => (url.endsWith('/auth/login') ? { status: 204 } : undefined))
    renderRoute(<LoginPage />, routes)
    await userEvent.type(screen.getByLabelText('Employee code'), ' EMP-00001 ')
    await userEvent.type(screen.getByLabelText('Password'), 'demo1234')
    await userEvent.click(screen.getByRole('button', { name: 'Sign in' }))
    await waitFor(() => expect(screen.getByText('History page')).toBeInTheDocument())
    expect(JSON.parse(fetch.mock.calls[0]![1]!.body as string)).toEqual({ employeeCode: 'EMP-00001', password: 'demo1234' })
  })
})
