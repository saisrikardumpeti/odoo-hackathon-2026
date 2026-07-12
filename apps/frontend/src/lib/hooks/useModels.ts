import { useQuery } from '@tanstack/react-query';
import { fetchModels } from '../api/models';
import { todoKeys } from './queryKeys';

export const useModels = () => {
  return useQuery({
    queryKey: todoKeys.models,
    queryFn: () => fetchModels(),
  });
};
