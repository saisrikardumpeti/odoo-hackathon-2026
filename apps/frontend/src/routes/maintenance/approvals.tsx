import { createFileRoute } from '@tanstack/react-router';
import { useMaintenanceList, useApproveMaintenance, useRejectMaintenance } from '#/lib/hooks/useMaintenance';
import { Button } from '#/components/ui/button';
import { useState } from 'react';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from '#/components/ui/dialog';
import { Textarea } from '#/components/ui/textarea';
import { Loader2 } from 'lucide-react';

export const Route = createFileRoute('/maintenance/approvals')({ component: ApprovalsPage });

function ApprovalsPage() {
  const { data, isLoading } = useMaintenanceList({ status: 'Pending' });
  const approveMutation = useApproveMaintenance();
  const rejectMutation = useRejectMaintenance();

  const [rejectTarget, setRejectTarget] = useState<{ id: string } | null>(null);
  const [rejectReason, setRejectReason] = useState('');

  const pendingRequests = data?.maintenance_requests || [];

  return (
    <div className="p-6 space-y-6">
      <div>
        <h1 className="text-2xl font-bold">Approval Queue</h1>
        <p className="text-sm text-muted-foreground">Pending maintenance requests awaiting your decision</p>
      </div>

      {isLoading ? (
        <div className="flex items-center justify-center h-48">
          <Loader2 className="size-6 animate-spin text-muted-foreground" />
        </div>
      ) : pendingRequests.length === 0 ? (
        <div className="rounded-lg border border-dashed p-12 text-center">
          <p className="text-muted-foreground">No pending maintenance requests</p>
        </div>
      ) : (
        <div className="space-y-3">
          {pendingRequests.map((req) => (
            <div key={req.id} className="rounded-lg border p-4 space-y-2">
              <div className="flex items-start justify-between">
                <div>
                  <p className="font-medium">{req.asset_name}</p>
                  <p className="text-xs text-muted-foreground">{req.asset_tag}</p>
                </div>
                <span className="text-xs text-muted-foreground">{req.priority}</span>
              </div>
              <p className="text-sm">{req.issue_description}</p>
              <div className="flex items-center gap-2 text-xs text-muted-foreground">
                <span>Raised by: {req.raised_by_name || 'Unknown'}</span>
                <span>|</span>
                <span>{new Date(req.created_at).toLocaleDateString()}</span>
              </div>
              <div className="flex gap-2 pt-1">
                <Button
                  size="sm"
                  onClick={() => approveMutation.mutate(req.id)}
                  disabled={approveMutation.isPending}
                >
                  {approveMutation.isPending ? 'Approving...' : 'Approve'}
                </Button>
                <Button
                  size="sm"
                  variant="destructive"
                  onClick={() => setRejectTarget({ id: req.id })}
                  disabled={rejectMutation.isPending}
                >
                  Reject
                </Button>
              </div>
            </div>
          ))}
        </div>
      )}

      <Dialog open={!!rejectTarget} onOpenChange={(open) => { if (!open) setRejectTarget(null); }}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Reject Maintenance Request</DialogTitle>
          </DialogHeader>
          <div className="space-y-4">
            <div className="space-y-2">
              <label className="text-sm font-medium">Reason (optional)</label>
              <Textarea
                value={rejectReason}
                onChange={(e) => setRejectReason(e.target.value)}
                placeholder="Enter rejection reason..."
              />
            </div>
            <div className="flex justify-end gap-2">
              <Button variant="outline" onClick={() => setRejectTarget(null)}>Cancel</Button>
              <Button
                variant="destructive"
                onClick={() => {
                  if (rejectTarget) {
                    rejectMutation.mutate({
                      id: rejectTarget.id,
                      reason: rejectReason || undefined,
                    });
                    setRejectTarget(null);
                    setRejectReason('');
                  }
                }}
                disabled={rejectMutation.isPending}
              >
                {rejectMutation.isPending ? 'Rejecting...' : 'Reject'}
              </Button>
            </div>
          </div>
        </DialogContent>
      </Dialog>
    </div>
  );
}
