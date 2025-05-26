import {calendar} from '@wails/go/models'

export type EventTime = {
  start: string
  end: string
}

export const toGo = (eventTime: EventTime): calendar.EventTime => {
  return new calendar.EventTime({
    start: eventTime.start,
    end: eventTime.end
  })
}
