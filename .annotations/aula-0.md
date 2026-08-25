`serverErr := make(chan error, 1)`

Isso cria um canal chamado `serverErr` que: 
- transporta valores do tipo *error*
- possui capacidade para armazenar 1 valor sem que o envio bloqueie
- começa vazio

---
 ### Context Package
 
 Package context defines the Context type, which carries deadlines, cancellation signals, and other request-scoped values across API boundaries and between processes

Incoming requests to a server should create a Context, and outgoing calls to servers should accept a Context. The chain of function calls between them must propagate the Context, optionally replacing it with a derived Context created using WithCancel, WithDeadline, WithTimeout, or WithValue.

--- 
### Em resumo sobre a `main`:

  inicia servidor
        ↓
  aguarda Ctrl+C, SIGTERM ou erro
        ↓
  recebe sinal?
        ↓
  encerra servidor com até 5 segundos
        ↓
  aguarda a finalização definitiva