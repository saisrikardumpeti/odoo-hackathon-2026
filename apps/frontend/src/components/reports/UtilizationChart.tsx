import { useUtilizationReport } from '#/lib/hooks/useReports';
import {
  BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, Legend,
} from 'recharts';
import { Skeleton } from '#/components/ui/skeleton';
import type { ReportFilters } from '#/lib/api/reports';

function UtilizationChart({ filters }: { filters?: ReportFilters }) {
  const { data, isLoading } = useUtilizationReport(filters);

  if (isLoading) {
    return <Skeleton className="h-80 w-full" />;
  }

  const items = data?.utilization ?? [];
  if (items.length === 0) {
    return <p className="py-12 text-center text-muted-foreground">No utilization data for the selected range.</p>;
  }

  const chartData = items.slice(0, 20).map((item) => ({
    name: `${item.asset_tag} - ${item.asset_name}`,
    Allocations: item.allocation_count,
    Bookings: item.booking_count,
    idle: item.days_idle ?? 0,
  }));

  return (
    <div className="space-y-4">
      <div className="text-sm text-muted-foreground">
        {items.filter((i) => i.days_idle !== null).length > 0 && (
          <span className="text-amber-600 dark:text-amber-400">
            ⚠ {items.filter((i) => i.days_idle !== null).length} asset(s) idle
          </span>
        )}
      </div>
      <ResponsiveContainer width="100%" height={400}>
        <BarChart data={chartData} layout="vertical" margin={{ left: 100, right: 20, top: 10, bottom: 10 }}>
          <CartesianGrid strokeDasharray="3 3" />
          <XAxis type="number" />
          <YAxis type="category" dataKey="name" width={180} tick={{ fontSize: 12 }} />
          <Tooltip />
          <Legend />
          <Bar dataKey="Allocations" fill="hsl(221.2 83.2% 53.3%)" />
          <Bar dataKey="Bookings" fill="hsl(142.1 76.2% 36.3%)" />
        </BarChart>
      </ResponsiveContainer>
    </div>
  );
}

export { UtilizationChart };
