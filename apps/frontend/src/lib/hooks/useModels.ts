import { useQuery } from '@tanstack/react-query';
import { fetchModels } from '../api/models';
import { queryKeys } from './queryKeys';

export const useModels = () => {
  return useQuery({
    queryKey: queryKeys.models,
    queryFn: () => fetchModels(),
  });
};
