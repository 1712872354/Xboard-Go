import { useState, useEffect } from 'react'
import { Outlet, NavLink, useLocation } from 'react-router-dom'
import {
  LayoutDashboard, Package, ShoppingCart, Link2, BarChart3,
  Ticket, User, Tag, Gift, Users, BookOpen, Bell,
  LogOut, Menu, Sun, Moon, ChevronDown, ChevronRight,
} from 'lucide-react'
import { useAuthStore } from '@/stores/auth'
import { useThemeStore } from '@/stores/theme'
import { useLogout } from '@/hooks/useAuth'
import { useIsMobile } from '@/hooks/use-mobile'
import { Button } from '@/components/ui/button'
import { Avatar, AvatarFallback } from '@/components/ui/avatar'
import { ScrollArea } from '@/components/ui/scroll-area'
import { Collapsible, CollapsibleContent, CollapsibleTrigger } from '@/components/ui/collapsible'
import {
  DropdownMenu, DropdownMenuContent, DropdownMenuItem,
  DropdownMenuLabel, DropdownMenuSeparator, DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { Sheet, SheetContent } from '@/components/ui/sheet'
import { cn } from '@/lib/utils'

interface NavItem {
  to: string
  icon: React.ComponentType<{ className?: string }>
  label: string
}

interface NavGroup {
  label: string
  items: NavItem[]
}

const navGroups: NavGroup[] = [
  {
    label: '概览',
    items: [
      { to: '/user/dashboard', icon: LayoutDashboard, label: '仪表盘' },
    ],
  },
  {
    label: '订阅',
    items: [
      { to: '/user/plans', icon: Package, label: '套餐' },
      { to: '/user/orders', icon: ShoppingCart, label: '订单' },
      { to: '/user/subscribe', icon: Link2, label: '订阅链接' },
    ],
  },
  {
    label: '工具',
    items: [
      { to: '/user/traffic', icon: BarChart3, label: '流量明细' },
      { to: '/user/tickets', icon: Ticket, label: '工单' },
      { to: '/user/coupons', icon: Tag, label: '优惠券' },
      { to: '/user/gift-cards', icon: Gift, label: '礼品卡' },
      { to: '/user/invite', icon: Users, label: '邀请返利' },
    ],
  },
  {
    label: '其他',
    items: [
      { to: '/user/knowledges', icon: BookOpen, label: '知识库' },
      { to: '/user/notices', icon: Bell, label: '公告' },
      { to: '/user/profile', icon: User, label: '个人资料' },
    ],
  },
]

function NavGroup({ group, collapsed }: { group: NavGroup; collapsed: boolean }) {
  const location = useLocation()
  const isActive = group.items.some((item) => location.pathname.startsWith(item.to))
  const [open, setOpen] = useState(isActive)

  if (collapsed) {
    return (
      <div className="px-2 py-1">
        {group.items.map((item) => (
          <NavLink
            key={item.to}
            to={item.to}
            className={({ isActive }) =>
              cn(
                'flex h-10 w-full items-center justify-center rounded-md text-sm transition-colors',
                isActive
                  ? 'bg-sidebar-accent text-sidebar-accent-foreground'
                  : 'text-sidebar-foreground/70 hover:bg-sidebar-accent/50 hover:text-sidebar-foreground',
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
      <CollapsibleTrigger className="flex h-10 w-full items-center justify-between rounded-md px-4 text-xs font-semibold uppercase tracking-wider text-sidebar-foreground/50 hover:bg-sidebar-accent/30 hover:text-sidebar-foreground/70">
        <span>{group.label}</span>
        <ChevronDown className={cn('h-3 w-3 transition-transform', open && 'rotate-180')} />
      </CollapsibleTrigger>
      <CollapsibleContent className="space-y-0.5 px-2">
        {group.items.map((item) => (
          <NavLink
            key={item.to}
            to={item.to}
            className={({ isActive }) =>
              cn(
                'flex h-9 w-full items-center gap-3 rounded-md px-3 text-sm font-medium transition-colors',
                isActive
                  ? 'bg-sidebar-accent text-sidebar-accent-foreground'
                  : 'text-sidebar-foreground/70 hover:bg-sidebar-accent/50 hover:text-sidebar-foreground',
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
      <div className="flex h-14 items-center border-b border-sidebar-border px-4">
        <div className={cn('flex items-center gap-2', collapsed && 'justify-center')}>
          <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 256 256" className="h-6 w-6 text-sidebar-foreground">
            <rect width="256" height="256" fill="none" />
            <line x1="208" y1="128" x2="128" y2="208" fill="none" stroke="currentColor" strokeLinecap="round" strokeLinejoin="round" strokeWidth="16" />
            <line x1="192" y1="40" x2="40" y2="192" fill="none" stroke="currentColor" strokeLinecap="round" strokeLinejoin="round" strokeWidth="16" />
          </svg>
          {!collapsed && <span className="font-semibold text-sidebar-foreground">XBoard</span>}
        </div>
      </div>
      <ScrollArea className="flex-1 py-2">
        <nav className="space-y-1">
          {navGroups.map((group) => (
            <NavGroup key={group.label} group={group} collapsed={collapsed} />
          ))}
        </nav>
      </ScrollArea>
    </div>
  )
}

export default function UserLayout() {
  const [collapsed, setCollapsed] = useState(false)
  const [open, setOpen] = useState(false)
  const isMobile = useIsMobile()
  const { theme, toggleTheme } = useThemeStore()
  const { user } = useAuthStore()
  const logout = useLogout()
  const location = useLocation()

  useEffect(() => {
    if (isMobile) setOpen(false)
  }, [location.pathname, isMobile])

  return (
    <div className="relative h-full overflow-hidden bg-background">
      {/* Desktop sidebar */}
      <aside
        className={cn(
          'fixed left-0 top-0 z-40 hidden h-svh flex-col border-r border-sidebar-border bg-sidebar transition-[width] duration-300 md:flex',
          collapsed ? 'w-14' : 'w-64',
        )}
      >
        <SidebarContent collapsed={collapsed} />
      </aside>

      {/* Mobile sidebar */}
      <Sheet open={open} onOpenChange={setOpen}>
        <SheetContent side="left" className="w-64 p-0 bg-sidebar border-r border-sidebar-border">
          <SidebarContent collapsed={false} />
        </SheetContent>
      </Sheet>

      {/* Main content */}
      <main
        className={cn(
          'flex h-svh flex-col overflow-hidden transition-[margin] duration-300',
          collapsed ? 'md:ml-14' : 'md:ml-64',
        )}
      >
        <header className="flex h-14 flex-none items-center justify-between border-b bg-background px-4 md:px-6">
          <div className="flex items-center gap-2">
            <Button
              variant="ghost"
              size="icon"
              className="h-8 w-8"
              onClick={() => {
                if (isMobile) setOpen(true)
                else setCollapsed(!collapsed)
              }}
            >
              <Menu className="h-4 w-4" />
            </Button>
            <nav className="hidden items-center gap-1 text-sm md:flex">
              {navGroups.map((group) => {
                const activeItem = group.items.find((item) => location.pathname.startsWith(item.to))
                if (!activeItem) return null
                return (
                  <div key={group.label} className="flex items-center gap-1">
                    <ChevronRight className="h-3 w-3 text-muted-foreground" />
                    <span className="text-muted-foreground">{group.label}</span>
                    <ChevronRight className="h-3 w-3 text-muted-foreground" />
                    <span className="font-medium">{activeItem.label}</span>
                  </div>
                )
              })}
            </nav>
          </div>
          <div className="flex items-center gap-2">
            <Button variant="ghost" size="icon" className="h-8 w-8" onClick={toggleTheme}>
              {theme === 'dark' ? <Sun className="h-4 w-4" /> : <Moon className="h-4 w-4" />}
            </Button>
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <Button variant="ghost" className="relative h-8 w-8 rounded-full">
                  <Avatar className="h-8 w-8">
                    <AvatarFallback>{user?.email?.[0]?.toUpperCase() || 'U'}</AvatarFallback>
                  </Avatar>
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end" className="w-56">
                <DropdownMenuLabel className="font-normal">
                  <div className="flex flex-col space-y-1">
                    <p className="text-sm font-medium leading-none">{user?.email}</p>
                  </div>
                </DropdownMenuLabel>
                <DropdownMenuSeparator />
                <DropdownMenuItem asChild>
                  <NavLink to="/user/profile"><User className="mr-2 h-4 w-4" />个人资料</NavLink>
                </DropdownMenuItem>
                <DropdownMenuItem asChild>
                  <NavLink to="/user/orders"><ShoppingCart className="mr-2 h-4 w-4" />我的订单</NavLink>
                </DropdownMenuItem>
                <DropdownMenuSeparator />
                <DropdownMenuItem onClick={logout}>
                  <LogOut className="mr-2 h-4 w-4" />退出登录
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
          </div>
        </header>
        <div className="flex-1 overflow-y-auto">
          <div className="px-4 py-6 md:px-6">
            <Outlet />
          </div>
        </div>
      </main>
    </div>
  )
}
