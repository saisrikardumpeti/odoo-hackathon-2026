import { createFileRoute, redirect, useNavigate } from '@tanstack/react-router';
import { useAuthStore } from '#/lib/stores/authStore';
import { useCategories } from '#/lib/hooks/useCategories';
import { useCreateAsset } from '#/lib/hooks/useAssets';
import { Button } from '#/components/ui/button';
import { Input } from '#/components/ui/input';
import { Label } from '#/components/ui/label';
import { Loader2, ArrowLeft, CheckCircle } from 'lucide-react';
import { useState, useMemo } from 'react';

export const Route = createFileRoute('/assets/new')({
  beforeLoad: () => {
    const { isAuthenticated, employee } = useAuthStore.getState();
    if (!isAuthenticated) throw redirect({ to: '/auth/login' });
    if (employee?.role !== 'Admin' && employee?.role !== 'AssetManager') {
      throw redirect({ to: '/assets' });
    }
  },
  component: RegisterAssetPage,
});

function RegisterAssetPage() {
  const navigate = useNavigate();
  const { data: categoriesData } = useCategories();
  const createAsset = useCreateAsset();

  const [name, setName] = useState('');
  const [categoryId, setCategoryId] = useState('');
  const [serialNumber, setSerialNumber] = useState('');
  const [acquisitionDate, setAcquisitionDate] = useState('');
  const [acquisitionCost, setAcquisitionCost] = useState('');
  const [condition, setCondition] = useState('');
  const [location, setLocation] = useState('');
  const [isBookable, setIsBookable] = useState(false);
  const [customFieldValues, setCustomFieldValues] = useState<Record<string, string>>({});
  const [createdTag, setCreatedTag] = useState<string | null>(null);

  const selectedCategory = useMemo(() => {
    if (!categoryId || !categoriesData?.categories) return null;
    return categoriesData.categories.find((c) => c.id === categoryId) ?? null;
  }, [categoryId, categoriesData]);

  const categoryCustomFields = useMemo(() => {
    if (!selectedCategory) return [];
    return Object.entries(selectedCategory.custom_fields).map(([key, value]) => ({
      key,
      defaultValue: String(value ?? ''),
      type: detectFieldType(value),
    }));
  }, [selectedCategory]);

  function detectFieldType(value: unknown): 'text' | 'number' | 'date' {
    if (typeof value === 'number') return 'number';
    if (typeof value === 'string' && !isNaN(Date.parse(value)) && value.includes('-')) return 'date';
    return 'text';
  }

  const handleCategoryChange = (id: string) => {
    setCategoryId(id);
    const cat = categoriesData?.categories?.find((c) => c.id === id);
    if (cat) {
      const initial: Record<string, string> = {};
      for (const [key, value] of Object.entries(cat.custom_fields)) {
        initial[key] = String(value ?? '');
      }
      setCustomFieldValues(initial);
    } else {
      setCustomFieldValues({});
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!name.trim() || !categoryId) return;

    const customFields: Record<string, unknown> = {};
    for (const field of categoryCustomFields) {
      const val = customFieldValues[field.key] ?? '';
      if (val === '' && field.defaultValue === '') continue;
      if (field.type === 'number') {
        customFields[field.key] = val ? parseFloat(val) : null;
      } else {
        customFields[field.key] = val || null;
      }
    }

    const result = await createAsset.mutateAsync({
      name: name.trim(),
      category_id: categoryId,
      serial_number: serialNumber.trim() || null,
      acquisition_date: acquisitionDate || null,
      acquisition_cost: acquisitionCost ? parseFloat(acquisitionCost) : null,
      condition: condition.trim() || null,
      location: location.trim() || null,
      is_bookable: isBookable,
      custom_fields: customFields,
    });

    setCreatedTag(result.asset.asset_tag);
  };

  if (createdTag) {
    return (
      <div className="p-8">
        <div className="mx-auto max-w-lg text-center">
          <CheckCircle className="mx-auto mb-4 size-12 text-green-600" />
          <h2 className="text-xl font-bold">Asset Registered</h2>
          <p className="text-muted-foreground mt-2">
            Asset <span className="font-mono font-medium text-foreground">{createdTag}</span> has been registered successfully.
          </p>
          <div className="mt-6 flex justify-center gap-3">
            <Button variant="outline" onClick={() => navigate({ to: '/assets' })}>
              Go to Directory
            </Button>
            <Button onClick={() => { setCreatedTag(null); setName(''); setCategoryId(''); setSerialNumber(''); setAcquisitionDate(''); setAcquisitionCost(''); setCondition(''); setLocation(''); setIsBookable(false); setCustomFieldValues({}); }}>
              Register Another
            </Button>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="p-8">
      <div className="mb-6">
        <Button variant="ghost" size="sm" onClick={() => navigate({ to: '/assets' })} className="mb-2">
          <ArrowLeft className="mr-1 size-4" /> Back to Directory
        </Button>
        <h1 className="text-2xl font-bold">Register Asset</h1>
        <p className="text-muted-foreground text-sm">Register a new asset in the system.</p>
      </div>

      <form onSubmit={handleSubmit} className="max-w-2xl space-y-5">
        <div className="space-y-2">
          <Label htmlFor="name">Asset Name *</Label>
          <Input id="name" value={name} onChange={(e) => setName(e.target.value)} placeholder="e.g. Dell Latitude 5540" required />
        </div>

        <div className="space-y-2">
          <Label htmlFor="category">Category *</Label>
          <select
            id="category"
            value={categoryId}
            onChange={(e) => handleCategoryChange(e.target.value)}
            required
            className="w-full rounded-lg border border-input bg-background px-3 py-2 text-sm outline-none focus-visible:ring-2 focus-visible:ring-ring"
          >
            <option value="">Select a category</option>
            {categoriesData?.categories?.map((cat) => (
              <option key={cat.id} value={cat.id}>{cat.name}</option>
            ))}
          </select>
        </div>

        {categoryCustomFields.length > 0 && (
          <div className="rounded-lg border p-4 space-y-4">
            <h3 className="text-sm font-semibold text-muted-foreground uppercase tracking-wide">
              Category Fields — {selectedCategory?.name}
            </h3>
            {categoryCustomFields.map((field) => (
              <div key={field.key} className="space-y-2">
                <Label htmlFor={`cf-${field.key}`} className="capitalize">
                  {field.key.replace(/_/g, ' ')}
                </Label>
                <Input
                  id={`cf-${field.key}`}
                  type={field.type === 'number' ? 'number' : field.type === 'date' ? 'date' : 'text'}
                  step={field.type === 'number' ? 'any' : undefined}
                  value={customFieldValues[field.key] ?? ''}
                  onChange={(e) => setCustomFieldValues((prev) => ({ ...prev, [field.key]: e.target.value }))}
                  placeholder={field.defaultValue || `Enter ${field.key.replace(/_/g, ' ')}`}
                />
              </div>
            ))}
          </div>
        )}

        <div className="grid grid-cols-2 gap-4">
          <div className="space-y-2">
            <Label htmlFor="serial">Serial Number</Label>
            <Input id="serial" value={serialNumber} onChange={(e) => setSerialNumber(e.target.value)} placeholder="e.g. SN-12345" />
          </div>
          <div className="space-y-2">
            <Label htmlFor="condition">Condition</Label>
            <Input id="condition" value={condition} onChange={(e) => setCondition(e.target.value)} placeholder="e.g. New, Good, Fair" />
          </div>
        </div>

        <div className="grid grid-cols-2 gap-4">
          <div className="space-y-2">
            <Label htmlFor="date">Acquisition Date</Label>
            <Input id="date" type="date" value={acquisitionDate} onChange={(e) => setAcquisitionDate(e.target.value)} />
          </div>
          <div className="space-y-2">
            <Label htmlFor="cost">Acquisition Cost</Label>
            <Input id="cost" type="number" step="0.01" min="0" value={acquisitionCost} onChange={(e) => setAcquisitionCost(e.target.value)} placeholder="e.g. 1299.99" />
          </div>
        </div>

        <div className="space-y-2">
          <Label htmlFor="location">Location</Label>
          <Input id="location" value={location} onChange={(e) => setLocation(e.target.value)} placeholder="e.g. Floor 3, Room 301" />
        </div>

        <label className="flex items-center gap-2 text-sm">
          <input
            type="checkbox"
            checked={isBookable}
            onChange={(e) => setIsBookable(e.target.checked)}
            className="size-4 rounded border-input accent-primary"
          />
          <span>Bookable as shared resource</span>
        </label>

        {createAsset.isError && (
          <p className="text-sm text-destructive">
            {createAsset.error instanceof Error ? createAsset.error.message : 'Failed to register asset'}
          </p>
        )}

        <Button type="submit" disabled={createAsset.isPending}>
          {createAsset.isPending && <Loader2 className="mr-2 size-4 animate-spin" />}
          Register Asset
        </Button>
      </form>
    </div>
  );
}
