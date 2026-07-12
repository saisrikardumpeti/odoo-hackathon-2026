import { useState } from 'react';
import { Building2, Plus, Pencil, ToggleLeft } from 'lucide-react';
import { useDepartments, useCreateDepartment, useUpdateDepartment, useDeactivateDepartment } from '#/lib/hooks/useDepartments';
import type { Department } from '#/lib/api/departments';
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

interface DepartmentForm {
  name: string;
  parent_department_id: string | null;
  head_employee_id: string | null;
}

const emptyForm: DepartmentForm = {
  name: '',
  parent_department_id: null,
  head_employee_id: null,
};

function DepartmentTab() {
  const { data, isLoading, isError } = useDepartments();
  const createDepartment = useCreateDepartment();
  const updateDepartment = useUpdateDepartment();
  const deactivateDepartment = useDeactivateDepartment();

  const [modalOpen, setModalOpen] = useState(false);
  const [editingDept, setEditingDept] = useState<Department | null>(null);
  const [form, setForm] = useState<DepartmentForm>(emptyForm);
  const [confirmDeactivate, setConfirmDeactivate] = useState<Department | null>(null);
  const [deactivateError, setDeactivateError] = useState<string | null>(null);

  const departments = data?.departments ?? [];

  const parentOptions = departments.filter((d) => d.id !== editingDept?.id);

  const openCreate = () => {
    setEditingDept(null);
    setForm(emptyForm);
    setModalOpen(true);
  };

  const openEdit = (dept: Department) => {
    setEditingDept(dept);
    setForm({
      name: dept.name,
      parent_department_id: dept.parent_department_id,
      head_employee_id: dept.head_employee_id,
    });
    setModalOpen(true);
  };

  const handleSave = async () => {
    if (editingDept) {
      await updateDepartment.mutateAsync({ id: editingDept.id, req: form });
    } else {
      await createDepartment.mutateAsync(form);
    }
    setModalOpen(false);
  };

  const handleDeactivate = async (force: boolean) => {
    if (!confirmDeactivate) return;
    setDeactivateError(null);
    try {
      const result = await deactivateDepartment.mutateAsync({ id: confirmDeactivate.id, force });
      if ('error' in result && result.requires_confirmation) {
        setDeactivateError(result.error ?? null);
        return;
      }
      setConfirmDeactivate(null);
    } catch (err: unknown) {
      const axiosErr = err as { response?: { data?: { error?: string; requires_confirmation?: boolean } } };
      if (axiosErr?.response?.data?.requires_confirmation) {
        setDeactivateError(axiosErr.response.data.error ?? null);
      } else {
        setConfirmDeactivate(null);
      }
    }
  };

  if (isLoading) {
    return <div className="py-8 text-center text-muted-foreground">Loading departments...</div>;
  }

  if (isError) {
    return <div className="py-8 text-center text-destructive">Failed to load departments.</div>;
  }

  const getParentName = (parentId: string | null): string => {
    if (!parentId) return '—';
    const parent = departments.find((d) => d.id === parentId);
    return parent?.name ?? '—';
  };

  return (
    <div>
      <div className="mb-4 flex items-center justify-between">
        <p className="text-sm text-muted-foreground">{departments.length} department(s)</p>
        <Button onClick={openCreate} size="sm">
          <Plus className="mr-1 size-4" /> Add Department
        </Button>
      </div>

      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>Name</TableHead>
            <TableHead>Parent</TableHead>
            <TableHead>Status</TableHead>
            <TableHead className="w-32">Actions</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {departments.length === 0 ? (
            <TableRow>
              <TableCell colSpan={4} className="text-center text-muted-foreground py-8">
                No departments yet. Click "Add Department" to create one.
              </TableCell>
            </TableRow>
          ) : (
            departments.map((dept) => (
              <TableRow key={dept.id}>
                <TableCell className="font-medium">
                  <div className="flex items-center gap-2">
                    <Building2 className="size-4 text-muted-foreground" />
                    {dept.name}
                  </div>
                </TableCell>
                <TableCell className="text-muted-foreground">{getParentName(dept.parent_department_id)}</TableCell>
                <TableCell>
                  <Badge variant={dept.status === 'Active' ? 'default' : 'secondary'}>
                    {dept.status}
                  </Badge>
                </TableCell>
                <TableCell>
                  <div className="flex items-center gap-1">
                    <Button variant="ghost" size="icon" onClick={() => openEdit(dept)}>
                      <Pencil className="size-4" />
                    </Button>
                    {dept.status === 'Active' && (
                      <Button variant="ghost" size="icon" onClick={() => setConfirmDeactivate(dept)}>
                        <ToggleLeft className="size-4" />
                      </Button>
                    )}
                  </div>
                </TableCell>
              </TableRow>
            ))
          )}
        </TableBody>
      </Table>

      <Dialog open={modalOpen} onOpenChange={setModalOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{editingDept ? 'Edit Department' : 'Create Department'}</DialogTitle>
            <DialogDescription>
              {editingDept ? 'Update the department details below.' : 'Add a new department to the organization.'}
            </DialogDescription>
          </DialogHeader>
          <div className="grid gap-4 py-4">
            <div className="grid gap-2">
              <Label htmlFor="dept-name">Name</Label>
              <Input
                id="dept-name"
                value={form.name}
                onChange={(e) => setForm({ ...form, name: e.target.value })}
                placeholder="e.g. Engineering"
              />
            </div>
            <div className="grid gap-2">
              <Label htmlFor="dept-parent">Parent Department</Label>
              <Select
                value={form.parent_department_id ?? '_none'}
                onValueChange={(v) => setForm({ ...form, parent_department_id: v === '_none' ? null : v })}
              >
                <SelectTrigger id="dept-parent">
                  <SelectValue placeholder="None (top-level)" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="_none">None (top-level)</SelectItem>
                  {parentOptions.map((d) => (
                    <SelectItem key={d.id} value={d.id}>{d.name}</SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setModalOpen(false)}>Cancel</Button>
            <Button onClick={handleSave} disabled={!form.name || createDepartment.isPending || updateDepartment.isPending}>
              {editingDept ? 'Save' : 'Create'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <Dialog open={confirmDeactivate !== null} onOpenChange={(o) => { if (!o) { setConfirmDeactivate(null); setDeactivateError(null); } }}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Deactivate Department</DialogTitle>
            <DialogDescription>
              {deactivateError
                ? 'This department has active employees. Deactivating may affect their records. Confirm to proceed.'
                : `Are you sure you want to deactivate "${confirmDeactivate?.name}"?`}
            </DialogDescription>
          </DialogHeader>
          {deactivateError && (
            <p className="text-sm text-destructive">{deactivateError}</p>
          )}
          <DialogFooter>
            <Button variant="outline" onClick={() => { setConfirmDeactivate(null); setDeactivateError(null); }}>
              Cancel
            </Button>
            <Button variant="destructive" onClick={() => handleDeactivate(deactivateError ? true : false)}>
              {deactivateError ? 'Force Deactivate' : 'Deactivate'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}

export { DepartmentTab };