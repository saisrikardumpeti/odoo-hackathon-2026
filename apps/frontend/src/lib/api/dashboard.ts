import axios from 'axios';

export interface KPIs {
  assets_available: number;
  assets_allocated: number;
  maintenance_today: number;
  active_bookings: number;
  pending_transfers: number;
  upcoming_returns: number;
}

export interface DashboardOverdueItem {
  type: string;
  id: string;
  asset_id: string;
  asset_tag: string;
  asset_name: string;
  employee_id: string | null;
  employee_name: string | null;
  expected_return_date: string | null;
  days_overdue: number;
}

export interface DashboardUpcomingItem {
  type: string;
  id: string;
  asset_id: string;
  asset_tag: string;
  asset_name: string;
  employee_id: string | null;
  employee_name: string | null;
  expected_date: string;
  days_until_due: number;
}

export interface KPIsResponse {
  kpis: KPIs;
}

export interface OverdueResponse {
  overdue: DashboardOverdueItem[];
}

export interface UpcomingResponse {
  upcoming: DashboardUpcomingItem[];
}

export interface RecentActivityItem {
  id: string;
  action: string;
  entity_type: string;
  actor_name: string | null;
  created_at: string;
}

export interface RecentActivityResponse {
  activity: RecentActivityItem[];
}

export const fetchKPIs = async (): Promise<KPIsResponse> => {
  const { data } = await axios.get('/api/v1/dashboard/kpis');
  return data;
};

export const fetchOverdue = async (): Promise<OverdueResponse> => {
  const { data } = await axios.get('/api/v1/dashboard/overdue');
  return data;
};

export const fetchRecentActivity = async (): Promise<RecentActivityResponse> => {
  const { data } = await axios.get('/api/v1/dashboard/activity');
  return data;
};

export const fetchUpcoming = async (windowDays?: number): Promise<UpcomingResponse> => {
  const params = windowDays ? { window_days: windowDays } : {};
  const { data } = await axios.get('/api/v1/dashboard/upcoming', { params });
  return data;
};
