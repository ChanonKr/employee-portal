import { Route, Routes } from 'react-router'
import { AppShell } from '@/components/AppShell'
import { AdminLayout } from '@/features/admin/AdminLayout'
import { DashboardPage } from '@/features/admin/DashboardPage'
import { EmployeeDetailPage } from '@/features/admin/EmployeeDetailPage'
import { EmployeesPage } from '@/features/admin/EmployeesPage'
import { SettingsPage } from '@/features/admin/SettingsPage'
import { SkillsAdminPage } from '@/features/admin/SkillsAdminPage'
import { SessionWatcher } from '@/components/SessionWatcher'
import { ChangePasswordPage } from '@/features/auth/ChangePasswordPage'
import { HomeRedirect, RedirectIfSignedIn, RequireAuth, RequireRole } from '@/features/auth/guards'
import { LoginPage } from '@/features/auth/LoginPage'
import { HistoryPage } from '@/features/history/HistoryPage'
import { ProfilePage } from '@/features/profile/ProfilePage'

export function App() {
  return (
    <>
      <SessionWatcher />
      <Routes>
        <Route
          path="/login"
          element={
            <RedirectIfSignedIn>
              <LoginPage />
            </RedirectIfSignedIn>
          }
        />
        <Route
          path="/change-password"
          element={
            <RequireAuth>
              <ChangePasswordPage />
            </RequireAuth>
          }
        />
        <Route
          element={
            <RequireAuth>
              <AppShell />
            </RequireAuth>
          }
        >
          {/* Admins have no profile of their own; they manage everyone else's under /admin. */}
          <Route
            path="/profile"
            element={
              <RequireRole role="employee">
                <ProfilePage />
              </RequireRole>
            }
          />
          <Route
            path="/history"
            element={
              <RequireRole role="employee">
                <HistoryPage />
              </RequireRole>
            }
          />
          <Route
            path="/admin"
            element={
              <RequireRole role="admin">
                <AdminLayout />
              </RequireRole>
            }
          >
            <Route index element={<DashboardPage />} />
            <Route path="employees" element={<EmployeesPage />} />
            <Route path="employees/:id" element={<EmployeeDetailPage />} />
            <Route path="skills" element={<SkillsAdminPage />} />
            <Route path="settings" element={<SettingsPage />} />
          </Route>
        </Route>
        <Route
          path="*"
          element={
            <RequireAuth>
              <HomeRedirect />
            </RequireAuth>
          }
        />
      </Routes>
    </>
  )
}
