## golang-p2p-messenger

A simple peer-to-peer messenger written in Go. It provides:

- A lightweight discovery service (STUN-like HTTP server) backed by Redis for registering peers and looking up their TCP/UDP addresses
- A CLI peer application with an interactive shell to:
  - start/register a peer with the discovery server
  - list or lookup peers
  - send text messages over TCP
  - send images over UDP in small packets and reassemble them on the receiver

### Features

- **Text messaging (TCP)**: fixed-length header framing, then payload
- **Image transfer (UDP)**: images are split into packets of 256 pixels, sent over UDP, and reassembled by the receiver. Basic ACK support exists in the receiver, but the sender currently runs without retry logic enabled
- **Discovery via HTTP**: peers register their `username`, `tcp_addr`, and `udp_addr` with the server and query other peers by username

## Project layout

- `stun/`: HTTP discovery server main
- `peer/`: CLI peer main (interactive shell)
- `cmd/root/`: interactive shell implementation (receivers, prompt, I/O)
- `cmd/peer/`: `peer` command and subcommands
  - `start`: register this peer on the discovery server
  - `get`: list peers or fetch one by username
  - `send text`: send a text message over TCP
  - `send image`: send an image over UDP
- `internal/`: reusable packages (`protocol`, `imgutil`, `stun`, etc.)

## Prerequisites

- **Go** 1.19+
- **Redis** 6/7 running locally on `localhost:6379`
  - Example (Docker):
    ```sh
    docker run --name redis -p 6379:6379 -d redis:7
    ```

## Build

```sh
go build -o bin/stun ./stun
go build -o bin/messenger ./peer
```

Or run directly with `go run` (shown below).

## Configuration

The peer CLI reads configuration with `viper`. Default config file is `config.yaml` in the working directory. Example:

```yaml
server: http://localhost:8080
tcp-port: "8083"
udp-port: "8084"
username: alice
```

You can also override at runtime using flags.

### Peer CLI flags

- Top-level:
  - `--tcp-port, -t`: TCP port to listen on (default `8081`)
  - `--udp-port, -u`: UDP port to listen on (default `8082`)
  - `--config, -c`: path to config file (default `config.yaml`)
- `peer` command (persistent across subcommands):
  - `--username, -n`: your username (required for `peer start` and for image sending metadata)
  - `--server, -s`: discovery server URL (default `http://localhost:8080`)

## Run

1) Start Redis (see prerequisites)

2) Start the discovery server

```sh
go run ./stun
```

This starts HTTP on `localhost:8080` with endpoints under `/peer/`.

3) Start a peer (interactive shell)

In terminal A:

```sh
go run ./peer --tcp-port 8083 --udp-port 8084
```

Register the peer with a username:

```
peer start -n alice
```

In terminal B:

```sh
go run ./peer --tcp-port 8085 --udp-port 8086
```

Register the second peer:

```
peer start -n bob
```

Now `alice` and `bob` are discoverable via the server.

## Usage examples

From `alice` shell:

- **List all peers**
  ```
  peer get --all
  ```

- **Lookup a single peer**
  ```
  peer get bob
  ```

- **Send text to `bob` (TCP)**
  ```
  peer send text bob "hello bob"
  ```
  On `bob`, the message appears in the console.

- **Send image to `bob` (UDP)**
  ```
  peer send image bob pic.jpg
  ```
  Supported formats for saving on the receiver side: `.png`, `.jpg`/`.jpeg`, `.gif`.
  The receiver writes the file as `new<original-filename>` (e.g., `newpic.jpg`).

- **Exit the peer shell**
  ```
  exit
  ```

## HTTP API (discovery server)

- `POST /peer/`
  - Request JSON:
    ```json
    { "udp_addr": "localhost:8084", "tcp_addr": "localhost:8083", "username": "alice" }
    ```
  - Responses:
    - `200 OK`: `{ "ok": true }`
    - `409 Conflict`: `{ "ok": false, "error": "username ... already exists" }`

- `GET /peer/`
  - Response: `{ "ok": true, "peers": [ { "username": "alice", "tcp_addr": "...", "udp_addr": "..." }, ... ] }`

- `GET /peer/{username}`
  - `200 OK` on success, `404 Not Found` if absent

## Protocol details

- **Text (TCP)**
  - Sender connects to target `tcp_addr`
  - Message format: 64-byte ASCII header containing the decimal length of the payload (left-padded with zeros), followed by the UTF-8 payload

- **Image (UDP)**
  - Sender connects to target `udp_addr`
  - Image is converted to RGBA matrix and chunked into packets of 256 pixels each
  - Packet schema:
    ```json
    {
      "Sender": "alice",
      "Filename": "pic.jpg",
      "Width": 1024,
      "Height": 768,
      "Row": 0,
      "Offset": 0,
      "Pixels": [ uint32 x 256 ]
    }
    ```
  - Receiver stores unique packets, acknowledges each with a small JSON ACK, and reassembles when all expected packets are received
  - Note: the optional retry/ACK handling in the sender is disabled by default; transfers are best-effort over UDP

## Notes and limitations

- Registration currently uses `localhost:<port>` for peer addresses; run peers on the same machine or adjust to your network environment
- No authentication, encryption, or NAT traversal. Intended for local demos and learning
- The receiver uses the output filename `new<original>` and relies on the original extension to determine the encoder

## License

See `LICENSE`.
