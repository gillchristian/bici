import {main} from '@wails/go/models'

export type EventTime = {
  start: string
  end: string
}

export const toGo = (eventTime: EventTime): main.EventTime => {
  return new main.EventTime({
    start: eventTime.start,
    end: eventTime.end
  })
}
