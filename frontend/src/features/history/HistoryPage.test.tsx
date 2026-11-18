import { fireEvent, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import type { HistoryItem } from '@/api/types'
import { mockFetch, renderRoute } from '@/test/render'
import { HistoryPage } from './HistoryPage'

afterEach(() => vi.unstubAllGlobals())

const item = (id: number, field: HistoryItem['field'], changedAt: string, extra: Partial<HistoryItem> = {}): HistoryItem => ({
  id, field, oldValue: 'old', newValue: 'new', changedAt, changedBy: { id: 2, role: 'employee' }, ...extra,
})

describe('HistoryPage', () => {
  it('shows the window from the API, groups by day, filters and loads more', async () => {
    const today = new Date().toISOString()
    const fetch = mockFetch((url) => {
      if (url.includes('/auth/me')) return { status: 200, body: { userId: 2, role: 'employee' } }
      const u = new URL(url, 'http://x')
      const page = Number(u.searchParams.get('page'))
      const field = u.searchParams.get('field')
      if (field === 'skills') {
        return { status: 200, body: { days: 30, hasOlder: false, page: { page: 1, pageSize: 20, hasNext: false },
          items: [item(9, 'skills', today, { oldValue: ['Go'], newValue: ['Go', 'React'] })] } }
      }
      return {
        status: 200,
        body: page === 1
          ? { days: 30, hasOlder: false, page: { page: 1, pageSize: 20, hasNext: true },
              items: [item(3, 'mobile_phone', today, { oldValue: '0811111111', newValue: '0812345678' }),
                      item(2, 'last_name', today, { changedBy: { id: 1, role: 'admin' } })] }
          : { days: 30, hasOlder: false, page: { page: 2, pageSize: 20, hasNext: false }, items: [item(1, 'address', '2026-10-20T08:00:00Z')] },
      }
    })

    renderRoute(<HistoryPage />)
    expect(await screen.findByText('Showing changes from the last 30 days.')).toBeInTheDocument()
    expect(screen.getByRole('heading', { name: 'Today' })).toBeInTheDocument()
    expect(screen.getByText('081 234 5678')).toBeInTheDocument()
    expect(screen.getByText('Changed by an admin')).toBeInTheDocument()

    await userEvent.click(screen.getByRole('button', { name: 'Load more' }))
    expect(await screen.findByText('Address', { selector: 'p' })).toBeInTheDocument()
    expect(screen.queryByRole('button', { name: 'Load more' })).not.toBeInTheDocument()

    await userEvent.click(screen.getByRole('button', { name: 'Skills' }))
    expect(await screen.findByText('React')).toBeInTheDocument()
    await waitFor(() => expect(fetch.mock.calls.some(([u]) => String(u).includes('field=skills'))).toBe(true))
    expect(screen.getByRole('button', { name: 'Skills' })).toHaveAttribute('aria-pressed', 'true')
  })

  it('shows photo changes as thumbnails and a placeholder for a deleted file', async () => {
    const today = new Date().toISOString()
    mockFetch((url) =>
      url.includes('/auth/me')
        ? { status: 200, body: { userId: 2, role: 'employee' } }
        : { status: 200, body: { days: 30, hasOlder: false, page: { page: 1, pageSize: 20, hasNext: false },
            items: [item(5, 'photo', today, { oldValue: '/api/v1/files/photos/old.jpg', newValue: '/api/v1/files/photos/new.jpg' })] } },
    )
    renderRoute(<HistoryPage />)
    const before = await screen.findByAltText('Previous photo')
    expect(before).toHaveAttribute('src', '/api/v1/files/photos/old.jpg')
    expect(screen.getByAltText('New photo')).toHaveAttribute('src', '/api/v1/files/photos/new.jpg')

    fireEvent.error(before) // the retention job deleted the old file
    expect(await screen.findByText('Previous photo is no longer kept')).toBeInTheDocument()
  })

  it('empty state uses the API window', async () => {
    mockFetch((url) =>
      url.includes('/auth/me')
        ? { status: 200, body: { userId: 2, role: 'employee' } }
        : { status: 200, body: { days: 14, hasOlder: true, items: [], page: { page: 1, pageSize: 20, hasNext: false } } },
    )
    renderRoute(<HistoryPage />)
    expect(await screen.findByText('No changes in the last 14 days')).toBeInTheDocument()
  })
})
