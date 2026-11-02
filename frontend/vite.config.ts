import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";

export default defineConfig({
  plugins: [react()],
  build: {
    // Go-сервер встраивает эту папку в бинарник (см. frontend.go); рядом лежит dist/.keep.
    outDir: "dist/app",
    emptyOutDir: true,
  },
  server: {
    // Слушаем все сетевые интерфейсы, чтобы открыть приложение с телефона по IP компьютера.
    host: true,
    port: 5173,
    proxy: {
      "/api": "http://127.0.0.1:8080",
    },
  },
});
