import {useEffect, useCallback, useState, useMemo} from 'react'
import {useAtom} from 'jotai'
import {OnFileDrop} from '@wails/runtime'
import {ImportEvents} from '@wails/go/main/App'
import * as E from 'fp-ts/Either'
import {pipe} from 'fp-ts/function'
import {format} from 'date-fns'

import {ScheduleAtom, EventsAtom} from '@/utils/calendar'
import type {WeekDict} from '@/utils/calendar'
import {WeekDay, WeekDayNum} from '@/utils/week-day'
import {parseSchedule, stringifySchedule} from '@/utils/calendar'
import * as C from '@/utils/calendar-event'
import {WEEK} from '@/models/week'
import {mkWeek} from '@/utils/calendar'

export const useScheduleWatcher = (callback: (day: WeekDay, schedule: string) => void) => {
  const [schedules, _setSchedules] = useAtom(ScheduleAtom)

  useEffect(() => {
    WEEK.forEach((day) => {
      callback(day, schedules[day])
    })
  }, [schedules, callback])
}

export const useFileDropHandler = () => {
  const [schedules, setSchedules] = useAtom(ScheduleAtom)

  useEffect(() => {
    OnFileDrop((_x, _y, paths) => {
      if (!paths || paths.length === 0) {
        console.error('No file path provided')
        return
      }

      console.log({path: paths[0]})

      ImportEvents(paths[0])
        .then((res) => {
          if (!Array.isArray(res)) {
            throw new Error('Failed to import events: invalid response')
          }
          return res
        })
        .then((events) => {
          const newSchedules = WEEK.reduce((acc, day) => {
            const dayEvents = events.filter((e) => e.weekday === WeekDayNum[day]).map(C.fromGo)
            const schedule = [
              ...new Set(
                `${schedules[day]}\n${stringifySchedule(dayEvents)}`
                  .trim()
                  .split('\n')
                  .filter(Boolean)
                  .sort()
              )
            ].join('\n')
            return {...acc, [day]: schedule}
          }, {} as WeekDict<string>)

          console.log({newSchedules})

          setSchedules((prev) => ({...prev, ...newSchedules}))
        })
        .catch((err) => {
          console.error('Error importing events:', err)
        })
    }, false)
  }, [schedules])
}

export const useScheduleParser = () => {
  const [_, setEvents] = useAtom(EventsAtom)
  const [errors, setErrors] = useState<WeekDict<string>>({
    [WeekDay.Sunday]: '',
    [WeekDay.Monday]: '',
    [WeekDay.Tuesday]: '',
    [WeekDay.Wednesday]: '',
    [WeekDay.Thursday]: '',
    [WeekDay.Friday]: '',
    [WeekDay.Saturday]: ''
  })

  const now = useMemo(() => new Date(), [])

  const weekArr = useMemo(() => mkWeek(now), [now])
  const week: WeekDict<Date> = useMemo(
    () =>
      weekArr.reduce((acc, day) => {
        const weekDay = WeekDay[format(day.date, 'EEEE') as keyof typeof WeekDay]
        acc[weekDay] = day.date
        return acc
      }, {} as WeekDict<Date>),
    [weekArr]
  )

  const doParse = useCallback(
    (day: WeekDay, schedule: string) => {
      if (schedule.trim() === '') {
        setEvents((prev: WeekDict<C.CalendarEvent[]>) => ({...prev, [day]: []}))
        setErrors((prev: WeekDict<string>) => ({...prev, [day]: ''}))
        return
      }

      pipe(
        schedule,
        parseSchedule(week[day]),
        E.match(
          (e) => {
            console.error(e)
            setErrors((prev: WeekDict<string>) => ({...prev, [day]: e}))
          },
          (es) => {
            setEvents((prev: WeekDict<C.CalendarEvent[]>) => ({...prev, [day]: es}))
            setErrors((prev: WeekDict<string>) => ({...prev, [day]: ''}))
          }
        )
      )
    },
    [week]
  )

  useScheduleWatcher(doParse)

  return {errors, week}
}
