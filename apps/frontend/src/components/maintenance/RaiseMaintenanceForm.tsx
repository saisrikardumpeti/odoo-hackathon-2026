import { useState } from 'react';
import { useNavigate } from '@tanstack/react-router';
import { useCreateMaintenance } from '#/lib/hooks/useMaintenance';
import { useAssets } from '#/lib/hooks/useAssets';
import { Button } from '#/components/ui/button';
import { Input } from '#/components/ui/input';
import { Textarea } from '#/components/ui/textarea';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '#/components/ui/select';
import type { MaintenancePriority } from '#/lib/api/maintenance';

function RaiseMaintenanceForm() {
  const navigate = useNavigate();
  const { data: assetsData } = useAssets();
  const createMutation = useCreateMaintenance();

  const [assetId, setAssetId] = useState('');
  const [issueDescription, setIssueDescription] = useState('');
  const [priority, setPriority] = useState<MaintenancePriority>('Medium');
  const [error, setError] = useState('');

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');

    if (!assetId || !issueDescription.trim()) {
      setError('Asset and issue description are required');
      return;
    }

    try {
      await createMutation.mutateAsync({
        asset_id: assetId,
        issue_description: issueDescription.trim(),
        priority,
      });
      navigate({ to: '/maintenance' });
    } catch {
      setError('Failed to create maintenance request');
    }
  };

  return (
    <form onSubmit={handleSubmit} className="space-y-6 max-w-lg">
      <div className="space-y-2">
        <label className="text-sm font-medium">Asset</label>
        <Select value={assetId} onValueChange={setAssetId}>
          <SelectTrigger>
            <SelectValue placeholder="Select an asset" />
          </SelectTrigger>
          <SelectContent>
            {assetsData?.assets?.map((asset) => (
              <SelectItem key={asset.id} value={asset.id}>
                {asset.asset_tag} - {asset.name}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      </div>

      <div className="space-y-2">
        <label className="text-sm font-medium">Issue Description</label>
        <Textarea
          value={issueDescription}
          onChange={(e) => setIssueDescription(e.target.value)}
          placeholder="Describe the issue..."
          rows={4}
        />
      </div>

      <div className="space-y-2">
        <label className="text-sm font-medium">Priority</label>
        <Select value={priority} onValueChange={(v: MaintenancePriority) => setPriority(v)}>
          <SelectTrigger>
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="Low">Low</SelectItem>
            <SelectItem value="Medium">Medium</SelectItem>
            <SelectItem value="High">High</SelectItem>
            <SelectItem value="Critical">Critical</SelectItem>
          </SelectContent>
        </Select>
      </div>

      {error && <p className="text-sm text-destructive">{error}</p>}

      <div className="flex gap-2">
        <Button type="submit" disabled={createMutation.isPending}>
          {createMutation.isPending ? 'Submitting...' : 'Submit Request'}
        </Button>
        <Button type="button" variant="outline" onClick={() => navigate({ to: '/maintenance' })}>
          Cancel
        </Button>
      </div>
    </form>
  );
}

export { RaiseMaintenanceForm };
