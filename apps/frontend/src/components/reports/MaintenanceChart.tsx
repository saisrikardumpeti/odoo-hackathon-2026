import { useMaintenanceFrequency } from '#/lib/hooks/useReports';
import { BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, PieChart, Pie, Cell, Legend } from 'recharts';
import { Skeleton } from '#/components/ui/skeleton';
import type { ReportFilters } from '#/lib/api/reports';

const COLORS = ['hsl(221.2 83.2% 53.3%)', 'hsl(142.1 76.2% 36.3%)', 'hsl(24.6 95% 53.1%)', 'hsl(262.1 83.3% 57.8%)', 'hsl(0 72.2% 50.6%)', 'hsl(187 100% 42%)'];

function MaintenanceChart({ filters }: { filters?: ReportFilters }) {
  const { data, isLoading } = useMaintenanceFrequency(filters);

  if (isLoading) {
    return <Skeleton className="h-80 w-full" />;
  }

  const byCategory = data?.by_category ?? [];
  if (byCategory.length === 0) {
    return <p className="py-12 text-center text-muted-foreground">No maintenance data for the selected range.</p>;
  }

  return (
    <div className="grid grid-cols-1 lg:grid-cols-2 gap-8">
      <div>
        <h3 className="text-sm font-medium mb-3">By Category</h3>
        <ResponsiveContainer width="100%" height={300}>
          <PieChart>
            <Pie
              data={byCategory}
              dataKey="count"
              nameKey="category_name"
              cx="50%"
              cy="50%"
              outerRadius={100}
              label={({ category_name, count }) => `${category_name} (${count})`}
            >
              {byCategory.map((_, idx) => (
                <Cell key={idx} fill={COLORS[idx % COLORS.length]} />
              ))}
            </Pie>
            <Tooltip />
            <Legend />
          </PieChart>
        </ResponsiveContainer>
      </div>

      <div>
        <h3 className="text-sm font-medium mb-3">By Asset (Top 10)</h3>
        <ResponsiveContainer width="100%" height={300}>
          <BarChart
            data={(data?.by_asset ?? []).slice(0, 10).map((i) => ({ name: `${i.asset_tag} - ${i.asset_name}`, count: i.count }))}
            layout="vertical"
            margin={{ left: 100, right: 20, top: 10, bottom: 10 }}
          >
            <CartesianGrid strokeDasharray="3 3" />
            <XAxis type="number" />
            <YAxis type="category" dataKey="name" width={180} tick={{ fontSize: 12 }} />
            <Tooltip />
            <Bar dataKey="count" fill="hsl(24.6 95% 53.1%)" />
          </BarChart>
        </ResponsiveContainer>
      </div>
    </div>
  );
}

export { MaintenanceChart };
