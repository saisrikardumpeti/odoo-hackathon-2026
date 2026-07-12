import { client } from './client';

export interface ActivityLog {
  id: string;
  actor_employee_id: string | null;
  actor_name: string | null;
  action: string;
  entity_type: string;
  entity_id: string | null;
  metadata: Record<string, unknown>;
  created_at: string;
}

export interface ActivityLogListResult {
  logs: ActivityLog[];
  total: number;
  page: number;
  page_size: number;
}

export interface ActivityLogFilters {
  actor?: string;
  action?: string;
  entity_type?: string;
  entity_id?: string;
  date_from?: string;
  date_to?: string;
  page?: number;
  page_size?: number;
}

export const listActivityLogs = async (
  params?: ActivityLogFilters
): Promise<ActivityLogListResult> => {
  const { data } = await client.get('/v1/activity-logs', { params });
  return data;
};
