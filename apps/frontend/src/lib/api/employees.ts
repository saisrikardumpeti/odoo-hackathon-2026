import { client } from './client';
import type { Employee } from './auth';

export interface ListEmployeesResponse {
  employees: Employee[];
}

export interface ListEmployeesParams {
  department_id?: string;
  role?: string;
  status?: string;
}

export interface UpdateEmployeeRequest {
  name?: string;
  department_id?: string | null;
  status?: string;
}

export interface UpdateRoleRequest {
  role: 'DepartmentHead' | 'AssetManager' | 'Employee';
}

export interface UpdateRoleResponse {
  message: string;
  from_role: string;
  to_role: string;
}

export const listEmployees = async (params?: ListEmployeesParams): Promise<ListEmployeesResponse> => {
  const { data } = await client.get('/v1/employees', { params });
  return data;
};

export const updateEmployee = async (id: string, req: UpdateEmployeeRequest): Promise<{ employee: Employee }> => {
  const { data } = await client.patch(`/v1/employees/${id}`, req);
  return data;
};

export const updateEmployeeRole = async (id: string, req: UpdateRoleRequest): Promise<UpdateRoleResponse> => {
  const { data } = await client.patch(`/v1/employees/${id}/role`, req);
  return data;
};