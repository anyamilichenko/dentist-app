import { useState, type FormEvent } from "react";
import { Link } from "react-router";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { api, refreshData, type Appointment } from "../api";
import {
  addMinutes,
  formatDayTitle,
  fromInputs,
  minutesBetween,
  SLOT_MINUTES,
  toDateInput,
  toTimeInput,
} from "../lib/dates";
import { formatDuration, formatPrice } from "../lib/format";
import { Modal } from "./Modal";
import { validatePatient } from "./PatientFields";
import { PatientPicker, type PatientChoice } from "./PatientPicker";
import { Field } from "./ui";

export type AppointmentDialogMode =
  | { kind: "create"; start: Date }
  | { kind: "edit"; appointment: Appointment };

type Props = {
  mode: AppointmentDialogMode;
  onClose: () => void;
};

export function AppointmentDialog({ mode, onClose }: Props) {
  const queryClient = useQueryClient();
  const existing = mode.kind === "edit" ? mode.appointment : null;
  const [initialStart, initialEnd] =
    mode.kind === "edit"
      ? [new Date(mode.appointment.starts_at), new Date(mode.appointment.ends_at)]
      : [mode.start, addMinutes(mode.start, SLOT_MINUTES)];

  const [patient, setPatient] = useState<PatientChoice>(
    existing ? { kind: "existing", patient: existing.patient } : null,
  );
  const [serviceId, setServiceId] = useState<number | null>(existing?.service_id ?? null);
  const [date, setDate] = useState(() => toDateInput(initialStart));
  const [start, setStart] = useState(() => toTimeInput(initialStart));
  const [end, setEnd] = useState(() => toTimeInput(initialEnd));
  const [note, setNote] = useState(existing?.note ?? "");
  const [formError, setFormError] = useState("");

  const services = useQuery({ queryKey: ["services"], queryFn: api.listServices });

  /** Время "HH:MM" плюс minutes. */
  function shiftTime(time: string, minutes: number): string {
    const d = fromInputs(date, time);
    return d ? toTimeInput(addMinutes(d, minutes)) : time;
  }

  // Выбор услуги подставляет её длительность.
  function changeService(id: number | null) {
    setServiceId(id);
    const service = services.data?.find((s) => s.id === id);
    if (service && start) setEnd(shiftTime(start, service.duration_minutes));
  }

  // Перенос начала сохраняет длительность записи.
  function changeStart(value: string) {
    const oldStart = fromInputs(date, start);
    const oldEnd = fromInputs(date, end);
    setStart(value);
    if (value && oldStart && oldEnd && oldEnd > oldStart) {
      setEnd(shiftTime(value, minutesBetween(oldStart, oldEnd)));
    }
  }

  const save = useMutation({
    mutationFn: async (v: { choice: NonNullable<PatientChoice>; startsAt: Date; endsAt: Date }) => {
      let patientId: number;
      if (v.choice.kind === "new") {
        const created = await api.createPatient(v.choice.input);
        // Если запись не создастся (например, время заняли), повторное нажатие не задублирует пациента.
        setPatient({ kind: "existing", patient: created });
        patientId = created.id;
      } else {
        patientId = v.choice.patient.id;
      }
      const input = {
        patient_id: patientId,
        service_id: serviceId,
        starts_at: v.startsAt.toISOString(),
        ends_at: v.endsAt.toISOString(),
        note,
      };
      return existing ? api.updateAppointment(existing.id, input) : api.createAppointment(input);
    },
    onSuccess: onClose,
    // Обновляем календарь в любом случае: при конфликте станет видно, кто занял время.
    onSettled: () => refreshData(queryClient),
  });

  const remove = useMutation({
    mutationFn: (id: number) => api.deleteAppointment(id),
    onSuccess: onClose,
    onSettled: () => refreshData(queryClient),
  });

  function submit(e: FormEvent) {
    e.preventDefault();
    setFormError("");
    const startsAt = fromInputs(date, start);
    const endsAt = fromInputs(date, end);

    if (!patient) return setFormError("Выберите пациента или добавьте нового");
    const patientError = patient.kind === "new" ? validatePatient(patient.input) : null;
    if (patientError) return setFormError(patientError);
    if (!startsAt || !endsAt) return setFormError("Укажите дату и время");
    if (endsAt <= startsAt) return setFormError("Время окончания должно быть позже начала");

    save.mutate({ choice: patient, startsAt, endsAt });
  }

  function confirmDelete() {
    if (existing && window.confirm("Удалить эту запись?")) remove.mutate(existing.id);
  }

  const error = formError || save.error?.message || remove.error?.message;
  const busy = save.isPending || remove.isPending;
  const day = fromInputs(date, start);

  return (
    <Modal
      title={existing ? "Запись" : "Новая запись"}
      onClose={onClose}
      footer={
        <>
          {existing && (
            <button type="button" className="btn btn-danger" onClick={confirmDelete} disabled={busy}>
              Удалить
            </button>
          )}
          <span className="spacer" />
          <button type="button" className="btn" onClick={onClose}>
            Отмена
          </button>
          <button type="submit" form="appointment-form" className="btn btn-primary" disabled={busy}>
            {save.isPending ? "Сохранение…" : existing ? "Сохранить" : "Записать"}
          </button>
        </>
      }
    >
      <form id="appointment-form" className="form" onSubmit={submit} noValidate>
        <section className="form-section">
          <h3 className="section-title">Пациент</h3>
          <PatientPicker value={patient} onChange={setPatient} />
        </section>

        <section className="form-section">
          <h3 className="section-title">Время{day && <span className="section-note"> · {formatDayTitle(day)}</span>}</h3>
          <div className="form-grid when-grid">
            <Field label="Дата">
              <input type="date" className="input" value={date} onChange={(e) => setDate(e.target.value)} />
            </Field>
            <Field label="Начало">
              <input type="time" step={300} className="input" value={start} onChange={(e) => changeStart(e.target.value)} />
            </Field>
            <Field label="Окончание">
              <input type="time" step={300} className="input" value={end} onChange={(e) => setEnd(e.target.value)} />
            </Field>
          </div>
        </section>

        <Field label="Услуга">
          <select
            className="input"
            value={serviceId ?? ""}
            onChange={(e) => changeService(e.target.value ? Number(e.target.value) : null)}
          >
            <option value="">Без услуги</option>
            {services.data?.map((s) => (
              <option key={s.id} value={s.id}>
                {s.name} · {formatDuration(s.duration_minutes)} · {formatPrice(s.price)}
              </option>
            ))}
          </select>
        </Field>
        {services.data?.length === 0 && (
          <p className="hint">
            Справочник услуг пуст — <Link to="/services">добавить услуги</Link>
          </p>
        )}

        <Field label="Комментарий">
          <textarea className="input" rows={2} value={note} onChange={(e) => setNote(e.target.value)} />
        </Field>

        {error && (
          <p className="form-error" role="alert">
            {error}
          </p>
        )}
      </form>
    </Modal>
  );
}
