export function clsxm(...classes: (string | boolean | undefined | null)[]) {
  return classes.filter(Boolean).join(' ')
}
