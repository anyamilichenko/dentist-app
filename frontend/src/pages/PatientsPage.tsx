import { useState, type FormEvent } from "react";
import { keepPreviousData, useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { api, refreshData, type Patient, type PatientInput } from "../api";
import { Modal } from "../components/Modal";
import { emptyPatient, PatientFields, validatePatient } from "../components/PatientFields";
import { cls } from "../components/ui";
import { useDebounced } from "../hooks";
import { formatShortDate, formatTime } from "../lib/dates";
import { fullName } from "../lib/format";

const LIMIT = 200;

export function PatientsPage() {
  const [search, setSearch] = useState("");
  const q = useDebounced(search.trim());
  // undefined — диалог закрыт, null — новый пациент
  const [editing, setEditing] = useState<Patient | null | undefined>(undefined);

  const patients = useQuery({
    queryKey: ["patients", q, LIMIT],
    queryFn: () => api.listPatients(q, LIMIT),
    placeholderData: keepPreviousData,
  });

  return (
    <main className="page">
      <div className="page-head">
        <h1>Пациенты</h1>
        <button type="button" className="btn btn-primary" onClick={() => setEditing(null)}>
          + Добавить
        </button>
      </div>

      <input
        className="input search"
        type="search"
        placeholder="Поиск по ФИО или телефону"
        value={search}
        onChange={(e) => setSearch(e.target.value)}
      />

      {patients.error && <p className="form-error">{patients.error.message}</p>}

      <div className="list">
        {patients.data?.map((p) => (
          <button key={p.id} type="button" className="list-row" onClick={() => setEditing(p)}>
            <span className="list-main">{fullName(p)}</span>
            <span className="list-side">{p.phone ?? "—"}</span>
          </button>
        ))}
        {patients.data?.length === 0 && (
          <div className="list-empty">
            {q ? "Никого не нашли" : "Пациентов пока нет. Добавьте первого или запишите прямо из календаря."}
          </div>
        )}
      </div>
      {patients.data?.length === LIMIT && <p className="hint">Показаны первые {LIMIT} — уточните поиск.</p>}

      {editing !== undefined && <PatientDialog patient={editing} onClose={() => setEditing(undefined)} />}
    </main>
  );
}

function PatientDialog({ patient, onClose }: { patient: Patient | null; onClose: () => void }) {
  const queryClient = useQueryClient();
  const [input, setInput] = useState<PatientInput>(() =>
    patient
      ? {
          last_name: patient.last_name,
          first_name: patient.first_name,
          middle_name: patient.middle_name ?? "",
          phone: patient.phone ?? "",
        }
      : emptyPatient,
  );
  const [formError, setFormError] = useState("");

  const history = useQuery({
    queryKey: ["patient-appointments", patient?.id],
    queryFn: () => api.patientAppointments(patient!.id),
    enabled: patient !== null,
  });

  const save = useMutation({
    mutationFn: (v: PatientInput) => (patient ? api.updatePatient(patient.id, v) : api.createPatient(v)),
    onSuccess: () => {
      refreshData(queryClient);
      onClose();
    },
  });

  const remove = useMutation({
    mutationFn: (id: number) => api.deletePatient(id),
    onSuccess: () => {
      refreshData(queryClient);
      onClose();
    },
  });

  function submit(e: FormEvent) {
    e.preventDefault();
    const err = validatePatient(input);
    setFormError(err ?? "");
    if (!err) save.mutate(input);
  }

  function confirmDelete() {
    if (patient && window.confirm(`Удалить пациента «${fullName(patient)}»?`)) remove.mutate(patient.id);
  }

  const error = formError || save.error?.message || remove.error?.message;
  const busy = save.isPending || remove.isPending;
  const now = new Date();

  return (
    <Modal
      title={patient ? fullName(patient) : "Новый пациент"}
      onClose={onClose}
      footer={
        <>
          {patient && (
            <button type="button" className="btn btn-danger" onClick={confirmDelete} disabled={busy}>
              Удалить
            </button>
          )}
          <span className="spacer" />
          <button type="button" className="btn" onClick={onClose}>
            Отмена
          </button>
          <button type="submit" form="patient-form" className="btn btn-primary" disabled={busy}>
            {save.isPending ? "Сохранение…" : "Сохранить"}
          </button>
        </>
      }
    >
      <form id="patient-form" className="form" onSubmit={submit} noValidate>
        <PatientFields value={input} onChange={setInput} autoFocus={!patient} />
        {patient?.phone && (
          <a className="btn btn-sm call-link" href={`tel:${patient.phone}`}>
            📞 Позвонить
          </a>
        )}
        {error && (
          <p className="form-error" role="alert">
            {error}
          </p>
        )}
      </form>

      {patient && (
        <section className="form-section">
          <h3 className="section-title">Записи</h3>
          {history.data?.length === 0 && <p className="hint">Записей пока нет</p>}
          <ul className="history">
            {history.data
              ?.slice()
              .reverse()
              .map((a) => {
                const start = new Date(a.starts_at);
                return (
                  <li key={a.id} className={cls(start > now && "upcoming")}>
                    <span>
                      {formatShortDate(start)}, {formatTime(start)}–{formatTime(new Date(a.ends_at))}
                    </span>
                    <span className="muted">{a.service?.name ?? "без услуги"}</span>
                  </li>
                );
              })}
          </ul>
        </section>
      )}
    </Modal>
  );
}
