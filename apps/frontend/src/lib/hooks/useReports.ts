import { useQuery } from '@tanstack/react-query';
import {
  fetchUtilizationReport,
  fetchMaintenanceFrequency,
  fetchRetirementWatchlist,
  fetchAllocationSummary,
  fetchBookingHeatmap,
} from '#/lib/api/reports';
import { queryKeys } from '#/lib/hooks/queryKeys';
import type { ReportFilters } from '#/lib/api/reports';

export const useUtilizationReport = (filters?: ReportFilters) => {
  return useQuery({
    queryKey: queryKeys.reports.utilization(filters as Record<string, unknown> | undefined),
    queryFn: () => fetchUtilizationReport(filters),
  });
};

export const useMaintenanceFrequency = (filters?: ReportFilters) => {
  return useQuery({
    queryKey: queryKeys.reports.maintenance(filters as Record<string, unknown> | undefined),
    queryFn: () => fetchMaintenanceFrequency(filters),
  });
};

export const useRetirementWatchlist = (filters?: ReportFilters) => {
  return useQuery({
    queryKey: queryKeys.reports.retirement(filters as Record<string, unknown> | undefined),
    queryFn: () => fetchRetirementWatchlist(filters),
  });
};

export const useAllocationSummary = () => {
  return useQuery({
    queryKey: queryKeys.reports.allocationSummary,
    queryFn: () => fetchAllocationSummary(),
  });
};

export const useBookingHeatmap = (filters?: ReportFilters) => {
  return useQuery({
    queryKey: queryKeys.reports.bookingHeatmap(filters as Record<string, unknown> | undefined),
    queryFn: () => fetchBookingHeatmap(filters),
  });
};
