import { useState } from 'react'
import { Outlet, NavLink, useLocation } from 'react-router-dom'
import {
  LayoutDashboard, Users, Package, ShoppingCart, Server, Settings,
  Tag, Gift, Hash, Bell, BookOpen, Mail, Puzzle, ClipboardList, Monitor,
  LogOut, Menu, Sun, Moon, ChevronLeft, ChevronDown, MessageSquare,
  Route, HardDrive, Shield, CreditCard, FileText,
} from 'lucide-react'
import { useAuthStore } from '@/stores/auth'
import { useThemeStore } from '@/stores/theme'
import { useLogout } from '@/hooks/useAuth'
import { useIsMobile } from '@/hooks/use-mobile'
import { Button } from '@/components/ui/button'
import { Avatar, AvatarFallback } from '@/components/ui/avatar'
import { Separator } from '@/components/ui/separator'
import { ScrollArea } from '@/components/ui/scroll-area'
import { Collapsible, CollapsibleContent, CollapsibleTrigger } from '@/components/ui/collapsible'
import {
  DropdownMenu, DropdownMenuContent, DropdownMenuItem,
  DropdownMenuLabel, DropdownMenuSeparator, DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { Sheet, SheetContent } from '@/components/ui/sheet'
import { cn } from '@/lib/utils'

const navGroups = [
  {
    label: '总览',
    items: [
      { to: '/admin/dashboard', icon: LayoutDashboard, label: '仪表盘' },
    ],
  },
  {
    label: '业务管理',
    items: [
      { to: '/admin/users', icon: Users, label: '用户管理' },
      { to: '/admin/plans', icon: Package, label: '套餐管理' },
      { to: '/admin/orders', icon: ShoppingCart, label: '订单管理' },
      { to: '/admin/tickets', icon: MessageSquare, label: '工单管理' },
    ],
  },
  {
    label: '节点管理',
    items: [
      { to: '/admin/nodes', icon: Server, label: '节点中心' },
      { to: '/admin/server-groups', icon: Shield, label: '权限组' },
      { to: '/admin/server-routes', icon: Route, label: '路由管理' },
      { to: '/admin/server-machines', icon: HardDrive, label: '服务器' },
    ],
  },
  {
    label: '营销工具',
    items: [
      { to: '/admin/coupons', icon: Tag, label: '优惠券' },
      { to: '/admin/gift-cards', icon: Gift, label: '礼品卡' },
      { to: '/admin/invite-codes', icon: Hash, label: '邀请码' },
    ],
  },
  {
    label: '内容管理',
    items: [
      { to: '/admin/notices', icon: Bell, label: '公告管理' },
      { to: '/admin/knowledges', icon: BookOpen, label: '知识库' },
      { to: '/admin/mail-templates', icon: Mail, label: '邮件模板' },
    ],
  },
  {
    label: '财务',
    items: [
      { to: '/admin/payment-gateways', icon: CreditCard, label: '支付配置' },
      { to: '/admin/traffic-reset-logs', icon: FileText, label: '流量重置日志' },
    ],
  },
  {
    label: '系统',
    items: [
      { to: '/admin/settings', icon: Settings, label: '系统设置' },
      { to: '/admin/plugins', icon: Puzzle, label: '插件管理' },
      { to: '/admin/audit-logs', icon: ClipboardList, label: '审计日志' },
    ],
  },
]

function NavGroup({ group, collapsed }: { group: typeof navGroups[0]; collapsed: boolean }) {
  const location = useLocation()
  const isActive = group.items.some((item) => location.pathname.startsWith(item.to))
  const [open, setOpen] = useState(isActive)

  if (collapsed) {
    return (
      <div className="space-y-1">
        {group.items.map((item) => (
          <NavLink
            key={item.to}
            to={item.to}
            className={({ isActive }) =>
              cn(
                'flex items-center justify-center rounded-lg p-2 text-sm transition-colors',
                isActive
                  ? 'bg-sidebar-accent text-sidebar-accent-foreground'
                  : 'text-sidebar-foreground hover:bg-sidebar-accent/50',
              )
            }
          >
            <item.icon className="h-4 w-4" />
          </NavLink>
        ))}
      </div>
    )
  }

  return (
    <Collapsible open={open} onOpenChange={setOpen}>
      <CollapsibleTrigger className="flex w-full items-center justify-between rounded-lg px-3 py-1.5 text-xs font-semibold text-muted-foreground hover:bg-sidebar-accent/30">
        {group.label}
        <ChevronDown className={cn('h-3 w-3 transition-transform', open && 'rotate-180')} />
      </CollapsibleTrigger>
      <CollapsibleContent className="space-y-1 mt-1">
        {group.items.map((item) => (
          <NavLink
            key={item.to}
            to={item.to}
            className={({ isActive }) =>
              cn(
                'flex items-center gap-3 rounded-lg px-3 py-2 text-sm font-medium transition-colors',
                isActive
                  ? 'bg-sidebar-accent text-sidebar-accent-foreground'
                  : 'text-sidebar-foreground hover:bg-sidebar-accent/50',
              )
            }
          >
            <item.icon className="h-4 w-4 shrink-0" />
            <span>{item.label}</span>
          </NavLink>
        ))}
      </CollapsibleContent>
    </Collapsible>
  )
}

function SidebarContent({ collapsed }: { collapsed: boolean }) {
  return (
    <div className="flex h-full flex-col">
      <div className="flex h-14 items-center px-4">
        {!collapsed && <span className="text-lg font-semibold">XBoard Admin</span>}
      </div>
      <Separator />
      <ScrollArea className="flex-1 py-2">
        <nav className="space-y-2 px-2">
          {navGroups.map((group) => (
            <NavGroup key={group.label} group={group} collapsed={collapsed} />
          ))}
        </nav>
      </ScrollArea>
    </div>
  )
}

export default function AdminLayout() {
  const [collapsed, setCollapsed] = useState(false)
  const [open, setOpen] = useState(false)
  const isMobile = useIsMobile()
  const { theme, toggleTheme } = useThemeStore()
  const { user } = useAuthStore()
  const logout = useLogout()

  return (
    <div className="flex h-screen overflow-hidden">
      <aside
        className={cn(
          'hidden lg:flex lg:flex-col border-r bg-sidebar transition-all duration-300',
          collapsed ? 'w-[68px]' : 'w-64',
        )}
      >
        <SidebarContent collapsed={collapsed} />
        <Separator />
        <div className="p-2">
          <Button variant="ghost" size="sm" className="w-full justify-center" onClick={() => setCollapsed(!collapsed)}>
            <ChevronLeft className={cn('h-4 w-4 transition-transform', collapsed && 'rotate-180')} />
          </Button>
        </div>
      </aside>

      {isMobile && (
        <Sheet open={open} onOpenChange={setOpen}>
          <SheetContent side="left" className="w-64 p-0 bg-sidebar">
            <SidebarContent collapsed={false} />
          </SheetContent>
        </Sheet>
      )}

      <div className="flex flex-1 flex-col overflow-hidden">
        <header className="flex h-14 items-center justify-between border-b px-4 bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60">
          <div className="flex items-center gap-2">
            {isMobile && (
              <Button variant="ghost" size="icon" onClick={() => setOpen(true)}>
                <Menu className="h-5 w-5" />
              </Button>
            )}
          </div>
          <div className="flex items-center gap-2">
            <Button variant="ghost" size="icon" onClick={toggleTheme}>
              {theme === 'dark' ? <Sun className="h-4 w-4" /> : <Moon className="h-4 w-4" />}
            </Button>
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <Button variant="ghost" className="relative h-8 w-8 rounded-full">
                  <Avatar className="h-8 w-8">
                    <AvatarFallback>{user?.email?.[0]?.toUpperCase() || 'A'}</AvatarFallback>
                  </Avatar>
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end">
                <DropdownMenuLabel>{user?.email}</DropdownMenuLabel>
                <DropdownMenuSeparator />
                <DropdownMenuItem onClick={logout}>
                  <LogOut className="mr-2 h-4 w-4" />
                  退出登录
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
          </div>
        </header>
        <main className="flex-1 overflow-y-auto p-4 lg:p-6">
          <Outlet />
        </main>
      </div>
    </div>
  )
}
