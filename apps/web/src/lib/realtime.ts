import { writable } from "svelte/store";

export type RealtimeEvent = { event: string; data: Record<string, unknown> | null };
export type ConnectionState = "disconnected" | "offline" | "reconnecting" | "connecting" | "expired" | "connected";
export const connection = writable<ConnectionState>("disconnected");
export const realtimeEvent = writable<RealtimeEvent | null>(null);
export const unread = writable(0);
let socket: WebSocket | null = null;

export function sendEvent(event: Record<string, unknown>) {
    if (socket?.readyState !== WebSocket.OPEN) return false;
    socket.send(JSON.stringify(event));
    return true;
}

export async function community<T>(path: string, method = "GET", body?: unknown): Promise<T> {
    const response = await fetch(`/api/community${path}`, {
        method,
        ...(body !== undefined ? { headers: { "Content-Type": "application/json" }, body: JSON.stringify(body) } : {}),
    });
    if (response.status === 204) return undefined as T;
    const result = await response.json();
    if (!response.ok) throw new Error(result.message || "The request could not be completed");
    return result;
}

export function startRealtime(onChange: () => void) {
    let stopped = false;
    let timer: ReturnType<typeof setTimeout>;
    let refresh: ReturnType<typeof setTimeout>;
    let attempts = 0;
    let connectedBefore = false;
    let needsRefresh = false;
    async function refreshUnread() {
        try {
            let page = 1,
                total = 0,
                count = 0;
            do {
                const result = await community<{ data: { unread: number }[]; total: number }>(
                    `/conversations?page=${page}`
                );
                count += result.data.length;
                total += result.data.reduce((sum, c) => sum + c.unread, 0);
                page++;
                if (count >= result.total || !result.data.length) break;
            } while (!stopped);
            if (!stopped) unread.set(total);
        } catch {
            /* The page-level API state remains available during reconnects. */
        }
    }
    async function connect() {
        if (stopped) return;
        if (!navigator.onLine) {
            connection.set("offline");
            return;
        }
        if (socket && (socket.readyState === WebSocket.OPEN || socket.readyState === WebSocket.CONNECTING)) return;
        connection.set(attempts ? "reconnecting" : "connecting");
        try {
            const response = await fetch("/api/realtime-ticket", { method: "POST" });
            if (response.status === 401) {
                connection.set("expired");
                return;
            }
            if (!response.ok) throw new Error("Connection unavailable");
            const { url } = await response.json();
            if (stopped) return;
            const ws = new WebSocket(url);
            socket = ws;
            ws.onopen = () => {
                attempts = 0;
                connection.set("connected");
                void refreshUnread();
                if (connectedBefore) onChange();
                connectedBefore = true;
            };
            ws.onmessage = ({ data }) => {
                let event: RealtimeEvent;
                try {
                    event = JSON.parse(data);
                } catch {
                    return;
                }
                realtimeEvent.set(event);
                if (!["presence", "typing", "ack", "error"].includes(event.event)) {
                    needsRefresh ||= event.event !== "read";
                    clearTimeout(refresh);
                    refresh = setTimeout(() => {
                        void refreshUnread();
                        if (needsRefresh) onChange();
                        needsRefresh = false;
                    }, 200);
                }
            };
            ws.onclose = () => {
                if (socket === ws) socket = null;
                retry();
            };
            ws.onerror = () => ws.close();
        } catch {
            retry();
        }
    }
    function retry() {
        if (stopped) return;
        if (socket && (socket.readyState === WebSocket.OPEN || socket.readyState === WebSocket.CONNECTING)) return;
        if (!navigator.onLine) {
            connection.set("offline");
            return;
        }
        connection.set("reconnecting");
        timer = setTimeout(connect, Math.min(30000, 1000 * 2 ** Math.min(attempts++, 5)) + Math.random() * 500);
    }
    const offline = () => {
        socket?.close();
        connection.set("offline");
    };
    const online = () => {
        clearTimeout(timer);
        void connect();
    };
    window.addEventListener("offline", offline);
    window.addEventListener("online", online);
    void connect();
    return () => {
        stopped = true;
        clearTimeout(timer);
        clearTimeout(refresh);
        window.removeEventListener("offline", offline);
        window.removeEventListener("online", online);
        socket?.close();
        socket = null;
        unread.set(0);
        realtimeEvent.set(null);
        connection.set("disconnected");
    };
}
