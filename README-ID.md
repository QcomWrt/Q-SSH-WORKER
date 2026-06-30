# Q-SSH-WORKER

🇺🇸 **English:** [README.md](README.md)

Q-SSH-WORKER adalah SSH Client dan transport engine modular yang ditulis sepenuhnya menggunakan bahasa Go.

Proyek ini dikembangkan sebagai pengganti modern berbagai utilitas SSH yang selama ini umum digunakan pada lingkungan tunneling, seperti **http.py**, **sshpass**, **corkscrew**, serta berbagai shell script dan helper binary lainnya.

Alih-alih menggabungkan banyak aplikasi berbeda, Q-SSH-WORKER mengintegrasikan seluruh proses mulai dari pembuatan koneksi jaringan, modifikasi transport, autentikasi SSH, hingga penyediaan SOCKS5 Proxy ke dalam satu aplikasi yang ringan, modular, dan mudah dikembangkan.

Q-SSH-WORKER merupakan SSH Engine utama yang digunakan oleh proyek **Q-LOAD**, namun tetap dapat dijalankan sebagai aplikasi mandiri (standalone).

---

# Mengapa Q-SSH-WORKER?

Pada umumnya, SSH tunneling membutuhkan beberapa aplikasi sekaligus, misalnya:

* sshpass
* http.py
* corkscrew
* shell script
* helper binary lainnya

Kombinasi tersebut cukup sulit dipelihara karena setiap aplikasi memiliki konfigurasi, dependensi, dan perilaku yang berbeda.

Q-SSH-WORKER menghilangkan seluruh ketergantungan tersebut dengan mengimplementasikan seluruh fungsinya secara native menggunakan Go.

Seluruh fitur seperti:

* HTTP CONNECT
* HTTP Payload Injection
* TLS
* WebSocket
* SOCKS5
* Auto Reconnect
* DNS Resolver

diimplementasikan langsung di dalam aplikasi tanpa memerlukan program eksternal.

Dengan demikian Q-SSH-WORKER **tidak lagi membutuhkan aplikasi seperti http.py maupun corkscrew** untuk membangun SSH Tunnel melalui proxy ataupun payload HTTP.

---

# Tujuan Proyek

Q-SSH-WORKER dirancang dengan tujuan:

* Membangun SSH Client modern menggunakan Go.
* Menggantikan utilitas seperti http.py, sshpass, corkscrew, dan helper serupa.
* Mendukung berbagai metode transport secara modular.
* Mendukung HTTP CONNECT Proxy.
* Mendukung HTTP Payload Injection.
* Mendukung TLS.
* Mendukung WebSocket.
* Mendukung gRPC pada pengembangan berikutnya.
* Menyediakan SOCKS5 Proxy lokal.
* Mendukung Auto Reconnect.
* Mudah diintegrasikan dengan Q-LOAD.
* Cross Platform.
* Minim dependensi eksternal.

---

# Arsitektur

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

Setiap layer memiliki tanggung jawab masing-masing sehingga dapat dikembangkan secara independen tanpa mempengaruhi layer lainnya.

---

# Struktur Direktori

```text
config/
    Loader dan parser konfigurasi.

network/
    Layer pembuat koneksi jaringan.
    - TCP
    - WebSocket
    - gRPC (planned)
    - DNS Resolver
    - Dialer

transport/
    Layer modifikasi koneksi.
    - HTTP
    - Payload
    - Proxy
    - TLS

ssh/
    SSH Client.
    - Authentication
    - Session
    - Client

socks/
    SOCKS5 Server.

worker/
    Worker Manager.
    Reconnect.
    Health Check.
    Statistics.

logger/
    Logging.

internal/
    Shared utilities.

examples/
    Contoh konfigurasi.
```

---

# Filosofi Desain

Q-SSH-WORKER menggunakan prinsip **Single Responsibility**.

Setiap package hanya memiliki satu tanggung jawab.

* **Network** bertugas membuat koneksi dasar.
* **Transport** memodifikasi koneksi tersebut.
* **SSH** melakukan autentikasi dan membuat SSH Session.
* **SOCKS5** menyediakan endpoint lokal.
* **Worker** mengelola reconnect, monitoring, dan statistik.

Dengan pendekatan ini, penambahan fitur baru seperti WebSocket, gRPC, ataupun transport lainnya tidak memerlukan perubahan besar pada implementasi SSH.

---

# Alur Kerja

Secara sederhana Q-SSH-WORKER bekerja seperti berikut:

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

Untuk koneksi WebSocket:

```text
WebSocket
      │
      ▼
TLS (optional)
      │
      ▼
SSH
```

Untuk pengembangan berikutnya:

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

# Konfigurasi

Konfigurasi menggunakan format JSON.

Contoh:

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

# Fitur

Fitur yang tersedia maupun sedang dikembangkan:

* Native SSH Password Authentication
* Native Public Key Authentication
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
* Tidak membutuhkan helper binary eksternal

---

# Roadmap

## Core

* [x] Redesain arsitektur proyek
* [ ] Loader konfigurasi
* [ ] Worker Manager

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

## Integrasi

* [ ] Integrasi dengan Q-LOAD

---

# Lisensi

Q-SSH-WORKER dirilis menggunakan lisensi MIT.

Silakan digunakan, dimodifikasi, maupun didistribusikan sesuai ketentuan lisensi MIT.
