import { createFileRoute, redirect } from '@tanstack/react-router';
import { useAuthStore } from '#/lib/stores/authStore';
import { useActivityLogs } from '#/lib/hooks/useActivityLogs';
import { useState } from 'react';
import type { ActivityLogFilters } from '#/lib/api/activityLogs';
import { ChevronDown, ChevronRight } from 'lucide-react';
import { Button } from '#/components/ui/button';
import { Input } from '#/components/ui/input';

export const Route = createFileRoute('/activity-logs/')({
  beforeLoad: () => {
    const { employee, isAuthenticated } = useAuthStore.getState();
    if (!isAuthenticated) {
      throw redirect({ to: '/auth/login' });
    }
    const role = employee?.role;
    if (!role || !['Admin', 'AssetManager', 'DepartmentHead'].includes(role)) {
      throw redirect({ to: '/' });
    }
  },
  component: ActivityLogsPage,
});

function ActivityLogsPage() {
  const [filters, setFilters] = useState<ActivityLogFilters>({ page: 1, page_size: 20 });
  const [expandedId, setExpandedId] = useState<string | null>(null);
  const { data, isLoading, isError } = useActivityLogs(filters);

  if (isLoading) {
    return (
      <div className="flex items-center justify-center p-8">
        <div className="size-6 animate-spin rounded-full border-2 border-primary border-t-transparent" />
      </div>
    );
  }

  if (isError) {
    return (
      <div className="p-8 text-destructive">Failed to load activity logs.</div>
    );
  }

  const totalPages = data ? Math.ceil(data.total / data.page_size) : 0;

  return (
    <div className="p-6">
      <h1 className="mb-6 text-2xl font-bold">Activity Logs</h1>

      <div className="mb-4 flex flex-wrap gap-2">
        <Input
          placeholder="Actor"
          className="w-40"
          value={filters.actor ?? ''}
          onChange={(e) => setFilters((f) => ({ ...f, actor: e.target.value || undefined, page: 1 }))}
        />
        <Input
          placeholder="Action"
          className="w-40"
          value={filters.action ?? ''}
          onChange={(e) => setFilters((f) => ({ ...f, action: e.target.value || undefined, page: 1 }))}
        />
        <Input
          placeholder="Entity type"
          className="w-40"
          value={filters.entity_type ?? ''}
          onChange={(e) => setFilters((f) => ({ ...f, entity_type: e.target.value || undefined, page: 1 }))}
        />
        <Input
          placeholder="Entity ID"
          className="w-40"
          value={filters.entity_id ?? ''}
          onChange={(e) => setFilters((f) => ({ ...f, entity_id: e.target.value || undefined, page: 1 }))}
        />
        <Input
          type="date"
          className="w-40"
          value={filters.date_from?.split('T')[0] ?? ''}
          onChange={(e) => setFilters((f) => ({
            ...f,
            date_from: e.target.value ? new Date(e.target.value).toISOString() : undefined,
            page: 1,
          }))}
        />
        <Input
          type="date"
          className="w-40"
          value={filters.date_to?.split('T')[0] ?? ''}
          onChange={(e) => setFilters((f) => ({
            ...f,
            date_to: e.target.value ? new Date(e.target.value).toISOString() : undefined,
            page: 1,
          }))}
        />
        <Button
          variant="outline"
          size="sm"
          onClick={() => setFilters({ page: 1, page_size: 20 })}
        >
          Clear
        </Button>
      </div>

      <div className="rounded-lg border">
        <div className="grid grid-cols-5 gap-4 border-b bg-muted/50 px-4 py-2 text-xs font-medium text-muted-foreground">
          <span>Timestamp</span>
          <span>Actor</span>
          <span>Action</span>
          <span>Entity</span>
          <span>Details</span>
        </div>

        {!data?.logs?.length ? (
          <p className="py-12 text-center text-muted-foreground">No activity logs found</p>
        ) : (
          data.logs.map((log) => (
            <div key={log.id}>
              <div className="grid grid-cols-5 gap-4 border-b px-4 py-2 text-sm hover:bg-accent/30">
                <span className="text-xs text-muted-foreground">
                  {new Date(log.created_at).toLocaleString()}
                </span>
                <span title={log.actor_employee_id ?? undefined}>
                  {log.actor_name ?? (log.actor_employee_id ? 'Unknown' : 'System')}
                </span>
                <span className="font-mono text-xs">{log.action}</span>
                <span className="text-xs">
                  {log.entity_type}{log.entity_id ? ` / ${log.entity_id.slice(0, 8)}...` : ''}
                </span>
                <button
                  onClick={() => setExpandedId(expandedId === log.id ? null : log.id)}
                  className="flex items-center gap-1 text-xs text-muted-foreground hover:text-foreground"
                >
                  {expandedId === log.id ? (
                    <ChevronDown className="size-3" />
                  ) : (
                    <ChevronRight className="size-3" />
                  )}
                  Metadata
                </button>
              </div>
              {expandedId === log.id && (
                <div className="border-b bg-muted/20 px-8 py-3">
                  <pre className="overflow-x-auto text-xs text-muted-foreground">
                    {JSON.stringify(log.metadata, null, 2)}
                  </pre>
                </div>
              )}
            </div>
          ))
        )}
      </div>

      {totalPages > 1 && (
        <div className="mt-4 flex items-center justify-center gap-2">
          <Button
            variant="outline"
            size="sm"
            disabled={filters.page! <= 1}
            onClick={() => setFilters((f) => ({ ...f, page: f.page! - 1 }))}
          >
            Previous
          </Button>
          <span className="text-sm text-muted-foreground">
            Page {filters.page} of {totalPages}
          </span>
          <Button
            variant="outline"
            size="sm"
            disabled={filters.page! >= totalPages}
            onClick={() => setFilters((f) => ({ ...f, page: f.page! + 1 }))}
          >
            Next
          </Button>
        </div>
      )}
    </div>
  );
}
