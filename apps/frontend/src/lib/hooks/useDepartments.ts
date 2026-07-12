import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { listDepartments, createDepartment, updateDepartment, deactivateDepartment } from '#/lib/api/departments';
import type { CreateDepartmentRequest, UpdateDepartmentRequest } from '#/lib/api/departments';
import { queryKeys } from './queryKeys';

export const useDepartments = () => {
  return useQuery({
    queryKey: queryKeys.departments.all,
    queryFn: () => listDepartments(),
  });
};

export const useCreateDepartment = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (req: CreateDepartmentRequest) => createDepartment(req),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.departments.all });
    },
  });
};

export const useUpdateDepartment = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, req }: { id: string; req: UpdateDepartmentRequest }) => updateDepartment(id, req),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.departments.all });
    },
  });
};

export const useDeactivateDepartment = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, force }: { id: string; force?: boolean }) => deactivateDepartment(id, force),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.departments.all });
    },
  });
};
