export type Role = 'employee' | 'admin'

export type Me = {
  userId: number
  employeeId: number | null
  employeeCode: string
  role: Role
  mustChangePassword: boolean
}

export type Skill = { id: number; name: string }

export type Profile = {
  id: number
  employeeCode: string
  firstName: string
  lastName: string
  photoUrl: string | null
  contact: { mobilePhone: string | null; email: string | null; address: string | null }
  skills: Skill[]
  version: number
}

export type HistoryField =
  | 'first_name'
  | 'last_name'
  | 'mobile_phone'
  | 'contact_email'
  | 'address'
  | 'photo'
  | 'skills'

export type HistoryItem = {
  id: number
  field: HistoryField
  oldValue: string | string[] | null
  newValue: string | string[] | null
  changedAt: string
  changedBy: { id: number; role: Role }
}

export type HistoryPage = {
  days: number | null // null = all time (admin view)
  hasOlder: boolean
  items: HistoryItem[]
  page: { page: number; pageSize: number; hasNext: boolean }
}

export type Version = { version: number }
