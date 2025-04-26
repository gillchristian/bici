import {format} from 'date-fns'

import type {CalendarEvent} from '@/utils/calendar'
import {WeekDayNum} from '@/utils/calendar'
import {clsxm} from '@/utils/clsxm'

// TODO:
//
// - Gray out past events
// - Show end time and/or duration?
// - Overlapping events?
// - Handle events from the previous day
export function CalendarEvent({
  event,
  mode = 'day'
}: {
  event: CalendarEvent
  mode?: 'day' | 'week'
}) {
  const start = new Date(event.time.start)
  const end = new Date(event.time.end)

  const startRow = timeToGridRow(start)
  const endRow = timeToGridRow(end)
  const span = Math.max(3, endRow - startRow)
  const rows = Math.floor(span / 3)

  return (
    <li
      className={clsxm(
        'relative mt-px sm:flex',
        mode === 'week' && `sm:col-start-${WeekDayNum[event.weekday]}`
      )}
      style={{gridRow: `${startRow} / span ${span}`}}
    >
      <div
        className={clsxm(
          'group absolute inset-1 flex flex-col overflow-y-hidden rounded-lg text-xs leading-normal',
          rows === 1 && 'px-2 py-0 justify-center',
          rows > 1 && 'p-2',
          `bg-${event.color}-50 hover:bg-${event.color}-100`
        )}
      >
        {rows > 1 ? (
          <>
            <p className={clsxm('order-1 font-semibold', `text-${event.color}-700`)}>
              {event.title}
            </p>

            {event.description && rows > 2 ? (
              <p
                className={clsxm(
                  'order-1',
                  `text-${event.color}-500 group-hover:text-${event.color}-700`
                )}
              >
                {event.description}
              </p>
            ) : null}

            <p className={clsxm(`text-${event.color}-500 group-hover:text-${event.color}-700`)}>
              <time dateTime={start.toISOString()}>{format(start, 'h:mm a')}</time>
            </p>
          </>
        ) : (
          <p className="flex gap-2">
            <span className={clsxm(`text-${event.color}-500 group-hover:text-${event.color}-700`)}>
              <time dateTime={start.toISOString()}>{format(start, 'h:mm a')}</time>
            </span>
            <span className={clsxm('order-1 font-semibold', `text-${event.color}-700`)}>
              {event.title}
            </span>
          </p>
        )}
      </div>
    </li>
  )
}

// 288 rows / 24 hours = 12 rows per hour
//
// 12 rows / 4 quarters = 3 rows per quarter
//
// +2 to account for the header offset
function timeToGridRow(time: Date): number {
  return time.getHours() * 12 + Math.floor(time.getMinutes() / 15) * 3 + 2
}
