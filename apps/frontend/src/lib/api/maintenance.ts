import { client } from './client';

export type MaintenanceStatus = 'Pending' | 'Approved' | 'Rejected' | 'TechnicianAssigned' | 'InProgress' | 'Resolved';
export type MaintenancePriority = 'Low' | 'Medium' | 'High' | 'Critical';

export interface MaintenanceRequest {
  id: string;
  asset_id: string;
  raised_by_employee_id: string;
  issue_description: string;
  priority: MaintenancePriority;
  photo_url: string | null;
  status: MaintenanceStatus;
  approved_by: string | null;
  approved_at: string | null;
  technician_name: string | null;
  resolved_at: string | null;
  resolution_notes: string | null;
  created_at: string;
  updated_at: string;
}

export interface MaintenanceDetail extends MaintenanceRequest {
  asset_tag: string;
  asset_name: string;
  raised_by_name: string | null;
  approved_by_name: string | null;
}

export interface ListMaintenanceResponse {
  maintenance_requests: MaintenanceDetail[];
  total_count: number;
}

export interface CreateMaintenanceRequest {
  asset_id: string;
  issue_description: string;
  priority: MaintenancePriority;
  photo_url?: string | null;
}

export interface ListMaintenanceParams {
  asset_id?: string;
  status?: string;
  priority?: string;
}

export const createMaintenance = async (req: CreateMaintenanceRequest): Promise<{ maintenance: MaintenanceRequest }> => {
  const { data } = await client.post('/v1/maintenance', req);
  return data;
};

export const listMaintenance = async (params?: ListMaintenanceParams): Promise<ListMaintenanceResponse> => {
  const { data } = await client.get('/v1/maintenance', { params });
  return data;
};

export const approveMaintenance = async (id: string): Promise<{ message: string }> => {
  const { data } = await client.patch(`/v1/maintenance/${id}/approve`);
  return data;
};

export const rejectMaintenance = async (id: string, reason?: string): Promise<{ message: string }> => {
  const { data } = await client.patch(`/v1/maintenance/${id}/reject`, { reason });
  return data;
};

export const assignTechnician = async (id: string, technicianName: string): Promise<{ message: string }> => {
  const { data } = await client.patch(`/v1/maintenance/${id}/assign-technician`, { technician_name: technicianName });
  return data;
};

export const startMaintenance = async (id: string): Promise<{ message: string }> => {
  const { data } = await client.patch(`/v1/maintenance/${id}/start`);
  return data;
};

export const resolveMaintenance = async (id: string, resolutionNotes: string): Promise<{ message: string }> => {
  const { data } = await client.patch(`/v1/maintenance/${id}/resolve`, { resolution_notes: resolutionNotes });
  return data;
};
