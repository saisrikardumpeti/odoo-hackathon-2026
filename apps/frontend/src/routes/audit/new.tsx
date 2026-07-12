import { createFileRoute, useNavigate } from '@tanstack/react-router';
import { useCreateAuditCycle, useAssignAuditors } from '#/lib/hooks/useAudit';
import { useDepartments } from '#/lib/hooks/useDepartments';
import { useEmployees } from '#/lib/hooks/useEmployees';
import { Button } from '#/components/ui/button';
import { useState } from 'react';
import { ArrowLeft, Plus, X } from 'lucide-react';

export const Route = createFileRoute('/audit/new')({ component: NewAuditCyclePage });

function NewAuditCyclePage() {
  const navigate = useNavigate();
  const { data: deptData } = useDepartments();
  const { data: empData } = useEmployees();
  const createCycle = useCreateAuditCycle();
  const assignAuditors = useAssignAuditors();

  const [name, setName] = useState('');
  const [scopeDepartmentId, setScopeDepartmentId] = useState('');
  const [scopeLocation, setScopeLocation] = useState('');
  const [startDate, setStartDate] = useState('');
  const [endDate, setEndDate] = useState('');
  const [selectedAuditors, setSelectedAuditors] = useState<string[]>([]);
  const [error, setError] = useState('');

  const departments = deptData?.departments ?? [];
  const employees = empData?.employees ?? [];

  const toggleAuditor = (id: string) => {
    setSelectedAuditors((prev) =>
      prev.includes(id) ? prev.filter((a) => a !== id) : [...prev, id],
    );
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');

    if (!name || !startDate || !endDate) {
      setError('Name, start date, and end date are required');
      return;
    }

    try {
      const result = await createCycle.mutateAsync({
        name,
        scope_department_id: scopeDepartmentId || null,
        scope_location: scopeLocation || null,
        start_date: startDate,
        end_date: endDate,
      });

      if (selectedAuditors.length > 0) {
        await assignAuditors.mutateAsync({
          cycleId: result.audit_cycle.id,
          req: { employee_ids: selectedAuditors },
        });
      }

      navigate({ to: '/audit/cycle/$id', params: { id: result.audit_cycle.id } });
    } catch {
      setError('Failed to create audit cycle');
    }
  };

  return (
    <div className="p-8 max-w-2xl">
      <div className="mb-6">
        <Button variant="ghost" size="sm" onClick={() => navigate({ to: '/audit' })}>
          <ArrowLeft className="size-4 mr-2" />
          Back to Audit
        </Button>
      </div>

      <h1 className="text-2xl font-bold mb-6">New Audit Cycle</h1>

      <form onSubmit={handleSubmit} className="space-y-6">
        <div className="space-y-2">
          <label className="text-sm font-medium">Cycle Name *</label>
          <input
            value={name}
            onChange={(e) => setName(e.target.value)}
            required
            className="w-full rounded-lg border border-input bg-background px-3 py-2 text-sm"
            placeholder="e.g. Q2 2026 Office Audit"
          />
        </div>

        <div className="grid grid-cols-2 gap-4">
          <div className="space-y-2">
            <label className="text-sm font-medium">Scope Department</label>
            <select
              value={scopeDepartmentId}
              onChange={(e) => setScopeDepartmentId(e.target.value)}
              className="w-full rounded-lg border border-input bg-background px-3 py-2 text-sm"
            >
              <option value="">All Departments</option>
              {departments.map((dept) => (
                <option key={dept.id} value={dept.id}>{dept.name}</option>
              ))}
            </select>
          </div>
          <div className="space-y-2">
            <label className="text-sm font-medium">Scope Location</label>
            <input
              value={scopeLocation}
              onChange={(e) => setScopeLocation(e.target.value)}
              className="w-full rounded-lg border border-input bg-background px-3 py-2 text-sm"
              placeholder="e.g. Building A"
            />
          </div>
        </div>

        <div className="grid grid-cols-2 gap-4">
          <div className="space-y-2">
            <label className="text-sm font-medium">Start Date *</label>
            <input
              type="date"
              value={startDate}
              onChange={(e) => setStartDate(e.target.value)}
              required
              className="w-full rounded-lg border border-input bg-background px-3 py-2 text-sm"
            />
          </div>
          <div className="space-y-2">
            <label className="text-sm font-medium">End Date *</label>
            <input
              type="date"
              value={endDate}
              onChange={(e) => setEndDate(e.target.value)}
              required
              className="w-full rounded-lg border border-input bg-background px-3 py-2 text-sm"
            />
          </div>
        </div>

        <div className="space-y-2">
          <label className="text-sm font-medium">Assign Auditors</label>
          <div className="max-h-48 overflow-y-auto rounded-lg border p-2 space-y-1">
            {employees.map((emp) => (
              <label key={emp.id} className="flex items-center gap-2 px-2 py-1.5 rounded hover:bg-muted cursor-pointer text-sm">
                <input
                  type="checkbox"
                  checked={selectedAuditors.includes(emp.id)}
                  onChange={() => toggleAuditor(emp.id)}
                  className="rounded"
                />
                <span>{emp.name}</span>
                <span className="text-xs text-muted-foreground ml-auto">{emp.role}</span>
              </label>
            ))}
          </div>
          {selectedAuditors.length > 0 && (
            <div className="flex flex-wrap gap-1 mt-2">
              {selectedAuditors.map((id) => {
                const emp = employees.find((e) => e.id === id);
                return (
                  <span key={id} className="inline-flex items-center gap-1 rounded-full bg-secondary px-2 py-0.5 text-xs">
                    {emp?.name}
                    <button type="button" onClick={() => toggleAuditor(id)} className="hover:text-destructive">
                      <X className="size-3" />
                    </button>
                  </span>
                );
              })}
            </div>
          )}
        </div>

        {error && <p className="text-sm text-destructive">{error}</p>}

        <Button type="submit" disabled={createCycle.isPending}>
          {createCycle.isPending ? 'Creating...' : 'Create Audit Cycle'}
        </Button>
      </form>
    </div>
  );
}
