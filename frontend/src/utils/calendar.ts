import {
  endOfMonth,
  format,
  getDay,
  isSameDay,
  isSameMonth,
  isToday,
  nextDay,
  previousDay,
  startOfMonth
} from 'date-fns'
import * as E from 'fp-ts/Either'
import * as A from 'fp-ts/Array'
import * as O from 'fp-ts/Option'
import {sequenceT} from 'fp-ts/Apply'
import {pipe} from 'fp-ts/function'
import {run as parse} from 'parser-ts/code-frame'
import * as C from 'parser-ts/char'
import * as S from 'parser-ts/string'
import * as P from 'parser-ts/Parser'

import {CalendarEvent, Color} from './calendar-event'
import {WeekDay} from './week-day'

export type Day = {
  date: Date
  isCurrentMonth: boolean
  isToday: boolean
  isSelected: boolean
}

const firstDayOfWeek = (date: Date) => (getDay(date) === 0 ? date : previousDay(date, 0))

const lastDayOfWeek = (date: Date) => (getDay(date) === 6 ? date : nextDay(date, 6))

export const mkMonth = (now: Date): Day[] => {
  const start = firstDayOfWeek(startOfMonth(now))
  const end = lastDayOfWeek(endOfMonth(now))

  const days: Day[] = []

  for (let dt = new Date(start); dt <= end; dt.setDate(dt.getDate() + 1)) {
    days.push({
      date: new Date(dt),
      isCurrentMonth: isSameMonth(dt, now),
      isToday: isToday(dt),
      isSelected: isSameDay(dt, now)
    })
  }

  return days
}

export const mkWeek = (now: Date): Day[] => {
  const start = firstDayOfWeek(now)
  const end = lastDayOfWeek(now)

  const days: Day[] = []

  for (let dt = new Date(start); dt <= end; dt.setDate(dt.getDate() + 1)) {
    days.push({
      date: new Date(dt),
      isCurrentMonth: isSameMonth(dt, now),
      isToday: isToday(dt),
      isSelected: isSameDay(dt, now)
    })
  }

  return days
}

const COLORS = {
  blue: null,
  pink: null,
  indigo: null,
  orange: null,
  amber: null,
  emerald: null,
  teal: null,
  cyan: null,
  purple: null,
  gray: null
}

const isColor = (color: string): color is Color => color in COLORS

const sequenceP = sequenceT(P.Applicative)

const space = C.char(' ')
const spaces1 = C.many1(space)
const newline = C.char('\n')
const dash = C.char('-')
const slash = C.char('/')
const colon = C.char(':')
const hash = C.char('#')
const word = S.many1(C.notOneOf(' \n\t#/'))
const words = pipe(
  P.sepBy1(spaces1, word),
  P.map((ws) => ws.join(' '))
)
const color: P.Parser<string, Color> = pipe(
  S.oneOf(A.Traversable)(Object.keys(COLORS)),
  P.filter(isColor)
)

const timeNumber = pipe(
  C.many1(C.digit),
  P.map((ds) => parseInt(ds, 10))
)

const time = pipe(
  sequenceP(timeNumber, colon, timeNumber),
  P.map(([h, _c, m]) => [h, m] as const)
)

const range = pipe(
  sequenceP(time, S.spaces, dash, S.spaces, time),
  P.map(([fh, _s1, _d, _s2, th]) => [fh, th] as const)
)

const maybeDescription = pipe(
  P.optional(sequenceP(spaces1, slash, spaces1, words)),
  P.map(O.map(([_s1, _t, _s2, d]) => d)),
  P.map(O.toUndefined)
)

const maybeColor = pipe(
  P.optional(sequenceP(spaces1, hash, color)),
  P.map(O.map(([_s, _h, c]) => c)),
  P.map(O.getOrElse(() => 'blue' as Color))
)

const event = (today: Date) =>
  pipe(
    sequenceP(range, spaces1, words, maybeDescription, maybeColor),
    P.map(([[from, to], _s1, title, description, color]) =>
      mkEvent(today)(title, from, to, color, description)
    )
  )

const events = (today: Date) =>
  pipe(P.sepBy1(P.many1(newline), event(today)), P.apFirst(P.eof<string>()))

const mkEvent =
  (day: Date) =>
  (
    title: string,
    [fh, fm]: readonly [number, number],
    [th, tm]: readonly [number, number],
    color: Color,
    description?: string
  ): CalendarEvent => ({
    id: title,
    title,
    description,
    time: {
      start: new Date(day.getFullYear(), day.getMonth(), day.getDate(), fh, fm).toISOString(),
      end: new Date(day.getFullYear(), day.getMonth(), day.getDate(), th, tm).toISOString()
    },
    color,
    weekday: getWeekDay(day),
    source: 'internal'
  })

export const parseSchedule =
  (today: Date) =>
  (schedule: string): E.Either<string, CalendarEvent[]> =>
    parse(events(today), schedule.trim())

export const stringifySchedule = (events: CalendarEvent[]): string => {
  return events
    .map((event) => {
      const start = new Date(event.time.start)
      const end = new Date(event.time.end)
      const timeStr = `${start.getHours()}:${start.getMinutes().toString().padStart(2, '0')} - ${end.getHours()}:${end.getMinutes().toString().padStart(2, '0')}`
      const title = event.title
      const description = event.description ? ` / ${event.description}` : ''
      const color = event.color !== 'blue' ? ` #${event.color}` : ''
      return `${timeStr} ${title}${description}${color}`
    })
    .join('\n')
}

export type EventTime = {
  start: string
  end: string
}

export const getWeekDay = (date: Date): WeekDay =>
  WeekDay[format(date, 'EEEE') as keyof typeof WeekDay]

export type WeekDict<T> = {
  [key in WeekDay]: T
}
