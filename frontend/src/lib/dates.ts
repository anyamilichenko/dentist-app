// Сетка календаря. Записи вне рабочих часов тоже видны: сетка расширяется под них.
export const SLOT_MINUTES = 30;
export const DAY_START_HOUR = 8;
export const DAY_END_HOUR = 21;

export function startOfDay(d: Date): Date {
  const r = new Date(d);
  r.setHours(0, 0, 0, 0);
  return r;
}

export function atHour(d: Date, hour: number): Date {
  const r = new Date(d);
  r.setHours(hour, 0, 0, 0);
  return r;
}

export function addDays(d: Date, days: number): Date {
  const r = new Date(d);
  r.setDate(r.getDate() + days);
  return r;
}

export function addMinutes(d: Date, minutes: number): Date {
  return new Date(d.getTime() + minutes * 60_000);
}

/** Понедельник недели, в которую попадает дата. */
export function startOfWeek(d: Date): Date {
  const r = startOfDay(d);
  return addDays(r, -((r.getDay() + 6) % 7));
}

export function isSameDay(a: Date, b: Date): boolean {
  return a.getFullYear() === b.getFullYear() && a.getMonth() === b.getMonth() && a.getDate() === b.getDate();
}

export function minutesBetween(a: Date, b: Date): number {
  return Math.round((b.getTime() - a.getTime()) / 60_000);
}

const timeFormat = new Intl.DateTimeFormat("ru-RU", { hour: "2-digit", minute: "2-digit" });
const weekdayFormat = new Intl.DateTimeFormat("ru-RU", { weekday: "short" });
const dayTitleFormat = new Intl.DateTimeFormat("ru-RU", { weekday: "long", day: "numeric", month: "long" });
const rangeFormat = new Intl.DateTimeFormat("ru-RU", { day: "numeric", month: "long", year: "numeric" });
const shortDateFormat = new Intl.DateTimeFormat("ru-RU", { day: "numeric", month: "short", year: "numeric" });

/** 09:30 */
export const formatTime = (d: Date) => timeFormat.format(d);
/** пн */
export const formatWeekday = (d: Date) => weekdayFormat.format(d);
/** понедельник, 3 ноября */
export const formatDayTitle = (d: Date) => dayTitleFormat.format(d);
/** 3–9 ноября 2026 г. */
export const formatDateRange = (a: Date, b: Date) => rangeFormat.formatRange(a, b);
/** 3 нояб. 2026 г. */
export const formatShortDate = (d: Date) => shortDateFormat.format(d);

const pad = (n: number) => String(n).padStart(2, "0");

/** Значение для <input type="date"> в местном времени. */
export function toDateInput(d: Date): string {
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}`;
}

/** Значение для <input type="time">. */
export function toTimeInput(d: Date): string {
  return `${pad(d.getHours())}:${pad(d.getMinutes())}`;
}

/** Дата из полей формы (местное время); null, если поля пустые или некорректные. */
export function fromInputs(date: string, time: string): Date | null {
  if (!date || !time) return null;
  const d = new Date(`${date}T${time}`);
  return Number.isNaN(d.getTime()) ? null : d;
}
