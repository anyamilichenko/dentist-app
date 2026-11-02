import type { QueryClient } from "@tanstack/react-query";

export type Patient = {
  id: number;
  last_name: string;
  first_name: string;
  middle_name: string | null;
  phone: string | null;
  created_at: string;
  updated_at: string;
};

export type PatientInput = {
  last_name: string;
  first_name: string;
  middle_name: string | null;
  phone: string | null;
};

export type Service = {
  id: number;
  name: string;
  duration_minutes: number;
  /** в рублях */
  price: number;
  created_at: string;
  updated_at: string;
};

export type ServiceInput = Pick<Service, "name" | "duration_minutes" | "price">;

export type Appointment = {
  id: number;
  patient_id: number;
  service_id: number | null;
  starts_at: string;
  ends_at: string;
  note: string;
  created_at: string;
  updated_at: string;
  patient: Patient;
  service: Service | null;
};

export type AppointmentInput = {
  patient_id: number;
  service_id: number | null;
  starts_at: string;
  ends_at: string;
  note: string;
};

export class ApiError extends Error {
  status: number;

  constructor(status: number, message: string) {
    super(message);
    this.status = status;
  }
}

async function request<T>(method: string, path: string, body?: unknown): Promise<T> {
  let res: Response;
  try {
    res = await fetch(`/api${path}`, {
      method,
      headers: body === undefined ? undefined : { "Content-Type": "application/json" },
      body: body === undefined ? undefined : JSON.stringify(body),
    });
  } catch {
    throw new ApiError(0, "Нет связи с сервером");
  }

  if (res.status === 204) return undefined as T;
  const data = await res.json().catch(() => null);
  if (!res.ok) {
    throw new ApiError(res.status, data?.error ?? `Ошибка сервера (${res.status})`);
  }
  return data as T;
}

const query = (params: Record<string, string>) => new URLSearchParams(params).toString();

export const api = {
  listPatients: (q: string, limit: number) =>
    request<Patient[]>("GET", `/patients?${query({ q, limit: String(limit) })}`),
  createPatient: (input: PatientInput) => request<Patient>("POST", "/patients", input),
  updatePatient: (id: number, input: PatientInput) => request<Patient>("PUT", `/patients/${id}`, input),
  deletePatient: (id: number) => request<void>("DELETE", `/patients/${id}`),

  listServices: () => request<Service[]>("GET", "/services"),
  createService: (input: ServiceInput) => request<Service>("POST", "/services", input),
  updateService: (id: number, input: ServiceInput) => request<Service>("PUT", `/services/${id}`, input),
  deleteService: (id: number) => request<void>("DELETE", `/services/${id}`),

  listAppointments: (from: Date, to: Date) =>
    request<Appointment[]>("GET", `/appointments?${query({ from: from.toISOString(), to: to.toISOString() })}`),
  patientAppointments: (patientId: number) =>
    request<Appointment[]>("GET", `/appointments?${query({ patient_id: String(patientId) })}`),
  createAppointment: (input: AppointmentInput) => request<Appointment>("POST", "/appointments", input),
  updateAppointment: (id: number, input: AppointmentInput) =>
    request<Appointment>("PUT", `/appointments/${id}`, input),
  deleteAppointment: (id: number) => request<void>("DELETE", `/appointments/${id}`),
};

/**
 * После любого изменения перезапрашиваем всё: записи содержат ФИО пациентов
 * и названия услуг, так что правка в справочнике должна отразиться в календаре.
 */
export function refreshData(queryClient: QueryClient) {
  return queryClient.invalidateQueries();
}
