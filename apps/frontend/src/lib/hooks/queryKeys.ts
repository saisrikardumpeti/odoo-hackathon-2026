export const queryKeys = {
  ping: ['ping'] as const,
  models: ['models'] as const,
  departments: {
    all: ['departments'] as const,
    detail: (id: string) => ['departments', id] as const,
  },
  categories: {
    all: ['categories'] as const,
    detail: (id: string) => ['categories', id] as const,
  },
  employees: {
    all: ['employees'] as const,
    filtered: (params?: Record<string, string>) => ['employees', params] as const,
  },
  assets: {
    all: ['assets'] as const,
    list: (params?: Record<string, unknown>) => ['assets', 'list', params] as const,
    detail: (id: string) => ['assets', id] as const,
    history: (id: string) => ['assets', id, 'history'] as const,
  },
  allocations: {
    all: ['allocations'] as const,
    my: ['allocations', 'my'] as const,
    overdue: ['allocations', 'overdue'] as const,
  },
  transfers: {
    all: ['transfers'] as const,
    pending: ['transfers', 'pending'] as const,
  },
};
