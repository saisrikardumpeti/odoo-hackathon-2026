import axios from 'axios';

export interface UtilizationItem {
  asset_id: string;
  asset_tag: string;
  asset_name: string;
  category_name: string;
  allocation_count: number;
  booking_count: number;
  total_activity: number;
  last_activity: string | null;
  days_idle: number | null;
}

export interface UtilizationResponse {
  utilization: UtilizationItem[];
}

export interface MaintenanceFrequencyItem {
  asset_id: string;
  asset_tag: string;
  asset_name: string;
  category_name: string;
  count: number;
}

export interface MaintenanceCategoryItem {
  category_name: string;
  count: number;
}

export interface MaintenanceFrequencyResponse {
  by_asset: MaintenanceFrequencyItem[];
  by_category: MaintenanceCategoryItem[];
}

export interface RetirementWatchlistItem {
  asset_id: string;
  asset_tag: string;
  asset_name: string;
  category_name: string;
  acquisition_date: string | null;
  age_years: number | null;
  status: string;
}

export interface RetirementWatchlistResponse {
  retirement_watchlist: RetirementWatchlistItem[];
}

export interface AllocationSummaryItem {
  department_name: string;
  department_id: string;
  asset_count: number;
}

export interface AllocationSummaryResponse {
  allocation_summary: AllocationSummaryItem[];
}

export interface BookingHeatmapItem {
  day_of_week: number;
  hour: number;
  count: number;
}

export interface BookingHeatmapResponse {
  heatmap: BookingHeatmapItem[];
}

export interface ReportFilters {
  from?: string;
  to?: string;
  idle_days?: number;
  age_years?: number;
}

export const fetchUtilizationReport = async (filters?: ReportFilters): Promise<UtilizationResponse> => {
  const { data } = await axios.get('/api/v1/reports/utilization', { params: filters });
  return data;
};

export const fetchMaintenanceFrequency = async (filters?: ReportFilters): Promise<MaintenanceFrequencyResponse> => {
  const { data } = await axios.get('/api/v1/reports/maintenance-frequency', { params: filters });
  return data;
};

export const fetchRetirementWatchlist = async (filters?: ReportFilters): Promise<RetirementWatchlistResponse> => {
  const { data } = await axios.get('/api/v1/reports/retirement-watchlist', { params: filters });
  return data;
};

export const fetchAllocationSummary = async (): Promise<AllocationSummaryResponse> => {
  const { data } = await axios.get('/api/v1/reports/allocation-summary');
  return data;
};

export const fetchBookingHeatmap = async (filters?: ReportFilters): Promise<BookingHeatmapResponse> => {
  const { data } = await axios.get('/api/v1/reports/booking-heatmap', { params: filters });
  return data;
};

export const getExportUrl = (reportType: string, filters?: ReportFilters): string => {
  const params = new URLSearchParams();
  params.set('type', reportType);
  if (filters?.from) params.set('from', filters.from);
  if (filters?.to) params.set('to', filters.to);
  if (filters?.idle_days) params.set('idle_days', String(filters.idle_days));
  if (filters?.age_years) params.set('age_years', String(filters.age_years));
  return `/api/v1/reports/export?${params.toString()}`;
};
