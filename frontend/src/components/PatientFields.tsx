import type { ChangeEvent } from "react";
import type { PatientInput } from "../api";
import { Field } from "./ui";

export const emptyPatient: PatientInput = { last_name: "", first_name: "", middle_name: "", phone: "" };

type Props = {
  value: PatientInput;
  onChange: (value: PatientInput) => void;
  autoFocus?: boolean;
};

export function PatientFields({ value, onChange, autoFocus }: Props) {
  const set = (field: keyof PatientInput) => (e: ChangeEvent<HTMLInputElement>) =>
    onChange({ ...value, [field]: e.target.value });

  return (
    <div className="form-grid">
      <Field label="Фамилия *">
        <input className="input" value={value.last_name} onChange={set("last_name")} autoFocus={autoFocus} autoComplete="off" />
      </Field>
      <Field label="Имя *">
        <input className="input" value={value.first_name} onChange={set("first_name")} autoComplete="off" />
      </Field>
      <Field label="Отчество">
        <input className="input" value={value.middle_name ?? ""} onChange={set("middle_name")} autoComplete="off" />
      </Field>
      <Field label="Телефон">
        <input
          className="input"
          type="tel"
          inputMode="tel"
          placeholder="+7 900 000-00-00"
          value={value.phone ?? ""}
          onChange={set("phone")}
          autoComplete="off"
        />
      </Field>
    </div>
  );
}

/** Сообщение об ошибке или null, если обязательные поля заполнены. */
export function validatePatient(p: PatientInput): string | null {
  if (!p.last_name.trim()) return "Укажите фамилию";
  if (!p.first_name.trim()) return "Укажите имя";
  return null;
}
