import { Outlet, createRootRoute } from '@tanstack/react-router'
import { TanStackRouterDevtoolsPanel } from '@tanstack/react-router-devtools'
import { TanStackDevtools } from '@tanstack/react-devtools'
import { ReactQueryDevtoolsPanel } from '@tanstack/react-query-devtools'
import '../index.css'
import { QueryClientProvider } from '@tanstack/react-query'
import { queryClient } from '#/lib/react-query'
import { AppSidebar } from '#/components/app-sidebar'
import { SidebarInset, SidebarProvider, SidebarTrigger } from '#/components/ui/sidebar'
import { useAuthStore } from '#/lib/stores/authStore'


export const Route = createRootRoute({
  component: RootComponent,
})

function RootComponent() {
  const isAuthenticated = useAuthStore((state) => state.isAuthenticated)

  return (
    <>
      <QueryClientProvider client={queryClient}>
        <SidebarProvider>
          {isAuthenticated && <AppSidebar />}
          <SidebarInset>
            {isAuthenticated && (
              <header className="flex h-12 md:hidden items-center gap-2 border-b px-4">
                <SidebarTrigger />
              </header>
            )}
            <Outlet />
          </SidebarInset>
          <TanStackDevtools
            plugins={[
              {
                name: 'TanStack Router',
                render: <TanStackRouterDevtoolsPanel />,
              },
              {
                name: 'TanStack React Query',
                render: <ReactQueryDevtoolsPanel />,
              }
            ]}
          />
        </SidebarProvider>
      </QueryClientProvider>
    </>
  )
}
