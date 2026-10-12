# PIHU WhatsApp MCP

PIHU WhatsApp MCP connects a WhatsApp account to PIHU through a local Go bridge and a Python [Model Context Protocol (MCP)](https://modelcontextprotocol.io/) server. It supports live QR pairing, connection and history-sync status, local message storage, and WhatsApp tools for searching, reading, and sending messages and media.

> This project connects to a personal WhatsApp account through WhatsApp Web multi-device using [`whatsmeow`](https://github.com/tulir/whatsmeow). Use it only with an account and data you are authorized to access. Messages and media can contain sensitive information; review any AI or MCP client access before enabling tools.

## What’s included

- **WhatsApp bridge** (`whatsapp-bridge/`): Go client that pairs with WhatsApp, exposes a loopback HTTP API for PIHU, and stores the WhatsApp session and message history locally.
- **MCP server** (`whatsapp-mcp-server/`): Python FastMCP server that exposes WhatsApp actions to PIHU and other compatible MCP clients.
- **Local PIHU auth and status page**: The bridge serves the pairing UI at `http://127.0.0.1:8080/mcp/whatsapp/auth`. It reads the QR and pairing/history-sync state directly from the local bridge API; no hosted relay or network service is required.

## Pair WhatsApp in PIHU OS

1. Open **PIHU OS → Settings → Connections → PIHU WhatsApp MCP**.
2. Start or refresh the WhatsApp connection. PIHU starts the local bridge and opens its local pairing page.
3. On your phone, open WhatsApp and choose **Linked devices → Link a device**, then scan the displayed QR code.
4. Keep PIHU open while WhatsApp connects and history synchronization runs. The page and PIHU settings show the live connection and sync state.

The pairing page, QR code, WhatsApp session, and message history remain on the computer. A successful pairing is persisted by WhatsApp’s local device session.

## Requirements

- **PIHU OS** for the integrated pairing experience and local bridge lifecycle.
- **Go 1.24.1 or newer** to build the bridge from source.
- **Python 3.11 or newer** and [uv](https://docs.astral.sh/uv/) to run the MCP server from source.
- **FFmpeg (optional)** for converting audio to WhatsApp-compatible Opus voice messages.
- **A WhatsApp account** with an available linked-device slot.

## Run from source

### 1. Start the Go bridge

```bash
cd whatsapp-bridge
go run .
```

The bridge listens on `127.0.0.1:8080`; it does not bind the pairing and message API to the network. It serves the local auth UI at `/mcp/whatsapp/auth` and provides the QR and connection state through its local API.

### 2. Start the Python MCP server

In a second terminal:

```bash
cd whatsapp-mcp-server
uv sync
uv run main.py
```

Configure your MCP client to launch `uv --directory /absolute/path/to/pihu-whatsapp-mcp/whatsapp-mcp-server run main.py`. In PIHU OS, the app manages the MCP server and Go bridge for you.

## Available MCP tools

| Tool | Description |
|---|---|
| `search_contacts` | Find contacts by name or phone number. |
| `list_chats` | Search and paginate chats, optionally including each chat’s last message. |
| `get_chat` | Read metadata for a chat JID. |
| `get_direct_chat_by_contact` | Find a direct chat using a phone number. |
| `get_contact_chats` | List chats for a contact JID. |
| `list_messages` | Search and paginate messages by date, sender, chat, or text, with optional surrounding context. |
| `get_message_context` | Retrieve messages around a selected message. |
| `get_last_interaction` | Get the latest message involving a contact. |
| `send_message` | Send text to a phone number or chat JID. |
| `send_file` | Send image, video, audio, or document files. |
| `send_audio_message` | Send audio as a WhatsApp voice message; FFmpeg can convert supported input formats. |
| `download_media` | Download media referenced by a message to a local file. |

## Local data and security

- WhatsApp session credentials and message history are stored by the bridge in its local `whatsapp-bridge/store/` directory. Keep this directory private and do not commit it.
- Pairing UI, QR code, session credentials, and message database stay on the local computer. The auth page binds only to the loopback bridge.
- The bridge HTTP API is bound to loopback and accepts CORS requests only from PIHU’s app origins and local development origins.
- Messages read through MCP may be shared with the connected AI client according to that client’s configuration. Only enable the tools and scopes you need, and treat message content as untrusted input.

## Troubleshooting

- **The local auth page cannot load:** confirm the bridge is running and open `http://127.0.0.1:8080/mcp/whatsapp/auth` from the same computer.
- **The local QR is missing:** confirm the bridge is running and that `http://127.0.0.1:8080/api/status` is reachable from PIHU.
- **WhatsApp is already linked:** the bridge reconnects using its local session and does not need a new QR. Use PIHU’s unlink/reset controls if you intend to pair again.
- **History is still syncing:** leave PIHU running; large histories may take time. The connection page reports the latest progress received from WhatsApp.
- **SQLite build errors on Windows:** `go-sqlite3` requires CGO and a C compiler. Install a compiler such as MSYS2, then set `CGO_ENABLED=1` before building.
- **Audio conversion fails:** install FFmpeg or provide an `.ogg` Opus file.

## Development checks

```bash
# Go bridge
cd whatsapp-bridge
gofmt -w *.go
go test ./...

# Python MCP server
cd ../whatsapp-mcp-server
uv run main.py
```

## License and upstream

This PIHU fork retains the upstream project’s [MIT license](LICENSE) and WhatsApp bridge dependencies. See the license for terms and attribution.
