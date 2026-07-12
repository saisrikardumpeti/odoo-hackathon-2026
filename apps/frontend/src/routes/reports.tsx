import { useState } from 'react';
import { createFileRoute, redirect } from '@tanstack/react-router';
import { useAuthStore } from '#/lib/stores/authStore';
import {
  useUtilizationReport,
  useMaintenanceFrequency,
  useRetirementWatchlist,
  useAllocationSummary,
  useBookingHeatmap,
} from '#/lib/hooks/useReports';
import { UtilizationChart } from '#/components/reports/UtilizationChart';
import { MaintenanceChart } from '#/components/reports/MaintenanceChart';
import { AllocationSummaryChart } from '#/components/reports/AllocationSummaryChart';
import { BookingHeatmapGrid } from '#/components/reports/BookingHeatmapGrid';
import { Tabs, TabsList, TabsTrigger, TabsContent } from '#/components/ui/tabs';
import { Input } from '#/components/ui/input';
import { Label } from '#/components/ui/label';
import { Button } from '#/components/ui/button';
import { Skeleton } from '#/components/ui/skeleton';
import { Download, Loader2 } from 'lucide-react';
import { downloadReportCSV } from '#/lib/api/reports';
import type { ReportFilters } from '#/lib/api/reports';

export const Route = createFileRoute('/reports')({
  beforeLoad: () => {
    const { isAuthenticated } = useAuthStore.getState();
    if (!isAuthenticated) throw redirect({ to: '/auth/login' });
  },
  component: ReportsPage,
});

const REPORT_TYPES = [
  { value: 'utilization', label: 'Utilization', exportType: 'utilization' },
  { value: 'maintenance', label: 'Maintenance Frequency', exportType: 'maintenance-frequency' },
  { value: 'allocation', label: 'Allocation Summary', exportType: 'allocation-summary' },
  { value: 'retirement', label: 'Retirement Watchlist', exportType: 'retirement-watchlist' },
  { value: 'heatmap', label: 'Booking Heatmap', exportType: 'booking-heatmap' },
] as const;

function ReportsPage() {
  const [activeTab, setActiveTab] = useState('utilization');
  const [fromDate, setFromDate] = useState('');
  const [toDate, setToDate] = useState('');
  const [exporting, setExporting] = useState(false);

  const filters: ReportFilters = {};
  if (fromDate) filters.from = new Date(fromDate).toISOString();
  if (toDate) filters.to = new Date(toDate).toISOString();

  const currentReport = REPORT_TYPES.find((r) => r.value === activeTab);

  const handleExport = async () => {
    setExporting(true);
    try {
      await downloadReportCSV(currentReport?.exportType ?? 'utilization', filters);
    } finally {
      setExporting(false);
    }
  };

  return (
    <div className="p-6 space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold">Reports & Analytics</h1>
          <p className="text-sm text-muted-foreground">Operational insights across all modules</p>
        </div>
      </div>

      <div className="flex flex-wrap items-end gap-4">
        <div>
          <Label htmlFor="from">From</Label>
          <Input id="from" type="date" value={fromDate} onChange={(e) => setFromDate(e.target.value)} className="w-44" />
        </div>
        <div>
          <Label htmlFor="to">To</Label>
          <Input id="to" type="date" value={toDate} onChange={(e) => setToDate(e.target.value)} className="w-44" />
        </div>
        <Button variant="outline" size="sm" onClick={handleExport} disabled={!currentReport || exporting}>
          {exporting ? <Loader2 className="size-4 mr-1 animate-spin" /> : <Download className="size-4 mr-1" />}
          Export CSV
        </Button>
      </div>

      <Tabs value={activeTab} onValueChange={setActiveTab}>
        <TabsList>
          {REPORT_TYPES.map((r) => (
            <TabsTrigger key={r.value} value={r.value}>{r.label}</TabsTrigger>
          ))}
        </TabsList>
        <TabsContent value="utilization" className="pt-6">
          <div className="rounded-lg border p-6">
            <h2 className="text-lg font-semibold mb-4">Asset Utilization Trends</h2>
            <UtilizationChart filters={filters} />
          </div>
        </TabsContent>
        <TabsContent value="maintenance" className="pt-6">
          <div className="rounded-lg border p-6">
            <h2 className="text-lg font-semibold mb-4">Maintenance Frequency</h2>
            <MaintenanceChart filters={filters} />
          </div>
        </TabsContent>
        <TabsContent value="allocation" className="pt-6">
          <div className="rounded-lg border p-6">
            <h2 className="text-lg font-semibold mb-4">Department Allocation Summary</h2>
            <AllocationSummaryChart />
          </div>
        </TabsContent>
        <TabsContent value="retirement" className="pt-6">
          <div className="rounded-lg border p-6">
            <h2 className="text-lg font-semibold mb-4">Assets Due for Retirement Review</h2>
            <RetirementWatchlistView filters={filters} />
          </div>
        </TabsContent>
        <TabsContent value="heatmap" className="pt-6">
          <div className="rounded-lg border p-6">
            <h2 className="text-lg font-semibold mb-4">Resource Booking Heatmap</h2>
            <BookingHeatmapGrid filters={filters} />
          </div>
        </TabsContent>
      </Tabs>
    </div>
  );
}

function RetirementWatchlistView({ filters }: { filters?: ReportFilters }) {
  const { data, isLoading } = useRetirementWatchlist(filters);

  if (isLoading) {
    return <Skeleton className="h-48 w-full" />;
  }

  const items = data?.retirement_watchlist ?? [];
  if (items.length === 0) {
    return <p className="py-8 text-center text-muted-foreground">No assets nearing retirement threshold.</p>;
  }

  return (
    <div className="overflow-x-auto">
      <table className="w-full text-sm">
        <thead>
          <tr className="border-b text-left">
            <th className="pb-2 font-medium">Asset</th>
            <th className="pb-2 font-medium">Category</th>
            <th className="pb-2 font-medium">Age (years)</th>
            <th className="pb-2 font-medium">Status</th>
          </tr>
        </thead>
        <tbody>
          {items.map((item) => (
            <tr key={item.asset_id} className="border-b last:border-0">
              <td className="py-2">{item.asset_tag} - {item.asset_name}</td>
              <td className="py-2 text-muted-foreground">{item.category_name}</td>
              <td className="py-2">{item.age_years?.toFixed(1) ?? '-'}</td>
              <td className="py-2">{item.status}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
