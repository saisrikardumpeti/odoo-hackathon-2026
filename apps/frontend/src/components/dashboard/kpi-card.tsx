import { cn } from '#/lib/utils';
import type { LucideIcon } from 'lucide-react';

function KpiCard({
  icon: Icon,
  label,
  value,
  className,
}: {
  icon: LucideIcon;
  label: string;
  value: number | string;
  className?: string;
}) {
  return (
    <div className={cn('flex items-center gap-4 rounded-lg border p-4', className)}>
      <div className="flex size-10 items-center justify-center rounded-md bg-primary/10 text-primary">
        <Icon className="size-5" />
      </div>
      <div>
        <p className="text-sm text-muted-foreground">{label}</p>
        <p className="text-2xl font-bold">{value}</p>
      </div>
    </div>
  );
}

export { KpiCard };
