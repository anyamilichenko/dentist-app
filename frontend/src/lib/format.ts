import type { PatientInput } from "../api";

type Name = Pick<PatientInput, "last_name" | "first_name" | "middle_name">;

/** Иванов Иван Иванович */
export function fullName(p: Name): string {
  return [p.last_name, p.first_name, p.middle_name].filter(Boolean).join(" ");
}

/** Иванов И. И. */
export function shortName(p: Name): string {
  const initials = [p.first_name, p.middle_name]
    .filter((s): s is string => Boolean(s))
    .map((s) => `${s[0]}.`)
    .join(" ");
  return `${p.last_name} ${initials}`;
}

const priceFormat = new Intl.NumberFormat("ru-RU", {
  style: "currency",
  currency: "RUB",
  maximumFractionDigits: 0,
});

export const formatPrice = (rub: number) => priceFormat.format(rub);

/** 1 ч 30 мин */
export function formatDuration(minutes: number): string {
  const h = Math.floor(minutes / 60);
  const m = minutes % 60;
  return [h > 0 && `${h} ч`, m > 0 && `${m} мин`].filter(Boolean).join(" ");
}
