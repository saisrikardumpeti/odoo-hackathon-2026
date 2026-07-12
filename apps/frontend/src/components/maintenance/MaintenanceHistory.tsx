import { useMaintenanceList } from '#/lib/hooks/useMaintenance';
import { Loader2, Wrench } from 'lucide-react';

function MaintenanceHistory({ assetId }: { assetId: string }) {
  const { data, isLoading } = useMaintenanceList({ asset_id: assetId });

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-24">
        <Loader2 className="size-5 animate-spin text-muted-foreground" />
      </div>
    );
  }

  const requests = data?.maintenance_requests || [];

  if (requests.length === 0) {
    return (
      <div className="flex items-center justify-center h-24 text-sm text-muted-foreground">
        No maintenance history
      </div>
    );
  }

  return (
    <div className="space-y-3">
      {requests.map((req) => (
        <div key={req.id} className="rounded-lg border p-3 space-y-1">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2">
              <Wrench className="size-4 text-muted-foreground" />
              <span className="text-sm font-medium">{req.status}</span>
            </div>
            <span className={`rounded-full px-2 py-0.5 text-[10px] font-medium ${
              req.priority === 'Critical' ? 'bg-red-100 text-red-700' :
              req.priority === 'High' ? 'bg-orange-100 text-orange-700' :
              req.priority === 'Medium' ? 'bg-yellow-100 text-yellow-700' :
              'bg-blue-100 text-blue-700'
            }`}>
              {req.priority}
            </span>
          </div>
          <p className="text-xs text-muted-foreground">{req.issue_description}</p>
          {req.resolution_notes && (
            <p className="text-xs text-muted-foreground">Resolution: {req.resolution_notes}</p>
          )}
          <p className="text-[10px] text-muted-foreground">
            {new Date(req.created_at).toLocaleDateString()}
          </p>
        </div>
      ))}
    </div>
  );
}

export { MaintenanceHistory };
