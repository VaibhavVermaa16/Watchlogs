# Watchlogs

**High-Performance Log Ingestion & Search Engine in Go**

Watchlogs is a lightweight, concurrency-safe log observability system designed for speed and predictability. It features an append-only storage engine, an in-memory inverted index, and an asynchronous ingestion pipeline tailored for real-time observability.

## 🚀 Key Features

- **Async Ingestion Pipeline:** Implemented a queue-based mechanism using **Go channels** to handle high concurrency without blocking the main execution path.
- **Smart Indexing & Bounded Memory:**
  - **Capped Indices:** To prevent memory bloat, we limit log ingestion per token. When limits are reached, the oldest log IDs are removed.
  - **Philosophy:** *Useful data > Complete data.* We prioritize recent, actionable insights over infinite history for observability.
- **Query Normalization:** Supports multi-word queries (e.g., "login Failed") agnostic to casing and punctuation.
- **Automatic Log Rotation:**
  - **Retention:** Logs older than 24 hours are discarded; the index is rebuilt automatically.
  - **Speed over Space:** We prefer deletion over compression for predictable performance.
- **Graceful Shutdown:** Ensures data in the channel is flushed to disk before exit to prevent data loss.

## 🛠 Architecture & Trade-offs

We made conscious design choices to balance **Latency vs. Durability**:

| Scenario | Behavior | Consequence |
| :--- | :--- | :--- |
| **Channel Fills** | Sender blocks; client waits. | **System Survives.** Backpressure slows flow but preserves data. |
| **Process Crash** | Crash before flush. | **Acceptable Risk.** We trade strict durability for lower latency. |
| **Disk Fills** | OS returns error; logs dropped. | **Corrupt State.** Recovery impossible; disk monitoring is required. |

## 🛡️ Correctness & Recovery

The system treats the **Disk as the Source of Truth**.

- **Crash Recovery:** Index and memory are rebuilt from disk on restart.
- **Partial Writes:** Broken lines or garbage JSON at the end of a file (caused by crashes during writes) are detected and ignored during rebuild.
- **Consistency Model:**
  - *Crash before write:* Log lost (Acceptable).
  - *Crash during write:* Garbage data ignored.
  - *Write done, index missing:* Fixed on rebuild.

## ⚡ Performance

**Search Complexity:**
`Cost ≈ size(t1) + size(t2) + intersections`

**Resource Management:**
- **Capped (Bounded):** Memory usage, index entries, search result size, channel buffer.
- **Grows (Until Rotation):** Total logs on disk, rebuild time.

## 📦 Core Components

1. **Ingestion Pipeline** (Async via Channels)
2. **Search Engine Core** (In-memory Inverted Index)
3. **Storage Layer** (Append-only, Restart-safe)
4. **Retention Manager** (24hr Rotation)

---
*Built with Go.*

## ✅ Quickstart

### Prerequisites
- Go 1.20+ installed

### Setup & Run
```bash
make setup
make run
```

### Build
```bash
make build
```

### Run Tests
```bash
make test
```

### Format Code
```bash
make fmt
```

### Clean
```bash
make clean
```