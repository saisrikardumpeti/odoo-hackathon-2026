import { client } from './client';

export interface AssetCategory {
  id: string;
  name: string;
  custom_fields: Record<string, unknown>;
  created_at: string;
  updated_at: string;
}

export interface ListCategoriesResponse {
  categories: AssetCategory[];
}

export interface CreateCategoryRequest {
  name: string;
  custom_fields?: Record<string, unknown>;
}

export interface UpdateCategoryRequest {
  name: string;
  custom_fields?: Record<string, unknown>;
}

export const listCategories = async (): Promise<ListCategoriesResponse> => {
  const { data } = await client.get('/v1/categories');
  return data;
};

export const createCategory = async (req: CreateCategoryRequest): Promise<{ category: AssetCategory }> => {
  const { data } = await client.post('/v1/categories', req);
  return data;
};

export const updateCategory = async (id: string, req: UpdateCategoryRequest): Promise<{ category: AssetCategory }> => {
  const { data } = await client.patch(`/v1/categories/${id}`, req);
  return data;
};