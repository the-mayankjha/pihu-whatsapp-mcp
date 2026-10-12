package main

import (
	"fmt"
	"net/url"
	"os/exec"
	"runtime"
	"sync"
)

// Pairing is served by the local bridge. Keep the small relay-shaped API here
// so all pairing/status transitions remain local and independent of PIHU web.
type pairingRelay struct {
	mu      sync.Mutex
	pageURL string
	opened  bool
}

var whatsappRelay = &pairingRelay{
	pageURL: "http://127.0.0.1:8080/mcp/whatsapp/auth",
}

func (relay *pairingRelay) begin() error { return nil }

func (relay *pairingRelay) openPage() error {
	relay.mu.Lock()
	pageURL := relay.pageURL
	relay.opened = true
	relay.mu.Unlock()

	parsed, err := url.Parse(pageURL)
	if err != nil || parsed.Scheme != "http" || parsed.Host != "127.0.0.1:8080" || parsed.Path != "/mcp/whatsapp/auth" {
		return fmt.Errorf("WhatsApp auth page must use the local PIHU bridge URL")
	}
	var command *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		command = exec.Command("open", pageURL)
	case "windows":
		command = exec.Command("rundll32", "url.dll,FileProtocolHandler", pageURL)
	default:
		command = exec.Command("xdg-open", pageURL)
	}
	if err := command.Start(); err != nil {
		relay.mu.Lock()
		relay.opened = false
		relay.mu.Unlock()
		return fmt.Errorf("could not open local WhatsApp auth page: %w", err)
	}
	return nil
}

func (relay *pairingRelay) reopenPage() error { return relay.openPage() }

func (relay *pairingRelay) reset() {
	relay.mu.Lock()
	relay.opened = false
	relay.mu.Unlock()
}

// Pairing state and QR images are now read directly from /api/status locally.
func (relay *pairingRelay) publish(_, _ string, _ map[string]any) error { return nil }
