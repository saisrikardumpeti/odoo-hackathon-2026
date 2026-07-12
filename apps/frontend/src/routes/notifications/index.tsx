import { createFileRoute, redirect, useNavigate } from '@tanstack/react-router';
import { useAuthStore } from '#/lib/stores/authStore';
import { useNotifications, useMarkNotificationRead, useMarkAllNotificationsRead } from '#/lib/hooks/useNotifications';
import { useState, useCallback } from 'react';
import { CheckCheck, ChevronLeft, ChevronRight, Bell, Package, Wrench, CalendarCheck, ArrowLeftRight, ClipboardCheck, AlertTriangle, Calendar } from 'lucide-react';
import { Button } from '#/components/ui/button';
import { cn } from '#/lib/utils';
import type { Notification } from '#/lib/api/notifications';

export const Route = createFileRoute('/notifications/')({
  beforeLoad: () => {
    if (!useAuthStore.getState().isAuthenticated) {
      throw redirect({ to: '/auth/login' });
    }
  },
  component: NotificationsPage,
});

const typeConfig: Record<string, { icon: typeof Bell; color: string; bg: string }> = {
  AssetAssigned: { icon: Package, color: 'text-blue-600', bg: 'bg-blue-100' },
  MaintenanceApproved: { icon: Wrench, color: 'text-green-600', bg: 'bg-green-100' },
  MaintenanceRejected: { icon: Wrench, color: 'text-red-600', bg: 'bg-red-100' },
  BookingConfirmed: { icon: CalendarCheck, color: 'text-emerald-600', bg: 'bg-emerald-100' },
  BookingCancelled: { icon: CalendarCheck, color: 'text-orange-600', bg: 'bg-orange-100' },
  BookingReminder: { icon: Calendar, color: 'text-purple-600', bg: 'bg-purple-100' },
  TransferApproved: { icon: ArrowLeftRight, color: 'text-indigo-600', bg: 'bg-indigo-100' },
  OverdueReturnAlert: { icon: AlertTriangle, color: 'text-red-600', bg: 'bg-red-100' },
  AuditDiscrepancyFlagged: { icon: ClipboardCheck, color: 'text-amber-600', bg: 'bg-amber-100' },
};

function getTypeConfig(type: string) {
  return typeConfig[type] ?? { icon: Bell, color: 'text-muted-foreground', bg: 'bg-muted' };
}

function getEntityLink(notif: Notification): string {
  if (!notif.related_entity_type || !notif.related_entity_id) return '/notifications';
  const eid = notif.related_entity_id;
  switch (notif.related_entity_type) {
    case 'asset': return `/assets/${eid}`;
    case 'booking': return '/resource-booking';
    case 'maintenance': return '/maintenance';
    case 'allocation': return '/allocation-transfer';
    case 'transfer': return '/allocation-transfer';
    case 'audit_cycle': return `/audit/cycle/${eid}`;
    case 'audit': return `/audit/cycle/${eid}`;
    default: return '/notifications';
  }
}

function timeGroup(dateStr: string): string {
  const now = new Date();
  const date = new Date(dateStr);
  const diffMs = now.getTime() - date.getTime();
  const diffDays = Math.floor(diffMs / 86_400_000);

  if (diffDays === 0) return 'Today';
  if (diffDays === 1) return 'Yesterday';
  if (diffDays < 7) return 'This Week';
  if (diffDays < 30) return 'This Month';
  return 'Older';
}

function groupByTime(notifs: Notification[]): [string, Notification[]][] {
  const groups = new Map<string, Notification[]>();
  for (const n of notifs) {
    const key = timeGroup(n.created_at);
    const arr = groups.get(key) ?? [];
    arr.push(n);
    groups.set(key, arr);
  }
  const order = ['Today', 'Yesterday', 'This Week', 'This Month', 'Older'];
  return order.filter((g) => groups.has(g)).map((g) => [g, groups.get(g)!]);
}

function NotificationSkeleton() {
  return (
    <div className="space-y-3">
      {Array.from({ length: 5 }).map((_, i) => (
        <div key={i} className="animate-pulse rounded-lg border p-4">
          <div className="flex items-start gap-3">
            <div className="size-8 rounded-full bg-muted" />
            <div className="flex-1 space-y-2">
              <div className="h-3 w-20 rounded bg-muted" />
              <div className="h-4 w-3/4 rounded bg-muted" />
              <div className="h-3 w-32 rounded bg-muted" />
            </div>
          </div>
        </div>
      ))}
    </div>
  );
}

function EmptyState() {
  return (
    <div className="flex flex-col items-center justify-center py-16 text-center">
      <div className="mb-4 flex size-16 items-center justify-center rounded-full bg-muted">
        <Bell className="size-8 text-muted-foreground/50" />
      </div>
      <h3 className="mb-1 text-lg font-semibold">All caught up!</h3>
      <p className="text-sm text-muted-foreground">You have no new notifications right now.</p>
    </div>
  );
}

function NotificationsPage() {
  const [page, setPage] = useState(1);
  const [activeFilter, setActiveFilter] = useState<string | null>(null);
  const { data, isLoading, isError, isFetching } = useNotifications({
    is_read: activeFilter === 'unread' ? 'false' : undefined,
    page,
    page_size: 20,
  });
  const markRead = useMarkNotificationRead();
  const markAllRead = useMarkAllNotificationsRead();

  const navigate = useNavigate();
  const handleMarkRead = useCallback((id: string) => {
    markRead.mutate(id);
  }, [markRead]);

  const handleNotificationClick = useCallback((notif: Notification) => {
    if (!notif.is_read) markRead.mutate(notif.id);
    const link = getEntityLink(notif);
    if (link !== '/notifications') navigate({ to: link as '/' | '/notifications' | '/resource-booking' | '/maintenance' | '/allocation-transfer' });
  }, [markRead, navigate]);

  const filters = [
    { key: null, label: 'All' },
    { key: 'unread', label: 'Unread' },
    ...Object.keys(typeConfig).map((key) => ({ key, label: key.replace(/([A-Z])/g, ' $1').trim() })),
  ];

  if (isError) {
    return (
      <div className="flex items-center justify-center p-16 text-destructive">
        <AlertTriangle className="mr-2 size-5" />
        Failed to load notifications.
      </div>
    );
  }

  const totalPages = data ? Math.ceil(data.total / data.page_size) : 0;
  const groups = data?.notifications ? groupByTime(data.notifications) : [];

  return (
    <div className="mx-auto max-w-5xl p-4 sm:p-6">
      <div className="mb-6 flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h1 className="text-2xl font-bold">Notifications</h1>
          {data && (
            <p className="mt-0.5 text-sm text-muted-foreground">
              {data.unread_count} unread of {data.total} total
            </p>
          )}
        </div>
        <div className="flex items-center gap-2">
          {data && data.unread_count > 0 && (
            <Button
              variant="outline"
              size="sm"
              onClick={() => markAllRead.mutate()}
              disabled={markAllRead.isPending}
            >
              <CheckCheck className="mr-1 size-4" />
              Mark all read
            </Button>
          )}
        </div>
      </div>

      <div className="mb-4 flex flex-wrap gap-1">
        {filters.map((f) => (
          <button
            key={f.key ?? 'all'}
            onClick={() => { setActiveFilter(f.key); setPage(1); }}
            className={cn(
              'whitespace-nowrap rounded-full px-3 py-1 text-xs font-medium transition-colors',
              activeFilter === f.key
                ? 'bg-primary text-primary-foreground'
                : 'bg-muted text-muted-foreground hover:bg-accent hover:text-accent-foreground'
            )}
          >
            {f.label}
          </button>
        ))}
      </div>

      {isLoading ? (
        <NotificationSkeleton />
      ) : !data?.notifications?.length ? (
        <EmptyState />
      ) : (
        <div className={cn('space-y-6', isFetching && 'opacity-60 transition-opacity')}>
          {groups.map(([group, notifs]) => (
            <section key={group}>
              <h2 className="mb-3 text-xs font-semibold uppercase tracking-wider text-muted-foreground">
                {group}
              </h2>
              <div className="space-y-2">
                {notifs.map((notif) => {
                  const cfg = getTypeConfig(notif.type);
                  const Icon = cfg.icon;

                  return (
                    <button
                      key={notif.id}
                      onClick={() => handleNotificationClick(notif)}
                      className={cn(
                        'group w-full rounded-lg border p-4 text-left transition-all',
                        !notif.is_read
                          ? 'border-primary/20 bg-accent/30 shadow-sm'
                          : 'hover:bg-accent/30'
                      )}
                    >
                      <div className="flex items-start gap-3">
                        <div className={cn('flex size-8 shrink-0 items-center justify-center rounded-full', cfg.bg)}>
                          <Icon className={cn('size-4', cfg.color)} />
                        </div>

                        <div className="min-w-0 flex-1">
                          <div className="flex items-start justify-between gap-2">
                            <div>
                              <span className={cn(
                                'inline-block rounded-full px-2 py-0.5 text-[10px] font-semibold uppercase tracking-wide',
                                cfg.bg, cfg.color
                              )}>
                                {notif.type.replace(/([A-Z])/g, ' $1').trim()}
                              </span>
                              <p className="mt-1 text-sm">{notif.message}</p>
                            </div>
                            {!notif.is_read && (
                              <span className="mt-1.5 size-2 shrink-0 rounded-full bg-primary" />
                            )}
                          </div>
                          <div className="mt-2 flex items-center gap-3">
                            <span className="text-xs text-muted-foreground">
                              {new Date(notif.created_at).toLocaleString()}
                            </span>
                          </div>
                        </div>
                      </div>
                    </button>
                  );
                })}
              </div>
            </section>
          ))}
        </div>
      )}

      {totalPages > 1 && (
        <div className="mt-8 flex items-center justify-center gap-2">
          <Button
            variant="outline"
            size="sm"
            disabled={page <= 1}
            onClick={() => setPage(page - 1)}
          >
            <ChevronLeft className="size-4" />
          </Button>
          {Array.from({ length: totalPages }, (_, i) => i + 1)
            .filter((p) => p === 1 || p === totalPages || Math.abs(p - page) <= 1)
            .map((p, idx, arr) => (
              <span key={p} className="flex items-center">
                {idx > 0 && arr[idx - 1] !== p - 1 ? (
                  <span className="px-1 text-muted-foreground">...</span>
                ) : null}
                <button
                  onClick={() => setPage(p)}
                  className={cn(
                    'flex size-8 items-center justify-center rounded-md text-sm transition-colors',
                    p === page
                      ? 'bg-primary text-primary-foreground'
                      : 'text-muted-foreground hover:bg-accent hover:text-accent-foreground'
                  )}
                >
                  {p}
                </button>
              </span>
            ))}
          <Button
            variant="outline"
            size="sm"
            disabled={page >= totalPages}
            onClick={() => setPage(page + 1)}
          >
            <ChevronRight className="size-4" />
          </Button>
        </div>
      )}
    </div>
  );
}
