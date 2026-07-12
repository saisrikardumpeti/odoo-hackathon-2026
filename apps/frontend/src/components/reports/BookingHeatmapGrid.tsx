import { useBookingHeatmap } from '#/lib/hooks/useReports';
import { Skeleton } from '#/components/ui/skeleton';
import { cn } from '#/lib/utils';
import type { ReportFilters } from '#/lib/api/reports';

const DAY_LABELS = ['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat'];
const HOURS = Array.from({ length: 14 }, (_, i) => i + 8);

function getIntensity(count: number, maxCount: number): string {
  if (maxCount === 0) return 'bg-muted';
  const ratio = count / maxCount;
  if (ratio === 0) return 'bg-muted';
  if (ratio <= 0.25) return 'bg-primary/20';
  if (ratio <= 0.5) return 'bg-primary/40';
  if (ratio <= 0.75) return 'bg-primary/60';
  return 'bg-primary/80';
}

function BookingHeatmapGrid({ filters }: { filters?: ReportFilters }) {
  const { data, isLoading } = useBookingHeatmap(filters);

  if (isLoading) {
    return <Skeleton className="h-80 w-full" />;
  }

  const heatmap = data?.heatmap ?? [];
  if (heatmap.length === 0) {
    return <p className="py-12 text-center text-muted-foreground">No booking data for the selected range.</p>;
  }

  const maxCount = Math.max(...heatmap.map((h) => h.count), 1);

  const grid: Record<string, number> = {};
  for (const item of heatmap) {
    grid[`${item.day_of_week}-${item.hour}`] = item.count;
  }

  return (
    <div className="overflow-x-auto">
      <div className="inline-block min-w-[600px]">
        <div className="grid" style={{ gridTemplateColumns: `60px repeat(${HOURS.length}, 1fr)`, gap: '2px' }}>
          <div className="text-xs text-muted-foreground p-1" />
          {HOURS.map((h) => (
            <div key={h} className="text-xs text-muted-foreground text-center p-1">
              {h.toString().padStart(2, '0')}:00
            </div>
          ))}
          {DAY_LABELS.map((day, dow) => (
            <>
              <div key={day} className="text-xs text-muted-foreground flex items-center p-1 font-medium">
                {day}
              </div>
              {HOURS.map((hour) => {
                const count = grid[`${dow}-${hour}`] ?? 0;
                return (
                  <div
                    key={`${dow}-${hour}`}
                    className={cn(
                      'aspect-square rounded-sm flex items-center justify-center text-xs font-medium',
                      count > 0 ? 'text-primary-foreground' : 'text-muted-foreground',
                      getIntensity(count, maxCount),
                    )}
                    title={`${day} ${hour}:00 - ${count} booking(s)`}
                  >
                    {count > 0 ? count : ''}
                  </div>
                );
              })}
            </>
          ))}
        </div>
      </div>
      <div className="mt-4 flex items-center gap-3 text-xs text-muted-foreground">
        <span>Low</span>
        <div className="flex gap-0.5">
          <div className="size-4 rounded-sm bg-muted" />
          <div className="size-4 rounded-sm bg-primary/20" />
          <div className="size-4 rounded-sm bg-primary/40" />
          <div className="size-4 rounded-sm bg-primary/60" />
          <div className="size-4 rounded-sm bg-primary/80" />
        </div>
        <span>High</span>
      </div>
    </div>
  );
}

export { BookingHeatmapGrid };
