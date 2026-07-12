import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import {
  listNotifications,
  markNotificationRead,
  markAllNotificationsRead,
  getUnreadCount,
} from '#/lib/api/notifications';
import { queryKeys } from './queryKeys';

export const useNotifications = (params?: { is_read?: string; page?: number; page_size?: number }) => {
  return useQuery({
    queryKey: queryKeys.notifications.list(params as Record<string, unknown> | undefined),
    queryFn: () => listNotifications(params),
  });
};

export const useUnreadCount = () => {
  return useQuery({
    queryKey: queryKeys.notifications.unreadCount,
    queryFn: getUnreadCount,
    refetchInterval: 30_000,
  });
};

export const useMarkNotificationRead = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => markNotificationRead(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.notifications.all });
    },
  });
};

export const useMarkAllNotificationsRead = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: () => markAllNotificationsRead(),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.notifications.all });
    },
  });
};
