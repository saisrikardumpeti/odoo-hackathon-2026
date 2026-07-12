import { useState } from 'react';
import { createFileRoute, redirect } from '@tanstack/react-router';
import { useAuthStore } from '#/lib/stores/authStore';
import { useMyBookings, useCancelBooking, useRescheduleBooking } from '#/lib/hooks/useBookings';
import { BookingStatusBadge } from '#/components/booking/BookingStatusBadge';
import { ConflictModal } from '#/components/booking/ConflictModal';
import { Button } from '#/components/ui/button';
import { Input } from '#/components/ui/input';
import { Label } from '#/components/ui/label';
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from '#/components/ui/dialog';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '#/components/ui/table';
import { Calendar, XCircle, RefreshCw, Loader2, CalendarCheck } from 'lucide-react';
import type { BookingDetail } from '#/lib/api/bookings';

export const Route = createFileRoute('/my-bookings')({
  beforeLoad: () => {
    const { isAuthenticated } = useAuthStore.getState();
    if (!isAuthenticated) throw redirect({ to: '/auth/login' });
  },
  component: MyBookingsPage,
});

function MyBookingsPage() {
  const { data, isLoading } = useMyBookings();
  const cancelMutation = useCancelBooking();
  const rescheduleMutation = useRescheduleBooking();

  const [rescheduleTarget, setRescheduleTarget] = useState<BookingDetail | null>(null);
  const [newStart, setNewStart] = useState('');
  const [newEnd, setNewEnd] = useState('');
  const [conflictOpen, setConflictOpen] = useState(false);
  const [conflictBookings, setConflictBookings] = useState<BookingDetail[]>([]);

  const bookings = data?.bookings ?? [];

  const handleCancel = async (id: string) => {
    await cancelMutation.mutateAsync(id);
  };

  const openReschedule = (b: BookingDetail) => {
    setRescheduleTarget(b);
    setNewStart(new Date(b.start_time).toISOString().slice(0, 16));
    setNewEnd(new Date(b.end_time).toISOString().slice(0, 16));
  };

  const handleReschedule = async () => {
    if (!rescheduleTarget || !newStart || !newEnd) return;
    try {
      await rescheduleMutation.mutateAsync({
        id: rescheduleTarget.id,
        req: {
          start_time: new Date(newStart).toISOString(),
          end_time: new Date(newEnd).toISOString(),
        },
      });
      setRescheduleTarget(null);
    } catch (err: unknown) {
      const axiosErr = err as {
        response?: { status?: number; data?: { error?: string; conflicting_bookings?: BookingDetail[] } };
      };
      if (axiosErr?.response?.status === 409 && axiosErr?.response?.data?.error === 'BookingOverlap') {
        setConflictBookings(axiosErr.response.data.conflicting_bookings ?? []);
        setConflictOpen(true);
      }
    }
  };

  const canCancel = (b: BookingDetail) => b.status === 'Upcoming' || b.status === 'Ongoing';
  const canReschedule = (b: BookingDetail) => b.status === 'Upcoming';

  return (
    <div className="p-8">
      <div className="mb-6">
        <h1 className="text-2xl font-bold flex items-center gap-2">
          <CalendarCheck className="size-6" /> My Bookings
        </h1>
        <p className="text-muted-foreground text-sm">View and manage your resource bookings.</p>
      </div>

      {isLoading ? (
        <div className="flex items-center justify-center py-16">
          <Loader2 className="size-8 animate-spin text-muted-foreground" />
        </div>
      ) : bookings.length === 0 ? (
        <div className="flex flex-col items-center justify-center py-20 text-muted-foreground">
          <Calendar className="size-12 mb-4 opacity-30" />
          <p className="text-lg">No bookings found.</p>
          <p className="text-sm">Create a booking from the Resource Booking page.</p>
        </div>
      ) : (
        <div className="rounded-lg border overflow-hidden">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Resource</TableHead>
                <TableHead>Date & Time</TableHead>
                <TableHead>Purpose</TableHead>
                <TableHead>Status</TableHead>
                <TableHead className="text-right">Actions</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {bookings.map((b) => (
                <TableRow key={b.id}>
                  <TableCell className="font-medium">
                    <span className="font-mono text-xs">{b.asset_tag}</span>
                    <br />
                    <span className="text-sm">{b.asset_name}</span>
                  </TableCell>
                  <TableCell>
                    <div className="text-sm">
                      {new Date(b.start_time).toLocaleDateString()}
                    </div>
                    <div className="text-xs text-muted-foreground">
                      {new Date(b.start_time).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}
                      {' – '}
                      {new Date(b.end_time).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}
                    </div>
                  </TableCell>
                  <TableCell className="text-sm text-muted-foreground">{b.purpose ?? '—'}</TableCell>
                  <TableCell>
                    <BookingStatusBadge status={b.status as 'Upcoming' | 'Ongoing' | 'Completed' | 'Cancelled'} />
                  </TableCell>
                  <TableCell className="text-right">
                    <div className="flex items-center justify-end gap-1">
                      {canReschedule(b) && (
                        <Button variant="ghost" size="icon" onClick={() => openReschedule(b)} title="Reschedule">
                          <RefreshCw className="size-4" />
                        </Button>
                      )}
                      {canCancel(b) && (
                        <Button
                          variant="ghost"
                          size="icon"
                          onClick={() => handleCancel(b.id)}
                          disabled={cancelMutation.isPending}
                          title="Cancel"
                        >
                          <XCircle className="size-4 text-destructive" />
                        </Button>
                      )}
                    </div>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </div>
      )}

      <Dialog open={!!rescheduleTarget} onOpenChange={(o) => { if (!o) setRescheduleTarget(null); }}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Reschedule Booking</DialogTitle>
          </DialogHeader>
          <div className="space-y-3 py-2">
            <div>
              <Label>New Start Time</Label>
              <Input
                type="datetime-local"
                value={newStart}
                onChange={(e) => setNewStart(e.target.value)}
              />
            </div>
            <div>
              <Label>New End Time</Label>
              <Input
                type="datetime-local"
                value={newEnd}
                onChange={(e) => setNewEnd(e.target.value)}
              />
            </div>
          </div>
          <DialogFooter showCloseButton>
            <Button onClick={handleReschedule} disabled={rescheduleMutation.isPending || !newStart || !newEnd}>
              {rescheduleMutation.isPending && <Loader2 className="mr-2 size-4 animate-spin" />}
              Save Changes
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <ConflictModal
        open={conflictOpen}
        onOpenChange={setConflictOpen}
        conflictingBookings={conflictBookings}
      />
    </div>
  );
}
