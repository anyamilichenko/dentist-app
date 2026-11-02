import { useSyncExternalStore } from "react";
import type { QueryClient } from "@tanstack/react-query";
import { refreshData } from "./api";

// Мгновенные обновления: сервер присылает событие после любого изменения данных
// (с любого устройства), и мы перезапрашиваем то, что сейчас на экране.

let connected = false;
const listeners = new Set<() => void>();

function setConnected(value: boolean) {
  if (connected === value) return;
  connected = value;
  listeners.forEach((notify) => notify());
}

/** true, пока открыт поток событий с сервера. */
export function useLiveConnected(): boolean {
  return useSyncExternalStore(
    (notify) => {
      listeners.add(notify);
      return () => {
        listeners.delete(notify);
      };
    },
    () => connected,
  );
}

/** Подключается к /api/events и держит соединение. Возвращает функцию отключения. */
export function connectLiveUpdates(queryClient: QueryClient): () => void {
  let source: EventSource | null = null;
  let retryTimer: ReturnType<typeof setTimeout> | undefined;
  let stopped = false;

  function open() {
    source = new EventSource("/api/events");
    source.onopen = () => {
      setConnected(true);
      // Пока связи не было, могли пропустить изменения — перезапрашиваем всё.
      refreshData(queryClient);
    };
    source.onmessage = () => {
      refreshData(queryClient);
    };
    source.onerror = () => {
      setConnected(false);
      // Обычно браузер переподключается сам. Но если сервер ответил ошибкой
      // (например, был выключен), EventSource закрывается насовсем — тогда пробуем сами.
      if (source?.readyState === EventSource.CLOSED && !stopped) {
        retryTimer = setTimeout(open, 5_000);
      }
    };
  }

  open();
  return () => {
    stopped = true;
    clearTimeout(retryTimer);
    source?.close();
    setConnected(false);
  };
}
