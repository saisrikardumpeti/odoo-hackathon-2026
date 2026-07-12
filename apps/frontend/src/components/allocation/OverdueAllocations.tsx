import { useOverdueAllocations } from '#/lib/hooks/useAllocations';
import { Badge } from '#/components/ui/badge';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '#/components/ui/table';
import { AlertTriangle, Loader2 } from 'lucide-react';
import { Link } from '@tanstack/react-router';

function OverdueAllocations({ compact = false }: { compact?: boolean }) {
  const { data, isLoading } = useOverdueAllocations();

  if (isLoading) {
    return (
      <div className="flex items-center gap-2 text-sm text-muted-foreground py-4">
        <Loader2 className="size-4 animate-spin" />
        Loading overdue allocations...
      </div>
    );
  }

  const allocations = data?.allocations ?? [];

  if (allocations.length === 0) {
    return (
      <div className="flex items-center gap-2 text-sm text-muted-foreground py-4">
        <AlertTriangle className="size-4" />
        No overdue allocations
      </div>
    );
  }

  if (compact) {
    return (
      <div className="space-y-2">
        <div className="flex items-center gap-2">
          <AlertTriangle className="size-4 text-destructive" />
          <span className="text-sm font-medium">{allocations.length} overdue allocation{allocations.length > 1 ? 's' : ''}</span>
        </div>
        <ul className="space-y-1">
          {allocations.slice(0, 5).map((a) => (
            <li key={a.id} className="flex items-center justify-between text-sm">
              <Link to="/assets/$id" params={{ id: a.asset_id }} className="text-primary hover:underline font-mono">
                {a.asset_tag}
              </Link>
              <span className="text-muted-foreground">{a.employee_name ?? a.department_name ?? 'Unknown'}</span>
            </li>
          ))}
        </ul>
        {allocations.length > 5 && (
          <Link to="/allocation-transfer/overdue" className="text-sm text-primary hover:underline block">
            View all {allocations.length} overdue
          </Link>
        )}
      </div>
    );
  }

  return (
    <div className="flex justify-center">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>Asset</TableHead>
            <TableHead>Holder</TableHead>
            <TableHead>Expected Return</TableHead>
            <TableHead>Status</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {allocations.map((a) => (
            <TableRow key={a.id}>
              <TableCell>
                <Link to="/assets/$id" params={{ id: a.asset_id }} className="text-primary hover:underline font-mono">
                  {a.asset_tag}
                </Link>
                <div className="text-xs text-muted-foreground">{a.asset_name}</div>
              </TableCell>
              <TableCell>{a.employee_name ?? a.department_name ?? '—'}</TableCell>
              <TableCell>{a.expected_return_date ?? '—'}</TableCell>
              <TableCell><Badge variant="destructive">Overdue</Badge></TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </div>
  );
}

export { OverdueAllocations };
