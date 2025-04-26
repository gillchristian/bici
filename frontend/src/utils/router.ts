import {atomWithStorage} from 'jotai/utils'

export type View = 'home' | 'settings'

export const ViewAtom = atomWithStorage<View>('view', 'home')
