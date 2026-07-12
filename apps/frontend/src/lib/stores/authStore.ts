import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import type { Employee } from '#/lib/api/auth';

interface AuthState {
  employee: Employee | null;
  accessToken: string | null;
  refreshToken: string | null;
  isAuthenticated: boolean;
  isInitialized: boolean;
  setAuth: (employee: Employee, accessToken: string, refreshToken: string) => void;
  setTokens: (accessToken: string, refreshToken: string) => void;
  setEmployee: (employee: Employee) => void;
  setInitialized: () => void;
  logout: () => void;
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set) => ({
      employee: null,
      accessToken: null,
      refreshToken: null,
      isAuthenticated: false,
      isInitialized: false,
      setAuth: (employee, accessToken, refreshToken) =>
        set({ employee, accessToken, refreshToken, isAuthenticated: true }),
      setTokens: (accessToken, refreshToken) =>
        set({ accessToken, refreshToken }),
      setEmployee: (employee) =>
        set({ employee }),
      setInitialized: () =>
        set({ isInitialized: true }),
      logout: () =>
        set({ employee: null, accessToken: null, refreshToken: null, isAuthenticated: false, isInitialized: true }),
    }),
    { name: 'auth-storage' },
  ),
);
