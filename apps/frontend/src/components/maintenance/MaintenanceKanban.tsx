import { useState } from 'react';
import { useMaintenanceList, useApproveMaintenance, useRejectMaintenance, useAssignTechnician, useStartMaintenance, useResolveMaintenance } from '#/lib/hooks/useMaintenance';
import type { MaintenanceDetail, MaintenanceStatus } from '#/lib/api/maintenance';
import { useAuthStore } from '#/lib/stores/authStore';
import { Button } from '#/components/ui/button';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '#/components/ui/dialog';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '#/components/ui/select';
import { Textarea } from '#/components/ui/textarea';
import { Input } from '#/components/ui/input';
import { Loader2, AlertCircle, CheckCircle2, Wrench, PlayCircle, CheckCircle } from 'lucide-react';

const STATUS_COLUMNS: { key: MaintenanceStatus; label: string }[] = [
  { key: 'Pending', label: 'Pending' },
  { key: 'Approved', label: 'Approved' },
  { key: 'TechnicianAssigned', label: 'Technician Assigned' },
  { key: 'InProgress', label: 'In Progress' },
  { key: 'Resolved', label: 'Resolved' },
];

const STATUS_ICONS: Record<string, React.ReactNode> = {
  Pending: <AlertCircle className="size-4" />,
  Approved: <CheckCircle2 className="size-4" />,
  Rejected: <AlertCircle className="size-4 text-destructive" />,
  TechnicianAssigned: <Wrench className="size-4" />,
  InProgress: <PlayCircle className="size-4" />,
  Resolved: <CheckCircle className="size-4" />,
};

const PRIORITY_COLORS: Record<string, string> = {
  Low: 'bg-blue-100 text-blue-700 dark:bg-blue-900 dark:text-blue-300',
  Medium: 'bg-yellow-100 text-yellow-700 dark:bg-yellow-900 dark:text-yellow-300',
  High: 'bg-orange-100 text-orange-700 dark:bg-orange-900 dark:text-orange-300',
  Critical: 'bg-red-100 text-red-700 dark:bg-red-900 dark:text-red-300',
};

function MaintenanceCard({
  item,
  onAction,
}: {
  item: MaintenanceDetail;
  onAction: (action: string, id: string) => void;
}) {
  return (
    <div className="rounded-lg border bg-card p-3 shadow-sm space-y-2">
      <div className="flex items-start justify-between gap-2">
        <div className="flex-1 min-w-0">
          <p className="text-sm font-medium truncate">{item.asset_name}</p>
          <p className="text-xs text-muted-foreground">{item.asset_tag}</p>
        </div>
        <span className={`shrink-0 rounded-full px-2 py-0.5 text-[10px] font-medium ${PRIORITY_COLORS[item.priority] || ''}`}>
          {item.priority}
        </span>
      </div>
      <p className="text-xs text-muted-foreground line-clamp-2">{item.issue_description}</p>
      <div className="flex items-center gap-2 text-[10px] text-muted-foreground">
        <span>by {item.raised_by_name || 'Unknown'}</span>
      </div>
      {item.technician_name && (
        <p className="text-[10px] text-muted-foreground">Technician: {item.technician_name}</p>
      )}
    </div>
  );
}

function MaintenanceKanban() {
  const role = useAuthStore((s) => s.employee?.role);
  const isManager = role === 'Admin' || role === 'AssetManager';
  const { data, isLoading } = useMaintenanceList();
  const approveMutation = useApproveMaintenance();
  const rejectMutation = useRejectMaintenance();
  const assignMutation = useAssignTechnician();
  const startMutation = useStartMaintenance();
  const resolveMutation = useResolveMaintenance();

  const [rejectDialog, setRejectDialog] = useState<{ id: string } | null>(null);
  const [rejectReason, setRejectReason] = useState('');
  const [assignDialog, setAssignDialog] = useState<{ id: string } | null>(null);
  const [technicianName, setTechnicianName] = useState('');
  const [resolveDialog, setResolveDialog] = useState<{ id: string } | null>(null);
  const [resolutionNotes, setResolutionNotes] = useState('');

  const columns: Record<string, MaintenanceDetail[]> = {
    Pending: [],
    Approved: [],
    TechnicianAssigned: [],
    InProgress: [],
    Rejected: [],
    Resolved: [],
  };

  if (data?.maintenance_requests) {
    for (const req of data.maintenance_requests) {
      if (columns[req.status]) {
        columns[req.status].push(req);
      }
    }
  }

  const handleAction = (action: string, id: string) => {
    switch (action) {
      case 'approve':
        approveMutation.mutate(id);
        break;
      case 'reject':
        setRejectDialog({ id });
        break;
      case 'assign':
        setAssignDialog({ id });
        break;
      case 'start':
        startMutation.mutate(id);
        break;
      case 'resolve':
        setResolveDialog({ id });
        break;
    }
  };

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-48">
        <Loader2 className="size-6 animate-spin text-muted-foreground" />
      </div>
    );
  }

  return (
    <>
      <div className="grid grid-cols-1 md:grid-cols-5 gap-4">
        {STATUS_COLUMNS.map((col) => {
          const items = columns[col.key] || [];
          return (
            <div key={col.key} className="flex flex-col gap-3">
              <div className="flex items-center gap-2 px-1">
                {STATUS_ICONS[col.key]}
                <h3 className="text-sm font-semibold">{col.label}</h3>
                <span className="ml-auto text-xs text-muted-foreground">{items.length}</span>
              </div>
              <div className="flex flex-col gap-2 min-h-[200px] rounded-lg bg-muted/30 p-2">
                {items.length === 0 && (
                  <p className="text-xs text-muted-foreground text-center py-8">No requests</p>
                )}
                {items.map((item) => (
                  <div key={item.id} className="space-y-1">
                    <MaintenanceCard item={item} onAction={handleAction} />
                    {isManager && item.status === 'Pending' && (
                      <div className="flex gap-1 px-1">
                        <Button
                          size="sm"
                          variant="default"
                          className="h-7 text-xs flex-1"
                          onClick={() => handleAction('approve', item.id)}
                          disabled={approveMutation.isPending}
                        >
                          Approve
                        </Button>
                        <Button
                          size="sm"
                          variant="destructive"
                          className="h-7 text-xs flex-1"
                          onClick={() => handleAction('reject', item.id)}
                          disabled={rejectMutation.isPending}
                        >
                          Reject
                        </Button>
                      </div>
                    )}
                    {isManager && item.status === 'Approved' && (
                      <Button
                        size="sm"
                        variant="outline"
                        className="h-7 text-xs w-full"
                        onClick={() => handleAction('assign', item.id)}
                        disabled={assignMutation.isPending}
                      >
                        Assign Technician
                      </Button>
                    )}
                    {item.status === 'TechnicianAssigned' && (
                      <Button
                        size="sm"
                        variant="outline"
                        className="h-7 text-xs w-full"
                        onClick={() => handleAction('start', item.id)}
                        disabled={startMutation.isPending}
                      >
                        Start Work
                      </Button>
                    )}
                    {item.status === 'InProgress' && (
                      <Button
                        size="sm"
                        variant="default"
                        className="h-7 text-xs w-full"
                        onClick={() => handleAction('resolve', item.id)}
                        disabled={resolveMutation.isPending}
                      >
                        Resolve
                      </Button>
                    )}
                  </div>
                ))}
              </div>
            </div>
          );
        })}
      </div>

      <Dialog open={!!rejectDialog} onOpenChange={(open) => { if (!open) setRejectDialog(null); }}>
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
              <Button variant="outline" onClick={() => setRejectDialog(null)}>Cancel</Button>
              <Button
                variant="destructive"
                onClick={() => {
                  if (rejectDialog) {
                    rejectMutation.mutate({ id: rejectDialog.id, reason: rejectReason || undefined });
                    setRejectDialog(null);
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

      <Dialog open={!!assignDialog} onOpenChange={(open) => { if (!open) setAssignDialog(null); }}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Assign Technician</DialogTitle>
          </DialogHeader>
          <div className="space-y-4">
            <div className="space-y-2">
              <label className="text-sm font-medium">Technician Name</label>
              <Input
                value={technicianName}
                onChange={(e) => setTechnicianName(e.target.value)}
                placeholder="Enter technician name..."
              />
            </div>
            <div className="flex justify-end gap-2">
              <Button variant="outline" onClick={() => setAssignDialog(null)}>Cancel</Button>
              <Button
                onClick={() => {
                  if (assignDialog && technicianName.trim()) {
                    assignMutation.mutate({ id: assignDialog.id, technicianName: technicianName.trim() });
                    setAssignDialog(null);
                    setTechnicianName('');
                  }
                }}
                disabled={assignMutation.isPending || !technicianName.trim()}
              >
                {assignMutation.isPending ? 'Assigning...' : 'Assign'}
              </Button>
            </div>
          </div>
        </DialogContent>
      </Dialog>

      <Dialog open={!!resolveDialog} onOpenChange={(open) => { if (!open) setResolveDialog(null); }}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Resolve Maintenance</DialogTitle>
          </DialogHeader>
          <div className="space-y-4">
            <div className="space-y-2">
              <label className="text-sm font-medium">Resolution Notes</label>
              <Textarea
                value={resolutionNotes}
                onChange={(e) => setResolutionNotes(e.target.value)}
                placeholder="Enter resolution notes..."
              />
            </div>
            <div className="flex justify-end gap-2">
              <Button variant="outline" onClick={() => setResolveDialog(null)}>Cancel</Button>
              <Button
                onClick={() => {
                  if (resolveDialog && resolutionNotes.trim()) {
                    resolveMutation.mutate({ id: resolveDialog.id, resolutionNotes: resolutionNotes.trim() });
                    setResolveDialog(null);
                    setResolutionNotes('');
                  }
                }}
                disabled={resolveMutation.isPending || !resolutionNotes.trim()}
              >
                {resolveMutation.isPending ? 'Resolving...' : 'Resolve'}
              </Button>
            </div>
          </div>
        </DialogContent>
      </Dialog>
    </>
  );
}

export { MaintenanceKanban };
