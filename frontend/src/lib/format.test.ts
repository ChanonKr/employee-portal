import { dayLabel, formatPhone, initials, timeAgo } from './format'

const now = new Date('2026-11-17T12:00:00')

it('formats Thai mobile numbers', () => {
  expect(formatPhone('0812345678')).toBe('081 234 5678')
  expect(formatPhone('+66812345678')).toBe('+66812345678')
  expect(formatPhone(null)).toBeNull()
})

it('builds initials', () => expect(initials('somchai', 'Jaidee')).toBe('SJ'))

it('relative times', () => {
  expect(timeAgo('2026-11-17T11:59:30', now)).toBe('just now')
  expect(timeAgo('2026-11-17T11:30:00', now)).toBe('30 minutes ago')
  expect(timeAgo('2026-11-17T10:00:00', now)).toBe('2 hours ago')
  expect(timeAgo('2026-11-16T09:00:00', now)).toBe('yesterday')
  expect(timeAgo('2026-11-01T09:00:00', now)).toBe('1 Nov 2026')
})

it('day labels', () => {
  expect(dayLabel('2026-11-17T01:00:00', now)).toBe('Today')
  expect(dayLabel('2026-11-16T23:00:00', now)).toBe('Yesterday')
  expect(dayLabel('2026-11-10T08:00:00', now)).toMatch(/10 Nov 2026/)
})
