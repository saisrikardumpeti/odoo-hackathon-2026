import { createFileRoute, redirect } from '@tanstack/react-router';
import { useAuthStore } from '#/lib/stores/authStore';
import { useState } from 'react';
import { useMyAllocations } from '#/lib/hooks/useAllocations';
import { useOverdueAllocations, usePendingTransfers, useApproveTransfer, useRejectTransfer } from '#/lib/hooks/useAllocations';
import { AllocateForm } from '#/components/allocation/AllocateForm';
import { OverdueAllocations } from '#/components/allocation/OverdueAllocations';
import { ReturnForm } from '#/components/allocation/ReturnForm';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '#/components/ui/tabs';
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogTrigger } from '#/components/ui/dialog';
import { Button } from '#/components/ui/button';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '#/components/ui/table';
import { Badge } from '#/components/ui/badge';
import {
  ArrowLeftRight, Undo2, CheckCircle, XCircle, Loader2, Package,
} from 'lucide-react';
import type { AllocationDetail } from '#/lib/api/allocations';

export const Route = createFileRoute('/allocation-transfer/')({
  beforeLoad: () => {
    if (!useAuthStore.getState().isAuthenticated) {
      throw redirect({ to: '/auth/login' });
    }
  },
  component: AllocationTransferPage,
});

function AllocationTransferPage() {
  const employee = useAuthStore((s) => s.employee);
  const isManager = employee?.role === 'Admin' || employee?.role === 'AssetManager' || employee?.role === 'DepartmentHead';

  const [activeTab, setActiveTab] = useState('allocate');
  const [returnAllocation, setReturnAllocation] = useState<AllocationDetail | null>(null);

  const { data: myAllocationsData, refetch: refetchMyAllocations } = useMyAllocations();
  const { data: overdueData } = useOverdueAllocations();
  const { data: pendingTransfersData, refetch: refetchTransfers } = usePendingTransfers();

  const approveTransferMutation = useApproveTransfer();
  const rejectTransferMutation = useRejectTransfer();

  const myAllocations = myAllocationsData?.allocations ?? [];
  const pendingTransfers = pendingTransfersData?.transfers ?? [];
  const overdueAllocations = overdueData?.allocations ?? [];

  return (
    <div className="p-8">
      <div className="mb-6">
        <h1 className="text-2xl font-bold flex items-center gap-2">
          <ArrowLeftRight className="size-6" /> Allocation & Transfer
        </h1>
        <p className="text-muted-foreground text-sm">Manage asset allocations, transfers, and returns.</p>
      </div>

      <Tabs value={activeTab} onValueChange={setActiveTab}>
        <TabsList>
          {isManager && <TabsTrigger value="allocate">Allocate</TabsTrigger>}
          <TabsTrigger value="my-allocations">My Allocations</TabsTrigger>
          <TabsTrigger value="overdue">
            Overdue
            {overdueAllocations.length > 0 && (
              <Badge variant="destructive" className="ml-2">{overdueAllocations.length}</Badge>
            )}
          </TabsTrigger>
          {isManager && (
            <TabsTrigger value="transfers">
              Transfers
              {pendingTransfers.length > 0 && (
                <Badge variant="default" className="ml-2">{pendingTransfers.length}</Badge>
              )}
            </TabsTrigger>
          )}
        </TabsList>

        {isManager && (
          <TabsContent value="allocate" className="mt-4">
            <AllocateForm />
          </TabsContent>
        )}

        <TabsContent value="my-allocations" className="mt-4">
          {myAllocations.length === 0 ? (
            <div className="flex flex-col items-center gap-2 py-16 text-muted-foreground">
              <Package className="size-12" />
              <p className="text-lg font-medium">No allocations</p>
              <p className="text-sm">You have no assets allocated to you.</p>
            </div>
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Asset</TableHead>
                  <TableHead>Allocated At</TableHead>
                  <TableHead>Expected Return</TableHead>
                  <TableHead>Status</TableHead>
                  <TableHead>Actions</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {myAllocations.map((a) => (
                  <TableRow key={a.id}>
                    <TableCell>
                      <span className="font-mono text-sm">{a.asset_tag}</span>
                      <div className="text-xs text-muted-foreground">{a.asset_name}</div>
                    </TableCell>
                    <TableCell className="text-sm">{new Date(a.allocated_at).toLocaleDateString()}</TableCell>
                    <TableCell className="text-sm">{a.expected_return_date ?? '—'}</TableCell>
                    <TableCell>
                      <Badge variant={a.status === 'Active' ? 'default' : 'secondary'}>
                        {a.status}
                      </Badge>
                    </TableCell>
                    <TableCell>
                      {a.status === 'Active' && (
                        <Dialog open={returnAllocation?.id === a.id} onOpenChange={(open) => setReturnAllocation(open ? a : null)}>
                          <DialogTrigger>
                            <Button variant="outline" size="sm">
                              <Undo2 className="mr-1 size-3" /> Return
                            </Button>
                          </DialogTrigger>
                          <DialogContent>
                            <DialogHeader>
                              <DialogTitle>Return Asset</DialogTitle>
                            </DialogHeader>
                            <ReturnForm
                              allocationId={a.id}
                              assetName={`${a.asset_tag} - ${a.asset_name}`}
                              onSuccess={() => { setReturnAllocation(null); refetchMyAllocations(); }}
                              onCancel={() => setReturnAllocation(null)}
                            />
                          </DialogContent>
                        </Dialog>
                      )}
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          )}
        </TabsContent>

        <TabsContent value="overdue" className="mt-4">
          <OverdueAllocations />
        </TabsContent>

        {isManager && (
          <TabsContent value="transfers" className="mt-4">
            {pendingTransfers.length === 0 ? (
              <div className="flex flex-col items-center gap-2 py-16 text-muted-foreground">
                <ArrowLeftRight className="size-12" />
                <p className="text-lg font-medium">No pending transfers</p>
                <p className="text-sm">All transfer requests have been processed.</p>
              </div>
            ) : (
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Asset</TableHead>
                    <TableHead>From</TableHead>
                    <TableHead>To</TableHead>
                    <TableHead>Requested By</TableHead>
                    <TableHead>Actions</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {pendingTransfers.map((t) => (
                    <TableRow key={t.id}>
                      <TableCell>
                        <span className="font-mono text-sm">{t.asset_tag}</span>
                        <div className="text-xs text-muted-foreground">{t.asset_name}</div>
                      </TableCell>
                      <TableCell className="text-sm">{t.from_employee_name ?? '—'}</TableCell>
                      <TableCell className="text-sm">{t.to_employee_name}</TableCell>
                      <TableCell className="text-sm">{t.requested_by_name}</TableCell>
                      <TableCell>
                        <div className="flex gap-2">
                          <Button
                            variant="default"
                            size="sm"
                            onClick={() => approveTransferMutation.mutate(t.id, { onSuccess: () => refetchTransfers() })}
                            disabled={approveTransferMutation.isPending}
                          >
                            {approveTransferMutation.isPending ? <Loader2 className="mr-1 size-3 animate-spin" /> : <CheckCircle className="mr-1 size-3" />}
                            Approve
                          </Button>
                          <Button
                            variant="outline"
                            size="sm"
                            onClick={() => rejectTransferMutation.mutate(t.id, { onSuccess: () => refetchTransfers() })}
                            disabled={rejectTransferMutation.isPending}
                          >
                            <XCircle className="mr-1 size-3" />
                            Reject
                          </Button>
                        </div>
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            )}
          </TabsContent>
        )}
      </Tabs>
    </div>
  );
}
