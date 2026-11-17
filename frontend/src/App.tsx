import { Navigate, Route, Routes } from 'react-router'
import { AppShell } from '@/components/AppShell'
import { SessionWatcher } from '@/components/SessionWatcher'
import { ChangePasswordPage } from '@/features/auth/ChangePasswordPage'
import { RedirectIfSignedIn, RequireAuth } from '@/features/auth/guards'
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
          <Route path="/profile" element={<ProfilePage />} />
          <Route path="/history" element={<HistoryPage />} />
        </Route>
        <Route path="*" element={<Navigate to="/profile" replace />} />
      </Routes>
    </>
  )
}
