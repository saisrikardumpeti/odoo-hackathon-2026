import { client } from './client';

export interface Employee {
  id: string;
  name: string;
  email: string;
  role: 'Admin' | 'DepartmentHead' | 'AssetManager' | 'Employee';
  department_id: string | null;
  status?: string;
}

export interface SignupResponse {
  message: string;
  employee_id: string;
}

export interface LoginResponse {
  access_token: string;
  refresh_token: string;
  employee: Employee;
}

export interface RefreshResponse {
  access_token: string;
  refresh_token: string;
  employee: Employee;
}

export interface MeResponse {
  employee: Employee;
}

export interface ForgotPasswordResponse {
  message: string;
}

export const signup = async (name: string, email: string, password: string): Promise<SignupResponse> => {
  const { data } = await client.post('/auth/signup', { name, email, password });
  return data;
};

export const login = async (email: string, password: string): Promise<LoginResponse> => {
  const { data } = await client.post('/auth/login', { email, password });
  return data;
};

export const refresh = async (refreshToken: string): Promise<RefreshResponse> => {
  const { data } = await client.post('/auth/refresh', { refresh_token: refreshToken });
  return data;
};

export const forgotPassword = async (email: string): Promise<ForgotPasswordResponse> => {
  const { data } = await client.post('/auth/forgot-password', { email });
  return data;
};

export const getMe = async (): Promise<MeResponse> => {
  const { data } = await client.get('/auth/me');
  return data;
};