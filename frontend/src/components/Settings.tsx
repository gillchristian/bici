import {HomeIcon} from '@heroicons/react/24/outline'
import {useAtom} from 'jotai'
import {format, isToday} from 'date-fns'

import {clsxm} from '@/utils/clsxm'
import {ViewAtom} from '@/utils/router'
import {useScheduleParser} from '@/models/calendar'
import {WEEK} from '@/models/week'

export const Settings = () => {
  const [_, setView] = useAtom(ViewAtom)

  const {errors, week, schedules, setSchedule} = useScheduleParser()

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

          <div className="grid grid-cols-4 gap-8">
            {WEEK.map((day) => {
              return (
                <div key={week[day].toISOString()}>
                  <label
                    htmlFor={week[day].toISOString()}
                    className={clsxm(
                      'block text-sm font-medium leading-6 text-gray-900',
                      isToday(week[day]) && 'underline'
                    )}
                  >
                    {format(week[day], 'EEEE')} {format(week[day], 'MMMM do')}
                  </label>
                  <div className="mt-2">
                    <textarea
                      id={week[day].toISOString()}
                      name={format(week[day], 'EEEE')}
                      rows={10}
                      className={clsxm(
                        'block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm',
                        'ring-1 ring-inset ring-gray-300 placeholder:text-gray-400',
                        'focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6'
                      )}
                      value={schedules[day]}
                      onChange={(e) => setSchedule(day, e.target.value)}
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
              )
            })}
          </div>
        </div>
      </div>
    </>
  )
}
