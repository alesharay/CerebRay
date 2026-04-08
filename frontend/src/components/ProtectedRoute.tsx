import { useEffect } from 'react'
import { useAuthStore } from '../store/authStore'
import { login } from '../api/auth'

export function ProtectedRoute({ children }: { children: React.ReactNode }) {
  const { user, loading, checkAuth } = useAuthStore()

  useEffect(() => {
    checkAuth()
  }, [checkAuth])

  if (loading) {
    return (
      <div className="flex h-screen items-center justify-center bg-zinc-950">
        <div className="text-zinc-400">Loading...</div>
      </div>
    )
  }

  if (!user) {
    login()
    return null
  }

  return <>{children}</>
}
