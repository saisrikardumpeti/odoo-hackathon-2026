import { useState, useRef, useEffect } from 'react';
import { Bell } from 'lucide-react';
import { useUnreadCount, useNotifications, useMarkNotificationRead, useMarkAllNotificationsRead } from '#/lib/hooks/useNotifications';
import { useAuthStore } from '#/lib/stores/authStore';
import { Link } from '@tanstack/react-router';
import { cn } from '#/lib/utils';

function NotificationBell() {
  const [isOpen, setIsOpen] = useState(false);
  const dropdownRef = useRef<HTMLDivElement>(null);
  const { data: unreadData } = useUnreadCount();
  const { data: notifData } = useNotifications({ is_read: 'false', page_size: 10 });
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
        className="relative rounded-md p-2 hover:bg-sidebar-accent hover:text-sidebar-accent-foreground transition-colors"
        aria-label="Notifications"
      >
        <Bell className="size-5" />
        {unreadCount > 0 && (
          <span className="absolute -right-0.5 -top-0.5 flex size-4 items-center justify-center rounded-full bg-red-500 text-[9px] font-bold text-white">
            {unreadCount > 9 ? '9+' : unreadCount}
          </span>
        )}
      </button>

      {isOpen && (
        <div className="absolute right-0 top-full z-50 mt-2 w-80 rounded-lg border bg-popover p-2 shadow-lg">
          <div className="flex items-center justify-between border-b px-2 pb-2">
            <span className="text-sm font-semibold">Notifications</span>
            {unreadCount > 0 && (
              <button
                onClick={() => markAllRead.mutate()}
                className="text-xs text-primary hover:underline"
              >
                Mark all read
              </button>
            )}
          </div>

          <div className="max-h-80 overflow-y-auto">
            {!notifData?.notifications?.length ? (
              <p className="py-6 text-center text-sm text-muted-foreground">No new notifications</p>
            ) : (
              notifData.notifications.map((notif) => (
                <button
                  key={notif.id}
                  onClick={() => {
                    markRead.mutate(notif.id);
                  }}
                  className={cn(
                    'w-full rounded-md px-2 py-2.5 text-left text-sm transition-colors hover:bg-accent',
                    !notif.is_read && 'bg-accent/50'
                  )}
                >
                  <p className="text-xs text-muted-foreground">{notif.type}</p>
                  <p className="text-sm">{notif.message}</p>
                  <p className="mt-0.5 text-xs text-muted-foreground">
                    {new Date(notif.created_at).toLocaleString()}
                  </p>
                </button>
              ))
            )}
          </div>

          <div className="border-t pt-2">
            <Link
              to="/notifications"
              onClick={() => setIsOpen(false)}
              className="block rounded-md px-2 py-1.5 text-center text-xs text-primary hover:bg-accent"
            >
              View all notifications
            </Link>
          </div>
        </div>
      )}
    </div>
  );
}

export { NotificationBell };
