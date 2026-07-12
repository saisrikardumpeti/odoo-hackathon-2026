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
  bookings: {
    all: ['bookings'] as const,
    my: ['bookings', 'my'] as const,
    byResource: (assetId: string, from?: string, to?: string) =>
      ['bookings', 'resource', assetId, from, to] as const,
    detail: (id: string) => ['bookings', id] as const,
  },
  maintenance: {
    all: ['maintenance'] as const,
    list: (params?: Record<string, string>) => ['maintenance', 'list', params] as const,
    detail: (id: string) => ['maintenance', id] as const,
    byAsset: (assetId: string) => ['maintenance', 'asset', assetId] as const,
  },
  notifications: {
    all: ['notifications'] as const,
    list: (params?: Record<string, unknown>) => ['notifications', 'list', params] as const,
    unreadCount: ['notifications', 'unread-count'] as const,
  },
  activityLogs: {
    all: ['activity-logs'] as const,
    list: (params?: Record<string, unknown>) => ['activity-logs', 'list', params] as const,
  },
  dashboard: {
    kpis: ['dashboard', 'kpis'] as const,
    overdue: ['dashboard', 'overdue'] as const,
    upcoming: (windowDays?: number) => ['dashboard', 'upcoming', windowDays] as const,
    activity: ['dashboard', 'activity'] as const,
  },
  audit: {
    cycles: {
      all: ['audit-cycles'] as const,
      detail: (id: string) => ['audit-cycles', id] as const,
    },
    items: (cycleId: string, myItems?: boolean) =>
      ['audit-items', cycleId, myItems] as const,
    discrepancies: {
      all: ['discrepancy-reports'] as const,
      filtered: (params?: Record<string, string>) =>
        ['discrepancy-reports', params] as const,
    },
  },
};
