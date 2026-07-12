import type { BookingDetail } from '#/lib/api/bookings';
import { cn } from '#/lib/utils';

interface BookingCalendarProps {
  bookings: BookingDetail[];
  selectedDate: Date;
  onSlotSelect: (start: string, end: string) => void;
}

const HOURS = Array.from({ length: 14 }, (_, i) => i + 7);
const DAYS = ['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat'];

function getWeekDates(date: Date): Date[] {
  const start = new Date(date);
  start.setDate(start.getDate() - start.getDay());
  start.setHours(0, 0, 0, 0);
  return Array.from({ length: 7 }, (_, i) => {
    const d = new Date(start);
    d.setDate(d.getDate() + i);
    return d;
  });
}

function isSameDay(a: Date, b: Date) {
  return a.getFullYear() === b.getFullYear()
    && a.getMonth() === b.getMonth()
    && a.getDate() === b.getDate();
}

function isSlotBooked(bookings: BookingDetail[], day: Date, hour: number): BookingDetail | undefined {
  return bookings.find((b) => {
    const bStart = new Date(b.start_time);
    const bEnd = new Date(b.end_time);
    if (!isSameDay(bStart, day)) return false;
    const slotStart = new Date(day);
    slotStart.setHours(hour, 0, 0, 0);
    const slotEnd = new Date(day);
    slotEnd.setHours(hour + 1, 0, 0, 0);
    return slotStart.getTime() < bEnd.getTime() && slotEnd.getTime() > bStart.getTime();
  });
}

function BookingCalendar({ bookings, selectedDate, onSlotSelect }: BookingCalendarProps) {
  const weekDates = getWeekDates(selectedDate);

  return (
    <div className="overflow-x-auto">
      <div className="grid grid-cols-[3.5rem_repeat(7,minmax(7rem,1fr))] border rounded-lg text-sm">
        <div className="sticky left-0 bg-background z-10 border-b p-2 font-medium text-muted-foreground text-center text-xs">
          Time
        </div>
        {weekDates.map((d, i) => (
          <div
            key={i}
            className={cn(
              'border-b border-l p-2 text-center font-medium text-xs',
              isSameDay(d, new Date()) && 'bg-primary/5 text-primary',
            )}
          >
            <div>{DAYS[i]}</div>
            <div className="text-muted-foreground">{d.getDate()}</div>
          </div>
        ))}

        {HOURS.map((hour) => (
          <>
            <div key={`label-${hour}`} className="sticky left-0 bg-background z-10 border-b p-1 text-[10px] text-muted-foreground text-center leading-none">
              {hour.toString().padStart(2, '0')}:00
            </div>
            {weekDates.map((day, dayIdx) => {
              const booking = isSlotBooked(bookings, day, hour);
              const slotStart = new Date(day);
              slotStart.setHours(hour, 0, 0, 0);
              const slotEnd = new Date(day);
              slotEnd.setHours(hour + 1, 0, 0, 0);

              return (
                <div
                  key={`${hour}-${dayIdx}`}
                  className={cn(
                    'border-b border-l min-h-[2.5rem] cursor-pointer hover:ring-2 hover:ring-primary/40 transition-all',
                    booking && (
                      booking.status === 'Ongoing'
                        ? 'bg-green-100 dark:bg-green-900/40'
                        : booking.status === 'Upcoming'
                          ? 'bg-blue-100 dark:bg-blue-900/40'
                          : 'bg-gray-100 dark:bg-gray-800/40'
                    ),
                  )}
                  onClick={() => onSlotSelect(slotStart.toISOString(), slotEnd.toISOString())}
                  title={booking ? `${booking.asset_name} (${booking.status}): ${booking.purpose ?? ''}` : `Click to book ${hour}:00-${hour + 1}:00`}
                >
                  {booking && (
                    <div className="px-1 py-0.5 text-[10px] leading-tight truncate">
                      <span className="font-medium">{booking.asset_name}</span>
                      {booking.purpose && <span>: {booking.purpose}</span>}
                    </div>
                  )}
                </div>
              );
            })}
          </>
        ))}
      </div>
    </div>
  );
}

export { BookingCalendar };
