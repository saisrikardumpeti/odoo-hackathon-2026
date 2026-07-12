import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarGroup,
  SidebarGroupContent,
  SidebarHeader,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  useSidebar,
} from '@/components/ui/sidebar'
import { Link, useNavigate } from '@tanstack/react-router'
import { useAuthStore } from '#/lib/stores/authStore'
import { Button } from '@/components/ui/button'
import {
  LayoutDashboard,
  Building2,
  Package,
  ArrowLeftRight,
  CalendarCheck,
  Wrench,
  ClipboardCheck,
  FileBarChart,
  Bell,
  LogOut,
  PanelLeftClose,
  PanelLeft,
} from 'lucide-react'

const navItems = [
  { label: 'Dashboard', href: '/', icon: LayoutDashboard },
  { label: 'Organization Setup', href: '/organization-setup', icon: Building2 },
  { label: 'Assets', href: '/assets', icon: Package },
  { label: 'Allocation & Transfer', href: '/allocation-transfer', icon: ArrowLeftRight },
  { label: 'Resource Booking', href: '/resource-booking', icon: CalendarCheck },
  { label: 'Maintenance', href: '/maintenance', icon: Wrench },
  { label: 'Audit', href: '/audit', icon: ClipboardCheck },
  { label: 'Reports', href: '/reports', icon: FileBarChart },
  { label: 'Notifications', href: '/notifications', icon: Bell },
]

function AppSidebar() {
  const navigate = useNavigate()
  const { employee, logout } = useAuthStore()
  const { state, toggleSidebar } = useSidebar()

  const handleLogout = () => {
    logout()
    navigate({ to: '/auth/login' })
  }

  return (
    <Sidebar collapsible="icon" variant="sidebar">
      <SidebarHeader className="border-b border-sidebar-border p-4">
        <div className="flex items-center justify-between">
          <Link to="/" className="text-lg font-bold truncate">
            {state === 'collapsed' ? 'AF' : 'AssetFlow'}
          </Link>
          <button
            onClick={toggleSidebar}
            className="rounded-md p-1 hover:bg-sidebar-accent hover:text-sidebar-accent-foreground"
          >
            {state === 'collapsed' ? <PanelLeft className="size-4" /> : <PanelLeftClose className="size-4" />}
          </button>
        </div>
      </SidebarHeader>

      <SidebarContent>
        <SidebarGroup>
          <SidebarGroupContent>
            <SidebarMenu>
              {navItems.map((item) => (
                <SidebarMenuItem key={item.href}>
                  <SidebarMenuButton
                    render={<Link to={item.href} />}
                    tooltip={item.label}
                  >
                    <item.icon />
                    <span>{item.label}</span>
                  </SidebarMenuButton>
                </SidebarMenuItem>
              ))}
            </SidebarMenu>
          </SidebarGroupContent>
        </SidebarGroup>
      </SidebarContent>

      <SidebarFooter className="border-t border-sidebar-border p-4">
        <div className="flex flex-col gap-2">
          <div className="flex items-center gap-2 truncate">
            <div className="flex size-7 shrink-0 items-center justify-center rounded-full bg-primary text-xs font-medium text-primary-foreground">
              {employee?.name?.charAt(0) ?? '?'}
            </div>
            <div className="flex flex-col truncate group-data-[collapsible=icon]:hidden">
              <span className="text-sm font-medium truncate">{employee?.name}</span>
              <span className="text-xs text-sidebar-foreground/60 truncate">{employee?.role}</span>
            </div>
          </div>
          <Button variant="ghost" size="sm" className="justify-start gap-2 group-data-[collapsible=icon]:hidden" onClick={handleLogout}>
            <LogOut className="size-4" />
            Sign out
          </Button>
        </div>
      </SidebarFooter>
    </Sidebar>
  )
}

export { AppSidebar }
