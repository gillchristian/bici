import {calendar} from '@wails/go/models'

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

type Source = 'external' | 'internal'

export type CalendarEvent = {
  id: string
  title: string
  description?: string
  time: EventTime
  color: Color
  weekday: W.WeekDay
  source: Source
}

export const toGo = (event: CalendarEvent): calendar.CalendarEvent => {
  return new calendar.CalendarEvent({
    id: event.id,
    title: event.title,
    description: event.description ?? '',
    time: event.time,
    weekday: W.toGo(event.weekday),
    color: event.color,
    source: event.source
  })
}

export const fromGo = (event: calendar.CalendarEvent): CalendarEvent => {
  return {
    id: event.id,
    title: event.title,
    description: event.description ?? '',
    time: event.time,
    color: event.color as Color,
    weekday: W.fromGo(event.weekday),
    source: event.source as Source
  }
}
