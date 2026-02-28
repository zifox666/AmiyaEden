export function humanizeNumber(num: number): string {
  const abs = Math.abs(num)

  if (abs >= 1e12) {
    return (num / 1e12).toFixed(2).replace(/\.00$/, '') + 't' // trillion
  }
  if (abs >= 1e9) {
    return (num / 1e9).toFixed(2).replace(/\.00$/, '') + 'b' // billion
  }
  if (abs >= 1e6) {
    return (num / 1e6).toFixed(2).replace(/\.00$/, '') + 'm' // million
  }
  if (abs >= 1e3) {
    return (num / 1e3).toFixed(2).replace(/\.00$/, '') + 'k' // thousand
  }
  return num.toString()
}
