import { useQuery } from '@tanstack/react-query';
import { fetchKPIs, fetchOverdue, fetchRecentActivity, fetchUpcoming } from '#/lib/api/dashboard';
import { queryKeys } from '#/lib/hooks/queryKeys';

export const useDashboardKPIs = () => {
  return useQuery({
    queryKey: queryKeys.dashboard.kpis,
    queryFn: () => fetchKPIs(),
  });
};

export const useDashboardOverdue = () => {
  return useQuery({
    queryKey: queryKeys.dashboard.overdue,
    queryFn: () => fetchOverdue(),
  });
};

export const useDashboardUpcoming = (windowDays?: number) => {
  return useQuery({
    queryKey: queryKeys.dashboard.upcoming(windowDays),
    queryFn: () => fetchUpcoming(windowDays),
  });
};

export const useDashboardRecentActivity = () => {
  return useQuery({
    queryKey: queryKeys.dashboard.activity,
    queryFn: () => fetchRecentActivity(),
  });
};
