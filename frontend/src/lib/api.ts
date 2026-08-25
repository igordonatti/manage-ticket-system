import { mockApi } from "./mock-api";
import type {
  Availability,
  CreateEventInput,
  Health,
  Reservation,
  TicketEvent,
} from "../types";

const apiMode = import.meta.env.VITE_API_MODE ?? "mock";
const baseUrl = (import.meta.env.VITE_API_BASE_URL ?? "http://localhost:8080").replace(
  /\/$/,
  "",
);

interface ApiErrorBody {
  code?: string;
  message?: string;
}

export class ApiError extends Error {
  constructor(
    message: string,
    readonly status: number,
    readonly code?: string,
  ) {
    super(message);
  }
}

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const controller = new AbortController();
  const timeout = window.setTimeout(() => controller.abort(), 5_000);

  try {
    const response = await fetch(`${baseUrl}${path}`, {
      ...init,
      signal: controller.signal,
      headers: {
        Accept: "application/json",
        ...init?.headers,
      },
    });

    if (!response.ok) {
      const body = (await response.json().catch(() => ({}))) as ApiErrorBody;
      throw new ApiError(
        body.message ?? `A API respondeu com HTTP ${response.status}`,
        response.status,
        body.code,
      );
    }

    return (await response.json()) as T;
  } catch (error) {
    if (error instanceof ApiError) throw error;
    if (error instanceof DOMException && error.name === "AbortError") {
      throw new ApiError("A API demorou mais de cinco segundos para responder", 0);
    }
    throw new ApiError("Não foi possível conectar à API Go", 0);
  } finally {
    window.clearTimeout(timeout);
  }
}

const httpApi = {
  health: () => request<Health>("/health"),

  async listEvents(): Promise<TicketEvent[]> {
    const response = await request<{ items: TicketEvent[] }>("/events");
    return response.items;
  },

  createEvent: (input: CreateEventInput) =>
    request<TicketEvent>("/events", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(input),
    }),

  getAvailability: (eventId: string) =>
    request<Availability>(`/events/${eventId}/availability`),

  createReservation: (eventId: string, quantity: number) =>
    request<Reservation>(`/events/${eventId}/reservations`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        "Idempotency-Key": crypto.randomUUID(),
      },
      body: JSON.stringify({ quantity }),
    }),
};

export const api = apiMode === "mock" ? mockApi : httpApi;
export const isMockMode = apiMode === "mock";

