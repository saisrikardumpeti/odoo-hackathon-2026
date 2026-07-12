import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { listAssets, getAsset, createAsset, getAssetHistory, uploadAssetDocument } from '#/lib/api/assets';
import type { ListAssetsParams, CreateAssetRequest } from '#/lib/api/assets';
import { queryKeys } from './queryKeys';

export const useAssets = (params?: ListAssetsParams) => {
  return useQuery({
    queryKey: queryKeys.assets.list(params as Record<string, unknown> | undefined),
    queryFn: () => listAssets(params),
  });
};

export const useAsset = (id: string) => {
  return useQuery({
    queryKey: queryKeys.assets.detail(id),
    queryFn: () => getAsset(id),
    enabled: !!id,
  });
};

export const useCreateAsset = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (req: CreateAssetRequest) => createAsset(req),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.assets.all });
    },
  });
};

export const useAssetHistory = (id: string) => {
  return useQuery({
    queryKey: queryKeys.assets.history(id),
    queryFn: () => getAssetHistory(id),
    enabled: !!id,
  });
};

export const useUploadAssetDocument = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, file, type }: { id: string; file: File; type: 'photo' | 'document' }) =>
      uploadAssetDocument(id, file, type),
    onSuccess: (_data, variables) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.assets.detail(variables.id) });
    },
  });
};
