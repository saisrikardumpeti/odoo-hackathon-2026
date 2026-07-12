import { createFileRoute, Link, redirect } from '@tanstack/react-router';
import { useAuthStore } from '#/lib/stores/authStore';
import { useAssets } from '#/lib/hooks/useAssets';
import { useCategories } from '#/lib/hooks/useCategories';
import { AssetStatusBadge } from '#/components/asset/AssetStatusBadge';
import type { AssetStatus } from '#/components/asset/AssetStatusBadge';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '#/components/ui/table';
import { Input } from '#/components/ui/input';
import { Button } from '#/components/ui/button';
import { Search, Plus, Package } from 'lucide-react';
import { useState } from 'react';

export const Route = createFileRoute('/assets/')({
  beforeLoad: () => {
    if (!useAuthStore.getState().isAuthenticated) {
      throw redirect({ to: '/auth/login' });
    }
  },
  component: AssetsPage,
});

const ASSET_STATUSES: AssetStatus[] = [
  'Available', 'Allocated', 'Reserved', 'UnderMaintenance', 'Lost', 'Retired', 'Disposed',
];

function AssetsPage() {
  const employee = useAuthStore((s) => s.employee);
  const canWrite = employee?.role === 'Admin' || employee?.role === 'AssetManager';

  const [searchTag, setSearchTag] = useState('');
  const [searchSerial, setSearchSerial] = useState('');
  const [filterCategory, setFilterCategory] = useState('');
  const [filterStatus, setFilterStatus] = useState('');
  const [filterLocation, setFilterLocation] = useState('');
  const [page, setPage] = useState(1);

  const { data, isLoading } = useAssets({
    asset_tag: searchTag || undefined,
    serial_number: searchSerial || undefined,
    category_id: filterCategory || undefined,
    status: filterStatus || undefined,
    location: filterLocation || undefined,
    page,
    page_size: 20,
  });

  const { data: categoriesData } = useCategories();

  return (
    <div className="p-8">
      <div className="mb-6 flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold">Asset Directory</h1>
          <p className="text-muted-foreground text-sm">Search and manage all registered assets.</p>
        </div>
        {canWrite && (
          <Link to="/assets/new">
            <Button>
              <Plus className="mr-2 size-4" /> Register Asset
            </Button>
          </Link>
        )}
      </div>

      <div className="mb-4 flex flex-wrap gap-3">
        <div className="relative flex-1 min-w-[200px]">
          <Search className="absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
          <Input
            placeholder="Search by asset tag..."
            value={searchTag}
            onChange={(e) => { setSearchTag(e.target.value); setPage(1); }}
            className="pl-9"
          />
        </div>
        <Input
          placeholder="Search by serial..."
          value={searchSerial}
          onChange={(e) => { setSearchSerial(e.target.value); setPage(1); }}
          className="max-w-[200px]"
        />
        <Input
          placeholder="Search by location..."
          value={filterLocation}
          onChange={(e) => { setFilterLocation(e.target.value); setPage(1); }}
          className="max-w-[200px]"
        />
        <select
          value={filterCategory}
          onChange={(e) => { setFilterCategory(e.target.value); setPage(1); }}
          className="rounded-lg border border-input bg-background px-3 py-2 text-sm outline-none focus-visible:ring-2 focus-visible:ring-ring max-w-[200px]"
        >
          <option value="">All Categories</option>
          {categoriesData?.categories?.map((cat) => (
            <option key={cat.id} value={cat.id}>{cat.name}</option>
          ))}
        </select>
        <select
          value={filterStatus}
          onChange={(e) => { setFilterStatus(e.target.value); setPage(1); }}
          className="rounded-lg border border-input bg-background px-3 py-2 text-sm outline-none focus-visible:ring-2 focus-visible:ring-ring max-w-[160px]"
        >
          <option value="">All Statuses</option>
          {ASSET_STATUSES.map((s) => (
            <option key={s} value={s}>{s}</option>
          ))}
        </select>
      </div>

      {isLoading ? (
        <div className="text-muted-foreground py-8 text-center">Loading assets...</div>
      ) : !data || data.assets.length === 0 ? (
        <div className="text-muted-foreground flex flex-col items-center gap-2 py-16">
          <Package className="size-12" />
          <p className="text-lg font-medium">No assets found</p>
          <p className="text-sm">Try adjusting your search or register a new asset.</p>
        </div>
      ) : (
        <>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Asset Tag</TableHead>
                <TableHead>Name</TableHead>
                <TableHead>Category</TableHead>
                <TableHead>Serial</TableHead>
                <TableHead>Status</TableHead>
                <TableHead>Location</TableHead>
                <TableHead>Bookable</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {data.assets.map((asset) => (
                <TableRow key={asset.id}>
                  <TableCell className="font-mono text-sm">
                    <Link to="/assets/$id" params={{ id: asset.id }} className="text-primary hover:underline">
                      {asset.asset_tag}
                    </Link>
                  </TableCell>
                  <TableCell>{asset.name}</TableCell>
                  <TableCell>{asset.category_name}</TableCell>
                  <TableCell className="text-muted-foreground text-sm">{asset.serial_number ?? '—'}</TableCell>
                  <TableCell><AssetStatusBadge status={asset.status as AssetStatus} /></TableCell>
                  <TableCell className="text-sm">{asset.location ?? '—'}</TableCell>
                  <TableCell>{asset.is_bookable ? 'Yes' : 'No'}</TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>

          <div className="mt-4 flex items-center justify-between text-sm text-muted-foreground">
            <span>{data.total_count} asset{(data.total_count ?? 0) !== 1 ? 's' : ''} found</span>
            <div className="flex gap-2">
              <Button
                variant="outline"
                size="sm"
                disabled={page <= 1}
                onClick={() => setPage((p) => Math.max(1, p - 1))}
              >
                Previous
              </Button>
              <Button
                variant="outline"
                size="sm"
                disabled={data.assets.length < 20}
                onClick={() => setPage((p) => p + 1)}
              >
                Next
              </Button>
            </div>
          </div>
        </>
      )}
    </div>
  );
}
