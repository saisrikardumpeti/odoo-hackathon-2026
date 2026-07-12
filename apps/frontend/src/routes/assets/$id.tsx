import { createFileRoute, redirect, Link } from '@tanstack/react-router';
import { useAuthStore } from '#/lib/stores/authStore';
import { useAsset, useAssetHistory } from '#/lib/hooks/useAssets';
import { AssetStatusBadge } from '#/components/asset/AssetStatusBadge';
import type { AssetStatus } from '#/components/asset/AssetStatusBadge';
import { Badge } from '#/components/ui/badge';
import { Button } from '#/components/ui/button';
import { Skeleton } from '#/components/ui/skeleton';
import { ArrowLeft, Package, Calendar, DollarSign, MapPin, Barcode, Hash, User, Building2, History } from 'lucide-react';

export const Route = createFileRoute('/assets/$id')({
  beforeLoad: () => {
    if (!useAuthStore.getState().isAuthenticated) {
      throw redirect({ to: '/auth/login' });
    }
  },
  component: AssetDetailPage,
});

function AssetDetailPage() {
  const { id } = Route.useParams();
  const employee = useAuthStore((s) => s.employee);
  const { data: assetData, isLoading: assetLoading } = useAsset(id);
  const { data: historyData, isLoading: historyLoading } = useAssetHistory(id);

  const showCost = employee?.role !== 'Employee';

  if (assetLoading) {
    return (
      <div className="p-8 space-y-4">
        <Skeleton className="h-8 w-48" />
        <Skeleton className="h-6 w-96" />
        <Skeleton className="h-64 w-full" />
      </div>
    );
  }

  const asset = assetData?.asset;
  if (!asset) {
    return (
      <div className="p-8 text-center text-muted-foreground">
        <Package className="mx-auto mb-4 size-12" />
        <p>Asset not found</p>
        <Button variant="link" asChild className="mt-2"><Link to="/assets">Back to Directory</Link></Button>
      </div>
    );
  }

  return (
    <div className="p-8">
      <Button variant="ghost" size="sm" asChild className="mb-4">
        <Link to="/assets"><ArrowLeft className="mr-1 size-4" /> Back to Directory</Link>
      </Button>

      <div className="mb-6 flex flex-wrap items-start justify-between gap-4">
        <div>
          <div className="flex items-center gap-3">
            <h1 className="text-2xl font-bold">{asset.name}</h1>
            <AssetStatusBadge status={asset.status as AssetStatus} />
            {asset.is_bookable && <Badge variant="outline">Bookable</Badge>}
          </div>
          <p className="font-mono text-sm text-muted-foreground">{asset.asset_tag}</p>
        </div>
      </div>

      <div className="grid gap-6 md:grid-cols-2">
        <div className="space-y-4 rounded-lg border p-4">
          <h2 className="text-sm font-semibold text-muted-foreground uppercase tracking-wide">Details</h2>

          <div className="space-y-3 text-sm">
            <div className="flex items-center gap-2">
              <Hash className="size-4 text-muted-foreground shrink-0" />
              <span className="text-muted-foreground">Category:</span>
              <span>{asset.category_name ?? asset.category_id}</span>
            </div>
            <div className="flex items-center gap-2">
              <Barcode className="size-4 text-muted-foreground shrink-0" />
              <span className="text-muted-foreground">Serial:</span>
              <span>{asset.serial_number ?? '—'}</span>
            </div>
            <div className="flex items-center gap-2">
              <Calendar className="size-4 text-muted-foreground shrink-0" />
              <span className="text-muted-foreground">Acquired:</span>
              <span>{asset.acquisition_date ?? '—'}</span>
            </div>
            {showCost && (
              <div className="flex items-center gap-2">
                <DollarSign className="size-4 text-muted-foreground shrink-0" />
                <span className="text-muted-foreground">Cost:</span>
                <span>{asset.acquisition_cost != null ? `$${Number(asset.acquisition_cost).toFixed(2)}` : '—'}</span>
              </div>
            )}
            <div className="flex items-center gap-2">
              <MapPin className="size-4 text-muted-foreground shrink-0" />
              <span className="text-muted-foreground">Location:</span>
              <span>{asset.location ?? '—'}</span>
            </div>
            <div className="flex items-center gap-2">
              <Package className="size-4 text-muted-foreground shrink-0" />
              <span className="text-muted-foreground">Condition:</span>
              <span className="capitalize">{asset.condition ?? '—'}</span>
            </div>
          </div>
        </div>

        <div className="space-y-4 rounded-lg border p-4">
          <h2 className="text-sm font-semibold text-muted-foreground uppercase tracking-wide">Holder</h2>

          <div className="space-y-3 text-sm">
            <div className="flex items-center gap-2">
              <User className="size-4 text-muted-foreground shrink-0" />
              <span className="text-muted-foreground">Holder:</span>
              <span>{asset.current_holder_name ?? 'Unallocated'}</span>
            </div>
            <div className="flex items-center gap-2">
              <Building2 className="size-4 text-muted-foreground shrink-0" />
              <span className="text-muted-foreground">Department:</span>
              <span>{asset.current_holder_department_name ?? '—'}</span>
            </div>
          </div>
        </div>
      </div>

      <div className="mt-8">
        <h2 className="mb-4 flex items-center gap-2 text-lg font-semibold">
          <History className="size-5" /> History Timeline
        </h2>

        {historyLoading ? (
          <div className="space-y-3">
            <Skeleton className="h-12 w-full" />
            <Skeleton className="h-12 w-full" />
            <Skeleton className="h-12 w-full" />
          </div>
        ) : !historyData || historyData.history.length === 0 ? (
          <p className="text-muted-foreground text-sm">No history events recorded for this asset.</p>
        ) : (
          <div className="relative space-y-0">
            {historyData.history.map((event, idx) => (
              <div key={idx} className="relative flex gap-4 pb-6 pl-6 last:pb-0">
                <div className="absolute left-[7px] top-2 h-full w-px bg-border last:hidden" />
                <div className={`absolute left-0 top-2 size-3.5 rounded-full border-2 ${
                  event.type === 'status_change' ? 'border-blue-500 bg-blue-50 dark:bg-blue-950' :
                  event.type === 'allocation' ? 'border-green-500 bg-green-50 dark:bg-green-950' :
                  'border-yellow-500 bg-yellow-50 dark:bg-yellow-950'
                }`} />
                <div className="flex-1">
                  <div className="flex items-center gap-2">
                    <Badge variant="outline" className="text-[10px] uppercase">
                      {event.type === 'status_change' ? 'Status' : event.type === 'allocation' ? 'Allocation' : 'Maintenance'}
                    </Badge>
                    <span className="text-xs text-muted-foreground">
                      {new Date(event.timestamp).toLocaleString()}
                    </span>
                  </div>
                  <pre className="mt-1 text-xs text-muted-foreground overflow-x-auto">
                    {JSON.stringify(event.data, null, 2)}
                  </pre>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
