# Minimal distributed transaction using two-phase commit (2PC)

```mermaid
sequenceDiagram
  autonumber
  participant C as Coordinator
  participant P1 as Participant 1
  participant P2 as Participant 2

  Note over C, P2: Two-Phase Commit (2PC)

  rect rgb(235, 245, 255)
    Note over C: Phase 1 — Prepare / Vote
    C->>P1: PREPARE(txId, writeSet)
    C->>P2: PREPARE(txId, writeSet)
    P1-->>C: VOTE-YES (prepared)
    P2-->>C: VOTE-YES (prepared)
    Note over C: All YES => decision = COMMIT
  end

  rect rgb(235, 255, 235)
    Note over C: Phase 2 — Decision
    C->>P1: COMMIT(txId)
    C->>P2: COMMIT(txId)
    P1-->>C: ACK
    P2-->>C: ACK
    Note over C,P1,P2: Transaction COMMITTED
  end

  %% --- Abort path (any NO or timeout) ---
  alt Any participant votes NO or times out
    rect rgb(255, 240, 240)
      Note over C: Phase 1 outcome => decision = ABORT
      C->>P1: ABORT(txId)
      C->>P2: ABORT(txId)
      P1-->>C: ACK
      P2-->>C: ACK
      Note over C,P1,P2: Transaction ABORTED
    end
  end

  %% --- Failure note: prepared participants may block ---
  Note over P1,P2: If a participant is PREPARED and loses contact with Coordinator,\n it may block waiting for the final decision (classic 2PC blocking).
```
