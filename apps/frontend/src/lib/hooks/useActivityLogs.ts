import { useQuery } from '@tanstack/react-query';
import { listActivityLogs } from '#/lib/api/activityLogs';
import type { ActivityLogFilters } from '#/lib/api/activityLogs';
import { queryKeys } from './queryKeys';

export const useActivityLogs = (params?: ActivityLogFilters) => {
  return useQuery({
    queryKey: queryKeys.activityLogs.list(params as Record<string, unknown> | undefined),
    queryFn: () => listActivityLogs(params),
  });
};
