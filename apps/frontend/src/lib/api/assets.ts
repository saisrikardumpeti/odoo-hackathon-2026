import { client } from './client';

export interface AssetListItem {
  id: string;
  asset_tag: string;
  name: string;
  category_name: string;
  serial_number: string | null;
  status: AssetStatus;
  location: string | null;
  current_holder_department_id: string | null;
  is_bookable: boolean;
}

export interface AssetDetail {
  id: string;
  asset_tag: string;
  name: string;
  category_id: string;
  category_name: string | null;
  serial_number: string | null;
  acquisition_date: string | null;
  acquisition_cost: number | null;
  condition: string | null;
  location: string | null;
  is_bookable: boolean;
  status: AssetStatus;
  current_holder_employee_id: string | null;
  current_holder_department_id: string | null;
  current_holder_name: string | null;
  current_holder_department_name: string | null;
  qr_code: string | null;
  custom_fields: Record<string, unknown>;
  created_at: string;
  updated_at: string;
}

export type AssetStatus = 'Available' | 'Allocated' | 'Reserved' | 'UnderMaintenance' | 'Lost' | 'Retired' | 'Disposed';

export interface HistoryEvent {
  timestamp: string;
  type: 'status_change' | 'allocation' | 'maintenance';
  data: Record<string, unknown>;
}

export interface ListAssetsResponse {
  assets: AssetListItem[];
  total_count: number;
  page: number;
  page_size: number;
}

export interface CreateAssetRequest {
  name: string;
  category_id: string;
  serial_number?: string | null;
  acquisition_date?: string | null;
  acquisition_cost?: number | null;
  condition?: string | null;
  location?: string | null;
  is_bookable: boolean;
  custom_fields?: Record<string, unknown>;
}

export interface ListAssetsParams {
  asset_tag?: string;
  serial_number?: string;
  category_id?: string;
  status?: string;
  department?: string;
  location?: string;
  is_bookable?: string;
  page?: number;
  page_size?: number;
}

export const listAssets = async (params?: ListAssetsParams): Promise<ListAssetsResponse> => {
  const { data } = await client.get('/v1/assets', { params });
  return data;
};

export const getAsset = async (id: string): Promise<{ asset: AssetDetail }> => {
  const { data } = await client.get(`/v1/assets/${id}`);
  return data;
};

export const createAsset = async (req: CreateAssetRequest): Promise<{ asset: AssetDetail }> => {
  const { data } = await client.post('/v1/assets', req);
  return data;
};

export const getAssetHistory = async (id: string): Promise<{ history: HistoryEvent[] }> => {
  const { data } = await client.get(`/v1/assets/${id}/history`);
  return data;
};

export const uploadAssetDocument = async (id: string, file: File, type: 'photo' | 'document'): Promise<{ document: { id: string; url: string; type: string } }> => {
  const formData = new FormData();
  formData.append('file', file);
  formData.append('type', type);
  const { data } = await client.post(`/v1/assets/${id}/documents`, formData, {
    headers: { 'Content-Type': 'multipart/form-data' },
  });
  return data;
};
