import { useQuery } from '@tanstack/react-query';
import { fetchPing } from '../api/ping';
import { queryKeys } from './queryKeys';

export const usePing = () => {
  return useQuery({
    queryKey: queryKeys.ping,
    queryFn: () => fetchPing(),
  });
};
