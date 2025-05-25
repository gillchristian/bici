import {useCallback, useEffect, useMemo, useState} from 'react'
import {HomeIcon} from '@heroicons/react/24/outline'
import {useAtom} from 'jotai'
import * as E from 'fp-ts/Either'
import {pipe} from 'fp-ts/function'

import {OnFileDrop} from '@wails/runtime'
import {ImportEvents} from '@wails/go/main/App'

import type {WeekDict} from '@/utils/calendar'
import {ScheduleAtom, EventsAtom} from '@/utils/calendar'
import {WeekDay, WeekDayNum} from '@/utils/week-day'
import {clsxm} from '@/utils/clsxm'
import {parseSchedule, stringifySchedule} from '@/utils/calendar'
import * as C from '@/utils/calendar-event'
import {getDay, addDays, format} from 'date-fns'
import {ViewAtom} from '@/utils/router'

const WEEK = [
  WeekDay.Sunday,
  WeekDay.Monday,
  WeekDay.Tuesday,
  WeekDay.Wednesday,
  WeekDay.Thursday,
  WeekDay.Friday,
  WeekDay.Saturday
]

const getWeekDay = (now: Date, day: WeekDayNum): Date => {
  const currentDay = getDay(now)
  const daysToAdd = (day - currentDay + 7) % 7
  return addDays(now, daysToAdd)
}

export const Settings = () => {
  const [schedules, setSchedules] = useAtom(ScheduleAtom)
  const [_, setEvents] = useAtom(EventsAtom)
  const [__, setView] = useAtom(ViewAtom)

  const now = useMemo(() => new Date(), [])

  const week: WeekDict<Date> = useMemo(
    () => ({
      [WeekDay.Sunday]: getWeekDay(now, WeekDayNum.Sunday),
      [WeekDay.Monday]: getWeekDay(now, WeekDayNum.Monday),
      [WeekDay.Tuesday]: getWeekDay(now, WeekDayNum.Tuesday),
      [WeekDay.Wednesday]: getWeekDay(now, WeekDayNum.Wednesday),
      [WeekDay.Thursday]: getWeekDay(now, WeekDayNum.Thursday),
      [WeekDay.Friday]: getWeekDay(now, WeekDayNum.Friday),
      [WeekDay.Saturday]: getWeekDay(now, WeekDayNum.Saturday)
    }),
    []
  )

  const [errors, setErrors] = useState<WeekDict<string>>({
    [WeekDay.Sunday]: '',
    [WeekDay.Monday]: '',
    [WeekDay.Tuesday]: '',
    [WeekDay.Wednesday]: '',
    [WeekDay.Thursday]: '',
    [WeekDay.Friday]: '',
    [WeekDay.Saturday]: ''
  })

  const doParse = useCallback((day: WeekDay, schedule: string) => {
    if (schedule.trim() === '') {
      setEvents((prev) => ({...prev, [day]: []}))
      setErrors((prev) => ({...prev, [day]: ''}))
      return
    }

    pipe(
      schedule,
      parseSchedule(week[day]),
      E.match(
        (e) => {
          console.error(e)
          setErrors((prev) => ({...prev, [day]: e}))
        },
        (es) => {
          setEvents((prev) => ({...prev, [day]: es}))
          setErrors((prev) => ({...prev, [day]: ''}))
        }
      )
    )
  }, [])

  useEffect(() => {
    OnFileDrop((_x, _y, paths) => {
      console.log({path: paths[0]})

      ImportEvents(paths[0])
        .then((res) =>
          Array.isArray(res) ? res : Promise.reject(new Error('Failed to import events'))
        )
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

          setSchedules((prev) => ({...prev, ...newSchedules}))
        })
        .catch((err) => {
          console.error(err)
        })
    }, false)
  }, [])

  // TODO: DRY !!!
  useEffect(() => {
    doParse(WeekDay.Sunday, schedules.Sunday)
  }, [schedules.Sunday])

  useEffect(() => {
    doParse(WeekDay.Monday, schedules.Monday)
  }, [schedules.Monday])

  useEffect(() => {
    doParse(WeekDay.Tuesday, schedules.Tuesday)
  }, [schedules.Tuesday])

  useEffect(() => {
    doParse(WeekDay.Wednesday, schedules.Wednesday)
  }, [schedules.Wednesday])

  useEffect(() => {
    doParse(WeekDay.Thursday, schedules.Thursday)
  }, [schedules.Thursday])

  useEffect(() => {
    doParse(WeekDay.Friday, schedules.Friday)
  }, [schedules.Friday])

  useEffect(() => {
    doParse(WeekDay.Saturday, schedules.Saturday)
  }, [schedules.Saturday])

  return (
    <>
      <div className="fixed top-3 right-3">
        <button
          type="button"
          className={clsxm(
            'text-sm font-semibold text-gray-600 shadow-sm cursor-pointer',
            'hover:text-gray-900 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-600'
          )}
          onClick={() => setView('home')}
        >
          <span className="sr-only">Go to home</span>

          <HomeIcon className="h-6 w-6" aria-hidden="true" />
        </button>
      </div>
      <div className="flex flex-col items-center">
        <div className="space-y-12 p-12 w-full">
          <h1 className="text-xl">Schedule</h1>

          <div className="grid grid-cols-4 gap-12">
            {WEEK.map((day) => (
              <div>
                <label htmlFor={day} className="block text-sm font-medium leading-6 text-gray-900">
                  {day} {format(week[day], 'do')}
                </label>
                <div className="mt-2">
                  <textarea
                    id={day}
                    name={day}
                    rows={10}
                    className={clsxm(
                      'block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm',
                      'ring-1 ring-inset ring-gray-300 placeholder:text-gray-400',
                      'focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6'
                    )}
                    value={schedules[day]}
                    onChange={(e) => setSchedules((prev) => ({...prev, [day]: e.target.value}))}
                  />
                </div>
                {errors[day] && (
                  <div className="mt-2 space-y-2">
                    <p className="text-sm text-red-600 font-bold">Failed to parse schedule</p>
                    <pre className="text-sm text-red-600 overflow-x-auto mt-1 bg-red-50 p-2 rounded-md">
                      {errors[day]}
                    </pre>
                  </div>
                )}
              </div>
            ))}
          </div>
        </div>
      </div>
    </>
  )
}
