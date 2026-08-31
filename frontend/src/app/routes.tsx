import React, { Suspense } from 'react'
import { createBrowserRouter, Navigate } from 'react-router-dom'
import { ProtectedRoute } from '@/components/auth/ProtectedRoute'

// 懒加载布局组件
const UserLayout = React.lazy(() => import('@/components/layout/UserLayout'))
const AdminLayout = React.lazy(() => import('@/components/layout/AdminLayout'))

// 懒加载用户页面
const UserLoginPage = React.lazy(() => import('@/pages/user/LoginPage'))
const UserRegisterPage = React.lazy(() => import('@/pages/user/RegisterPage'))
const UserDashboardPage = React.lazy(() => import('@/pages/user/DashboardPage'))
const UserPlansPage = React.lazy(() => import('@/pages/user/PlansPage'))
const UserOrdersPage = React.lazy(() => import('@/pages/user/OrdersPage'))
const UserSubscribePage = React.lazy(() => import('@/pages/user/SubscribePage'))
const UserTrafficPage = React.lazy(() => import('@/pages/user/TrafficPage'))
const UserTicketsPage = React.lazy(() => import('@/pages/user/TicketsPage'))
const UserProfilePage = React.lazy(() => import('@/pages/user/ProfilePage'))
const UserCouponsPage = React.lazy(() => import('@/pages/user/CouponsPage'))
const UserGiftCardsPage = React.lazy(() => import('@/pages/user/GiftCardsPage'))
const UserInvitePage = React.lazy(() => import('@/pages/user/InvitePage'))
const UserKnowledgesPage = React.lazy(() => import('@/pages/user/KnowledgesPage'))
const UserNoticesPage = React.lazy(() => import('@/pages/user/NoticesPage'))

// 懒加载管理员页面
const AdminLoginPage = React.lazy(() => import('@/pages/admin/LoginPage'))
const AdminDashboardPage = React.lazy(() => import('@/pages/admin/DashboardPage'))
const AdminUsersPage = React.lazy(() => import('@/pages/admin/UsersPage'))
const AdminPlansPage = React.lazy(() => import('@/pages/admin/PlansPage'))
const AdminOrdersPage = React.lazy(() => import('@/pages/admin/OrdersPage'))
const AdminNodesPage = React.lazy(() => import('@/pages/admin/NodesPage'))
const AdminSettingsPage = React.lazy(() => import('@/pages/admin/SettingsPage'))
const AdminNoticesPage = React.lazy(() => import('@/pages/admin/NoticesPage'))
const AdminKnowledgesPage = React.lazy(() => import('@/pages/admin/KnowledgesPage'))
const AdminMailTemplatesPage = React.lazy(() => import('@/pages/admin/MailTemplatesPage'))
const AdminCouponsPage = React.lazy(() => import('@/pages/admin/CouponsPage'))
const AdminGiftCardsPage = React.lazy(() => import('@/pages/admin/GiftCardsPage'))
const AdminInviteCodesPage = React.lazy(() => import('@/pages/admin/InviteCodesPage'))
const AdminPluginsPage = React.lazy(() => import('@/pages/admin/PluginsPage'))
const AdminAuditLogsPage = React.lazy(() => import('@/pages/admin/AuditLogsPage'))
const AdminTicketsPage = React.lazy(() => import('@/pages/admin/TicketsPage'))
const AdminServerGroupsPage = React.lazy(() => import('@/pages/admin/ServerGroupsPage'))
const AdminServerRoutesPage = React.lazy(() => import('@/pages/admin/ServerRoutesPage'))
const AdminServerMachinesPage = React.lazy(() => import('@/pages/admin/ServerMachinesPage'))
const AdminPaymentGatewaysPage = React.lazy(() => import('@/pages/admin/PaymentGatewaysPage'))
const AdminTrafficResetLogsPage = React.lazy(() => import('@/pages/admin/TrafficResetLogsPage'))

// 加载中组件
const Loading = () => <div className="flex items-center justify-center h-screen">加载中...</div>

export const router = createBrowserRouter([
  { path: '/', element: <Navigate to="/user/dashboard" replace /> },
  { 
    path: '/user/login', 
    element: (
      <Suspense fallback={<Loading />}>
        <UserLoginPage />
      </Suspense>
    ) 
  },
  { 
    path: '/user/register', 
    element: (
      <Suspense fallback={<Loading />}>
        <UserRegisterPage />
      </Suspense>
    ) 
  },
  { 
    path: '/admin/login', 
    element: (
      <Suspense fallback={<Loading />}>
        <AdminLoginPage />
      </Suspense>
    ) 
  },
  {
    path: '/user',
    element: (
      <ProtectedRoute>
        <Suspense fallback={<Loading />}>
          <UserLayout />
        </Suspense>
      </ProtectedRoute>
    ),
    children: [
      { index: true, element: <Navigate to="/user/dashboard" replace /> },
      { 
        path: 'dashboard', 
        element: (
          <Suspense fallback={<Loading />}>
            <UserDashboardPage />
          </Suspense>
        ) 
      },
      { 
        path: 'plans', 
        element: (
          <Suspense fallback={<Loading />}>
            <UserPlansPage />
          </Suspense>
        ) 
      },
      { 
        path: 'orders', 
        element: (
          <Suspense fallback={<Loading />}>
            <UserOrdersPage />
          </Suspense>
        ) 
      },
      { 
        path: 'subscribe', 
        element: (
          <Suspense fallback={<Loading />}>
            <UserSubscribePage />
          </Suspense>
        ) 
      },
      { 
        path: 'traffic', 
        element: (
          <Suspense fallback={<Loading />}>
            <UserTrafficPage />
          </Suspense>
        ) 
      },
      { 
        path: 'tickets', 
        element: (
          <Suspense fallback={<Loading />}>
            <UserTicketsPage />
          </Suspense>
        ) 
      },
      { 
        path: 'profile', 
        element: (
          <Suspense fallback={<Loading />}>
            <UserProfilePage />
          </Suspense>
        ) 
      },
      { 
        path: 'coupons', 
        element: (
          <Suspense fallback={<Loading />}>
            <UserCouponsPage />
          </Suspense>
        ) 
      },
      { 
        path: 'gift-cards', 
        element: (
          <Suspense fallback={<Loading />}>
            <UserGiftCardsPage />
          </Suspense>
        ) 
      },
      { 
        path: 'invite', 
        element: (
          <Suspense fallback={<Loading />}>
            <UserInvitePage />
          </Suspense>
        ) 
      },
      { 
        path: 'knowledges', 
        element: (
          <Suspense fallback={<Loading />}>
            <UserKnowledgesPage />
          </Suspense>
        ) 
      },
      { 
        path: 'notices', 
        element: (
          <Suspense fallback={<Loading />}>
            <UserNoticesPage />
          </Suspense>
        ) 
      },
    ],
  },
  {
    path: '/admin',
    element: (
      <ProtectedRoute role="admin">
        <Suspense fallback={<Loading />}>
          <AdminLayout />
        </Suspense>
      </ProtectedRoute>
    ),
    children: [
      { index: true, element: <Navigate to="/admin/dashboard" replace /> },
      { 
        path: 'dashboard', 
        element: (
          <Suspense fallback={<Loading />}>
            <AdminDashboardPage />
          </Suspense>
        ) 
      },
      { 
        path: 'users', 
        element: (
          <Suspense fallback={<Loading />}>
            <AdminUsersPage />
          </Suspense>
        ) 
      },
      { 
        path: 'plans', 
        element: (
          <Suspense fallback={<Loading />}>
            <AdminPlansPage />
          </Suspense>
        ) 
      },
      { 
        path: 'orders', 
        element: (
          <Suspense fallback={<Loading />}>
            <AdminOrdersPage />
          </Suspense>
        ) 
      },
      { 
        path: 'nodes', 
        element: (
          <Suspense fallback={<Loading />}>
            <AdminNodesPage />
          </Suspense>
        ) 
      },
      { 
        path: 'settings', 
        element: (
          <Suspense fallback={<Loading />}>
            <AdminSettingsPage />
          </Suspense>
        ) 
      },
      { 
        path: 'notices', 
        element: (
          <Suspense fallback={<Loading />}>
            <AdminNoticesPage />
          </Suspense>
        ) 
      },
      { 
        path: 'knowledges', 
        element: (
          <Suspense fallback={<Loading />}>
            <AdminKnowledgesPage />
          </Suspense>
        ) 
      },
      { 
        path: 'mail-templates', 
        element: (
          <Suspense fallback={<Loading />}>
            <AdminMailTemplatesPage />
          </Suspense>
        ) 
      },
      { 
        path: 'coupons', 
        element: (
          <Suspense fallback={<Loading />}>
            <AdminCouponsPage />
          </Suspense>
        ) 
      },
      { 
        path: 'gift-cards', 
        element: (
          <Suspense fallback={<Loading />}>
            <AdminGiftCardsPage />
          </Suspense>
        ) 
      },
      { 
        path: 'invite-codes', 
        element: (
          <Suspense fallback={<Loading />}>
            <AdminInviteCodesPage />
          </Suspense>
        ) 
      },
      { 
        path: 'plugins', 
        element: (
          <Suspense fallback={<Loading />}>
            <AdminPluginsPage />
          </Suspense>
        ) 
      },
      {
        path: 'audit-logs',
        element: (
          <Suspense fallback={<Loading />}>
            <AdminAuditLogsPage />
          </Suspense>
        )
      },
      {
        path: 'tickets',
        element: (
          <Suspense fallback={<Loading />}>
            <AdminTicketsPage />
          </Suspense>
        )
      },
      {
        path: 'server-groups',
        element: (
          <Suspense fallback={<Loading />}>
            <AdminServerGroupsPage />
          </Suspense>
        )
      },
      {
        path: 'server-routes',
        element: (
          <Suspense fallback={<Loading />}>
            <AdminServerRoutesPage />
          </Suspense>
        )
      },
      {
        path: 'server-machines',
        element: (
          <Suspense fallback={<Loading />}>
            <AdminServerMachinesPage />
          </Suspense>
        )
      },
      {
        path: 'payment-gateways',
        element: (
          <Suspense fallback={<Loading />}>
            <AdminPaymentGatewaysPage />
          </Suspense>
        )
      },
      {
        path: 'traffic-reset-logs',
        element: (
          <Suspense fallback={<Loading />}>
            <AdminTrafficResetLogsPage />
          </Suspense>
        )
      },
    ],
  },
  { path: '*', element: <Navigate to="/user/dashboard" replace /> },
])
