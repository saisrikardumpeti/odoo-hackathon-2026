import { client } from './client';

export interface Booking {
  id: string;
  resource_asset_id: string;
  booked_by_employee_id: string;
  start_time: string;
  end_time: string;
  purpose: string | null;
  status: 'Upcoming' | 'Ongoing' | 'Completed' | 'Cancelled';
  created_at: string;
  updated_at: string;
}

export interface BookingDetail extends Booking {
  asset_name: string;
  asset_tag: string;
  booked_by_name: string | null;
}

export interface CreateBookingRequest {
  resource_asset_id: string;
  start_time: string;
  end_time: string;
  purpose?: string;
}

export interface RescheduleBookingRequest {
  start_time: string;
  end_time: string;
}

export interface BookingOverlapError {
  error: 'BookingOverlap';
  message: string;
  conflicting_bookings: BookingDetail[];
}

export const listResourceBookings = async (
  assetId: string,
  from?: string,
  to?: string,
): Promise<{ bookings: BookingDetail[] }> => {
  const params: Record<string, string> = {};
  if (from) params.from = from;
  if (to) params.to = to;
  const { data } = await client.get(`/v1/resources/${assetId}/bookings`, { params });
  return data;
};

export const createBooking = async (req: CreateBookingRequest): Promise<{ booking: Booking }> => {
  const { data } = await client.post('/v1/bookings', req);
  return data;
};

export const cancelBooking = async (id: string): Promise<{ message: string }> => {
  const { data } = await client.patch(`/v1/bookings/${id}/cancel`);
  return data;
};

export const rescheduleBooking = async (
  id: string,
  req: RescheduleBookingRequest,
): Promise<{ message: string }> => {
  const { data } = await client.patch(`/v1/bookings/${id}/reschedule`, req);
  return data;
};

export const listMyBookings = async (): Promise<{ bookings: BookingDetail[] }> => {
  const { data } = await client.get('/v1/bookings/my');
  return data;
};
