import {
  Activity,
  ArrowUpRight,
  CalendarDays,
  Check,
  ChevronRight,
  CircleDollarSign,
  Clock3,
  LoaderCircle,
  MapPin,
  Plus,
  Radio,
  RefreshCw,
  ServerOff,
  Ticket,
  Users,
  X,
  Zap,
} from "lucide-react";
import { useCallback, useEffect, useMemo, useState, type FormEvent } from "react";
import { api, isMockMode } from "./lib/api";
import type {
  Availability,
  CreateEventInput,
  Health,
  Reservation,
  TicketEvent,
} from "./types";

const currency = new Intl.NumberFormat("pt-BR", {
  style: "currency",
  currency: "BRL",
});

const dateTime = new Intl.DateTimeFormat("pt-BR", {
  day: "2-digit",
  month: "short",
  year: "numeric",
  hour: "2-digit",
  minute: "2-digit",
});

const statusLabels: Record<TicketEvent["status"], string> = {
  draft: "Rascunho",
  on_sale: "À venda",
  sold_out: "Esgotado",
  closed: "Encerrado",
};

interface EventFormState {
  name: string;
  venue: string;
  starts_at: string;
  sales_start_at: string;
  capacity: string;
  price: string;
}

const initialEventForm: EventFormState = {
  name: "",
  venue: "",
  starts_at: "",
  sales_start_at: "",
  capacity: "",
  price: "",
};

function App() {
  const [events, setEvents] = useState<TicketEvent[]>([]);
  const [health, setHealth] = useState<Health | null>(null);
  const [selectedId, setSelectedId] = useState<string | null>(null);
  const [availability, setAvailability] = useState<Availability | null>(null);
  const [loading, setLoading] = useState(true);
  const [availabilityLoading, setAvailabilityLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [eventModalOpen, setEventModalOpen] = useState(false);
  const [reservationModalOpen, setReservationModalOpen] = useState(false);
  const [eventForm, setEventForm] = useState<EventFormState>(initialEventForm);
  const [quantity, setQuantity] = useState(1);
  const [submitting, setSubmitting] = useState(false);
  const [receipt, setReceipt] = useState<Reservation | null>(null);

  const loadDashboard = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const [nextHealth, nextEvents] = await Promise.all([
        api.health(),
        api.listEvents(),
      ]);
      setHealth(nextHealth);
      setEvents(nextEvents);
      setSelectedId((current) => current ?? nextEvents[0]?.id ?? null);
    } catch (loadError) {
      setHealth(null);
      setError(loadError instanceof Error ? loadError.message : "Falha inesperada");
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    void loadDashboard();
  }, [loadDashboard]);

  const loadAvailability = useCallback(async (eventId: string) => {
    setAvailabilityLoading(true);
    try {
      setAvailability(await api.getAvailability(eventId));
    } catch (loadError) {
      setAvailability(null);
      setError(loadError instanceof Error ? loadError.message : "Falha inesperada");
    } finally {
      setAvailabilityLoading(false);
    }
  }, []);

  useEffect(() => {
    if (selectedId) void loadAvailability(selectedId);
  }, [loadAvailability, selectedId]);

  const selectedEvent = events.find((event) => event.id === selectedId) ?? null;

  const totals = useMemo(
    () => ({
      events: events.length,
      capacity: events.reduce((sum, event) => sum + event.capacity, 0),
      onSale: events.filter((event) => event.status === "on_sale").length,
    }),
    [events],
  );

  async function handleCreateEvent(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setSubmitting(true);
    setError(null);

    const input: CreateEventInput = {
      name: eventForm.name.trim(),
      venue: eventForm.venue.trim(),
      starts_at: new Date(eventForm.starts_at).toISOString(),
      sales_start_at: new Date(eventForm.sales_start_at).toISOString(),
      capacity: Number(eventForm.capacity),
      price_cents: Math.round(Number(eventForm.price.replace(",", ".")) * 100),
    };

    try {
      const created = await api.createEvent(input);
      setEvents((current) => [created, ...current]);
      setSelectedId(created.id);
      setEventForm(initialEventForm);
      setEventModalOpen(false);
    } catch (submitError) {
      setError(submitError instanceof Error ? submitError.message : "Falha inesperada");
    } finally {
      setSubmitting(false);
    }
  }

  async function handleReserve(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!selectedEvent) return;
    setSubmitting(true);
    setError(null);

    try {
      const reservation = await api.createReservation(selectedEvent.id, quantity);
      setReceipt(reservation);
      await loadAvailability(selectedEvent.id);
    } catch (submitError) {
      setError(submitError instanceof Error ? submitError.message : "Falha inesperada");
    } finally {
      setSubmitting(false);
    }
  }

  function closeReservationModal() {
    setReservationModalOpen(false);
    setReceipt(null);
    setQuantity(1);
  }

  return (
    <div className="app-shell">
      <div className="grain" aria-hidden="true" />

      <header className="topbar">
        <a className="brand" href="#top" aria-label="Ir para o início">
          <span className="brand-mark">
            <Ticket size={20} strokeWidth={1.8} />
          </span>
          <span>
            <strong>ENTRELINHAS</strong>
            <small>sala de controle</small>
          </span>
        </a>

        <div className="environment">
          <span className={`signal ${health ? "signal-online" : "signal-offline"}`} />
          <span>{isMockMode ? "laboratório simulado" : "API Go"}</span>
          <b>{health ? "conectada" : "indisponível"}</b>
        </div>
      </header>

      <main id="top">
        <section className="hero reveal">
          <div className="hero-copy">
            <span className="eyebrow">
              <Radio size={14} /> inventário em tempo real
            </span>
            <h1>
              Toda venda deixa
              <em> um sinal.</em>
            </h1>
            <p>
              Observe eventos, estoque e reservas enquanto constrói o espaço operacional
              que sustenta cada ingresso.
            </p>
          </div>

          <div className="hero-actions">
            <button className="button button-primary" onClick={() => setEventModalOpen(true)}>
              <Plus size={18} /> Criar evento
            </button>
            <button className="button button-quiet" onClick={() => void loadDashboard()}>
              <RefreshCw size={17} className={loading ? "spin" : ""} /> Atualizar
            </button>
          </div>
        </section>

        <section className="metrics reveal delay-1" aria-label="Resumo da operação">
          <Metric
            icon={<Activity size={18} />}
            label="Eventos mapeados"
            value={String(totals.events).padStart(2, "0")}
            note={`${totals.onSale} com vendas abertas`}
          />
          <Metric
            icon={<Users size={18} />}
            label="Capacidade total"
            value={totals.capacity.toLocaleString("pt-BR")}
            note="lugares nos espaços"
          />
          <Metric
            icon={<Zap size={18} />}
            label="Estado da API"
            value={health?.status === "ok" ? "Estável" : "Em espera"}
            note={`Redis: ${health?.redis ?? "sem leitura"}`}
          />
        </section>

        {error && (
          <div className="error-banner" role="alert">
            <ServerOff size={19} />
            <span>{error}</span>
            <button onClick={() => setError(null)} aria-label="Fechar aviso">
              <X size={17} />
            </button>
          </div>
        )}

        <section className="workspace reveal delay-2">
          <div className="event-list-panel">
            <div className="section-heading">
              <div>
                <span className="section-index">01</span>
                <h2>Próximos eventos</h2>
              </div>
              <span className="section-note">selecione para inspecionar</span>
            </div>

            {loading ? (
              <div className="loading-state">
                <LoaderCircle className="spin" /> Consultando o espaço…
              </div>
            ) : events.length === 0 ? (
              <div className="empty-state">
                <Ticket size={30} />
                <h3>Nenhum evento no espaço</h3>
                <p>Crie o primeiro evento para iniciar o laboratório.</p>
              </div>
            ) : (
              <div className="event-list">
                {events.map((event, index) => (
                  <button
                    key={event.id}
                    className={`event-row ${selectedId === event.id ? "event-row-selected" : ""}`}
                    onClick={() => setSelectedId(event.id)}
                  >
                    <span className="event-number">{String(index + 1).padStart(2, "0")}</span>
                    <span className="event-main">
                      <strong>{event.name}</strong>
                      <small>
                        <MapPin size={13} /> {event.venue}
                      </small>
                    </span>
                    <span className="event-date">
                      {dateTime.format(new Date(event.starts_at))}
                    </span>
                    <span className={`status status-${event.status}`}>
                      {statusLabels[event.status]}
                    </span>
                    <ChevronRight className="row-arrow" size={18} />
                  </button>
                ))}
              </div>
            )}
          </div>

          <aside className="inspection-panel">
            <div className="section-heading section-heading-dark">
              <div>
                <span className="section-index">02</span>
                <h2>Leitura do espaço</h2>
              </div>
            </div>

            {selectedEvent ? (
              <>
                <div className="inspection-title">
                  <p>{selectedEvent.venue}</p>
                  <h3>{selectedEvent.name}</h3>
                  <div className="event-meta">
                    <span>
                      <CalendarDays size={15} />
                      {dateTime.format(new Date(selectedEvent.starts_at))}
                    </span>
                    <span>
                      <CircleDollarSign size={15} />
                      {currency.format(selectedEvent.price_cents / 100)}
                    </span>
                  </div>
                </div>

                {availabilityLoading || !availability ? (
                  <div className="availability-loading">
                    <LoaderCircle className="spin" /> Lendo inventário
                  </div>
                ) : (
                  <AvailabilityDial availability={availability} />
                )}

                <button
                  className="button button-accent button-full"
                  onClick={() => setReservationModalOpen(true)}
                  disabled={selectedEvent.status !== "on_sale" || !availability?.available}
                >
                  Simular reserva <ArrowUpRight size={18} />
                </button>

                <div className="space-key">
                  <span>chave de afinidade</span>
                  <code>{`{${selectedEvent.id.slice(0, 8)}}:inventory`}</code>
                </div>
              </>
            ) : (
              <div className="empty-inspection">Selecione um evento para inspecionar.</div>
            )}
          </aside>
        </section>

        <section className="learning-strip reveal delay-3">
          <span className="section-index">03</span>
          <div>
            <span className="eyebrow">próximo experimento</span>
            <h2>Faça o endpoint de saúde responder.</h2>
          </div>
          <p>
            Mude o frontend para modo HTTP e veja este painel reconhecer o seu primeiro
            processo Go.
          </p>
          <a href="https://pkg.go.dev/net/http" target="_blank" rel="noreferrer">
            Aula 0 <ArrowUpRight size={16} />
          </a>
        </section>
      </main>

      <footer>
        <span>LAB / SPACE-BASED ARCHITECTURE</span>
        <span>Campo Grande · MS</span>
        <span>versão {health?.version ?? "—"}</span>
      </footer>

      {eventModalOpen && (
        <Modal title="Abrir um novo evento" eyebrow="cadastro operacional" onClose={() => setEventModalOpen(false)}>
          <form className="form" onSubmit={handleCreateEvent}>
            <label className="field field-wide">
              <span>Nome do evento</span>
              <input
                required
                minLength={3}
                value={eventForm.name}
                onChange={(event) => setEventForm({ ...eventForm, name: event.target.value })}
                placeholder="Ex.: Festival do Cerrado"
              />
            </label>
            <label className="field field-wide">
              <span>Local</span>
              <input
                required
                minLength={3}
                value={eventForm.venue}
                onChange={(event) => setEventForm({ ...eventForm, venue: event.target.value })}
                placeholder="Espaço e cidade"
              />
            </label>
            <label className="field">
              <span>Data do evento</span>
              <input
                required
                type="datetime-local"
                value={eventForm.starts_at}
                onChange={(event) => setEventForm({ ...eventForm, starts_at: event.target.value })}
              />
            </label>
            <label className="field">
              <span>Início das vendas</span>
              <input
                required
                type="datetime-local"
                value={eventForm.sales_start_at}
                onChange={(event) =>
                  setEventForm({ ...eventForm, sales_start_at: event.target.value })
                }
              />
            </label>
            <label className="field">
              <span>Capacidade</span>
              <input
                required
                min={1}
                type="number"
                value={eventForm.capacity}
                onChange={(event) => setEventForm({ ...eventForm, capacity: event.target.value })}
                placeholder="1000"
              />
            </label>
            <label className="field">
              <span>Preço em reais</span>
              <input
                required
                min={0}
                step="0.01"
                inputMode="decimal"
                value={eventForm.price}
                onChange={(event) => setEventForm({ ...eventForm, price: event.target.value })}
                placeholder="85,00"
              />
            </label>
            <button className="button button-primary field-wide" disabled={submitting}>
              {submitting ? <LoaderCircle className="spin" size={18} /> : <Plus size={18} />}
              Criar evento
            </button>
          </form>
        </Modal>
      )}

      {reservationModalOpen && selectedEvent && (
        <Modal title={receipt ? "Reserva registrada" : "Simular uma reserva"} eyebrow={selectedEvent.name} onClose={closeReservationModal}>
          {receipt ? (
            <div className="receipt">
              <span className="receipt-check"><Check size={28} /></span>
              <p>O espaço separou {receipt.quantity} ingresso(s).</p>
              <dl>
                <div><dt>Reserva</dt><dd>{receipt.id}</dd></div>
                <div><dt>Expira</dt><dd>{dateTime.format(new Date(receipt.expires_at))}</dd></div>
                <div><dt>Status</dt><dd>{receipt.status}</dd></div>
              </dl>
              <button className="button button-primary button-full" onClick={closeReservationModal}>
                Concluir
              </button>
            </div>
          ) : (
            <form className="reservation-form" onSubmit={handleReserve}>
              <div className="quantity-picker">
                <button type="button" onClick={() => setQuantity((value) => Math.max(1, value - 1))}>−</button>
                <div><strong>{quantity}</strong><span>ingresso(s)</span></div>
                <button type="button" onClick={() => setQuantity((value) => Math.min(10, value + 1))}>+</button>
              </div>
              <div className="reservation-summary">
                <span>Total simulado</span>
                <strong>{currency.format((selectedEvent.price_cents * quantity) / 100)}</strong>
              </div>
              <p className="form-note"><Clock3 size={15} /> A reserva ficará ativa por dez minutos.</p>
              <button className="button button-accent button-full" disabled={submitting}>
                {submitting ? <LoaderCircle className="spin" size={18} /> : <Zap size={18} />}
                Reservar atomicamente
              </button>
            </form>
          )}
        </Modal>
      )}
    </div>
  );
}

function Metric({
  icon,
  label,
  value,
  note,
}: {
  icon: React.ReactNode;
  label: string;
  value: string;
  note: string;
}) {
  return (
    <article className="metric">
      <div className="metric-icon">{icon}</div>
      <span>{label}</span>
      <strong>{value}</strong>
      <small>{note}</small>
    </article>
  );
}

function AvailabilityDial({ availability }: { availability: Availability }) {
  const percentage = Math.round((availability.available / availability.capacity) * 100);
  const dialStyle = {
    "--availability": `${percentage * 3.6}deg`,
  } as React.CSSProperties;

  return (
    <div className="availability">
      <div className="dial" style={dialStyle}>
        <div>
          <strong>{percentage}%</strong>
          <span>livre</span>
        </div>
      </div>
      <div className="availability-stats">
        <div><span>Disponíveis</span><strong>{availability.available.toLocaleString("pt-BR")}</strong></div>
        <div><span>Reservados</span><strong>{availability.reserved.toLocaleString("pt-BR")}</strong></div>
        <div><span>Vendidos</span><strong>{availability.sold.toLocaleString("pt-BR")}</strong></div>
        <div><span>Versão</span><strong>#{availability.version}</strong></div>
      </div>
    </div>
  );
}

function Modal({
  title,
  eyebrow,
  onClose,
  children,
}: {
  title: string;
  eyebrow: string;
  onClose: () => void;
  children: React.ReactNode;
}) {
  return (
    <div className="modal-backdrop" role="presentation" onMouseDown={(event) => event.target === event.currentTarget && onClose()}>
      <section className="modal" role="dialog" aria-modal="true" aria-labelledby="modal-title">
        <button className="modal-close" onClick={onClose} aria-label="Fechar janela"><X size={20} /></button>
        <span className="eyebrow">{eyebrow}</span>
        <h2 id="modal-title">{title}</h2>
        {children}
      </section>
    </div>
  );
}

export default App;
