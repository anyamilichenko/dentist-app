import { useState, type FormEvent } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { api, refreshData, type Service, type ServiceInput } from "../api";
import { Modal } from "../components/Modal";
import { cls, Field } from "../components/ui";
import { formatDuration, formatPrice } from "../lib/format";

const QUICK_DURATIONS = [15, 30, 45, 60, 90, 120];

export function ServicesPage() {
  // undefined — диалог закрыт, null — новая услуга
  const [editing, setEditing] = useState<Service | null | undefined>(undefined);
  const services = useQuery({ queryKey: ["services"], queryFn: api.listServices });

  return (
    <main className="page">
      <div className="page-head">
        <h1>Услуги</h1>
        <button type="button" className="btn btn-primary" onClick={() => setEditing(null)}>
          + Добавить
        </button>
      </div>

      {services.error && <p className="form-error">{services.error.message}</p>}

      <div className="list">
        {services.data?.map((s) => (
          <button key={s.id} type="button" className="list-row service-row" onClick={() => setEditing(s)}>
            <span className="list-main">{s.name}</span>
            <span className="list-side">{formatDuration(s.duration_minutes)}</span>
            <span className="list-side price">{formatPrice(s.price)}</span>
          </button>
        ))}
        {services.data?.length === 0 && (
          <div className="list-empty">Услуг пока нет. Добавьте, например, «Консультация, 30 минут».</div>
        )}
      </div>

      {editing !== undefined && <ServiceDialog service={editing} onClose={() => setEditing(undefined)} />}
    </main>
  );
}

function ServiceDialog({ service, onClose }: { service: Service | null; onClose: () => void }) {
  const queryClient = useQueryClient();
  const [name, setName] = useState(service?.name ?? "");
  const [duration, setDuration] = useState(String(service?.duration_minutes ?? 30));
  const [price, setPrice] = useState(service ? String(service.price) : "");
  const [formError, setFormError] = useState("");

  const save = useMutation({
    mutationFn: (v: ServiceInput) => (service ? api.updateService(service.id, v) : api.createService(v)),
    onSuccess: () => {
      refreshData(queryClient);
      onClose();
    },
  });

  const remove = useMutation({
    mutationFn: (id: number) => api.deleteService(id),
    onSuccess: () => {
      refreshData(queryClient);
      onClose();
    },
  });

  function submit(e: FormEvent) {
    e.preventDefault();
    const input: ServiceInput = {
      name: name.trim(),
      duration_minutes: Number(duration),
      price: price.trim() === "" ? 0 : Number(price),
    };
    let err = "";
    if (!input.name) err = "Укажите название услуги";
    else if (!Number.isInteger(input.duration_minutes) || input.duration_minutes < 5 || input.duration_minutes > 720)
      err = "Длительность — целое число минут от 5 до 720";
    else if (!Number.isInteger(input.price) || input.price < 0) err = "Цена — целое неотрицательное число рублей";
    setFormError(err);
    if (!err) save.mutate(input);
  }

  function confirmDelete() {
    if (service && window.confirm(`Удалить услугу «${service.name}»?`)) remove.mutate(service.id);
  }

  const error = formError || save.error?.message || remove.error?.message;
  const busy = save.isPending || remove.isPending;

  return (
    <Modal
      title={service ? service.name : "Новая услуга"}
      onClose={onClose}
      footer={
        <>
          {service && (
            <button type="button" className="btn btn-danger" onClick={confirmDelete} disabled={busy}>
              Удалить
            </button>
          )}
          <span className="spacer" />
          <button type="button" className="btn" onClick={onClose}>
            Отмена
          </button>
          <button type="submit" form="service-form" className="btn btn-primary" disabled={busy}>
            {save.isPending ? "Сохранение…" : "Сохранить"}
          </button>
        </>
      }
    >
      <form id="service-form" className="form" onSubmit={submit} noValidate>
        <Field label="Название *">
          <input className="input" value={name} onChange={(e) => setName(e.target.value)} autoFocus={!service} />
        </Field>
        <div className="form-grid">
          <Field label="Длительность, мин *">
            <input
              className="input"
              type="number"
              inputMode="numeric"
              min={5}
              max={720}
              step={5}
              value={duration}
              onChange={(e) => setDuration(e.target.value)}
            />
          </Field>
          <Field label="Цена, ₽">
            <input
              className="input"
              type="number"
              inputMode="numeric"
              min={0}
              step={100}
              placeholder="0"
              value={price}
              onChange={(e) => setPrice(e.target.value)}
            />
          </Field>
        </div>
        <div className="quick-picks">
          {QUICK_DURATIONS.map((m) => (
            <button
              key={m}
              type="button"
              className={cls("btn btn-sm", Number(duration) === m && "active")}
              onClick={() => setDuration(String(m))}
            >
              {formatDuration(m)}
            </button>
          ))}
        </div>
        {error && (
          <p className="form-error" role="alert">
            {error}
          </p>
        )}
      </form>
    </Modal>
  );
}
