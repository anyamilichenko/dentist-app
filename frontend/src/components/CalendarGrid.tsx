import { useEffect, useRef, type CSSProperties } from "react";
import type { Appointment } from "../api";
import { useNow } from "../hooks";
import {
  addMinutes,
  atHour,
  DAY_END_HOUR,
  DAY_START_HOUR,
  formatTime,
  formatWeekday,
  isSameDay,
  minutesBetween,
  SLOT_MINUTES,
} from "../lib/dates";
import { fullName, shortName } from "../lib/format";
import { cls } from "./ui";

const SLOT_HEIGHT = 40; // px на один слот
const PX_PER_MINUTE = SLOT_HEIGHT / SLOT_MINUTES;

// Цвет записи зависит от услуги, чтобы день читался с одного взгляда.
const PALETTE = [
  { bg: "#e0f2fe", border: "#0284c7" },
  { bg: "#dcfce7", border: "#16a34a" },
  { bg: "#fef3c7", border: "#d97706" },
  { bg: "#fce7f3", border: "#db2777" },
  { bg: "#ede9fe", border: "#7c3aed" },
  { bg: "#ccfbf1", border: "#0d9488" },
];
const NO_SERVICE = { bg: "#f1f5f9", border: "#64748b" };

type Props = {
  days: Date[];
  appointments: Appointment[];
  onSlotClick: (start: Date) => void;
  onAppointmentClick: (appointment: Appointment) => void;
};

export function CalendarGrid({ days, appointments, onSlotClick, onAppointmentClick }: Props) {
  const now = useNow();
  const [startHour, endHour] = visibleHours(appointments);
  const hours = Array.from({ length: endHour - startHour }, (_, i) => startHour + i);
  const scrollRef = useRef<HTMLDivElement>(null);

  // При смене дня/недели прокручиваем к текущему времени (если виден сегодняшний день)
  // или к первой записи. Обновление данных по таймеру прокрутку не трогает.
  const daysKey = days.map((d) => d.toDateString()).join();
  useEffect(() => {
    const el = scrollRef.current;
    if (!el) return;
    const minuteOfDay = (d: Date) => d.getHours() * 60 + d.getMinutes();
    const starts = appointments
      .map((a) => new Date(a.starts_at))
      .filter((s) => days.some((d) => isSameDay(s, d)))
      .map(minuteOfDay);
    const target = days.some((d) => isSameDay(d, now))
      ? minuteOfDay(now)
      : starts.length > 0
        ? Math.min(...starts)
        : startHour * 60;
    el.scrollTop = Math.max(0, (target - startHour * 60) * PX_PER_MINUTE - 2 * SLOT_HEIGHT);
  }, [daysKey]);

  return (
    <div className="calendar" ref={scrollRef} style={{ "--days": days.length } as CSSProperties}>
      {days.length > 1 && (
        <div className="calendar-head">
          <div />
          {days.map((d) => (
            <div key={d.toISOString()} className={cls("calendar-day-title", isSameDay(d, now) && "today")}>
              <span className="weekday">{formatWeekday(d)}</span>
              <span className="daynum">{d.getDate()}</span>
            </div>
          ))}
        </div>
      )}
      <div className="calendar-body">
        <div className="calendar-times">
          {hours.map((h) => (
            <div key={h} className="calendar-hour" style={{ height: 60 * PX_PER_MINUTE }}>
              <span>{String(h).padStart(2, "0")}:00</span>
            </div>
          ))}
        </div>
        {days.map((day) => (
          <DayColumn
            key={day.toISOString()}
            day={day}
            startHour={startHour}
            endHour={endHour}
            now={now}
            appointments={appointments.filter((a) => isSameDay(new Date(a.starts_at), day))}
            onSlotClick={onSlotClick}
            onAppointmentClick={onAppointmentClick}
          />
        ))}
      </div>
    </div>
  );
}

type DayColumnProps = {
  day: Date;
  startHour: number;
  endHour: number;
  now: Date;
  appointments: Appointment[];
  onSlotClick: (start: Date) => void;
  onAppointmentClick: (appointment: Appointment) => void;
};

function DayColumn({ day, startHour, endHour, now, appointments, onSlotClick, onAppointmentClick }: DayColumnProps) {
  const dayStart = atHour(day, startHour);
  const totalMinutes = (endHour - startHour) * 60;
  const slotCount = Math.floor(totalMinutes / SLOT_MINUTES);
  const isToday = isSameDay(day, now);
  const nowOffset = minutesBetween(dayStart, now);

  return (
    <div className={cls("calendar-col", isToday && "today")}>
      {Array.from({ length: slotCount }, (_, i) => {
        const start = addMinutes(dayStart, i * SLOT_MINUTES);
        return (
          <button
            key={i}
            type="button"
            className={cls("slot", start.getMinutes() === 0 && "hour-start", addMinutes(start, SLOT_MINUTES) <= now && "past")}
            style={{ height: SLOT_HEIGHT }}
            onClick={() => onSlotClick(start)}
            aria-label={`Записать на ${formatTime(start)}`}
          >
            <span>+ {formatTime(start)}</span>
          </button>
        );
      })}

      {appointments.map((a) => (
        <AppointmentBlock key={a.id} appointment={a} dayStart={dayStart} onClick={() => onAppointmentClick(a)} />
      ))}

      {isToday && nowOffset >= 0 && nowOffset <= totalMinutes && (
        <div className="now-line" style={{ top: nowOffset * PX_PER_MINUTE }} />
      )}
    </div>
  );
}

function AppointmentBlock({ appointment: a, dayStart, onClick }: { appointment: Appointment; dayStart: Date; onClick: () => void }) {
  const start = new Date(a.starts_at);
  const end = new Date(a.ends_at);
  const top = Math.max(0, minutesBetween(dayStart, start)) * PX_PER_MINUTE;
  const height = Math.max(minutesBetween(start, end) * PX_PER_MINUTE, 20);
  const color = a.service ? PALETTE[a.service.id % PALETTE.length] : NO_SERVICE;
  const time = `${formatTime(start)}–${formatTime(end)}`;

  return (
    <button
      type="button"
      className={cls("appt", height < 36 && "compact")}
      style={{ top, height: height - 2, background: color.bg, borderLeftColor: color.border }}
      onClick={onClick}
      title={[time, fullName(a.patient), a.service?.name].filter(Boolean).join(" · ")}
    >
      <span className="appt-time">{time}</span>
      <span className="appt-name">{shortName(a.patient)}</span>
      {a.service && <span className="appt-service">{a.service.name}</span>}
    </button>
  );
}

/** Рабочие часы, расширенные так, чтобы были видны все записи. */
function visibleHours(appointments: Appointment[]): [number, number] {
  let start = DAY_START_HOUR;
  let end = DAY_END_HOUR;
  for (const a of appointments) {
    const s = new Date(a.starts_at);
    const e = new Date(a.ends_at);
    start = Math.min(start, s.getHours());
    end = Math.max(end, isSameDay(s, e) ? e.getHours() + (e.getMinutes() > 0 ? 1 : 0) : 24);
  }
  return [start, end];
}
