import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { listCategories, createCategory, updateCategory } from '#/lib/api/categories';
import type { CreateCategoryRequest, UpdateCategoryRequest } from '#/lib/api/categories';
import { queryKeys } from './queryKeys';

export const useCategories = () => {
  return useQuery({
    queryKey: queryKeys.categories.all,
    queryFn: () => listCategories(),
  });
};

export const useCreateCategory = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (req: CreateCategoryRequest) => createCategory(req),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.categories.all });
    },
  });
};

export const useUpdateCategory = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, req }: { id: string; req: UpdateCategoryRequest }) => updateCategory(id, req),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.categories.all });
    },
  });
};
