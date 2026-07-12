import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import {
  createCycle,
  listCycles,
  getCycle,
  assignAuditors,
  listItems,
  patchItem,
  closeCycle,
  listDiscrepancyReports,
  resolveDiscrepancy,
} from '#/lib/api/audit';
import type {
  CreateCycleRequest,
  AssignAuditorsRequest,
  PatchItemRequest,
  ListDiscrepancyReportsParams,
} from '#/lib/api/audit';
import { queryKeys } from './queryKeys';

export const useAuditCycles = () => {
  return useQuery({
    queryKey: queryKeys.audit.cycles.all,
    queryFn: listCycles,
  });
};

export const useAuditCycle = (id: string) => {
  return useQuery({
    queryKey: queryKeys.audit.cycles.detail(id),
    queryFn: () => getCycle(id),
    enabled: !!id,
  });
};

export const useCreateAuditCycle = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (req: CreateCycleRequest) => createCycle(req),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.audit.cycles.all });
    },
  });
};

export const useAssignAuditors = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ cycleId, req }: { cycleId: string; req: AssignAuditorsRequest }) =>
      assignAuditors(cycleId, req),
    onSuccess: (_data, variables) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.audit.cycles.detail(variables.cycleId) });
    },
  });
};

export const useAuditItems = (cycleId: string, myItems?: boolean) => {
  return useQuery({
    queryKey: queryKeys.audit.items(cycleId, myItems),
    queryFn: () => listItems(cycleId, myItems),
    enabled: !!cycleId,
  });
};

export const usePatchAuditItem = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ itemId, req }: { itemId: string; req: PatchItemRequest }) =>
      patchItem(itemId, req),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.audit.cycles.all });
      queryClient.invalidateQueries({ queryKey: queryKeys.audit.discrepancies.all });
    },
  });
};

export const useCloseAuditCycle = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (cycleId: string) => closeCycle(cycleId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.audit.cycles.all });
      queryClient.invalidateQueries({ queryKey: queryKeys.assets.all });
    },
  });
};

export const useDiscrepancyReports = (params?: ListDiscrepancyReportsParams) => {
  return useQuery({
    queryKey: queryKeys.audit.discrepancies.filtered(params as Record<string, string> | undefined),
    queryFn: () => listDiscrepancyReports(params),
  });
};

export const useResolveDiscrepancy = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => resolveDiscrepancy(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.audit.discrepancies.all });
    },
  });
};
