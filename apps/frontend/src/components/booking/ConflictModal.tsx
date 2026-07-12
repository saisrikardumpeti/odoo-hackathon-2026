import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription } from '#/components/ui/dialog';
import { Calendar } from 'lucide-react';
import type { BookingDetail } from '#/lib/api/bookings';

interface ConflictModalProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  conflictingBookings: BookingDetail[];
}

function formatSlot(b: BookingDetail) {
  const start = new Date(b.start_time);
  const end = new Date(b.end_time);
  return `${start.toLocaleDateString()} ${start.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })} – ${end.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}`;
}

function ConflictModal({ open, onOpenChange, conflictingBookings }: ConflictModalProps) {
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <Calendar className="size-5 text-destructive" />
            Time Slot Conflict
          </DialogTitle>
          <DialogDescription>
            This time range overlaps with the following existing booking(s). Please choose a different time.
          </DialogDescription>
        </DialogHeader>
        <div className="space-y-2">
          {conflictingBookings.map((b) => (
            <div key={b.id} className="rounded-lg border border-destructive/20 bg-destructive/5 p-3 text-sm">
              <p className="font-medium text-destructive">{b.asset_name} ({b.asset_tag})</p>
              <p className="text-muted-foreground mt-1">{formatSlot(b)}</p>
              {b.purpose && (
                <p className="text-muted-foreground mt-0.5 text-xs">{b.purpose}</p>
              )}
              <p className="text-muted-foreground mt-0.5 text-xs">
                Booked by: {b.booked_by_name ?? 'Unknown'}
              </p>
            </div>
          ))}
        </div>
      </DialogContent>
    </Dialog>
  );
}

export { ConflictModal };
