import { useState } from "react";
import { keepPreviousData, useQuery } from "@tanstack/react-query";
import { api, type Patient, type PatientInput } from "../api";
import { useDebounced } from "../hooks";
import { fullName } from "../lib/format";
import { PatientFields } from "./PatientFields";

export type PatientChoice =
  | { kind: "existing"; patient: Patient }
  | { kind: "new"; input: PatientInput }
  | null;

type Props = {
  value: PatientChoice;
  onChange: (value: PatientChoice) => void;
};

/** Выбор пациента из базы с поиском или ввод нового. */
export function PatientPicker({ value, onChange }: Props) {
  const [search, setSearch] = useState("");
  const q = useDebounced(search.trim());
  const results = useQuery({
    queryKey: ["patients", q, 8],
    queryFn: () => api.listPatients(q, 8),
    enabled: value === null,
    placeholderData: keepPreviousData,
  });

  if (value?.kind === "existing") {
    const p = value.patient;
    return (
      <div className="chip">
        <div className="chip-text">
          <div className="chip-title">{fullName(p)}</div>
          {p.phone && (
            <a className="chip-sub" href={`tel:${p.phone}`}>
              {p.phone}
            </a>
          )}
        </div>
        <button type="button" className="btn btn-sm" onClick={() => onChange(null)}>
          Сменить
        </button>
      </div>
    );
  }

  if (value?.kind === "new") {
    return (
      <div className="new-patient">
        <div className="new-patient-head">
          <span>Новый пациент</span>
          <button type="button" className="link-btn" onClick={() => onChange(null)}>
            ← выбрать из базы
          </button>
        </div>
        <PatientFields value={value.input} onChange={(input) => onChange({ kind: "new", input })} autoFocus />
      </div>
    );
  }

  return (
    <div>
      <input
        className="input"
        type="search"
        placeholder="Поиск по ФИО или телефону"
        value={search}
        onChange={(e) => setSearch(e.target.value)}
        autoFocus
      />
      <div className="picker-results">
        {results.data?.map((p) => (
          <button
            key={p.id}
            type="button"
            className="picker-item"
            onClick={() => onChange({ kind: "existing", patient: p })}
          >
            <span>{fullName(p)}</span>
            <span className="muted">{p.phone}</span>
          </button>
        ))}
        {results.data?.length === 0 && q && <div className="picker-empty">Никого не нашли</div>}
        <button
          type="button"
          className="picker-item picker-new"
          onClick={() => onChange({ kind: "new", input: draftFromSearch(search) })}
        >
          + Новый пациент{search.trim() && ` «${search.trim()}»`}
        </button>
      </div>
    </div>
  );
}

/** Черновик нового пациента из строки поиска: слова — ФИО, слова с цифрами — телефон. */
function draftFromSearch(search: string): PatientInput {
  const words = search.trim().split(/\s+/).filter(Boolean);
  const names = words.filter((w) => !/\d/.test(w)).map((w) => w.charAt(0).toUpperCase() + w.slice(1));
  return {
    last_name: names[0] ?? "",
    first_name: names[1] ?? "",
    middle_name: names.slice(2).join(" "),
    phone: words.filter((w) => /\d/.test(w)).join(" "),
  };
}
