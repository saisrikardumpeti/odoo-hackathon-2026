import { useState } from 'react';
import { useReturnAllocation } from '#/lib/hooks/useAllocations';
import { Button } from '#/components/ui/button';
import { Textarea } from '#/components/ui/textarea';
import { Undo2, Loader2 } from 'lucide-react';

interface ReturnFormProps {
  allocationId: string;
  assetName: string;
  onSuccess: () => void;
  onCancel: () => void;
}

function ReturnForm({ allocationId, assetName, onSuccess, onCancel }: ReturnFormProps) {
  const [notes, setNotes] = useState('');
  const mutation = useReturnAllocation();

  const handleSubmit = async () => {
    try {
      await mutation.mutateAsync({
        id: allocationId,
        req: { return_condition_notes: notes || null },
      });
      onSuccess();
    } catch {
      // error handled by mutation
    }
  };

  return (
    <div className="space-y-4">
      <div className="flex items-center gap-2">
        <Undo2 className="size-5 text-muted-foreground" />
        <h3 className="font-medium">Return Asset: {assetName}</h3>
      </div>
      <p className="text-sm text-muted-foreground">
        Confirm the return and add any condition notes about the asset's state.
      </p>
      <Textarea
        placeholder="Return condition notes (optional)..."
        value={notes}
        onChange={(e) => setNotes(e.target.value)}
        rows={3}
      />
      <div className="flex justify-end gap-2">
        <Button variant="outline" onClick={onCancel} disabled={mutation.isPending}>
          Cancel
        </Button>
        <Button onClick={handleSubmit} disabled={mutation.isPending}>
          {mutation.isPending && <Loader2 className="mr-2 size-4 animate-spin" />}
          Confirm Return
        </Button>
      </div>
      {mutation.isError && (
        <p className="text-sm text-destructive">Failed to return asset. Please try again.</p>
      )}
    </div>
  );
}

export { ReturnForm };
