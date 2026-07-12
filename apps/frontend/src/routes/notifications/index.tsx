import { createFileRoute, redirect } from '@tanstack/react-router';
import { useAuthStore } from '#/lib/stores/authStore';
import { useNotifications, useMarkNotificationRead, useMarkAllNotificationsRead } from '#/lib/hooks/useNotifications';
import { useState } from 'react';
import { CheckCheck, ChevronLeft, ChevronRight } from 'lucide-react';
import { Button } from '#/components/ui/button';
import { cn } from '#/lib/utils';

export const Route = createFileRoute('/notifications/')({
  beforeLoad: () => {
    if (!useAuthStore.getState().isAuthenticated) {
      throw redirect({ to: '/auth/login' });
    }
  },
  component: NotificationsPage,
});

function NotificationsPage() {
  const [page, setPage] = useState(1);
  const [unreadOnly, setUnreadOnly] = useState(false);
  const { data, isLoading, isError } = useNotifications({
    is_read: unreadOnly ? 'false' : undefined,
    page,
    page_size: 20,
  });
  const markRead = useMarkNotificationRead();
  const markAllRead = useMarkAllNotificationsRead();

  if (isLoading) {
    return (
      <div className="flex items-center justify-center p-8">
        <div className="size-6 animate-spin rounded-full border-2 border-primary border-t-transparent" />
      </div>
    );
  }

  if (isError) {
    return (
      <div className="p-8 text-destructive">Failed to load notifications.</div>
    );
  }

  const totalPages = data ? Math.ceil(data.total / data.page_size) : 0;

  return (
    <div className="mx-auto max-w-3xl p-6">
      <div className="mb-6 flex items-center justify-between">
        <h1 className="text-2xl font-bold">Notifications</h1>
        <div className="flex items-center gap-2">
          <Button
            variant={unreadOnly ? 'default' : 'outline'}
            size="sm"
            onClick={() => { setUnreadOnly(!unreadOnly); setPage(1); }}
          >
            Unread only
          </Button>
          <Button
            variant="outline"
            size="sm"
            onClick={() => markAllRead.mutate()}
            disabled={markAllRead.isPending}
          >
            <CheckCheck className="mr-1 size-4" />
            Mark all read
          </Button>
        </div>
      </div>

      <div className="space-y-2">
        {!data?.notifications?.length ? (
          <p className="py-12 text-center text-muted-foreground">No notifications</p>
        ) : (
          data.notifications.map((notif) => (
            <button
              key={notif.id}
              onClick={() => markRead.mutate(notif.id)}
              className={cn(
                'w-full rounded-lg border p-4 text-left transition-colors hover:bg-accent',
                !notif.is_read && 'border-primary/20 bg-accent/30'
              )}
            >
              <div className="flex items-start justify-between gap-4">
                <div className="min-w-0 flex-1">
                  <span className="inline-block rounded-full bg-primary/10 px-2 py-0.5 text-xs font-medium text-primary">
                    {notif.type}
                  </span>
                  <p className="mt-1 text-sm">{notif.message}</p>
                  <p className="mt-1 text-xs text-muted-foreground">
                    {new Date(notif.created_at).toLocaleString()}
                  </p>
                </div>
                {!notif.is_read && (
                  <span className="mt-1 size-2 shrink-0 rounded-full bg-primary" />
                )}
              </div>
            </button>
          ))
        )}
      </div>

      {totalPages > 1 && (
        <div className="mt-6 flex items-center justify-center gap-2">
          <Button
            variant="outline"
            size="sm"
            disabled={page <= 1}
            onClick={() => setPage(page - 1)}
          >
            <ChevronLeft className="size-4" />
          </Button>
          <span className="text-sm text-muted-foreground">
            Page {page} of {totalPages}
          </span>
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
