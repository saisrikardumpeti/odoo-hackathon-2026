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
};
