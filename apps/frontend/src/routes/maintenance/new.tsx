import { createFileRoute } from '@tanstack/react-router';
import { RaiseMaintenanceForm } from '#/components/maintenance/RaiseMaintenanceForm';

export const Route = createFileRoute('/maintenance/new')({ component: NewMaintenancePage });

function NewMaintenancePage() {
  return (
    <div className="min-h-full flex items-center justify-center p-8">
      <div className="w-full max-w-2xl space-y-8">
        <div className="text-center">
          <h1 className="text-3xl font-bold">Raise Maintenance Request</h1>
          <p className="text-muted-foreground mt-1">Submit a new maintenance request for an asset</p>
        </div>
        <RaiseMaintenanceForm />
      </div>
    </div>
  );
}
