import { useAllocationSummary } from '#/lib/hooks/useReports';
import { PieChart, Pie, Cell, Tooltip, Legend, ResponsiveContainer } from 'recharts';
import { Skeleton } from '#/components/ui/skeleton';

const COLORS = ['hsl(221.2 83.2% 53.3%)', 'hsl(142.1 76.2% 36.3%)', 'hsl(24.6 95% 53.1%)', 'hsl(262.1 83.3% 57.8%)', 'hsl(0 72.2% 50.6%)'];

function AllocationSummaryChart() {
  const { data, isLoading } = useAllocationSummary();

  if (isLoading) {
    return <Skeleton className="h-80 w-full" />;
  }

  const items = data?.allocation_summary ?? [];
  if (items.length === 0) {
    return <p className="py-12 text-center text-muted-foreground">No allocation data available.</p>;
  }

  return (
    <div className="grid grid-cols-1 lg:grid-cols-2 gap-8">
      <div>
        <ResponsiveContainer width="100%" height={300}>
          <PieChart>
            <Pie
              data={items}
              dataKey="asset_count"
              nameKey="department_name"
              cx="50%"
              cy="50%"
              outerRadius={100}
              label={({ department_name, asset_count }) => `${department_name} (${asset_count})`}
            >
              {items.map((_, idx) => (
                <Cell key={idx} fill={COLORS[idx % COLORS.length]} />
              ))}
            </Pie>
            <Tooltip />
            <Legend />
          </PieChart>
        </ResponsiveContainer>
      </div>

      <div className="space-y-2">
        <h3 className="text-sm font-medium mb-2">Department-wise Allocation</h3>
        <div className="rounded-lg border divide-y">
          {items.map((item) => (
            <div key={item.department_id} className="flex items-center justify-between px-4 py-3">
              <span className="font-medium">{item.department_name}</span>
              <span className="text-sm text-muted-foreground">{item.asset_count} asset(s)</span>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}

export { AllocationSummaryChart };
