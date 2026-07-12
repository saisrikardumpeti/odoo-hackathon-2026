import { cva, type VariantProps } from 'class-variance-authority';
import { cn } from '#/lib/utils';

const statusVariants = cva('inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium', {
  variants: {
    status: {
      Available: 'bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400',
      Allocated: 'bg-blue-100 text-blue-800 dark:bg-blue-900/30 dark:text-blue-400',
      Reserved: 'bg-purple-100 text-purple-800 dark:bg-purple-900/30 dark:text-purple-400',
      UnderMaintenance: 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900/30 dark:text-yellow-400',
      Lost: 'bg-red-100 text-red-800 dark:bg-red-900/30 dark:text-red-400',
      Retired: 'bg-orange-100 text-orange-800 dark:bg-orange-900/30 dark:text-orange-400',
      Disposed: 'bg-gray-100 text-gray-800 dark:bg-gray-900/30 dark:text-gray-400',
    },
  },
  defaultVariants: {
    status: 'Available',
  },
});

type StatusVariants = VariantProps<typeof statusVariants>;
type AssetStatus = NonNullable<StatusVariants['status']>;

interface AssetStatusBadgeProps {
  status: AssetStatus;
  className?: string;
}

function AssetStatusBadge({ status, className }: AssetStatusBadgeProps) {
  return (
    <span className={cn(statusVariants({ status }), className)}>
      {status}
    </span>
  );
}

export { AssetStatusBadge };
export type { AssetStatus };
