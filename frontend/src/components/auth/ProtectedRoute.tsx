import { useState, useEffect } from 'react'
import { Navigate, useLocation } from 'react-router-dom'
import { useAuthStore } from '@/stores/auth'
import { useUserProfile } from '@/hooks/useAuth'

interface Props {
  children: React.ReactNode
  role?: 'user' | 'admin'
}

export function ProtectedRoute({ children, role }: Props) {
  const { isAuthenticated, token } = useAuthStore()
  const { data: user, isLoading } = useUserProfile()
  const location = useLocation()
  const [isStoreReady, setIsStoreReady] = useState(false)

  useEffect(() => {
    setIsStoreReady(true)
  }, [])

  // Wait for Zustand persist store to hydrate from localStorage on page load
  if (!isStoreReady) {
    return (
      <div className="flex h-screen items-center justify-center">
        <div className="h-8 w-8 animate-spin rounded-full border-4 border-primary border-t-transparent" />
      </div>
    )
  }

  if (!token || !isAuthenticated) {
    const loginPath = role === 'admin' ? '/admin/login' : '/user/login'
    return <Navigate to={loginPath} state={{ from: location }} replace />
  }

  if (isLoading) {
    return (
      <div className="flex h-screen items-center justify-center">
        <div className="h-8 w-8 animate-spin rounded-full border-4 border-primary border-t-transparent" />
      </div>
    )
  }

  if (role === 'admin' && user?.role !== 'admin') {
    return <Navigate to="/user/dashboard" replace />
  }

  return <>{children}</>
}
