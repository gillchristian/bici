import {main} from '@wails/go/models'

import {EventTime} from './event-time'
import * as W from './week-day'

export type Color =
  | 'blue'
  | 'pink'
  | 'indigo'
  | 'orange'
  | 'amber'
  | 'emerald'
  | 'teal'
  | 'cyan'
  | 'purple'
  | 'gray'

export type CalendarEvent = {
  id: string
  title: string
  description?: string
  time: EventTime
  color: Color
  weekday: W.WeekDay
}

export const toGo = (event: CalendarEvent): main.CalendarEvent => {
  return new main.CalendarEvent({
    id: event.id,
    title: event.title,
    description: event.description ?? '',
    time: event.time,
    weekday: W.toGo(event.weekday)
  })
}
