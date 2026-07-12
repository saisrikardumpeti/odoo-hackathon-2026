import { useAuthStore } from '#/lib/stores/authStore';
import type { ReactNode } from 'react';

interface RequireRoleProps {
  roles: string[];
  children: ReactNode;
  fallback?: ReactNode;
}

function RequireRole({ roles, children, fallback }: RequireRoleProps) {
  const role = useAuthStore((state) => state.employee?.role);

  if (!role || !roles.includes(role)) {
    return fallback ?? null;
  }

  return <>{children}</>;
}

export { RequireRole };
