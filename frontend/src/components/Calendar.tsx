import {useEffect, useMemo, useRef, useCallback} from 'react'
import {CogIcon, ArrowDownTrayIcon} from '@heroicons/react/24/outline'
import {format} from 'date-fns'
import {useAtom, useAtomValue} from 'jotai'

import {ExportEvents} from '@wails/go/main/App'

import {clsxm} from '@/utils/clsxm'
import {mkWeek, EventsAtom, ScheduleAtom, getWeekDay} from '@/utils/calendar'
import {ViewAtom} from '@/utils/router'
import * as CE from '@/utils/calendar-event'

import {Week} from './Week'

type Props = {}

export function Calendar({}: Props) {
  const container = useRef<HTMLDivElement>(null)
  const containerNav = useRef<HTMLDivElement>(null)
  const containerOffset = useRef<HTMLDivElement>(null)

  const now = useMemo(() => new Date(), [])
  const week = useMemo(() => mkWeek(now), [])
  const today = useMemo(() => getWeekDay(now), [])

  const schedules = useAtomValue(ScheduleAtom)

  const events = useAtomValue(EventsAtom)
  const [_, setView] = useAtom(ViewAtom)

  useEffect(() => {
    if (!container.current || !containerNav.current || !containerOffset.current) {
      return
    }

    const currentMinute = new Date().getHours() * 60

    const containerHeight =
      container.current.scrollHeight -
      containerNav.current.offsetHeight -
      containerOffset.current.offsetHeight

    container.current.scrollTop = (containerHeight * currentMinute) / 1440
  }, [])

  const exportEvents = useCallback(() => {
    const events_ = Object.values(events)
      .flat()
      .map((event) => CE.toGo(event))

    ExportEvents(events_).then((message) => {
      console.log({message})
      if (message === '') {
        console.log('Events exported successfully')
      } else {
        console.error('Failed to export events')
        console.error(message)
      }
    })
  }, [events])

  return (
    <div className="flex h-full flex-col">
      <header className="flex flex-none items-center justify-between border-b border-gray-200 px-6 py-4">
        <div>
          <h1 className="text-base font-semibold leading-6 text-gray-900">
            <time dateTime={format(now, 'yyyy-MM-dd')} className="sm:hidden">
              {format(now, 'MMM d, yyyy')}
            </time>
            <time dateTime={format(now, 'yyyy-MM-dd')} className="hidden sm:inline">
              {format(now, 'MMMM d, yyyy')}
            </time>
          </h1>
          <p className="mt-1 text-sm text-gray-500">{format(now, 'EEEE')}</p>
        </div>
        <div className="flex items-center">
          <div className="hidden md:ml-4 md:flex md:items-center">
            <button
              type="button"
              className={clsxm(
                'ml-6 px-3 py-2 text-sm font-semibold text-gray-600 shadow-sm cursor-pointer',
                'hover:text-gray-900 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-600'
              )}
              onClick={exportEvents}
            >
              <span className="sr-only">Export events</span>
              <ArrowDownTrayIcon className="h-6 w-6" aria-hidden="true" />
            </button>
            <button
              type="button"
              className={clsxm(
                'ml-6 px-3 py-2 text-sm font-semibold text-gray-600 shadow-sm cursor-pointer',
                'hover:text-gray-900 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-600'
              )}
              onClick={() => setView('settings')}
            >
              <span className="sr-only">Go to settings</span>
              <CogIcon className="h-6 w-6" aria-hidden="true" />
            </button>
          </div>

          <div className="relative ml-6 md:hidden flex items-center gap-4">
            <button
              className="-mx-2 flex items-center rounded-full border border-transparent p-2 text-gray-400 hover:text-gray-500"
              onClick={() => setView('settings')}
            >
              <span className="sr-only">Open menu</span>
              <CogIcon className="h-5 w-5" aria-hidden="true" />
            </button>
            <button
              className="-mx-2 flex items-center rounded-full border border-transparent p-2 text-gray-400 hover:text-gray-500"
              onClick={exportEvents}
            >
              <span className="sr-only">Export events</span>
              <ArrowDownTrayIcon className="h-5 w-5" aria-hidden="true" />
            </button>
          </div>
        </div>
      </header>

      <Week
        events={events}
        containerRef={container}
        containerNavRef={containerNav}
        containerOffsetRef={containerOffset}
        week={week}
      />
    </div>
  )
}
