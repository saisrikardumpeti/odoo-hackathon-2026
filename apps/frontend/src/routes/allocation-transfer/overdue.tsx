import { createFileRoute, redirect, Link } from '@tanstack/react-router';
import { useAuthStore } from '#/lib/stores/authStore';
import { OverdueAllocations } from '#/components/allocation/OverdueAllocations';
import { Button } from '#/components/ui/button';
import { ArrowLeft, AlertTriangle } from 'lucide-react';

export const Route = createFileRoute('/allocation-transfer/overdue')({
  beforeLoad: () => {
    if (!useAuthStore.getState().isAuthenticated) {
      throw redirect({ to: '/auth/login' });
    }
  },
  component: OverduePage,
});

function OverduePage() {
  return (
    <div className="p-8">
      <div className="mb-6 flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold flex items-center gap-2">
            <AlertTriangle className="size-6 text-destructive" /> Overdue Allocations
          </h1>
          <p className="text-muted-foreground text-sm">Assets that are past their expected return date.</p>
        </div>
        <Link to="/allocation-transfer">
          <Button variant="ghost">
            <ArrowLeft className="mr-1 size-4" /> Back
          </Button>
        </Link>
      </div>
      <OverdueAllocations />
    </div>
  );
}
