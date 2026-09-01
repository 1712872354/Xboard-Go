import { useState, useEffect } from 'react'
import { Outlet, NavLink, useLocation, useNavigate } from 'react-router-dom'
import {
  LayoutDashboard, Users, Package, ShoppingCart, Server, Settings,
  Tag, Gift, Bell, BookOpen, Mail, Puzzle, Shield, UserPlus,
  Smartphone, FileText, Send, AtSign, Globe, CreditCard,
  LogOut, Menu, Sun, Moon, ChevronDown, MessageSquare,
  Search, X, ChevronRight, HardDrive, Route,
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
import { Input } from '@/components/ui/input'
import { cn } from '@/lib/utils'

interface NavItem {
  to: string
  icon: React.ComponentType<{ className?: string }>
  label: string
}

interface NavGroup {
  label: string
  icon: React.ComponentType<{ className?: string }>
  items: NavItem[]
}

const navGroups: NavGroup[] = [
  {
    label: '仪表盘',
    icon: LayoutDashboard,
    items: [
      { to: '/admin/dashboard', icon: LayoutDashboard, label: '仪表盘' },
    ],
  },
  {
    label: '系统管理',
    icon: Settings,
    items: [
      { to: '/admin/settings', icon: Settings, label: '系统配置' },
      { to: '/admin/plugins', icon: Puzzle, label: '插件管理' },
      { to: '/admin/notices', icon: Bell, label: '公告管理' },
      { to: '/admin/payment-gateways', icon: CreditCard, label: '支付配置' },
      { to: '/admin/knowledges', icon: BookOpen, label: '知识库管理' },
      { to: '/admin/mail-templates', icon: Mail, label: '邮件模板' },
    ],
  },
  {
    label: '节点管理',
    icon: Server,
    items: [
      { to: '/admin/server-machines', icon: HardDrive, label: '服务器管理' },
      { to: '/admin/nodes', icon: Server, label: '节点管理' },
      { to: '/admin/server-groups', icon: Users, label: '权限组管理' },
      { to: '/admin/server-routes', icon: Route, label: '路由管理' },
    ],
  },
  {
    label: '订阅管理',
    icon: Package,
    items: [
      { to: '/admin/plans', icon: Package, label: '套餐管理' },
      { to: '/admin/orders', icon: ShoppingCart, label: '订单管理' },
      { to: '/admin/coupons', icon: Tag, label: '优惠券管理' },
      { to: '/admin/gift-cards', icon: Gift, label: '礼品卡管理' },
    ],
  },
  {
    label: '用户管理',
    icon: Users,
    items: [
      { to: '/admin/users', icon: Users, label: '用户管理' },
      { to: '/admin/tickets', icon: MessageSquare, label: '工单管理' },
      { to: '/admin/traffic-reset-logs', icon: FileText, label: '流量重置日志' },
      { to: '/admin/audit-logs', icon: FileText, label: '审计日志' },
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

export default function AdminLayout() {
  const [collapsed, setCollapsed] = useState(false)
  const [open, setOpen] = useState(false)
  const [searchOpen, setSearchOpen] = useState(false)
  const isMobile = useIsMobile()
  const { theme, toggleTheme } = useThemeStore()
  const { user } = useAuthStore()
  const logout = useLogout()
  const location = useLocation()
  const navigate = useNavigate()

  // 键盘快捷键 Ctrl+K 打开搜索
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if ((e.metaKey || e.ctrlKey) && e.key === 'k') {
        e.preventDefault()
        setSearchOpen(true)
      }
      if (e.key === 'Escape') {
        setSearchOpen(false)
      }
    }
    document.addEventListener('keydown', handleKeyDown)
    return () => document.removeEventListener('keydown', handleKeyDown)
  }, [])

  // 移动端自动关闭侧边栏
  useEffect(() => {
    if (isMobile) {
      setOpen(false)
    }
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
        {/* Header */}
        <header className="flex h-14 flex-none items-center justify-between border-b bg-background px-4 md:px-6">
          <div className="flex items-center gap-2">
            <Button
              variant="ghost"
              size="icon"
              className="h-8 w-8"
              onClick={() => {
                if (isMobile) {
                  setOpen(true)
                } else {
                  setCollapsed(!collapsed)
                }
              }}
            >
              <Menu className="h-4 w-4" />
            </Button>

            {/* 面包屑导航 */}
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
            {/* 搜索按钮 */}
            <Button
              variant="ghost"
              size="sm"
              className="hidden h-8 gap-2 text-muted-foreground md:flex"
              onClick={() => setSearchOpen(true)}
            >
              <Search className="h-3.5 w-3.5" />
              <span className="text-xs">搜索菜单...</span>
              <kbd className="pointer-events-none ml-2 hidden h-5 select-none items-center gap-1 rounded border bg-muted px-1.5 font-mono text-[10px] font-medium opacity-100 sm:flex">
                <span className="text-xs">⌘</span>K
              </kbd>
            </Button>

            {/* 主题切换 */}
            <Button variant="ghost" size="icon" className="h-8 w-8" onClick={toggleTheme}>
              {theme === 'dark' ? <Sun className="h-4 w-4" /> : <Moon className="h-4 w-4" />}
            </Button>

            {/* 用户菜单 */}
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <Button variant="ghost" className="relative h-8 w-8 rounded-full">
                  <Avatar className="h-8 w-8">
                    <AvatarFallback>{user?.email?.[0]?.toUpperCase() || 'A'}</AvatarFallback>
                  </Avatar>
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end" className="w-56">
                <DropdownMenuLabel className="font-normal">
                  <div className="flex flex-col space-y-1">
                    <p className="text-sm font-medium leading-none">{user?.email}</p>
                    <p className="text-xs leading-none text-muted-foreground">管理员</p>
                  </div>
                </DropdownMenuLabel>
                <DropdownMenuSeparator />
                <DropdownMenuItem onClick={() => navigate('/admin/settings')}>
                  <Settings className="mr-2 h-4 w-4" />
                  系统设置
                </DropdownMenuItem>
                <DropdownMenuSeparator />
                <DropdownMenuItem onClick={logout}>
                  <LogOut className="mr-2 h-4 w-4" />
                  退出登录
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
          </div>
        </header>

        {/* Content area */}
        <div className="flex-1 overflow-y-auto">
          <div className="px-4 py-6 md:px-6">
            <Outlet />
          </div>
        </div>
      </main>

      {/* Search dialog */}
      {searchOpen && (
        <div className="fixed inset-0 z-50 flex items-start justify-center pt-[20vh]">
          <div className="fixed inset-0 bg-background/80 backdrop-blur-sm" onClick={() => setSearchOpen(false)} />
          <div className="relative w-full max-w-md rounded-lg border bg-background p-4 shadow-lg">
            <div className="flex items-center gap-2">
              <Search className="h-4 w-4 text-muted-foreground" />
              <Input
                placeholder="搜索菜单和功能..."
                className="border-0 bg-transparent shadow-none focus-visible:ring-0"
                autoFocus
              />
              <Button variant="ghost" size="icon" className="h-6 w-6" onClick={() => setSearchOpen(false)}>
                <X className="h-4 w-4" />
              </Button>
            </div>
            <div className="mt-4 space-y-2">
              {navGroups.flatMap((group) =>
                group.items.map((item) => (
                  <NavLink
                    key={item.to}
                    to={item.to}
                    onClick={() => setSearchOpen(false)}
                    className="flex items-center gap-3 rounded-md px-3 py-2 text-sm hover:bg-accent"
                  >
                    <item.icon className="h-4 w-4 text-muted-foreground" />
                    <span>{item.label}</span>
                    <span className="ml-auto text-xs text-muted-foreground">{group.label}</span>
                  </NavLink>
                ))
              )}
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
