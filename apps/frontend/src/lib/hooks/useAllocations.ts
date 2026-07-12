import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import {
  createAllocation,
  returnAllocation,
  listOverdueAllocations,
  listMyAllocations,
  createTransfer,
  approveTransfer,
  rejectTransfer,
  listPendingTransfers,
} from '#/lib/api/allocations';
import type {
  CreateAllocationRequest,
  ReturnAllocationRequest,
  CreateTransferRequest,
} from '#/lib/api/allocations';
import { queryKeys } from './queryKeys';

export const useCreateAllocation = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (req: CreateAllocationRequest) => createAllocation(req),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.allocations.all });
      queryClient.invalidateQueries({ queryKey: queryKeys.assets.all });
    },
  });
};

export const useReturnAllocation = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, req }: { id: string; req: ReturnAllocationRequest }) => returnAllocation(id, req),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.allocations.all });
      queryClient.invalidateQueries({ queryKey: queryKeys.assets.all });
    },
  });
};

export const useOverdueAllocations = () => {
  return useQuery({
    queryKey: queryKeys.allocations.overdue,
    queryFn: listOverdueAllocations,
  });
};

export const useMyAllocations = () => {
  return useQuery({
    queryKey: queryKeys.allocations.my,
    queryFn: listMyAllocations,
  });
};

export const useCreateTransfer = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (req: CreateTransferRequest) => createTransfer(req),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.transfers.all });
    },
  });
};

export const useApproveTransfer = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => approveTransfer(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.transfers.all });
      queryClient.invalidateQueries({ queryKey: queryKeys.allocations.all });
      queryClient.invalidateQueries({ queryKey: queryKeys.assets.all });
    },
  });
};

export const useRejectTransfer = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => rejectTransfer(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.transfers.all });
    },
  });
};

export const usePendingTransfers = () => {
  return useQuery({
    queryKey: queryKeys.transfers.pending,
    queryFn: listPendingTransfers,
  });
};
