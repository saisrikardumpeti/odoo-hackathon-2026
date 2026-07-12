import { useState } from 'react';
import { Tags, Plus, Pencil, X, GripVertical } from 'lucide-react';
import { useCategories, useCreateCategory, useUpdateCategory } from '#/lib/hooks/useCategories';
import type { AssetCategory } from '#/lib/api/categories';
import { Button } from '#/components/ui/button';
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

interface CustomField {
  key: string;
  type: 'text' | 'number' | 'date';
  value: string;
}

interface CategoryForm {
  name: string;
  customFields: CustomField[];
}

const emptyForm: CategoryForm = {
  name: '',
  customFields: [],
};

function CategoryTab() {
  const { data, isLoading, isError } = useCategories();
  const createCategory = useCreateCategory();
  const updateCategory = useUpdateCategory();

  const [modalOpen, setModalOpen] = useState(false);
  const [editingCat, setEditingCat] = useState<AssetCategory | null>(null);
  const [form, setForm] = useState<CategoryForm>(emptyForm);

  const categories = data?.categories ?? [];

  const openCreate = () => {
    setEditingCat(null);
    setForm(emptyForm);
    setModalOpen(true);
  };

  const openEdit = (cat: AssetCategory) => {
    setEditingCat(cat);
    const fields: CustomField[] = Object.entries(cat.custom_fields).map(([key, value]) => ({
      key,
      type: detectType(value as string) as CustomField['type'],
      value: String(value ?? ''),
    }));
    setForm({ name: cat.name, customFields: fields });
    setModalOpen(true);
  };

  const detectType = (val: string): string => {
    if (!isNaN(Number(val)) && val !== '') return 'number';
    if (!isNaN(Date.parse(val))) return 'date';
    return 'text';
  };

  const addField = () => {
    setForm((prev) => ({
      ...prev,
      customFields: [...prev.customFields, { key: '', type: 'text', value: '' }],
    }));
  };

  const removeField = (idx: number) => {
    setForm((prev) => ({
      ...prev,
      customFields: prev.customFields.filter((_, i) => i !== idx),
    }));
  };

  const updateField = (idx: number, field: Partial<CustomField>) => {
    setForm((prev) => ({
      ...prev,
      customFields: prev.customFields.map((f, i) => (i === idx ? { ...f, ...field } : f)),
    }));
  };

  const handleSave = async () => {
    const customFields: Record<string, unknown> = {};
    form.customFields.forEach((f) => {
      if (f.key) {
        let val: unknown = f.value;
        if (f.type === 'number') val = Number(f.value);
        else if (f.type === 'date') val = f.value;
        customFields[f.key] = val;
      }
    });

    if (editingCat) {
      await updateCategory.mutateAsync({ id: editingCat.id, req: { name: form.name, custom_fields: customFields } });
    } else {
      await createCategory.mutateAsync({ name: form.name, custom_fields: customFields });
    }
    setModalOpen(false);
  };

  if (isLoading) {
    return <div className="py-8 text-center text-muted-foreground">Loading categories...</div>;
  }

  if (isError) {
    return <div className="py-8 text-center text-destructive">Failed to load categories.</div>;
  }

  return (
    <div>
      <div className="mb-4 flex items-center justify-between">
        <p className="text-sm text-muted-foreground">{categories.length} categor(ies)</p>
        <Button onClick={openCreate} size="sm">
          <Plus className="mr-1 size-4" /> Add Category
        </Button>
      </div>

      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>Name</TableHead>
            <TableHead>Custom Fields</TableHead>
            <TableHead className="w-20">Actions</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {categories.length === 0 ? (
            <TableRow>
              <TableCell colSpan={3} className="text-center text-muted-foreground py-8">
                No categories yet. Click "Add Category" to create one.
              </TableCell>
            </TableRow>
          ) : (
            categories.map((cat) => {
              const fieldEntries = Object.entries(cat.custom_fields);
              return (
                <TableRow key={cat.id}>
                  <TableCell className="font-medium">
                    <div className="flex items-center gap-2">
                      <Tags className="size-4 text-muted-foreground" />
                      {cat.name}
                    </div>
                  </TableCell>
                  <TableCell className="text-muted-foreground text-sm">
                    {fieldEntries.length === 0
                      ? 'No custom fields'
                      : fieldEntries.map(([k, v]) => (
                          <span key={k} className="mr-2 inline-block rounded bg-muted px-2 py-0.5 text-xs">
                            {k}: {String(v ?? '')}
                          </span>
                        ))}
                  </TableCell>
                  <TableCell>
                    <Button variant="ghost" size="icon" onClick={() => openEdit(cat)}>
                      <Pencil className="size-4" />
                    </Button>
                  </TableCell>
                </TableRow>
              );
            })
          )}
        </TableBody>
      </Table>

      <Dialog open={modalOpen} onOpenChange={setModalOpen}>
        <DialogContent className="max-w-lg">
          <DialogHeader>
            <DialogTitle>{editingCat ? 'Edit Category' : 'Create Category'}</DialogTitle>
            <DialogDescription>
              {editingCat ? 'Update the category details below.' : 'Add a new asset category.'}
            </DialogDescription>
          </DialogHeader>
          <div className="grid gap-4 py-4">
            <div className="grid gap-2">
              <Label htmlFor="cat-name">Name</Label>
              <Input
                id="cat-name"
                value={form.name}
                onChange={(e) => setForm({ ...form, name: e.target.value })}
                placeholder="e.g. Electronics"
              />
            </div>

            <div className="grid gap-2">
              <div className="flex items-center justify-between">
                <Label>Custom Fields</Label>
                <Button variant="outline" size="sm" onClick={addField}>
                  <Plus className="mr-1 size-3" /> Add Field
                </Button>
              </div>
              <div className="grid gap-2">
                {form.customFields.length === 0 && (
                  <p className="text-sm text-muted-foreground">No custom fields defined.</p>
                )}
                {form.customFields.map((field, idx) => (
                  <div key={idx} className="flex items-start gap-2 rounded border p-2">
                    <GripVertical className="mt-2 size-4 shrink-0 text-muted-foreground" />
                    <div className="grid flex-1 gap-2">
                      <Input
                        placeholder="Field name"
                        value={field.key}
                        onChange={(e) => updateField(idx, { key: e.target.value })}
                      />
                      <div className="flex gap-2">
                        <Select
                          value={field.type}
                          onValueChange={(v) => updateField(idx, { type: v as CustomField['type'] })}
                        >
                          <SelectTrigger className="w-28">
                            <SelectValue />
                          </SelectTrigger>
                          <SelectContent>
                            <SelectItem value="text">Text</SelectItem>
                            <SelectItem value="number">Number</SelectItem>
                            <SelectItem value="date">Date</SelectItem>
                          </SelectContent>
                        </Select>
                        <Input
                          placeholder="Value"
                          value={field.value}
                          onChange={(e) => updateField(idx, { value: e.target.value })}
                        />
                      </div>
                    </div>
                    <Button variant="ghost" size="icon" className="mt-2 shrink-0" onClick={() => removeField(idx)}>
                      <X className="size-4" />
                    </Button>
                  </div>
                ))}
              </div>
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setModalOpen(false)}>Cancel</Button>
            <Button onClick={handleSave} disabled={!form.name || createCategory.isPending || updateCategory.isPending}>
              {editingCat ? 'Save' : 'Create'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}

export { CategoryTab };