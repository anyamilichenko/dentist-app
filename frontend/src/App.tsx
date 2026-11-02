import { useEffect } from "react";
import { Link, Navigate, NavLink, Route, Routes } from "react-router";
import { useQueryClient } from "@tanstack/react-query";
import { cls, ToothIcon } from "./components/ui";
import { connectLiveUpdates, useLiveConnected } from "./live";
import { CalendarPage } from "./pages/CalendarPage";
import { PatientsPage } from "./pages/PatientsPage";
import { ServicesPage } from "./pages/ServicesPage";

export function App() {
  const queryClient = useQueryClient();
  useEffect(() => connectLiveUpdates(queryClient), [queryClient]);
  const live = useLiveConnected();

  return (
    <>
      <header className="app-header">
        <Link to="/" className="brand">
          <ToothIcon />
          <span className="brand-name">Запись</span>
        </Link>
        <nav className="nav">
          <NavLink to="/" end>
            Календарь
          </NavLink>
          <NavLink to="/patients">Пациенты</NavLink>
          <NavLink to="/services">Услуги</NavLink>
        </nav>
        <span
          className={cls("live-dot", live && "on")}
          role="status"
          title={live ? "На связи: изменения с других устройств приходят сразу" : "Нет связи с сервером"}
          aria-label={live ? "На связи" : "Нет связи с сервером"}
        />
      </header>
      <Routes>
        <Route path="/" element={<CalendarPage />} />
        <Route path="/patients" element={<PatientsPage />} />
        <Route path="/services" element={<ServicesPage />} />
        <Route path="*" element={<Navigate to="/" replace />} />
      </Routes>
    </>
  );
}
