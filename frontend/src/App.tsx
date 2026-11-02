import { Link, Navigate, NavLink, Route, Routes } from "react-router";
import { ToothIcon } from "./components/ui";
import { CalendarPage } from "./pages/CalendarPage";
import { PatientsPage } from "./pages/PatientsPage";
import { ServicesPage } from "./pages/ServicesPage";

export function App() {
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
