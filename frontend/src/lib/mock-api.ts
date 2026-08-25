import type {
  Availability,
  CreateEventInput,
  Health,
  Reservation,
  TicketEvent,
} from "../types";

const pause = (milliseconds = 260) =>
  new Promise((resolve) => window.setTimeout(resolve, milliseconds));

const events: TicketEvent[] = [
  {
    id: "c951183c-370c-4a6d-b9f2-dfa437a3f3b8",
    name: "Órbita Sonora",
    venue: "Galpão 67 — Campo Grande",
    starts_at: "2026-10-17T23:00:00-04:00",
    sales_start_at: "2026-09-01T10:00:00-04:00",
    capacity: 2400,
    price_cents: 14500,
    status: "on_sale",
  },
  {
    id: "f804fa44-10a1-41fb-97bf-25e3c5fb21cb",
    name: "Festival Ipê Amarelo",
    venue: "Parque das Nações Indígenas",
    starts_at: "2026-11-07T16:00:00-04:00",
    sales_start_at: "2026-09-14T09:00:00-04:00",
    capacity: 8000,
    price_cents: 8900,
    status: "on_sale",
  },
  {
    id: "1e605a4c-320e-4700-944a-e6957a5d468b",
    name: "Sessão Meia-Noite",
    venue: "Cine Teatro Glauce Rocha",
    starts_at: "2026-09-12T23:30:00-04:00",
    sales_start_at: "2026-08-28T12:00:00-04:00",
    capacity: 760,
    price_cents: 5200,
    status: "draft",
  },
];

const availability = new Map<string, Availability>([
  [
    events[0].id,
    {
      event_id: events[0].id,
      capacity: 2400,
      available: 1712,
      reserved: 188,
      sold: 500,
      version: 689,
    },
  ],
  [
    events[1].id,
    {
      event_id: events[1].id,
      capacity: 8000,
      available: 6198,
      reserved: 302,
      sold: 1500,
      version: 1803,
    },
  ],
  [
    events[2].id,
    {
      event_id: events[2].id,
      capacity: 760,
      available: 760,
      reserved: 0,
      sold: 0,
      version: 1,
    },
  ],
]);

export const mockApi = {
  async health(): Promise<Health> {
    await pause(120);
    return { status: "ok", redis: "not_checked", version: "mock" };
  },

  async listEvents(): Promise<TicketEvent[]> {
    await pause();
    return structuredClone(events);
  },

  async createEvent(input: CreateEventInput): Promise<TicketEvent> {
    await pause(420);
    const created: TicketEvent = {
      ...input,
      id: crypto.randomUUID(),
      status: "draft",
    };
    events.unshift(created);
    availability.set(created.id, {
      event_id: created.id,
      capacity: created.capacity,
      available: created.capacity,
      reserved: 0,
      sold: 0,
      version: 1,
    });
    return structuredClone(created);
  },

  async getAvailability(eventId: string): Promise<Availability> {
    await pause(180);
    const current = availability.get(eventId);
    if (!current) throw new Error("Evento não encontrado");
    return structuredClone(current);
  },

  async createReservation(
    eventId: string,
    quantity: number,
  ): Promise<Reservation> {
    await pause(520);
    const current = availability.get(eventId);
    if (!current || current.available < quantity) {
      throw new Error("Estoque insuficiente para esta reserva");
    }
    current.available -= quantity;
    current.reserved += quantity;
    current.version += 1;
    const now = new Date();
    return {
      id: crypto.randomUUID(),
      event_id: eventId,
      quantity,
      status: "pending",
      created_at: now.toISOString(),
      expires_at: new Date(now.getTime() + 10 * 60 * 1000).toISOString(),
    };
  },
};

