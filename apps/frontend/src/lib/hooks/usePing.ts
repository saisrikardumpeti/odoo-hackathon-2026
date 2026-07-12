import { useQuery } from '@tanstack/react-query';
import { fetchPing } from '../api/ping';
import { todoKeys } from './queryKeys';

export const usePing = () => {
  return useQuery({
    queryKey: todoKeys.ping,
    queryFn: () => fetchPing(),
  });
};
