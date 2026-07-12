import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import {
  listResourceBookings,
  createBooking,
  cancelBooking,
  rescheduleBooking,
  listMyBookings,
} from '#/lib/api/bookings';
import type {
  CreateBookingRequest,
  RescheduleBookingRequest,
} from '#/lib/api/bookings';
import { queryKeys } from './queryKeys';

export const useResourceBookings = (assetId: string, from?: string, to?: string) => {
  return useQuery({
    queryKey: queryKeys.bookings.byResource(assetId, from, to),
    queryFn: () => listResourceBookings(assetId, from, to),
    enabled: !!assetId,
  });
};

export const useCreateBooking = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (req: CreateBookingRequest) => createBooking(req),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.bookings.all });
    },
  });
};

export const useCancelBooking = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => cancelBooking(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.bookings.all });
    },
  });
};

export const useRescheduleBooking = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, req }: { id: string; req: RescheduleBookingRequest }) =>
      rescheduleBooking(id, req),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.bookings.all });
    },
  });
};

export const useMyBookings = () => {
  return useQuery({
    queryKey: queryKeys.bookings.my,
    queryFn: listMyBookings,
  });
};
