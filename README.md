# Ledger Events

> Go backend service responsible for managing accounting events within the Ledger platform.

## Overview

Ledger Events is a backend service responsible for processing and exposing accounting events for the Ledger ecosystem.

The service is designed as an independent backend component, isolating accounting event processing from other services in the platform.

The project was developed in Go and follows a modular structure that separates the application entry point, configuration, internal implementation and supporting infrastructure.

## Architecture

The project is organized around a modular backend structure:

```text
ledger-events/
├── cmd/
├── config/
├── docs/
├── internal/
├── scripts/
├── .env
├── Makefile
├── docker-compose.yaml
├── go.mod
└── go.sum
```

### `cmd/`

Contains the application entry point and service initialization.

### `config/`

Contains application configuration and configuration-related components.

### `internal/`

Contains the application's internal implementation and keeps implementation details encapsulated within the service.

### `docs/`

Contains project documentation and architectural resources.

### `scripts/`

Contains supporting scripts and infrastructure resources used during development and deployment.

## Responsibilities

Ledger Events is responsible for the accounting-event layer of the Ledger ecosystem.

Its main responsibility is to provide a dedicated backend component for accounting events and their processing.

The service helps isolate accounting concerns from other components of the platform, allowing them to evolve independently.

## Role in the Ledger Ecosystem

Ledger Events is part of a broader set of backend services designed around accounting and ledger processing.

```text
                         ┌───────────────────┐
                         │   Ledger Config   │
                         │   Configuration   │
                         └─────────┬─────────┘
                                   │
                                   │
                         ┌─────────▼─────────┐
                         │   Ledger Events   │
                         │                   │
                         │     Go API        │
                         └─────────┬─────────┘
                                   │
                                   │ Events
                                   ▼
                         ┌───────────────────┐
                         │   Ledger Worker   │
                         │                   │
                         │ Async Processing  │
                         └───────────────────┘
```

The separation of responsibilities allows the different components of the Ledger platform to evolve independently.

## Project Structure

```text
cmd/
    Application entry point

config/
    Application configuration

docs/
    Documentation and architectural resources

internal/
    Application implementation

scripts/
    Development and infrastructure resources
```

## Design Principles

The project follows principles commonly used in modern backend systems:

- **Separation of concerns** — different application responsibilities are isolated.
- **Encapsulation** — internal implementation details remain within the service.
- **Service isolation** — accounting event processing is implemented as an independent backend component.
- **Modularity** — application components are organized into dedicated packages.
- **Cloud-native development** — the repository includes container and infrastructure resources for local development and deployment.

## Local Development

### Requirements

- Go
- Docker
- Docker Compose

### Clone

```bash
git clone https://github.com/clodoaldomarques/ledger-events.git
cd ledger-events
```

### Install dependencies

```bash
go mod download
```

### Run the application

```bash
go run ./cmd/...
```

### Run tests

```bash
go test ./...
```

### Run with Docker Compose

```bash
docker compose up
```

## Development Automation

The repository includes a `Makefile` to simplify common development and operational tasks.

To inspect the available commands:

```bash
make
```

## Technology Stack

| Technology | Purpose |
|---|---|
| Go | Backend service |
| Docker | Containerization |
| Docker Compose | Local development environment |
| Make | Development automation |

## Engineering Concepts

This project demonstrates practical backend engineering concepts including:

- Go backend development
- API development
- Service-oriented architecture
- Separation of concerns
- Modular application structure
- Configuration management
- Containerization
- Local development automation
- Distributed backend architecture

## Relationship with Other Projects

Ledger Events is part of a broader portfolio of Go backend projects focused on distributed systems and cloud-native architecture.

### Core SDK

Provides reusable infrastructure components such as AWS integrations, messaging and observability.

### Ledger Worker

Provides asynchronous processing of ledger-related events using a worker-based architecture.

### Ledger Config

Provides configuration capabilities for the Ledger ecosystem.

Together, these projects demonstrate different aspects of building distributed backend systems with Go.

## Project Status

This project is part of my Go backend engineering portfolio and serves as a practical exploration of accounting-event processing, service-oriented architecture and distributed backend systems.

The project is intended primarily as an engineering study and portfolio project rather than a production-ready financial platform.

## Author

**Clodoaldo Marques**

Backend Software Engineer focused on Go, Microservices, Distributed Systems and Cloud-Native architectures.

- GitHub: https://github.com/clodoaldomarques
- LinkedIn: https://www.linkedin.com/in/clodoaldomarques/