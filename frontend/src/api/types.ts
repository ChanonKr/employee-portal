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

// ── admin

export type AdminSkill = { id: number; name: string; isActive: boolean; sortOrder: number; usedBy: number }

export type AdminSettings = {
  historyDays: number
  maxHistoryDays: number // retention: changes older than this are deleted
  updatedBy: { id: number; employeeCode: string } | null
  updatedAt: string | null
}

export type EmployeeSummary = {
  id: number
  employeeCode: string
  firstName: string
  lastName: string
  photoUrl: string | null
  role: Role
  isActive: boolean
}

export type EmployeeList = {
  items: EmployeeSummary[]
  page: { page: number; pageSize: number; hasNext: boolean }
  total: number
}

export type EmployeeDetail = Omit<Profile, 'skills'> & {
  skills: (Skill & { isActive: boolean })[]
  role: Role
  isActive: boolean
}

export type DashboardStats = {
  employees: number
  completeProfiles: number
  topSkills: { id: number; name: string; usedBy: number }[]
  recentChanges: {
    id: number
    employeeId: number
    employeeCode: string
    name: string
    field: HistoryField
    changedAt: string
    changedByRole: Role
  }[]
}

export type CreatedAccount = { id: number; employeeCode: string; temporaryPassword: string }
