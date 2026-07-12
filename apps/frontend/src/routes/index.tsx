import { createFileRoute, redirect, useNavigate } from '@tanstack/react-router';
import {
  AlertTriangle,
  CalendarClock,
  CalendarX,
  ClipboardList,
  History,
  Package,
  Plus,
  UserCheck,
  Wrench,
} from 'lucide-react';
import { KpiCard } from '#/components/dashboard/kpi-card';
import { useDashboardKPIs, useDashboardOverdue, useDashboardRecentActivity, useDashboardUpcoming } from '#/lib/hooks/useDashboard';
import { useAuthStore } from '#/lib/stores/authStore';
import { cn } from '#/lib/utils';

export const Route = createFileRoute('/')({
  beforeLoad: () => {
    if (!useAuthStore.getState().isAuthenticated) {
      throw redirect({ to: '/auth/login' })
    }
  },
  component: Home,
})

function formatActivityAction(action: string): string {
  const parts = action.split('.')
  return parts.map((p) => p.charAt(0).toUpperCase() + p.slice(1)).join(' ')
}

function formatActivityTime(createdAt: string): string {
  const date = new Date(createdAt)
  const now = new Date()
  const diffMs = now.getTime() - date.getTime()
  const diffMins = Math.floor(diffMs / 60000)
  if (diffMins < 1) return 'Just now'
  if (diffMins < 60) return `${diffMins}m ago`
  const diffHours = Math.floor(diffMins / 60)
  if (diffHours < 24) return `${diffHours}h ago`
  const diffDays = Math.floor(diffHours / 24)
  if (diffDays < 7) return `${diffDays}d ago`
  return date.toLocaleDateString()
}

function Home() {
  const employee = useAuthStore((state) => state.employee)
  const navigate = useNavigate()

  const { data: kpisData, isLoading: kpisLoading } = useDashboardKPIs()
  const { data: overdueData, isLoading: overdueLoading } = useDashboardOverdue()
  const { data: upcomingData, isLoading: upcomingLoading } = useDashboardUpcoming()
  const { data: activityData, isLoading: activityLoading } = useDashboardRecentActivity()

  const kpis = kpisData?.kpis

  return (
    <div className="p-6 space-y-8">
      <div>
        <h1 className="text-2xl font-bold">Dashboard</h1>
        <p className="text-muted-foreground text-sm">Welcome back, {employee?.name}</p>
      </div>

      <section>
        <h2 className="mb-4 text-lg font-semibold">Overview</h2>
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-6">
          <KpiCard
            icon={Package}
            label="Assets Available"
            value={kpisLoading ? '...' : (kpis?.assets_available ?? 0)}
          />
          <KpiCard
            icon={UserCheck}
            label="Assets Allocated"
            value={kpisLoading ? '...' : (kpis?.assets_allocated ?? 0)}
          />
          <KpiCard
            icon={Wrench}
            label="Maintenance Today"
            value={kpisLoading ? '...' : (kpis?.maintenance_today ?? 0)}
          />
          <KpiCard
            icon={CalendarClock}
            label="Active Bookings"
            value={kpisLoading ? '...' : (kpis?.active_bookings ?? 0)}
          />
          <KpiCard
            icon={ClipboardList}
            label="Pending Transfers"
            value={kpisLoading ? '...' : (kpis?.pending_transfers ?? 0)}
          />
          <KpiCard
            icon={CalendarX}
            label="Upcoming Returns"
            value={kpisLoading ? '...' : (kpis?.upcoming_returns ?? 0)}
          />
        </div>
      </section>

      <section>
        <h2 className="mb-4 text-lg font-semibold flex items-center gap-2 text-red-600">
          <AlertTriangle className="size-5" />
          Overdue Returns
        </h2>
        {overdueLoading ? (
          <p className="text-sm text-muted-foreground">Loading...</p>
        ) : overdueData?.overdue && overdueData.overdue.length > 0 ? (
          <div className="overflow-x-auto rounded-lg border border-red-200">
            <table className="w-full text-sm">
              <thead>
                <tr className="bg-red-50 text-left text-red-700">
                  <th className="p-3 font-medium">Asset</th>
                  <th className="p-3 font-medium">Employee</th>
                  <th className="p-3 font-medium">Expected Return</th>
                  <th className="p-3 font-medium">Days Overdue</th>
                </tr>
              </thead>
              <tbody>
                {overdueData.overdue.map((item) => (
                  <tr key={item.id} className="border-t border-red-100">
                    <td className="p-3">
                      <span className="font-medium">{item.asset_tag}</span>
                      <span className="ml-2 text-muted-foreground">{item.asset_name}</span>
                    </td>
                    <td className="p-3">{item.employee_name ?? '—'}</td>
                    <td className="p-3">{item.expected_return_date ?? '—'}</td>
                    <td className="p-3">
                      <span className="inline-flex items-center rounded-full bg-red-100 px-2 py-0.5 text-xs font-medium text-red-700">
                        {item.days_overdue} day{item.days_overdue !== 1 ? 's' : ''}
                      </span>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        ) : (
          <p className="text-sm text-muted-foreground">No overdue returns.</p>
        )}
      </section>

      <section>
        <h2 className="mb-4 text-lg font-semibold flex items-center gap-2">
          <CalendarClock className="size-5 text-blue-600" />
          Upcoming Returns
        </h2>
        {upcomingLoading ? (
          <p className="text-sm text-muted-foreground">Loading...</p>
        ) : upcomingData?.upcoming && upcomingData.upcoming.length > 0 ? (
          <div className="overflow-x-auto rounded-lg border">
            <table className="w-full text-sm">
              <thead>
                <tr className="bg-muted text-left text-muted-foreground">
                  <th className="p-3 font-medium">Asset</th>
                  <th className="p-3 font-medium">Employee</th>
                  <th className="p-3 font-medium">Expected Return</th>
                  <th className="p-3 font-medium">Days Until Due</th>
                </tr>
              </thead>
              <tbody>
                {upcomingData.upcoming.map((item) => (
                  <tr key={item.id} className="border-t">
                    <td className="p-3">
                      <span className="font-medium">{item.asset_tag}</span>
                      <span className="ml-2 text-muted-foreground">{item.asset_name}</span>
                    </td>
                    <td className="p-3">{item.employee_name ?? '—'}</td>
                    <td className="p-3">{item.expected_date}</td>
                    <td className="p-3">
                      <span className={cn(
                        'inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium',
                        item.days_until_due <= 2
                          ? 'bg-amber-100 text-amber-700'
                          : 'bg-blue-100 text-blue-700',
                      )}>
                        {item.days_until_due} day{item.days_until_due !== 1 ? 's' : ''}
                      </span>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        ) : (
          <p className="text-sm text-muted-foreground">No upcoming returns.</p>
        )}
      </section>

      <section>
        <h2 className="mb-4 text-lg font-semibold">Quick Actions</h2>
        <div className="flex flex-wrap gap-3">
          <button
            onClick={() => navigate({ to: '/assets/new' })}
            className="inline-flex items-center gap-2 rounded-lg border bg-primary px-4 py-2 text-sm font-medium text-primary-foreground hover:bg-primary/90"
          >
            <Plus className="size-4" />
            Register Asset
          </button>
          <button
            onClick={() => navigate({ to: '/resource-booking' })}
            className="inline-flex items-center gap-2 rounded-lg border bg-primary px-4 py-2 text-sm font-medium text-primary-foreground hover:bg-primary/90"
          >
            <CalendarClock className="size-4" />
            Book Resource
          </button>
          <button
            onClick={() => navigate({ to: '/maintenance/new' })}
            className="inline-flex items-center gap-2 rounded-lg border bg-primary px-4 py-2 text-sm font-medium text-primary-foreground hover:bg-primary/90"
          >
            <Wrench className="size-4" />
            Raise Maintenance Request
          </button>
        </div>
      </section>

      <section>
        <h2 className="mb-4 text-lg font-semibold flex items-center gap-2">
          <History className="size-5 text-muted-foreground" />
          Recent Activity
        </h2>
        {activityLoading ? (
          <p className="text-sm text-muted-foreground">Loading...</p>
        ) : activityData?.activity && activityData.activity.length > 0 ? (
          <div className="overflow-x-auto rounded-lg border">
            <table className="w-full text-sm">
              <thead>
                <tr className="bg-muted text-left text-muted-foreground">
                  <th className="p-3 font-medium">Action</th>
                  <th className="p-3 font-medium">Type</th>
                  <th className="p-3 font-medium">Actor</th>
                  <th className="p-3 font-medium">When</th>
                </tr>
              </thead>
              <tbody>
                {activityData.activity.map((item) => (
                  <tr key={item.id} className="border-t">
                    <td className="p-3 font-medium">{formatActivityAction(item.action)}</td>
                    <td className="p-3 text-muted-foreground">{item.entity_type}</td>
                    <td className="p-3">{item.actor_name ?? 'System'}</td>
                    <td className="p-3 text-muted-foreground">{formatActivityTime(item.created_at)}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        ) : (
          <p className="text-sm text-muted-foreground">No recent activity.</p>
        )}
      </section>
    </div>
  )
}