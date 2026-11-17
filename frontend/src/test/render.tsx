import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { render } from '@testing-library/react'
import type { ReactElement } from 'react'
import { MemoryRouter, Route, Routes } from 'react-router'

/** Render one route with fresh React Query state. */
type RouteOptions = { path?: string; at?: string; extra?: { path: string; element: ReactElement }[] }

export function renderRoute(element: ReactElement, opts: RouteOptions = {}) {
  const { path = '/', extra = [] } = opts
  const at = opts.at ?? path
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false }, mutations: { retry: false } } })
  return render(
    <QueryClientProvider client={qc}>
      <MemoryRouter initialEntries={[at]}>
        <Routes>
          <Route path={path} element={element} />
          {extra.map((r) => (
            <Route key={r.path} path={r.path} element={r.element} />
          ))}
        </Routes>
      </MemoryRouter>
    </QueryClientProvider>,
  )
}

type Handler = (url: string, init: RequestInit) => { status: number; body?: unknown } | undefined

/** Replace fetch with a small router of canned responses. */
export function mockFetch(handler: Handler) {
  const fn = vi.fn(async (input: RequestInfo | URL, init: RequestInit = {}) => {
    const url = typeof input === 'string' ? input : input.toString()
    const res = handler(url, init) ?? { status: 404, body: { error: { code: 'NOT_FOUND', message: 'Not found.' } } }
    return new Response(res.body === undefined ? null : JSON.stringify(res.body), {
      status: res.status,
      headers: { 'Content-Type': 'application/json' },
    })
  })
  vi.stubGlobal('fetch', fn)
  return fn
}
