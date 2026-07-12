import { client } from './client';
import type { Employee } from './auth';

export interface AuditCycle {
  id: string;
  name: string;
  scope_department_id: string | null;
  scope_location: string | null;
  start_date: string;
  end_date: string;
  status: 'Draft' | 'Active' | 'Closed';
  created_by: string;
  closed_at: string | null;
  created_at: string;
  updated_at: string;
}

export interface AuditCycleDetail extends AuditCycle {
  scope_department_name: string | null;
  created_by_name: string | null;
  assigned_auditors: Employee[];
  item_count: number;
  verified_count: number;
  missing_count: number;
  damaged_count: number;
}

export interface AuditItem {
  id: string;
  audit_cycle_id: string;
  asset_id: string;
  auditor_id: string | null;
  result: 'Verified' | 'Missing' | 'Damaged' | null;
  notes: string | null;
  verified_at: string | null;
  created_at: string;
  updated_at: string;
  asset_tag: string;
  asset_name: string;
  asset_status: string;
  asset_location: string | null;
}

export interface DiscrepancyReport {
  id: string;
  audit_cycle_id: string;
  asset_id: string;
  audit_item_id: string;
  issue_type: 'Missing' | 'Damaged';
  resolved: boolean;
  resolved_by: string | null;
  resolved_at: string | null;
  created_at: string;
  updated_at: string;
  cycle_name: string;
  asset_tag: string;
  asset_name: string;
  resolved_by_name: string | null;
}

export interface CreateCycleRequest {
  name: string;
  scope_department_id?: string | null;
  scope_location?: string | null;
  start_date: string;
  end_date: string;
}

export interface CreateCycleResponse {
  audit_cycle: AuditCycle;
}

export interface ListCyclesResponse {
  audit_cycles: AuditCycleDetail[];
}

export interface GetCycleResponse {
  audit_cycle: AuditCycleDetail;
}

export interface ListItemsResponse {
  items: AuditItem[];
}

export interface AssignAuditorsRequest {
  employee_ids: string[];
}

export interface PatchItemRequest {
  result?: 'Verified' | 'Missing' | 'Damaged' | null;
  notes?: string | null;
}

export interface ListDiscrepancyReportsParams {
  cycle_id?: string;
  resolved?: string;
}

export interface ListDiscrepancyReportsResponse {
  discrepancy_reports: DiscrepancyReport[];
}

export const createCycle = async (req: CreateCycleRequest): Promise<CreateCycleResponse> => {
  const { data } = await client.post('/v1/audit-cycles', req);
  return data;
};

export const listCycles = async (): Promise<ListCyclesResponse> => {
  const { data } = await client.get('/v1/audit-cycles');
  return data;
};

export const getCycle = async (id: string): Promise<GetCycleResponse> => {
  const { data } = await client.get(`/v1/audit-cycles/${id}`);
  return data;
};

export const assignAuditors = async (cycleId: string, req: AssignAuditorsRequest): Promise<{ message: string }> => {
  const { data } = await client.post(`/v1/audit-cycles/${cycleId}/auditors`, req);
  return data;
};

export const listItems = async (cycleId: string, myItems?: boolean): Promise<ListItemsResponse> => {
  const params: Record<string, string> = {};
  if (myItems) params.my_items = 'true';
  const { data } = await client.get(`/v1/audit-cycles/${cycleId}/items`, { params });
  return data;
};

export const patchItem = async (itemId: string, req: PatchItemRequest): Promise<{ message: string }> => {
  const { data } = await client.patch(`/v1/audit-items/${itemId}`, req);
  return data;
};

export const closeCycle = async (cycleId: string): Promise<{ message: string }> => {
  const { data } = await client.patch(`/v1/audit-cycles/${cycleId}/close`);
  return data;
};

export const listDiscrepancyReports = async (params?: ListDiscrepancyReportsParams): Promise<ListDiscrepancyReportsResponse> => {
  const { data } = await client.get('/v1/discrepancy-reports', { params });
  return data;
};

export const resolveDiscrepancy = async (id: string): Promise<{ message: string }> => {
  const { data } = await client.patch(`/v1/discrepancy-reports/${id}/resolve`);
  return data;
};
