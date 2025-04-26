import {useAtomValue} from 'jotai'

import {Calendar} from './components/Calendar'
import {Settings} from './components/Settings'
import {ViewAtom} from './utils/router'

function App() {
  const view = useAtomValue(ViewAtom)

  if (view === 'settings') {
    return <Settings />
  }

  if (view === 'home') {
    return <Calendar />
  }

  return null
}

export default App
