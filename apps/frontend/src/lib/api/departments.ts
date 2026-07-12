import { client } from './client';

export interface Department {
  id: string;
  name: string;
  parent_department_id: string | null;
  head_employee_id: string | null;
  status: 'Active' | 'Inactive';
  created_at: string;
  updated_at: string;
}

export interface ListDepartmentsResponse {
  departments: Department[];
}

export interface CreateDepartmentRequest {
  name: string;
  parent_department_id?: string | null;
  head_employee_id?: string | null;
}

export interface UpdateDepartmentRequest {
  name: string;
  parent_department_id?: string | null;
  head_employee_id?: string | null;
}

export interface DeactivateDepartmentRequest {
  force: boolean;
}

export const listDepartments = async (): Promise<ListDepartmentsResponse> => {
  const { data } = await client.get('/v1/departments');
  return data;
};

export const createDepartment = async (req: CreateDepartmentRequest): Promise<{ department: Department }> => {
  const { data } = await client.post('/v1/departments', req);
  return data;
};

export const updateDepartment = async (id: string, req: UpdateDepartmentRequest): Promise<{ department: Department }> => {
  const { data } = await client.patch(`/v1/departments/${id}`, req);
  return data;
};

export const deactivateDepartment = async (id: string, force: boolean = false): Promise<{ message: string } & Partial<{ error: string; active_employee_count: number; requires_confirmation: boolean }>> => {
  const { data } = await client.patch(`/v1/departments/${id}/deactivate`, { force });
  return data;
};