import { useState, useRef, useEffect } from 'react';
import { Bell, CheckCheck, Package, Wrench, CalendarCheck, ArrowLeftRight, ClipboardCheck, AlertTriangle, Calendar } from 'lucide-react';
import { useUnreadCount, useNotifications, useMarkNotificationRead, useMarkAllNotificationsRead } from '#/lib/hooks/useNotifications';
import { useAuthStore } from '#/lib/stores/authStore';
import { useNavigate } from '@tanstack/react-router';
import { cn } from '#/lib/utils';
import type { Notification } from '#/lib/api/notifications';

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
  if (!notif.related_entity_type || !notif.related_entity_id) return '';
  const eid = notif.related_entity_id;
  switch (notif.related_entity_type) {
    case 'asset': return `/assets/${eid}`;
    case 'booking': return '/resource-booking';
    case 'maintenance': return '/maintenance';
    case 'allocation': return '/allocation-transfer';
    case 'transfer': return '/allocation-transfer';
    case 'audit_cycle': return `/audit/cycle/${eid}`;
    case 'audit': return `/audit/cycle/${eid}`;
    default: return '';
  }
}

function timeAgo(dateStr: string): string {
  const now = new Date();
  const date = new Date(dateStr);
  const diffMs = now.getTime() - date.getTime();
  const mins = Math.floor(diffMs / 60_000);
  if (mins < 1) return 'just now';
  if (mins < 60) return `${mins}m ago`;
  const hours = Math.floor(mins / 60);
  if (hours < 24) return `${hours}h ago`;
  const days = Math.floor(hours / 24);
  if (days === 1) return 'yesterday';
  return `${days}d ago`;
}

function NotificationBell() {
  const [isOpen, setIsOpen] = useState(false);
  const dropdownRef = useRef<HTMLDivElement>(null);
  const navigate = useNavigate();
  const { data: unreadData } = useUnreadCount();
  const { data: notifData } = useNotifications({ is_read: 'false', page_size: 5 });
  const markRead = useMarkNotificationRead();
  const markAllRead = useMarkAllNotificationsRead();

  const isAuthenticated = useAuthStore((state) => state.isAuthenticated);

  useEffect(() => {
    function handleClickOutside(event: MouseEvent) {
      if (dropdownRef.current && !dropdownRef.current.contains(event.target as Node)) {
        setIsOpen(false);
      }
    }
    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, []);

  if (!isAuthenticated) return null;

  const unreadCount = unreadData?.unread_count ?? 0;

  return (
    <div className="relative" ref={dropdownRef}>
      <button
        onClick={() => setIsOpen(!isOpen)}
        className={cn(
          'relative rounded-md p-2 transition-colors hover:bg-sidebar-accent hover:text-sidebar-accent-foreground',
          isOpen && 'bg-sidebar-accent text-sidebar-accent-foreground'
        )}
        aria-label={`Notifications (${unreadCount} unread)`}
      >
        <Bell className="size-5" />
        {unreadCount > 0 && (
          <span className="absolute -right-0.5 -top-0.5 flex size-4.5 items-center justify-center rounded-full bg-red-500 text-[9px] font-bold text-white ring-2 ring-sidebar">
            {unreadCount > 9 ? '9+' : unreadCount}
          </span>
        )}
      </button>

      {isOpen && (
        <div className="absolute right-0 top-full z-50 mt-2 w-80 rounded-lg border bg-popover shadow-lg">
          <div className="flex items-center justify-between border-b px-3 py-2.5">
            <span className="text-sm font-semibold">Notifications</span>
            {unreadCount > 0 && (
              <button
                onClick={() => markAllRead.mutate()}
                className="text-xs font-medium text-primary hover:underline"
              >
                Mark all read
              </button>
            )}
          </div>

          <div className="max-h-80 overflow-y-auto">
            {!notifData?.notifications?.length ? (
              <div className="flex flex-col items-center py-8 text-center">
                <Bell className="mb-2 size-6 text-muted-foreground/40" />
                <p className="text-sm text-muted-foreground">No new notifications</p>
              </div>
            ) : (
              notifData.notifications.map((notif) => {
                const cfg = getTypeConfig(notif.type);
                const Icon = cfg.icon;
                const link = getEntityLink(notif);

                return (
                  <div
                    key={notif.id}
                    className={cn(
                      'relative border-b last:border-b-0 transition-colors hover:bg-accent/50',
                      !notif.is_read && 'bg-accent/20'
                    )}
                  >
                    <button
                      onClick={() => {
                        markRead.mutate(notif.id);
                        if (link) navigate({ to: link as any });
                        setIsOpen(false);
                      }}
                      className="flex w-full items-start gap-2.5 px-3 py-2.5 text-left"
                    >
                      <div className={cn('flex size-7 shrink-0 items-center justify-center rounded-full', cfg.bg)}>
                        <Icon className={cn('size-3.5', cfg.color)} />
                      </div>
                      <div className="min-w-0 flex-1">
                        <p className="line-clamp-2 text-sm leading-snug">{notif.message}</p>
                        <p className="mt-0.5 text-[11px] text-muted-foreground">{timeAgo(notif.created_at)}</p>
                      </div>
                      {!notif.is_read && (
                        <span className="mt-1.5 size-1.5 shrink-0 rounded-full bg-primary" />
                      )}
                    </button>
                  </div>
                );
              })
            )}
          </div>

          <div className="border-t px-2 py-2">
            <button
              onClick={() => { setIsOpen(false); navigate({ to: '/notifications' }); }}
              className="w-full rounded-md px-2 py-1.5 text-center text-xs font-medium text-primary transition-colors hover:bg-accent"
            >
              View all notifications
            </button>
          </div>
        </div>
      )}
    </div>
  );
}

export { NotificationBell };
