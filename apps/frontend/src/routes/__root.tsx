import { Outlet, createRootRoute } from '@tanstack/react-router'
import { TanStackRouterDevtoolsPanel } from '@tanstack/react-router-devtools'
import { TanStackDevtools } from '@tanstack/react-devtools'
import { ReactQueryDevtoolsPanel } from '@tanstack/react-query-devtools'
import '../index.css'
import { QueryClientProvider } from '@tanstack/react-query'
import { queryClient } from '#/lib/react-query'


export const Route = createRootRoute({
  component: RootComponent,
})

function RootComponent() {
  return (
    <>
      <QueryClientProvider client={queryClient}>
        <Outlet />
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
      </QueryClientProvider>
    </>
  )
}
