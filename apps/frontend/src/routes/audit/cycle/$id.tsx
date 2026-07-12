import { createFileRoute, useNavigate } from '@tanstack/react-router';
import { useAuthStore } from '#/lib/stores/authStore';
import { useAuditCycle, useAuditItems, usePatchAuditItem, useCloseAuditCycle } from '#/lib/hooks/useAudit';
import { Button } from '#/components/ui/button';
import { Badge } from '#/components/ui/badge';
import { ArrowLeft, TriangleAlert, CheckCircle, XCircle, AlertTriangle } from 'lucide-react';
import { useState } from 'react';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '#/components/ui/dialog';

export const Route = createFileRoute('/audit/cycle/$id')({ component: CycleDetailPage });

const statusBadgeVariant = (status: string) => {
  switch (status) {
    case 'Draft': return 'secondary' as const;
    case 'Active': return 'default' as const;
    case 'Closed': return 'outline' as const;
    default: return 'secondary' as const;
  }
};

const resultIcon = (result: string | null) => {
  switch (result) {
    case 'Verified': return <CheckCircle className="size-4 text-green-600" />;
    case 'Missing': return <XCircle className="size-4 text-destructive" />;
    case 'Damaged': return <AlertTriangle className="size-4 text-amber-600" />;
    default: return null;
  }
};

function CycleDetailPage() {
  const { id } = Route.useParams();
  const navigate = useNavigate();
  const employee = useAuthStore((s) => s.employee);
  const isAdmin = employee?.role === 'Admin';

  const { data: cycleData, isLoading: cycleLoading } = useAuditCycle(id);
  const [showMyItems, setShowMyItems] = useState(false);
  const { data: itemsData, isLoading: itemsLoading } = useAuditItems(id, showMyItems);
  const patchItem = usePatchAuditItem();
  const closeMutation = useCloseAuditCycle();
  const [showCloseDialog, setShowCloseDialog] = useState(false);
  const [notesInput, setNotesInput] = useState<Record<string, string>>({});

  if (cycleLoading) {
    return <div className="p-8">Loading...</div>;
  }

  const cycle = cycleData?.audit_cycle;
  if (!cycle) {
    return <div className="p-8 text-destructive">Audit cycle not found</div>;
  }

  const items = itemsData?.items ?? [];

  const handleResult = (itemId: string, result: 'Verified' | 'Missing' | 'Damaged') => {
    patchItem.mutate({
      itemId,
      req: { result, notes: notesInput[itemId] || null },
    });
  };

  return (
    <div className="p-8">
      <div className="mb-6">
        <Button variant="ghost" size="sm" onClick={() => navigate({ to: '/audit' })}>
          <ArrowLeft className="size-4 mr-2" />
          Back to Audit
        </Button>
      </div>

      <div className="flex items-start justify-between mb-8">
        <div>
          <h1 className="text-2xl font-bold">{cycle.name}</h1>
          <div className="flex items-center gap-3 mt-1 text-sm text-muted-foreground">
            <Badge variant={statusBadgeVariant(cycle.status)}>{cycle.status}</Badge>
            <span>{cycle.start_date} &mdash; {cycle.end_date}</span>
            <span>Scope: {cycle.scope_department_name ?? cycle.scope_location ?? 'All Assets'}</span>
          </div>
          <div className="flex gap-4 mt-3 text-sm">
            <span className="text-green-600">{cycle.verified_count ?? 0} Verified</span>
            <span className="text-destructive">{cycle.missing_count ?? 0} Missing</span>
            <span className="text-amber-600">{cycle.damaged_count ?? 0} Damaged</span>
            <span>{cycle.item_count ?? 0} Total</span>
          </div>
        </div>
        <div className="flex gap-2 items-start">
          {cycle.assigned_auditors?.some((a) => a.id === employee?.id) && (
            <Button
              variant={showMyItems ? 'default' : 'outline'}
              onClick={() => setShowMyItems(!showMyItems)}
            >
              {showMyItems ? 'Show All Items' : 'My Assigned Items'}
            </Button>
          )}
          {isAdmin && cycle.status !== 'Closed' && (
            <Button variant="destructive" onClick={() => setShowCloseDialog(true)}>
              Close Cycle
            </Button>
          )}
        </div>
      </div>

      {itemsLoading ? (
        <div className="text-muted-foreground">Loading items...</div>
      ) : items.length === 0 ? (
        <div className="text-center py-12 text-muted-foreground">
          <ClipboardCheck className="size-12 mx-auto mb-4 opacity-20" />
          <p>No audit items found{showMyItems ? ' assigned to you' : ''}.</p>
        </div>
      ) : (
        <div className="rounded-lg border">
          <table className="w-full">
            <thead>
              <tr className="border-b bg-muted/50 text-sm">
                <th className="text-left p-3 font-medium">Asset Tag</th>
                <th className="text-left p-3 font-medium">Asset Name</th>
                <th className="text-left p-3 font-medium">Status</th>
                <th className="text-left p-3 font-medium">Location</th>
                <th className="text-left p-3 font-medium">Result</th>
                <th className="text-left p-3 font-medium">Notes</th>
                {cycle.status !== 'Closed' && !showMyItems && (
                  <th className="text-left p-3 font-medium">Actions</th>
                )}
              </tr>
            </thead>
            <tbody>
              {items.map((item) => (
                <tr key={item.id} className="border-b text-sm hover:bg-muted/30">
                  <td className="p-3 font-mono text-xs">{item.asset_tag}</td>
                  <td className="p-3 font-medium">{item.asset_name}</td>
                  <td className="p-3"><Badge variant="outline">{item.asset_status}</Badge></td>
                  <td className="p-3 text-muted-foreground">{item.asset_location ?? '-'}</td>
                  <td className="p-3">
                    <div className="flex items-center gap-2">
                      {resultIcon(item.result)}
                      <span>{item.result ?? 'Pending'}</span>
                    </div>
                  </td>
                  <td className="p-3 max-w-[200px]">
                    {cycle.status !== 'Closed' ? (
                      <input
                        value={notesInput[item.id] ?? item.notes ?? ''}
                        onChange={(e) => setNotesInput((prev) => ({ ...prev, [item.id]: e.target.value }))}
                        placeholder="Add notes..."
                        className="w-full rounded border border-input bg-background px-2 py-1 text-xs"
                      />
                    ) : (
                      <span className="text-muted-foreground text-xs">{item.notes ?? '-'}</span>
                    )}
                  </td>
                  {cycle.status !== 'Closed' && (
                    <td className="p-3">
                      <div className="flex gap-1">
                        <button
                          onClick={() => handleResult(item.id, 'Verified')}
                          disabled={patchItem.isPending}
                          className="inline-flex items-center gap-1 rounded px-2 py-1 text-xs font-medium bg-green-100 text-green-700 hover:bg-green-200 disabled:opacity-50"
                        >
                          <CheckCircle className="size-3" />
                          Verify
                        </button>
                        <button
                          onClick={() => handleResult(item.id, 'Missing')}
                          disabled={patchItem.isPending}
                          className="inline-flex items-center gap-1 rounded px-2 py-1 text-xs font-medium bg-red-100 text-red-700 hover:bg-red-200 disabled:opacity-50"
                        >
                          <XCircle className="size-3" />
                          Missing
                        </button>
                        <button
                          onClick={() => handleResult(item.id, 'Damaged')}
                          disabled={patchItem.isPending}
                          className="inline-flex items-center gap-1 rounded px-2 py-1 text-xs font-medium bg-amber-100 text-amber-700 hover:bg-amber-200 disabled:opacity-50"
                        >
                          <AlertTriangle className="size-3" />
                          Damaged
                        </button>
                      </div>
                    </td>
                  )}
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      <Dialog open={showCloseDialog} onOpenChange={setShowCloseDialog}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle className="text-destructive flex items-center gap-2">
              <TriangleAlert className="size-5" />
              Close Audit Cycle
            </DialogTitle>
            <DialogDescription className="space-y-3">
              <p>This action is <strong>irreversible</strong>. Are you sure you want to close this cycle?</p>
              <ul className="list-disc pl-4 text-sm space-y-1">
                <li>All audit items will be <strong>locked</strong> from further edits</li>
                <li>Assets confirmed as <strong>Missing</strong> will have their status bulk-updated to <strong>Lost</strong></li>
                <li>Damaged items will <strong>not</strong> have their status automatically changed</li>
              </ul>
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button variant="outline" onClick={() => setShowCloseDialog(false)}>Cancel</Button>
            <Button
              variant="destructive"
              disabled={closeMutation.isPending}
              onClick={() => {
                closeMutation.mutate(id, {
                  onSuccess: () => setShowCloseDialog(false),
                });
              }}
            >
              {closeMutation.isPending ? 'Closing...' : 'Confirm Close Cycle'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}

function ClipboardCheck({ className }: { className?: string }) {
  return (
    <svg
      xmlns="http://www.w3.org/2000/svg"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
      className={className}
    >
      <rect width="8" height="4" x="8" y="2" rx="1" ry="1" />
      <path d="M16 4h2a2 2 0 0 1 2 2v14a2 2 0 0 1-2 2H6a2 2 0 0 1-2-2V6a2 2 0 0 1 2-2h2" />
      <path d="m9 14 2 2 4-4" />
    </svg>
  );
}
