import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";

export default defineConfig({
  plugins: [react()],
  server: {
    // Слушаем все сетевые интерфейсы, чтобы открыть приложение с телефона по IP компьютера.
    host: true,
    port: 5173,
    proxy: {
      "/api": "http://127.0.0.1:8080",
    },
  },
});
