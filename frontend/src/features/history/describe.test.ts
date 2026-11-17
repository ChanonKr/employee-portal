import type { HistoryItem } from '@/api/types'
import { displayValue, skillDiff } from './describe'

const item = (o: HistoryItem['oldValue'], n: HistoryItem['newValue']): HistoryItem => ({
  id: 1, field: 'skills', oldValue: o, newValue: n, changedAt: '', changedBy: { id: 1, role: 'employee' },
})

it('diffs skills by name', () => {
  expect(skillDiff(item(['Go', 'React', 'Flutter'], ['Go', 'React', 'CI/CD']))).toEqual({ added: ['CI/CD'], removed: ['Flutter'] })
  expect(skillDiff(item([], ['Go']))).toEqual({ added: ['Go'], removed: [] })
})

it('formats values for display', () => {
  expect(displayValue('mobile_phone', '0812345678')).toBe('081 234 5678')
  expect(displayValue('photo', 'abc.jpg')).toBeNull()
  expect(displayValue('address', null)).toBeNull()
})
