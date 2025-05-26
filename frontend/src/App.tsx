import {useAtomValue} from 'jotai'

import {Calendar} from './components/Calendar'
import {Settings} from './components/Settings'
import {ViewAtom} from './utils/router'
import {useFileDropHandler} from '@/models/calendar'

function App() {
  const view = useAtomValue(ViewAtom)

  useFileDropHandler()

  if (view === 'settings') {
    return <Settings />
  }

  if (view === 'home') {
    return <Calendar />
  }

  return null
}

export default App
