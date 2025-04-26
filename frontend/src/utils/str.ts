export function utf8ToBase64(str: string) {
  const encoder = new TextEncoder()
  const uint8Array = encoder.encode(str)
  return btoa(String.fromCharCode(...Array.from(uint8Array)))
}

export function base64ToUtf8(base64: string) {
  const binaryString = atob(base64)
  const uint8Array = Uint8Array.from(binaryString, (char) => char.charCodeAt(0))
  const decoder = new TextDecoder()
  return decoder.decode(uint8Array)
}
