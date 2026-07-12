import { useState, useCallback } from 'react';
import { createFileRoute, redirect } from '@tanstack/react-router';
import { useAuthStore } from '#/lib/stores/authStore';
import { useAssets } from '#/lib/hooks/useAssets';
import { useResourceBookings, useCreateBooking } from '#/lib/hooks/useBookings';
import { BookingCalendar } from '#/components/booking/BookingCalendar';
import { ConflictModal } from '#/components/booking/ConflictModal';
import { Button } from '#/components/ui/button';
import { Input } from '#/components/ui/input';
import { Label } from '#/components/ui/label';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '#/components/ui/select';
import { Calendar, ChevronLeft, ChevronRight, Loader2 } from 'lucide-react';
import type { BookingDetail } from '#/lib/api/bookings';

export const Route = createFileRoute('/resource-booking')({
  beforeLoad: () => {
    const { isAuthenticated } = useAuthStore.getState();
    if (!isAuthenticated) throw redirect({ to: '/auth/login' });
  },
  component: ResourceBookingPage,
});

function ResourceBookingPage() {
  const [selectedAssetId, setSelectedAssetId] = useState('');
  const [weekOffset, setWeekOffset] = useState(0);
  const [startTime, setStartTime] = useState('');
  const [endTime, setEndTime] = useState('');
  const [purpose, setPurpose] = useState('');
  const [overlapWarning, setOverlapWarning] = useState<string | null>(null);
  const [conflictOpen, setConflictOpen] = useState(false);
  const [conflictBookings, setConflictBookings] = useState<BookingDetail[]>([]);

  const today = new Date();
  const weekStart = new Date(today);
  weekStart.setDate(weekStart.getDate() + weekOffset * 7);
  const weekEnd = new Date(weekStart);
  weekEnd.setDate(weekEnd.getDate() + 7);

  const { data: assetsData } = useAssets({ is_bookable: 'true', page_size: 100 });
  const { data: bookingsData } = useResourceBookings(
    selectedAssetId,
    weekStart.toISOString(),
    weekEnd.toISOString(),
  );
  const createBookingMutation = useCreateBooking();

  const assets = assetsData?.assets ?? [];
  const bookings = bookingsData?.bookings ?? [];
  const selectedAsset = assets.find((a) => a.id === selectedAssetId);

  const checkOverlap = useCallback(
    (start: string, end: string): boolean => {
      if (!start || !end || bookings.length === 0) return false;
      const s = new Date(start).getTime();
      const e = new Date(end).getTime();
      for (const b of bookings) {
        const bs = new Date(b.start_time).getTime();
        const be = new Date(b.end_time).getTime();
        if (s < be && e > bs) {
          return true;
        }
      }
      return false;
    },
    [bookings],
  );

  const handleSlotSelect = (start: string, end: string) => {
    setStartTime(start);
    setEndTime(end);
    if (checkOverlap(start, end)) {
      setOverlapWarning('This time slot overlaps with an existing booking.');
    } else {
      setOverlapWarning(null);
    }
  };

  const handleCreateBooking = async () => {
    if (!selectedAssetId || !startTime || !endTime) return;
    try {
      await createBookingMutation.mutateAsync({
        resource_asset_id: selectedAssetId,
        start_time: startTime,
        end_time: endTime,
        purpose: purpose || undefined,
      });
      setStartTime('');
      setEndTime('');
      setPurpose('');
      setOverlapWarning(null);
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

  const hasOverlap = overlapWarning !== null;

  return (
    <div className="p-8">
      <div className="mb-6">
        <h1 className="text-2xl font-bold">Resource Booking</h1>
        <p className="text-muted-foreground text-sm">Select a resource and book a time slot.</p>
      </div>

      <div className="flex flex-wrap items-end gap-4 mb-6">
        <div className="flex-1 min-w-[200px]">
          <Label htmlFor="resource">Resource</Label>
          <Select value={selectedAssetId} onValueChange={(v) => { setSelectedAssetId(v ?? ''); setWeekOffset(0); }}>
            <SelectTrigger id="resource" className="w-full">
              <SelectValue placeholder="Select a bookable resource">
                {selectedAsset ? `${selectedAsset.asset_tag} - ${selectedAsset.name}` : null}
              </SelectValue>
            </SelectTrigger>
            <SelectContent>
              {assets.length === 0 && (
                <SelectItem value="" disabled>No bookable resources</SelectItem>
              )}
              {assets.map((a) => (
                <SelectItem key={a.id} value={a.id}>
                  <span className="font-mono">{a.asset_tag}</span> - {a.name}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>

        <div className="flex items-center gap-2">
          <Button variant="outline" size="icon" onClick={() => setWeekOffset(weekOffset - 1)}>
            <ChevronLeft className="size-4" />
          </Button>
          <span className="text-sm font-medium whitespace-nowrap w-40 text-center">
            {weekStart.toLocaleDateString()} – {weekEnd.toLocaleDateString()}
          </span>
          <Button variant="outline" size="icon" onClick={() => setWeekOffset(weekOffset + 1)}>
            <ChevronRight className="size-4" />
          </Button>
          <Button variant="ghost" size="sm" onClick={() => setWeekOffset(0)}>Today</Button>
        </div>
      </div>

      {selectedAssetId ? (
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
          <div className="lg:col-span-2">
            <h2 className="text-lg font-semibold mb-3 flex items-center gap-2">
              <Calendar className="size-5" /> Weekly Calendar
            </h2>
            <BookingCalendar
              bookings={bookings}
              selectedDate={weekStart}
              onSlotSelect={handleSlotSelect}
            />
          </div>

          <div className="space-y-4">
            <h2 className="text-lg font-semibold">New Booking</h2>
            <div className="space-y-3 rounded-lg border p-4">
              <div>
                <Label htmlFor="start">Start Time</Label>
                <Input
                  id="start"
                  type="datetime-local"
                  value={startTime ? new Date(startTime).toISOString().slice(0, 16) : ''}
                  onChange={(e) => {
                    const val = e.target.value;
                    if (val) {
                      const iso = new Date(val).toISOString();
                      setStartTime(iso);
                      if (endTime) {
                        if (checkOverlap(iso, endTime)) {
                          setOverlapWarning('This time slot overlaps with an existing booking.');
                        } else {
                          setOverlapWarning(null);
                        }
                      }
                    }
                  }}
                />
              </div>
              <div>
                <Label htmlFor="end">End Time</Label>
                <Input
                  id="end"
                  type="datetime-local"
                  value={endTime ? new Date(endTime).toISOString().slice(0, 16) : ''}
                  onChange={(e) => {
                    const val = e.target.value;
                    if (val) {
                      const iso = new Date(val).toISOString();
                      setEndTime(iso);
                      if (startTime) {
                        if (checkOverlap(startTime, iso)) {
                          setOverlapWarning('This time slot overlaps with an existing booking.');
                        } else {
                          setOverlapWarning(null);
                        }
                      }
                    }
                  }}
                />
              </div>
              <div>
                <Label htmlFor="purpose">Purpose</Label>
                <Input
                  id="purpose"
                  placeholder="e.g. Team meeting"
                  value={purpose}
                  onChange={(e) => setPurpose(e.target.value)}
                />
              </div>

              {overlapWarning && (
                <p className="text-sm text-destructive flex items-center gap-1">
                  <Calendar className="size-3 shrink-0" />
                  {overlapWarning}
                </p>
              )}

              <Button
                onClick={handleCreateBooking}
                disabled={!selectedAssetId || !startTime || !endTime || hasOverlap || createBookingMutation.isPending}
                className="w-full"
              >
                {createBookingMutation.isPending && <Loader2 className="mr-2 size-4 animate-spin" />}
                Book Slot
              </Button>

              {createBookingMutation.isSuccess && (
                <p className="text-sm text-green-600 dark:text-green-400">Booking created successfully!</p>
              )}
            </div>
          </div>
        </div>
      ) : (
        <div className="flex flex-col items-center justify-center py-20 text-muted-foreground">
          <Calendar className="size-12 mb-4 opacity-30" />
          <p className="text-lg">Select a resource to view its booking calendar.</p>
        </div>
      )}

      <ConflictModal
        open={conflictOpen}
        onOpenChange={setConflictOpen}
        conflictingBookings={conflictBookings}
      />
    </div>
  );
}
