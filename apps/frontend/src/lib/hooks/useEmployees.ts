import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { listEmployees, updateEmployee, updateEmployeeRole } from '#/lib/api/employees';
import type { ListEmployeesParams, UpdateEmployeeRequest, UpdateRoleRequest } from '#/lib/api/employees';
import { queryKeys } from './queryKeys';

export const useEmployees = (params?: ListEmployeesParams) => {
  return useQuery({
    queryKey: queryKeys.employees.filtered(params as Record<string, string> | undefined),
    queryFn: () => listEmployees(params),
  });
};

export const useUpdateEmployee = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, req }: { id: string; req: UpdateEmployeeRequest }) => updateEmployee(id, req),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.employees.all });
    },
  });
};

export const useUpdateEmployeeRole = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, req }: { id: string; req: UpdateRoleRequest }) => updateEmployeeRole(id, req),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.employees.all });
    },
  });
};
