import http from "k6/http";
import { check } from "k6";
import exec from "k6/execution";

const baseUrl = __ENV.BASE_URL || "http://localhost:8080";
const eventId = __ENV.EVENT_ID;

if (!eventId) {
  throw new Error("Defina EVENT_ID com o evento que receberá a carga.");
}

export const options = {
  scenarios: {
    abertura_de_vendas: {
      executor: "shared-iterations",
      vus: Number(__ENV.VUS || 100),
      iterations: Number(__ENV.ITERATIONS || 500),
      maxDuration: "45s",
    },
  },
  thresholds: {
    http_req_duration: ["p(95)<500"],
    http_req_failed: ["rate<0.90"],
  },
};

export default function () {
  const requestId = `${exec.scenario.iterationInTest}-${exec.vu.idInTest}`;
  const response = http.post(
    `${baseUrl}/events/${eventId}/reservations`,
    JSON.stringify({ quantity: 1 }),
    {
      headers: {
        "Content-Type": "application/json",
        "Idempotency-Key": `k6-${requestId}`,
      },
    },
  );

  check(response, {
    "reserva criada ou estoque esgotado": (res) =>
      res.status === 201 || res.status === 409,
  });
}

