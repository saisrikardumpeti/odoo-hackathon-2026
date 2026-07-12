import { createFileRoute } from '@tanstack/react-router';
import { RaiseMaintenanceForm } from '#/components/maintenance/RaiseMaintenanceForm';

export const Route = createFileRoute('/maintenance/new')({ component: NewMaintenancePage });

function NewMaintenancePage() {
  return (
    <div className="p-6 space-y-6">
      <div>
        <h1 className="text-2xl font-bold">Raise Maintenance Request</h1>
        <p className="text-sm text-muted-foreground">Submit a new maintenance request for an asset</p>
      </div>
      <RaiseMaintenanceForm />
    </div>
  );
}
