import { createFileRoute, Link } from '@tanstack/react-router';
import { useAuthStore } from '#/lib/stores/authStore';
import { useAuditCycles, useCloseAuditCycle, useDiscrepancyReports, useResolveDiscrepancy } from '#/lib/hooks/useAudit';
import { Button } from '#/components/ui/button';
import { Badge } from '#/components/ui/badge';
import { ClipboardCheck, Plus, TriangleAlert } from 'lucide-react';
import { useState } from 'react';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '#/components/ui/dialog';

export const Route = createFileRoute('/audit/')({ component: AuditPage });

const statusBadgeVariant = (status: string) => {
  switch (status) {
    case 'Draft': return 'secondary' as const;
    case 'Active': return 'default' as const;
    case 'Closed': return 'outline' as const;
    default: return 'secondary' as const;
  }
};

function AuditPage() {
  const employee = useAuthStore((s) => s.employee);
  const isAdmin = employee?.role === 'Admin';
  const isManager = employee?.role === 'Admin' || employee?.role === 'AssetManager';
  const [activeTab, setActiveTab] = useState<'cycles' | 'discrepancies'>('cycles');
  const [closeCycleId, setCloseCycleId] = useState<string | null>(null);

  return (
    <div className="p-8">
      <div className="mb-8 flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold flex items-center gap-2">
            <ClipboardCheck className="size-6" />
            Asset Audit
          </h1>
          <p className="text-muted-foreground text-sm">Manage audit cycles and discrepancy reports</p>
        </div>
        <div className="flex gap-2">
          {isAdmin && (
            <Link to="/audit/new">
              <Button>
                <Plus className="size-4 mr-2" />
                New Audit Cycle
              </Button>
            </Link>
          )}
        </div>
      </div>

      <div className="flex gap-4 border-b mb-6">
        <button
          onClick={() => setActiveTab('cycles')}
          className={`pb-2 text-sm font-medium border-b-2 transition-colors ${activeTab === 'cycles' ? 'border-primary text-primary' : 'border-transparent text-muted-foreground hover:text-foreground'}`}
        >
          Audit Cycles
        </button>
        <button
          onClick={() => setActiveTab('discrepancies')}
          className={`pb-2 text-sm font-medium border-b-2 transition-colors ${activeTab === 'discrepancies' ? 'border-primary text-primary' : 'border-transparent text-muted-foreground hover:text-foreground'}`}
        >
          Discrepancy Reports
        </button>
      </div>

      {activeTab === 'cycles' ? (
        <CyclesTab isAdmin={isAdmin} />
      ) : (
        <DiscrepanciesTab isManager={isManager} />
      )}
    </div>
  );
}

function CyclesTab({ isAdmin }: { isAdmin: boolean }) {
  const { data, isLoading } = useAuditCycles();
  const closeMutation = useCloseAuditCycle();
  const [closeCycleId, setCloseCycleId] = useState<string | null>(null);

  if (isLoading) {
    return <div className="text-muted-foreground">Loading...</div>;
  }

  const cycles = data?.audit_cycles ?? [];

  if (cycles.length === 0) {
    return (
      <div className="text-center py-12 text-muted-foreground">
        <ClipboardCheck className="size-12 mx-auto mb-4 opacity-20" />
        <p className="text-lg font-medium">No audit cycles yet</p>
        <p className="text-sm">Create your first audit cycle to start tracking asset verifications.</p>
      </div>
    );
  }

  return (
    <>
      <div className="rounded-lg border">
        <table className="w-full">
          <thead>
            <tr className="border-b bg-muted/50 text-sm">
              <th className="text-left p-3 font-medium">Name</th>
              <th className="text-left p-3 font-medium">Scope</th>
              <th className="text-left p-3 font-medium">Period</th>
              <th className="text-left p-3 font-medium">Status</th>
              <th className="text-left p-3 font-medium">Items</th>
              <th className="text-left p-3 font-medium">Verified</th>
              <th className="text-left p-3 font-medium">Issues</th>
              <th className="text-left p-3 font-medium">Actions</th>
            </tr>
          </thead>
          <tbody>
            {cycles.map((cycle) => (
              <tr key={cycle.id} className="border-b text-sm hover:bg-muted/30">
                <td className="p-3">
                  <Link to="/audit/cycle/$id" params={{ id: cycle.id }} className="font-medium hover:underline">
                    {cycle.name}
                  </Link>
                </td>
                <td className="p-3 text-muted-foreground">
                  {cycle.scope_department_name ?? cycle.scope_location ?? 'All Assets'}
                </td>
                <td className="p-3 text-muted-foreground">
                  {cycle.start_date} &mdash; {cycle.end_date}
                </td>
                <td className="p-3">
                  <Badge variant={statusBadgeVariant(cycle.status)}>{cycle.status}</Badge>
                </td>
                <td className="p-3">{cycle.item_count}</td>
                <td className="p-3 text-green-600">{cycle.verified_count ?? 0}</td>
                <td className="p-3">
                  {((cycle.missing_count ?? 0) + (cycle.damaged_count ?? 0)) > 0 && (
                    <span className="text-destructive flex items-center gap-1">
                      <TriangleAlert className="size-3" />
                      {(cycle.missing_count ?? 0) + (cycle.damaged_count ?? 0)}
                    </span>
                  )}
                </td>
                <td className="p-3">
                  <div className="flex gap-1">
                    <Link to="/audit/cycle/$id" params={{ id: cycle.id }}>
                      <Button variant="outline" size="sm">View</Button>
                    </Link>
                    {isAdmin && cycle.status !== 'Closed' && (
                      <Button
                        variant="destructive"
                        size="sm"
                        onClick={() => setCloseCycleId(cycle.id)}
                      >
                        Close
                      </Button>
                    )}
                  </div>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      <Dialog open={!!closeCycleId} onOpenChange={(o) => { if (!o) setCloseCycleId(null); }}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle className="text-destructive">Close Audit Cycle</DialogTitle>
            <DialogDescription>
              This action is <strong>irreversible</strong>. All audit items in this cycle will be locked from further edits.
              Assets confirmed as <strong>Missing</strong> will have their status bulk-updated to <strong>Lost</strong>.
              Damaged items will not be automatically updated.
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button variant="outline" onClick={() => setCloseCycleId(null)}>Cancel</Button>
            <Button
              variant="destructive"
              disabled={closeMutation.isPending}
              onClick={() => {
                if (closeCycleId) {
                  closeMutation.mutate(closeCycleId, {
                    onSuccess: () => setCloseCycleId(null),
                  });
                }
              }}
            >
              {closeMutation.isPending ? 'Closing...' : 'Confirm Close Cycle'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  );
}

function DiscrepanciesTab({ isManager }: { isManager: boolean }) {
  const [filterResolved, setFilterResolved] = useState<string>('false');
  const { data, isLoading } = useDiscrepancyReports({ resolved: filterResolved || undefined });
  const resolveMutation = useResolveDiscrepancy();

  if (isLoading) {
    return <div className="text-muted-foreground">Loading...</div>;
  }

  const reports = data?.discrepancy_reports ?? [];

  return (
    <div>
      <div className="flex items-center gap-2 mb-4">
        <label className="text-sm font-medium">Filter:</label>
        <select
          value={filterResolved}
          onChange={(e) => setFilterResolved(e.target.value)}
          className="rounded-lg border border-input bg-background px-3 py-1.5 text-sm"
        >
          <option value="false">Unresolved</option>
          <option value="true">Resolved</option>
          <option value="">All</option>
        </select>
      </div>

      {reports.length === 0 ? (
        <div className="text-center py-12 text-muted-foreground">
          <TriangleAlert className="size-12 mx-auto mb-4 opacity-20" />
          <p>No discrepancy reports found</p>
        </div>
      ) : (
        <div className="rounded-lg border">
          <table className="w-full">
            <thead>
              <tr className="border-b bg-muted/50 text-sm">
                <th className="text-left p-3 font-medium">Asset</th>
                <th className="text-left p-3 font-medium">Cycle</th>
                <th className="text-left p-3 font-medium">Issue Type</th>
                <th className="text-left p-3 font-medium">Status</th>
                <th className="text-left p-3 font-medium">Resolved By</th>
                <th className="text-left p-3 font-medium">Actions</th>
              </tr>
            </thead>
            <tbody>
              {reports.map((report) => (
                <tr key={report.id} className="border-b text-sm hover:bg-muted/30">
                  <td className="p-3">
                    <div className="font-medium">{report.asset_name}</div>
                    <div className="text-xs text-muted-foreground">{report.asset_tag}</div>
                  </td>
                  <td className="p-3 text-muted-foreground">{report.cycle_name}</td>
                  <td className="p-3">
                    <Badge variant={report.issue_type === 'Missing' ? 'destructive' : 'secondary'}>
                      {report.issue_type}
                    </Badge>
                  </td>
                  <td className="p-3">
                    {report.resolved ? (
                      <Badge variant="outline" className="text-green-600">Resolved</Badge>
                    ) : (
                      <Badge variant="secondary">Open</Badge>
                    )}
                  </td>
                  <td className="p-3 text-muted-foreground">{report.resolved_by_name ?? '-'}</td>
                  <td className="p-3">
                    {!report.resolved && isManager && (
                      <Button
                        variant="outline"
                        size="sm"
                        disabled={resolveMutation.isPending}
                        onClick={() => resolveMutation.mutate(report.id)}
                      >
                        Resolve
                      </Button>
                    )}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
}
