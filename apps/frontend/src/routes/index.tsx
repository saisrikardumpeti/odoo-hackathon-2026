import { usePing } from '#/lib/hooks/usePing'
import { Link, createFileRoute, redirect } from '@tanstack/react-router'
import { useAuthStore } from '#/lib/stores/authStore'
import { Button } from '#/components/ui/button'

export const Route = createFileRoute('/')({
  beforeLoad: () => {
    if (!useAuthStore.getState().isAuthenticated) {
      throw redirect({ to: '/auth/login' })
    }
  },
  component: Home,
})

function Home() {
  const employee = useAuthStore((state) => state.employee)
  const { data, isLoading, isError } = usePing()

  if (isLoading) {
    return <div className="p-8">Loading...</div>
  }

  return (
    <div className="p-8">
      <div className="mb-8">
        <h1 className="text-2xl font-bold">Dashboard</h1>
        <p className="text-muted-foreground text-sm">Welcome back, {employee?.name}</p>
      </div>
      {isError && <p className="text-destructive">Failed to load data</p>}
      {data && <pre className="text-sm">{JSON.stringify(data, null, 2)}</pre>}
    </div>
  )
}
