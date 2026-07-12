import { useState } from 'react';
import { useAssets } from '#/lib/hooks/useAssets';
import { useEmployees } from '#/lib/hooks/useEmployees';
import { useDepartments } from '#/lib/hooks/useDepartments';
import { useCreateAllocation, useCreateTransfer } from '#/lib/hooks/useAllocations';
import { Button } from '#/components/ui/button';
import { Input } from '#/components/ui/input';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '#/components/ui/select';
import { ConflictModal } from '#/components/allocation/ConflictModal';
import { Search, Loader2, User, Building2 } from 'lucide-react';

function AllocateForm() {
  const [selectedAssetId, setSelectedAssetId] = useState('');
  const [holderType, setHolderType] = useState<'employee' | 'department'>('employee');
  const [holderId, setHolderId] = useState('');
  const [expectedReturnDate, setExpectedReturnDate] = useState('');
  const [searchAsset, setSearchAsset] = useState('');
  const [transferToEmployeeId, setTransferToEmployeeId] = useState('');

  const [conflictOpen, setConflictOpen] = useState(false);
  const [conflictAllocationId, setConflictAllocationId] = useState('');
  const [conflictHolderName, setConflictHolderName] = useState('');

  const { data: assetsData } = useAssets({ status: 'Available', asset_tag: searchAsset || undefined, page_size: 50 });
  const { data: employeesData } = useEmployees();
  const { data: departmentsData } = useDepartments();

  const createAllocationMutation = useCreateAllocation();
  const createTransferMutation = useCreateTransfer();

  const assets = assetsData?.assets ?? [];
  const employees = employeesData?.employees ?? [];
  const departments = departmentsData?.departments ?? [];

  const selectedAsset = assets.find((a) => a.id === selectedAssetId);

  const handleAllocate = async () => {
    if (!selectedAssetId || !holderId) return;

    try {
      await createAllocationMutation.mutateAsync({
        asset_id: selectedAssetId,
        employee_id: holderType === 'employee' ? holderId : null,
        department_id: holderType === 'department' ? holderId : null,
        expected_return_date: expectedReturnDate || null,
      });
      setSelectedAssetId('');
      setHolderId('');
      setExpectedReturnDate('');
    } catch (err: unknown) {
      const axiosErr = err as {
        response?: { status?: number; data?: { error?: string; message?: string; current_holder?: { id?: string; employee_name?: string; department_name?: string } } };
      };
      if (axiosErr?.response?.status === 409 && axiosErr?.response?.data?.error === 'AlreadyAllocated') {
        setConflictAllocationId(axiosErr.response.data.current_holder?.id ?? '');
        setConflictHolderName(axiosErr.response.data.message ?? 'currently held');
        setConflictOpen(true);
      }
    }
  };

  const handleRequestTransfer = (allocationId: string) => {
    setConflictOpen(false);
    if (allocationId && transferToEmployeeId) {
      createTransferMutation.mutate({
        allocation_id: allocationId,
        to_employee_id: transferToEmployeeId,
      });
    }
  };

  return (
    <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
      <div className="space-y-4">
        <h3 className="text-sm font-semibold text-muted-foreground uppercase tracking-wide">Asset Selection</h3>

        <div className="relative">
          <Search className="absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
          <Input
            placeholder="Search available assets..."
            value={searchAsset}
            onChange={(e) => setSearchAsset(e.target.value)}
            className="pl-9"
          />
        </div>

        <Select value={selectedAssetId} onValueChange={(v) => setSelectedAssetId(v ?? '')}>
          <SelectTrigger className="w-full">
            <SelectValue placeholder="Select an asset">
              {selectedAsset ? `${selectedAsset.asset_tag} - ${selectedAsset.name}` : null}
            </SelectValue>
          </SelectTrigger>
          <SelectContent>
            {assets.length === 0 && (
              <SelectItem value="" disabled>No available assets found</SelectItem>
            )}
            {assets.map((a) => (
              <SelectItem key={a.id} value={a.id}>
                <span className="font-mono">{a.asset_tag}</span> - {a.name}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      </div>

      <div className="space-y-4">
        <h3 className="text-sm font-semibold text-muted-foreground uppercase tracking-wide">Assign To</h3>

        <div className="flex gap-2">
          <Button
            variant={holderType === 'employee' ? 'default' : 'outline'}
            onClick={() => { setHolderType('employee'); setHolderId(''); }}
            className="flex-1"
          >
            <User className="mr-1 size-4" /> Employee
          </Button>
          <Button
            variant={holderType === 'department' ? 'default' : 'outline'}
            onClick={() => { setHolderType('department'); setHolderId(''); }}
            className="flex-1"
          >
            <Building2 className="mr-1 size-4" /> Department
          </Button>
        </div>

        {holderType === 'employee' ? (
          <Select value={holderId} onValueChange={(v) => setHolderId(v ?? '')}>
            <SelectTrigger className="w-full">
              <SelectValue placeholder="Select employee">
                {employees.find((e) => e.id === holderId)?.name ?? null}
              </SelectValue>
            </SelectTrigger>
            <SelectContent>
              {employees.map((e) => (
                <SelectItem key={e.id} value={e.id}>{e.name} ({e.email})</SelectItem>
              ))}
            </SelectContent>
          </Select>
        ) : (
          <Select value={holderId} onValueChange={(v) => setHolderId(v ?? '')}>
            <SelectTrigger className="w-full">
              <SelectValue placeholder="Select department">
                {departments.find((d) => d.id === holderId)?.name ?? null}
              </SelectValue>
            </SelectTrigger>
            <SelectContent>
              {departments.map((d) => (
                <SelectItem key={d.id} value={d.id}>{d.name}</SelectItem>
              ))}
            </SelectContent>
          </Select>
        )}

        <Input
          type="date"
          placeholder="Expected return date"
          value={expectedReturnDate}
          onChange={(e) => setExpectedReturnDate(e.target.value)}
        />
      </div>

      <div className="md:col-span-2 space-y-4">
        <h3 className="text-sm font-semibold text-muted-foreground uppercase tracking-wide">Transfer On Conflict</h3>

        <Select value={transferToEmployeeId} onValueChange={(v) => setTransferToEmployeeId(v ?? '')}>
          <SelectTrigger className="w-full">
            <SelectValue placeholder="Transfer to employee (optional)">
              {employees.find((e) => e.id === transferToEmployeeId)?.name ?? null}
            </SelectValue>
          </SelectTrigger>
          <SelectContent>
            {employees.map((e) => (
              <SelectItem key={e.id} value={e.id}>{e.name}</SelectItem>
            ))}
          </SelectContent>
        </Select>
        <p className="text-xs text-muted-foreground">
          If the asset is already allocated elsewhere, a transfer request to this employee will be offered instead of blocking you.
        </p>
      </div>

      <div className="md:col-span-2">
        <Button
          onClick={handleAllocate}
          disabled={!selectedAssetId || !holderId || createAllocationMutation.isPending}
          className="w-full"
        >
          {createAllocationMutation.isPending && <Loader2 className="mr-2 size-4 animate-spin" />}
          Allocate Asset
        </Button>

        {createAllocationMutation.isError && !conflictOpen && (
          <p className="mt-2 text-sm text-destructive">Failed to allocate asset. Please try again.</p>
        )}
      </div>

      <ConflictModal
        open={conflictOpen}
        onOpenChange={setConflictOpen}
        holderName={conflictHolderName}
        assetId={selectedAssetId}
        allocationId={conflictAllocationId}
        onRequestTransfer={handleRequestTransfer}
      />
    </div>
  );
}

export { AllocateForm };
