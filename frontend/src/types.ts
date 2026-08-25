export type EventStatus = "draft" | "on_sale" | "sold_out" | "closed";

export interface TicketEvent {
  id: string;
  name: string;
  venue: string;
  starts_at: string;
  sales_start_at: string;
  capacity: number;
  price_cents: number;
  status: EventStatus;
}

export interface CreateEventInput {
  name: string;
  venue: string;
  starts_at: string;
  sales_start_at: string;
  capacity: number;
  price_cents: number;
}

export interface Availability {
  event_id: string;
  capacity: number;
  available: number;
  reserved: number;
  sold: number;
  version: number;
}

export interface Reservation {
  id: string;
  event_id: string;
  quantity: number;
  status: "pending" | "confirmed" | "cancelled" | "expired";
  expires_at: string;
  created_at: string;
}

export interface Health {
  status: "ok" | "degraded";
  redis: "ok" | "unavailable" | "not_checked";
  version: string;
}

