import { useState } from 'react';
import { Users, Pencil, ShieldAlert } from 'lucide-react';
import { useEmployees, useUpdateEmployee, useUpdateEmployeeRole } from '#/lib/hooks/useEmployees';
import { useDepartments } from '#/lib/hooks/useDepartments';
import type { Employee } from '#/lib/api/auth';
import { Button } from '#/components/ui/button';
import { Badge } from '#/components/ui/badge';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '#/components/ui/table';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogFooter,
} from '#/components/ui/dialog';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '#/components/ui/select';
import { Input } from '#/components/ui/input';
import { Label } from '#/components/ui/label';

function EmployeeTab() {
  const [filters, setFilters] = useState<{ department_id: string | null; role: string | null; status: string | null }>({
    department_id: null,
    role: null,
    status: null,
  });

  const { data: deptData } = useDepartments();
  const { data, isLoading, isError } = useEmployees({
    ...(filters.department_id ? { department_id: filters.department_id } : {}),
    ...(filters.role ? { role: filters.role } : {}),
    ...(filters.status ? { status: filters.status } : {}),
  });

  const updateEmployee = useUpdateEmployee();
  const updateRole = useUpdateEmployeeRole();

  const [editModalOpen, setEditModalOpen] = useState(false);
  const [roleModalOpen, setRoleModalOpen] = useState(false);
  const [selectedEmp, setSelectedEmp] = useState<Employee | null>(null);
  const [editForm, setEditForm] = useState({ name: '', status: '' });
  const [newRole, setNewRole] = useState<string | null>(null);

  const departments = deptData?.departments ?? [];
  const employees = data?.employees ?? [];

  const roleColor = (role: string) => {
    switch (role) {
      case 'Admin': return 'destructive' as const;
      case 'DepartmentHead': return 'default' as const;
      case 'AssetManager': return 'secondary' as const;
      default: return 'outline' as const;
    }
  };

  const openEdit = (emp: Employee) => {
    setSelectedEmp(emp);
    setEditForm({ name: emp.name, status: emp.status ?? 'Active' });
    setEditModalOpen(true);
  };

  const openRoleChange = (emp: Employee) => {
    setSelectedEmp(emp);
    setNewRole(null);
    setRoleModalOpen(true);
  };

  const handleEditSave = async () => {
    if (!selectedEmp) return;
    await updateEmployee.mutateAsync({
      id: selectedEmp.id,
      req: {
        name: editForm.name,
        department_id: selectedEmp.department_id,
        status: editForm.status,
      },
    });
    setEditModalOpen(false);
  };

  const handleRoleChange = async () => {
    if (!selectedEmp || !newRole) return;
    await updateRole.mutateAsync({
      id: selectedEmp.id,
      req: { role: newRole as 'DepartmentHead' | 'AssetManager' | 'Employee' },
    });
    setRoleModalOpen(false);
  };

  const getDepartmentName = (deptId: string | null): string => {
    if (!deptId) return '—';
    return departments.find((d) => d.id === deptId)?.name ?? '—';
  };

  return (
    <div>
      <div className="mb-4 flex flex-wrap items-center gap-3">
        <Select
          value={filters.department_id ?? ''}
          onValueChange={(v) => setFilters((f) => ({ ...f, department_id: v || null }))}
        >
          <SelectTrigger className="w-48">
            <SelectValue placeholder="All Departments" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="">All Departments</SelectItem>
            {departments.map((d) => (
              <SelectItem key={d.id} value={d.id}>{d.name}</SelectItem>
            ))}
          </SelectContent>
        </Select>

        <Select
          value={filters.role ?? ''}
          onValueChange={(v) => setFilters((f) => ({ ...f, role: v || null }))}
        >
          <SelectTrigger className="w-40">
            <SelectValue placeholder="All Roles" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="">All Roles</SelectItem>
            <SelectItem value="Admin">Admin</SelectItem>
            <SelectItem value="DepartmentHead">Department Head</SelectItem>
            <SelectItem value="AssetManager">Asset Manager</SelectItem>
            <SelectItem value="Employee">Employee</SelectItem>
          </SelectContent>
        </Select>

        <Select
          value={filters.status ?? ''}
          onValueChange={(v) => setFilters((f) => ({ ...f, status: v || null }))}
        >
          <SelectTrigger className="w-36">
            <SelectValue placeholder="All Statuses" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="">All Statuses</SelectItem>
            <SelectItem value="Active">Active</SelectItem>
            <SelectItem value="Inactive">Inactive</SelectItem>
          </SelectContent>
        </Select>
      </div>

      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>Name</TableHead>
            <TableHead>Email</TableHead>
            <TableHead>Department</TableHead>
            <TableHead>Role</TableHead>
            <TableHead>Status</TableHead>
            <TableHead className="w-40">Actions</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {isLoading ? (
            <TableRow>
              <TableCell colSpan={6} className="text-center text-muted-foreground py-8">
                Loading employees...
              </TableCell>
            </TableRow>
          ) : isError ? (
            <TableRow>
              <TableCell colSpan={6} className="text-center text-destructive py-8">
                Failed to load employees.
              </TableCell>
            </TableRow>
          ) : employees.length === 0 ? (
            <TableRow>
              <TableCell colSpan={6} className="text-center text-muted-foreground py-8">
                No employees found.
              </TableCell>
            </TableRow>
          ) : (
            employees.map((emp) => (
              <TableRow key={emp.id}>
                <TableCell className="font-medium">
                  <div className="flex items-center gap-2">
                    <Users className="size-4 text-muted-foreground" />
                    {emp.name}
                  </div>
                </TableCell>
                <TableCell className="text-muted-foreground">{emp.email}</TableCell>
                <TableCell className="text-muted-foreground">{getDepartmentName(emp.department_id)}</TableCell>
                <TableCell>
                  <Badge variant={roleColor(emp.role)}>{emp.role}</Badge>
                </TableCell>
                <TableCell>
                  <Badge variant={emp.status === 'Active' ? 'default' : 'secondary'}>
                    {emp.status ?? 'Active'}
                  </Badge>
                </TableCell>
                <TableCell>
                  <div className="flex items-center gap-1">
                    <Button variant="ghost" size="sm" onClick={() => openEdit(emp)}>
                      <Pencil className="mr-1 size-3" /> Edit
                    </Button>
                    <Button variant="outline" size="sm" onClick={() => openRoleChange(emp)}>
                      <ShieldAlert className="mr-1 size-3" /> Promote / Role
                    </Button>
                  </div>
                </TableCell>
              </TableRow>
            ))
          )}
        </TableBody>
      </Table>

      <Dialog open={editModalOpen} onOpenChange={setEditModalOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Edit Employee</DialogTitle>
            <DialogDescription>Update employee name and status.</DialogDescription>
          </DialogHeader>
          <div className="grid gap-4 py-4">
            <div className="grid gap-2">
              <Label htmlFor="emp-name">Name</Label>
              <Input id="emp-name" value={editForm.name} onChange={(e) => setEditForm({ ...editForm, name: e.target.value })} />
            </div>
            <div className="grid gap-2">
              <Label htmlFor="emp-status">Status</Label>
              <Select value={editForm.status} onValueChange={(v) => setEditForm({ ...editForm, status: v || 'Active' })}>
                <SelectTrigger id="emp-status">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="Active">Active</SelectItem>
                  <SelectItem value="Inactive">Inactive</SelectItem>
                </SelectContent>
              </Select>
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setEditModalOpen(false)}>Cancel</Button>
            <Button onClick={handleEditSave} disabled={updateEmployee.isPending}>Save</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <Dialog open={roleModalOpen} onOpenChange={setRoleModalOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Change Role — {selectedEmp?.name}</DialogTitle>
            <DialogDescription>
              This action is audited. Select the new role for this employee.
              Current role: <strong>{selectedEmp?.role}</strong>
            </DialogDescription>
          </DialogHeader>
          <div className="grid gap-4 py-4">
            <div className="grid gap-2">
              <Label htmlFor="emp-role">New Role</Label>
              <Select value={newRole ?? ''} onValueChange={(v) => setNewRole(v)}>
                <SelectTrigger id="emp-role">
                  <SelectValue placeholder="Select a role..." />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="DepartmentHead">Department Head</SelectItem>
                  <SelectItem value="AssetManager">Asset Manager</SelectItem>
                  <SelectItem value="Employee">Employee</SelectItem>
                </SelectContent>
              </Select>
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setRoleModalOpen(false)}>Cancel</Button>
            <Button
              variant="default"
              onClick={handleRoleChange}
              disabled={!newRole || newRole === selectedEmp?.role || updateRole.isPending}
            >
              Confirm Role Change
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}

export { EmployeeTab };