import { mockFetch } from '@/test/render'
import { api, ApiError, errorMessage, setUnauthorizedHandler } from './client'
import { safeNext } from '@/features/auth/api'

afterEach(() => {
  vi.unstubAllGlobals()
  setUnauthorizedHandler(null)
})

describe('api client', () => {
  it('sends JSON and If-Match, returns the body', async () => {
    const fetch = mockFetch(() => ({ status: 200, body: { version: 8 } }))
    const res = await api.put<{ version: number }>('/profile/contact/mobile', { value: '0812345678' }, { version: 7 })
    expect(res.version).toBe(8)
    const [url, init] = fetch.mock.calls[0]!
    expect(url).toBe('/api/v1/profile/contact/mobile')
    expect((init!.headers as Record<string, string>)['If-Match']).toBe('"7"')
    expect(init!.body).toBe('{"value":"0812345678"}')
  })

  it('turns the error envelope into ApiError', async () => {
    mockFetch(() => ({
      status: 400,
      body: { error: { code: 'VALIDATION_ERROR', message: 'Some fields are invalid.', fields: { mobile: 'must be 9–15 digits' }, requestId: 'r1' } },
    }))
    const err = (await api.get('/x').catch((e: unknown) => e)) as ApiError
    expect(err).toBeInstanceOf(ApiError)
    expect(err.code).toBe('VALIDATION_ERROR')
    expect(err.requestId).toBe('r1')
    expect(errorMessage(err)).toBe('Mobile must be 9–15 digits')
  })

  it('calls the unauthorized handler on 401, except for login and me', async () => {
    const onUnauthorized = vi.fn()
    setUnauthorizedHandler(onUnauthorized)
    mockFetch(() => ({ status: 401, body: { error: { code: 'UNAUTHORIZED', message: 'Please sign in.' } } }))

    await api.get('/auth/me').catch(() => {})
    await api.post('/auth/login', {}).catch(() => {})
    expect(onUnauthorized).not.toHaveBeenCalled()

    await api.get('/profile').catch(() => {})
    expect(onUnauthorized).toHaveBeenCalledOnce()
  })

  it('reports network failures', async () => {
    vi.stubGlobal('fetch', vi.fn().mockRejectedValue(new TypeError('offline')))
    const err = (await api.get('/x').catch((e: unknown) => e)) as ApiError
    expect(err.code).toBe('NETWORK_ERROR')
  })
})

describe('safeNext', () => {
  it.each([
    [null, '/profile'],
    ['/history', '/history'],
    ['/admin/skills?x=1', '/admin/skills?x=1'],
    ['https://evil.example', '/profile'],
    ['//evil.example', '/profile'],
    ['/login', '/profile'],
  ])('%s → %s', (next, want) => expect(safeNext(next)).toBe(want))
})
