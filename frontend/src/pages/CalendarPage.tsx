import { useState } from "react";
import { keepPreviousData, useQuery } from "@tanstack/react-query";
import { api, type Appointment } from "../api";
import { AppointmentDialog, type AppointmentDialogMode } from "../components/AppointmentDialog";
import { CalendarGrid } from "../components/CalendarGrid";
import { cls } from "../components/ui";
import { useMediaQuery } from "../hooks";
import { useLiveConnected } from "../live";
import {
  addDays,
  formatDateRange,
  formatDayTitle,
  formatWeekday,
  isSameDay,
  startOfDay,
  startOfWeek,
} from "../lib/dates";

export function CalendarPage() {
  // На телефоне показываем один день, на компьютере — неделю.
  const isMobile = useMediaQuery("(max-width: 720px)");
  const [anchor, setAnchor] = useState(() => startOfDay(new Date()));
  const [dialog, setDialog] = useState<AppointmentDialogMode | null>(null);

  const weekStart = startOfWeek(anchor);
  const weekDays = Array.from({ length: 7 }, (_, i) => addDays(weekStart, i));

  const live = useLiveConnected();
  const appointments = useQuery({
    queryKey: ["appointments", weekStart.toISOString()],
    queryFn: () => api.listAppointments(weekStart, addDays(weekStart, 7)),
    placeholderData: keepPreviousData,
    // Изменения с других устройств приходят через поток событий (live.ts);
    // если он оборвался, подстраховываемся опросом.
    refetchInterval: live ? false : 15_000,
  });
  const list = appointments.data ?? [];

  const step = isMobile ? 1 : 7;
  const title = isMobile ? formatDayTitle(anchor) : formatDateRange(weekStart, weekDays[6]);
  const status = appointments.isError
    ? "Нет связи с сервером"
    : appointments.isPending || appointments.isPlaceholderData
      ? "Загрузка…"
      : "";

  return (
    <main className="calendar-page">
      <div className="calendar-toolbar">
        <div className="btn-group">
          <button
            type="button"
            className="btn icon-btn"
            onClick={() => setAnchor(addDays(anchor, -step))}
            aria-label={isMobile ? "Предыдущий день" : "Предыдущая неделя"}
          >
            ‹
          </button>
          <button type="button" className="btn" onClick={() => setAnchor(startOfDay(new Date()))}>
            Сегодня
          </button>
          <button
            type="button"
            className="btn icon-btn"
            onClick={() => setAnchor(addDays(anchor, step))}
            aria-label={isMobile ? "Следующий день" : "Следующая неделя"}
          >
            ›
          </button>
        </div>
        <h1 className="calendar-title">{title}</h1>
        {status && <span className={cls("sync-status", appointments.isError && "error")}>{status}</span>}
      </div>

      {isMobile && <WeekStrip days={weekDays} selected={anchor} appointments={list} onSelect={setAnchor} />}

      <CalendarGrid
        days={isMobile ? [anchor] : weekDays}
        appointments={list}
        onSlotClick={(start) => setDialog({ kind: "create", start })}
        onAppointmentClick={(appointment) => setDialog({ kind: "edit", appointment })}
      />

      {dialog && <AppointmentDialog mode={dialog} onClose={() => setDialog(null)} />}
    </main>
  );
}

type WeekStripProps = {
  days: Date[];
  selected: Date;
  appointments: Appointment[];
  onSelect: (day: Date) => void;
};

/** Полоска дней недели для телефона с числом записей на каждый день. */
function WeekStrip({ days, selected, appointments, onSelect }: WeekStripProps) {
  const today = new Date();
  return (
    <div className="week-strip">
      {days.map((d) => {
        const count = appointments.filter((a) => isSameDay(new Date(a.starts_at), d)).length;
        return (
          <button
            key={d.toISOString()}
            type="button"
            className={cls("week-strip-day", isSameDay(d, selected) && "selected", isSameDay(d, today) && "today")}
            onClick={() => onSelect(d)}
          >
            <span className="weekday">{formatWeekday(d)}</span>
            <span className="daynum">{d.getDate()}</span>
            <span className="count">{count > 0 ? count : " "}</span>
          </button>
        );
      })}
    </div>
  );
}
