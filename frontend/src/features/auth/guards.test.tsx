import { screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { mockFetch, renderRoute } from '@/test/render'
import { RedirectIfSignedIn, RequireAuth, RequireRole } from './guards'
import { LoginPage } from './LoginPage'

afterEach(() => vi.unstubAllGlobals())

const me = { userId: 9, employeeId: 9, employeeCode: 'ADMIN-001', role: 'admin', mustChangePassword: false }

describe('sign-in flow with real guards', () => {
  it('returns to ?next after signing in (not always /profile)', async () => {
    let signedIn = false
    mockFetch((url) => {
      if (url.endsWith('/auth/login')) {
        signedIn = true
        return { status: 204 }
      }
      if (url.endsWith('/auth/me')) return signedIn ? { status: 200, body: me } : { status: 401, body: { error: { code: 'UNAUTHORIZED' } } }
      return undefined
    })
    renderRoute(
      <RedirectIfSignedIn>
        <LoginPage />
      </RedirectIfSignedIn>,
      {
        path: '/login',
        at: '/login?next=%2Fadmin%2Fskills',
        extra: [
          { path: '/admin/skills', element: <RequireAuth><RequireRole role="admin"><p>Admin skills</p></RequireRole></RequireAuth> },
          { path: '/profile', element: <p>Profile</p> },
        ],
      },
    )
    await userEvent.type(await screen.findByLabelText('Employee code'), 'ADMIN-001')
    await userEvent.type(screen.getByLabelText('Password'), 'demo1234')
    await userEvent.click(screen.getByRole('button', { name: 'Sign in' }))
    expect(await screen.findByText('Admin skills')).toBeInTheDocument()
  })

  it('sends a temporary-password session to /change-password', async () => {
    mockFetch((url) => (url.endsWith('/auth/me') ? { status: 200, body: { ...me, mustChangePassword: true } } : undefined))
    renderRoute(<RequireAuth><p>Profile</p></RequireAuth>, {
      path: '/profile',
      extra: [{ path: '/change-password', element: <p>Set a new password</p> }],
    })
    expect(await screen.findByText('Set a new password')).toBeInTheDocument()
  })

  it('keeps employees out of admin pages', async () => {
    mockFetch((url) => (url.endsWith('/auth/me') ? { status: 200, body: { ...me, role: 'employee' } } : undefined))
    renderRoute(<RequireAuth><RequireRole role="admin"><p>Admin</p></RequireRole></RequireAuth>, {
      path: '/admin',
      extra: [{ path: '/profile', element: <p>Profile</p> }],
    })
    expect(await screen.findByText('Profile')).toBeInTheDocument()
    expect(screen.queryByText('Admin')).not.toBeInTheDocument()
  })
})

describe('admin home', () => {
  it('sends admins to /admin after signing in, not to a profile', async () => {
    mockFetch((url) => (url.endsWith('/auth/me') ? { status: 200, body: me } : undefined))
    renderRoute(<RedirectIfSignedIn><p>Login</p></RedirectIfSignedIn>, {
      path: '/login',
      extra: [
        { path: '/admin', element: <p>Admin home</p> },
        { path: '/profile', element: <p>Profile</p> },
      ],
    })
    expect(await screen.findByText('Admin home')).toBeInTheDocument()
  })

  it('keeps admins out of the employee profile page', async () => {
    mockFetch((url) => (url.endsWith('/auth/me') ? { status: 200, body: me } : undefined))
    renderRoute(<RequireAuth><RequireRole role="employee"><p>Profile</p></RequireRole></RequireAuth>, {
      path: '/profile',
      extra: [{ path: '/admin', element: <p>Admin home</p> }],
    })
    expect(await screen.findByText('Admin home')).toBeInTheDocument()
    expect(screen.queryByText('Profile')).not.toBeInTheDocument()
  })
})
