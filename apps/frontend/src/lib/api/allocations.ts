import { client } from './client';

export interface Allocation {
  id: string;
  asset_id: string;
  employee_id: string | null;
  department_id: string | null;
  allocated_by: string;
  allocated_at: string;
  expected_return_date: string | null;
  returned_at: string | null;
  return_condition_notes: string | null;
  status: 'Active' | 'Returned' | 'Overdue';
  created_at: string;
  updated_at: string;
}

export interface AllocationDetail extends Allocation {
  asset_tag: string;
  asset_name: string;
  employee_name: string | null;
  department_name: string | null;
  allocated_by_name: string;
}

export interface TransferRequest {
  id: string;
  asset_id: string;
  allocation_id: string;
  from_employee_id: string | null;
  to_employee_id: string;
  requested_by: string;
  status: 'Requested' | 'Approved' | 'Rejected';
  approved_by: string | null;
  approved_at: string | null;
  created_at: string;
  updated_at: string;
}

export interface TransferRequestDetail extends TransferRequest {
  asset_tag: string;
  asset_name: string;
  from_employee_name: string | null;
  to_employee_name: string;
  requested_by_name: string;
}

export interface CreateAllocationRequest {
  asset_id: string;
  employee_id?: string | null;
  department_id?: string | null;
  expected_return_date?: string | null;
}

export interface CreateTransferRequest {
  allocation_id: string;
  to_employee_id: string;
}

export interface ReturnAllocationRequest {
  return_condition_notes?: string | null;
}

export interface ConflictError {
  error: 'AlreadyAllocated';
  message: string;
  current_holder: AllocationDetail;
}

export const createAllocation = async (req: CreateAllocationRequest): Promise<{ allocation: Allocation }> => {
  const { data } = await client.post('/v1/allocations', req);
  return data;
};

export const returnAllocation = async (id: string, req: ReturnAllocationRequest): Promise<{ message: string }> => {
  const { data } = await client.post(`/v1/allocations/${id}/return`, req);
  return data;
};

export const listOverdueAllocations = async (): Promise<{ allocations: AllocationDetail[] }> => {
  const { data } = await client.get('/v1/allocations/overdue');
  return data;
};

export const listMyAllocations = async (): Promise<{ allocations: AllocationDetail[] }> => {
  const { data } = await client.get('/v1/allocations/my');
  return data;
};

export const createTransfer = async (req: CreateTransferRequest): Promise<{ transfer: TransferRequest }> => {
  const { data } = await client.post('/v1/transfers', req);
  return data;
};

export const approveTransfer = async (id: string): Promise<{ message: string; new_allocation_id: string }> => {
  const { data } = await client.patch(`/v1/transfers/${id}/approve`);
  return data;
};

export const rejectTransfer = async (id: string): Promise<{ message: string }> => {
  const { data } = await client.patch(`/v1/transfers/${id}/reject`);
  return data;
};

export const listPendingTransfers = async (): Promise<{ transfers: TransferRequestDetail[] }> => {
  const { data } = await client.get('/v1/transfers/pending');
  return data;
};
