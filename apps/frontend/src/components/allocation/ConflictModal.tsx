import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription, DialogFooter } from '#/components/ui/dialog';
import { Button } from '#/components/ui/button';
import { User, ArrowLeftRight } from 'lucide-react';

interface ConflictModalProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  holderName: string;
  assetId: string;
  allocationId: string;
  onRequestTransfer: (allocationId: string) => void;
}

function ConflictModal({ open, onOpenChange, holderName, allocationId, onRequestTransfer }: ConflictModalProps) {
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <User className="size-5 text-destructive" />
            Asset Already Allocated
          </DialogTitle>
          <DialogDescription>
            This asset is currently held by <strong>{holderName}</strong>. You cannot allocate it until it is returned.
          </DialogDescription>
        </DialogHeader>
        <div className="rounded-lg bg-muted p-4 text-sm">
          <p className="font-medium">Suggest an action:</p>
          <p className="text-muted-foreground mt-1">
            If you need this asset for someone else, you can request a transfer. The current holder or a manager will need to approve it.
          </p>
        </div>
        <DialogFooter showCloseButton>
          <Button
            onClick={() => onRequestTransfer(allocationId)}
            className="gap-2"
          >
            <ArrowLeftRight className="size-4" />
            Request Transfer
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

export { ConflictModal };
