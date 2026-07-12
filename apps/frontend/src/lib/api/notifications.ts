import { client } from './client';

export interface Notification {
  id: string;
  employee_id: string;
  type: string;
  message: string;
  related_entity_type: string | null;
  related_entity_id: string | null;
  is_read: boolean;
  created_at: string;
}

export interface NotificationListResult {
  notifications: Notification[];
  total: number;
  page: number;
  page_size: number;
  unread_count: number;
}

export interface UnreadCountResult {
  unread_count: number;
}

export const listNotifications = async (
  params?: { is_read?: string; page?: number; page_size?: number }
): Promise<NotificationListResult> => {
  const { data } = await client.get('/v1/notifications', { params });
  return data;
};

export const markNotificationRead = async (id: string): Promise<{ message: string }> => {
  const { data } = await client.patch(`/v1/notifications/${id}/read`);
  return data;
};

export const markAllNotificationsRead = async (): Promise<{ message: string }> => {
  const { data } = await client.patch('/v1/notifications/read-all');
  return data;
};

export const getUnreadCount = async (): Promise<UnreadCountResult> => {
  const { data } = await client.get('/v1/notifications/unread-count');
  return data;
};
