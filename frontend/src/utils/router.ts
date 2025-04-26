import {atom} from 'jotai'

export type View = 'home' | 'settings'

export const ViewAtom = atom<View>('home')
