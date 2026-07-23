# CBC Chain Explorer

High-precision blockchain explorer for **CBC Chain (Caerulean ByteChains)**, supporting the **Dynamic Consensus Framework (DCF)**, **Dynamic Voting Framework (DVF)**, **Proof of Inference (PoI)**, and **Proof of Stake (PoS)**.

---

## Table of Contents

- [Features](#features)
- [Quick Start](#quick-start)
- [Configuration](#configuration)
- [Architecture & API Catalog](#architecture--api-catalog)
- [License](#license)

---

## Features

- **CBC Chain Integration**
  - Live indexing of blocks, extrinsics, events, and consensus telemetry.
  - Native support for DCF (Dynamic Consensus Framework) PoS/PoI Trust Scores.
  - Native support for DVF (Dynamic Voting Framework) BFT finality quorums.
  - Validator Leaderboard (Alice, Bob, Charlie).
- **Extensible Architecture**
  - High-density REST API layer on port `4399`.
  - React/Next.js frontend interface on port `3000`.

---

## Quick Start

### Prerequisites

- Linux / macOS
- Go 1.23+
- Node.js 18+
- Docker & Docker Compose

### Starting the Explorer Stack

```bash
# Clean restart nodes, database, explorer backend, and React UI
./clean-restart.sh
```

- **Frontend Dashboard**: [http://localhost:3000](http://localhost:3000)
- **Explorer API**: `http://localhost:4399/api/scan`

---

## Configuration

Initialize or edit `configs/config.yaml`:

```yaml
server:
  http:
    addr: 0.0.0.0:4399
    timeout: 10s
database:
  mysql:
    api: "mysql://root:helloload@127.0.0.1:3306/cbc_explorer?writeTimeout=3s&parseTime=true&loc=Local&charset=utf8mb4,utf8"
redis:
  proto: tcp
  addr: 127.0.0.1:6379
UI:
  enable_substrate: true
  enable_evm: false
```

---

## Service Control

- **Start All**: `./clean-restart.sh`
- **Stop All**: `./stop-cbc-services.sh`
- **Build Binary**: `./build.sh build`

---

## LICENSE

MIT / GPL-3.0
