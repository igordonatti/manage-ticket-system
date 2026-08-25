 ## Caminho de implementação

  1. MVP local: React, Go, Redis standalone e PostgreSQL com Docker Compose.
  2. Concorrência: reserva atômica, idempotência, expiração e teste contra overselling.
  3. Assincronismo: outbox, NATS JetStream e persistence worker.
  4. Distribuição: Redis Cluster e particionamento por event_id.
  5. Resiliência: múltiplas processing units, reinicializações e recuperação de backlog.
  6. Operação: OpenTelemetry, Prometheus e Grafana. OpenTelemetry para Go
     (https://opentelemetry.io/docs/languages/go/)

  7. Elasticidade: Kubernetes e HPA baseado em latência, requisições ou tamanho do backlog, não apenas
     CPU. O HPA suporta métricas customizadas. Kubernetes HPA
     (https://kubernetes.io/docs/concepts/workloads/autoscaling/horizontal-pod-autoscale/)

  8. Validação: k6 simulando abertura de vendas e concentração de tráfego em um único evento. Testes de
     API com k6 (https://grafana.com/docs/k6/latest/testing-guides/api-load-testing/)