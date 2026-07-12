import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import {
  createMaintenance,
  listMaintenance,
  approveMaintenance,
  rejectMaintenance,
  assignTechnician,
  startMaintenance,
  resolveMaintenance,
} from '#/lib/api/maintenance';
import type { CreateMaintenanceRequest, ListMaintenanceParams } from '#/lib/api/maintenance';
import { queryKeys } from './queryKeys';

export const useMaintenanceList = (params?: ListMaintenanceParams) => {
  return useQuery({
    queryKey: queryKeys.maintenance.list(params as Record<string, string> | undefined),
    queryFn: () => listMaintenance(params),
  });
};

export const useCreateMaintenance = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (req: CreateMaintenanceRequest) => createMaintenance(req),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.maintenance.all });
      queryClient.invalidateQueries({ queryKey: queryKeys.assets.all });
    },
  });
};

export const useApproveMaintenance = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => approveMaintenance(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.maintenance.all });
      queryClient.invalidateQueries({ queryKey: queryKeys.assets.all });
    },
  });
};

export const useRejectMaintenance = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, reason }: { id: string; reason?: string }) => rejectMaintenance(id, reason),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.maintenance.all });
    },
  });
};

export const useAssignTechnician = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, technicianName }: { id: string; technicianName: string }) =>
      assignTechnician(id, technicianName),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.maintenance.all });
    },
  });
};

export const useStartMaintenance = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => startMaintenance(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.maintenance.all });
    },
  });
};

export const useResolveMaintenance = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, resolutionNotes }: { id: string; resolutionNotes: string }) =>
      resolveMaintenance(id, resolutionNotes),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.maintenance.all });
      queryClient.invalidateQueries({ queryKey: queryKeys.assets.all });
    },
  });
};
