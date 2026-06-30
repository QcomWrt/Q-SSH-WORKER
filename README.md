# Q-SSH-WORKER

> 🌐 **Documentation**
>
> - 🇺🇸 English (Current)
> - 🇮🇩 [Bahasa Indonesia](README-ID.md)

Q-SSH-WORKER is a modular SSH client and transport engine written entirely in Go.

The project is designed as a modern replacement for various standalone SSH tunneling utilities commonly used in tunneling environments, such as **http.py**, **sshpass**, **corkscrew**, shell scripts, and other helper binaries.

Instead of combining multiple external applications, Q-SSH-WORKER integrates the entire workflow—from network connection, transport processing, SSH authentication, to providing a local SOCKS5 proxy—into a single lightweight, modular, and extensible application.

Q-SSH-WORKER serves as the primary SSH engine for the **Q-LOAD** ecosystem while also being fully usable as a standalone application.

---

# Why Q-SSH-WORKER?

Traditional SSH tunneling solutions usually depend on several external programs, including:

* sshpass
* http.py
* corkscrew
* shell scripts
* additional helper binaries

Managing multiple external tools makes deployment and maintenance more complicated.

Q-SSH-WORKER removes these dependencies by implementing everything natively in Go.

Features such as:

* HTTP CONNECT
* HTTP Payload Injection
* TLS
* WebSocket
* SOCKS5
* Automatic Reconnect
* DNS Resolver

are implemented directly inside the application.

As a result, **Q-SSH-WORKER no longer requires external utilities such as http.py or corkscrew** to establish SSH tunnels through HTTP proxies or custom payloads.

---

# Project Goals

Q-SSH-WORKER is designed with the following goals:

* Build a modern SSH client in Go.
* Replace utilities such as http.py, sshpass, corkscrew, and similar helper programs.
* Provide a modular transport pipeline.
* Support HTTP CONNECT proxy.
* Support HTTP payload injection.
* Support TLS.
* Support WebSocket.
* Support gRPC in future releases.
* Provide a local SOCKS5 proxy.
* Support automatic reconnect.
* Integrate seamlessly with Q-LOAD.
* Cross-platform.
* Minimal external dependencies.

---

# Architecture

```text
                Configuration
                      │
                      ▼
                 Network Layer
        (TCP / WebSocket / gRPC)
                      │
                      ▼
               Transport Layer
      (Proxy → HTTP → Payload → TLS)
                      │
                      ▼
                  SSH Client
      (Authentication & SSH Session)
                      │
                      ▼
                SOCKS5 Server
                      │
                      ▼
                    Q-LOAD
```

Each layer has a single responsibility, allowing independent development and maintenance without affecting the others.

---

# Directory Structure

```text
config/
    Configuration loader and parser.

network/
    Network connection layer.
    - TCP
    - WebSocket
    - gRPC (planned)
    - DNS Resolver
    - Dialer

transport/
    Connection modifier layer.
    - HTTP
    - Payload
    - Proxy
    - TLS

ssh/
    SSH client implementation.
    - Authentication
    - Session
    - Client

socks/
    SOCKS5 server.

worker/
    Worker manager.
    Reconnect.
    Health monitoring.
    Statistics.

logger/
    Logging.

internal/
    Shared utilities.

examples/
    Configuration examples.
```

---

# Design Philosophy

Q-SSH-WORKER follows the **Single Responsibility Principle**.

Each package is responsible for only one task.

* **Network** creates the underlying connection.
* **Transport** modifies the connection.
* **SSH** performs authentication and establishes SSH sessions.
* **SOCKS5** exposes the local proxy interface.
* **Worker** manages reconnect, monitoring, and statistics.

This modular architecture allows new transports such as WebSocket or gRPC to be added without changing the SSH implementation.

---

# Workflow

A typical TCP connection works as follows:

```text
TCP Connection
        │
        ▼
HTTP CONNECT (optional)
        │
        ▼
Payload Injection (optional)
        │
        ▼
TLS (optional)
        │
        ▼
SSH Handshake
        │
        ▼
SOCKS5 Server
```

For WebSocket:

```text
WebSocket
      │
      ▼
TLS (optional)
      │
      ▼
SSH
```

Future gRPC support:

```text
gRPC
  │
  ▼
TLS
  │
  ▼
SSH
```

---

# Configuration

Configuration is stored in JSON format.

Example:

```json
{
    "listen": {
        "host": "127.0.0.1",
        "port": 1080
    },

    "ssh": {
        "host": "server.example.com",
        "port": 22,
        "username": "user",
        "password": "password"
    },

    "network": {
        "type": "tcp"
    },

    "transport": {
        "tls": false,
        "host": "",
        "path": "/",
        "sni": ""
    },

    "payload": {
        "enable": false
    },

    "worker": {
        "max_retry": 0
    }
}
```

---

# Features

Current and planned features include:

* Native SSH Password Authentication
* Native SSH Public Key Authentication
* Native Keyboard Interactive Authentication
* Native HTTP CONNECT Proxy
* Native HTTP Payload Injection
* Native TLS
* Native WebSocket
* Native DNS Resolver
* SOCKS5 Server
* Worker Manager
* Automatic Reconnect
* Health Monitoring
* Lightweight Memory Usage
* Modular Architecture
* Cross Platform
* No external helper binaries required

---

# Roadmap

## Core

* [x] Project architecture redesign
* [ ] Configuration loader
* [ ] Worker manager

## Network

* [ ] TCP
* [ ] WebSocket
* [ ] gRPC

## Transport

* [ ] HTTP CONNECT
* [ ] Payload Injection
* [ ] TLS

## SSH

* [ ] SSH Client
* [ ] Authentication
* [ ] Session

## SOCKS

* [ ] SOCKS5 Server

## Integration

* [ ] Q-LOAD integration

---

# License

Q-SSH-WORKER is released under the MIT License.

You are free to use, modify, and distribute this project under the terms of the MIT License.
