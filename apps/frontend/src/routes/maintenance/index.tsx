import { createFileRoute, Link } from '@tanstack/react-router';
import { useAuthStore } from '#/lib/stores/authStore';
import { MaintenanceKanban } from '#/components/maintenance/MaintenanceKanban';
import { Button } from '#/components/ui/button';
import { Plus } from 'lucide-react';

export const Route = createFileRoute('/maintenance/')({ component: MaintenancePage });

function MaintenancePage() {
  const isManager = useAuthStore((s) => {
    const role = s.employee?.role;
    return role === 'Admin' || role === 'AssetManager';
  });

  return (
    <div className="p-6 space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold">Maintenance</h1>
          <p className="text-sm text-muted-foreground">Manage maintenance requests</p>
        </div>
        <div className="flex items-center gap-2">
          {isManager && (
            <Link to="/maintenance/approvals">
              <Button variant="outline" size="sm">Approvals</Button>
            </Link>
          )}
          <Link to="/maintenance/new">
            <Button size="sm">
              <Plus className="size-4 mr-1" />
              Raise Request
            </Button>
          </Link>
        </div>
      </div>
      <MaintenanceKanban />
    </div>
  );
}
