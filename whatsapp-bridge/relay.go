package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"time"

	"rsc.io/qr"
)

type relaySession struct {
	ID          string `json:"id"`
	WriterToken string `json:"writer_token"`
	ViewerToken string `json:"viewer_token"`
}

type pairingRelay struct {
	mu      sync.Mutex
	apiURL  string
	pageURL string
	session relaySession
	client  *http.Client
	opened  bool
}

var whatsappRelay = &pairingRelay{
	apiURL:  strings.TrimRight(envOr("PIHU_WHATSAPP_RELAY_URL", "https://pihu.nfks.co.in/api/v2/whatsapp/sessions"), "/"),
	pageURL: strings.TrimRight(envOr("PIHU_WHATSAPP_AUTH_PAGE", "https://pihu.nfks.co.in/mcp/whatsapp/auth"), "/"),
	client:  &http.Client{Timeout: 5 * time.Second},
}

func envOr(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

func (relay *pairingRelay) begin() error {
	relay.mu.Lock()
	defer relay.mu.Unlock()
	if relay.session.ID != "" {
		return nil
	}
	request, err := http.NewRequest(http.MethodPost, relay.apiURL, bytes.NewReader([]byte(`{}`)))
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "application/json")
	response, err := relay.client.Do(request)
	if err != nil {
		return fmt.Errorf("create pairing relay: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusCreated {
		return fmt.Errorf("create pairing relay returned HTTP %d", response.StatusCode)
	}
	if err := json.NewDecoder(response.Body).Decode(&relay.session); err != nil {
		return fmt.Errorf("read pairing relay response: %w", err)
	}
	if relay.session.ID == "" || relay.session.WriterToken == "" || relay.session.ViewerToken == "" {
		return fmt.Errorf("pairing relay returned an incomplete session")
	}
	return nil
}

func (relay *pairingRelay) openPage() error {
	relay.mu.Lock()
	if relay.session.ID == "" {
		relay.mu.Unlock()
		return fmt.Errorf("WhatsApp pairing page is not ready")
	}
	if relay.opened {
		relay.mu.Unlock()
		return nil
	}
	relay.opened = true
	relay.mu.Unlock()

	pageURL, err := relay.authURL()
	if err != nil {
		relay.mu.Lock()
		relay.opened = false
		relay.mu.Unlock()
		return err
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
		return fmt.Errorf("could not open WhatsApp auth page: %w", err)
	}
	return nil
}

func (relay *pairingRelay) reopenPage() error {
	relay.mu.Lock()
	relay.opened = false
	relay.mu.Unlock()
	return relay.openPage()
}

func (relay *pairingRelay) authURL() (string, error) {
	relay.mu.Lock()
	session := relay.session
	relay.mu.Unlock()
	if session.ID == "" || session.ViewerToken == "" {
		return "", fmt.Errorf("WhatsApp pairing session is not ready")
	}
	page, err := url.Parse(relay.pageURL)
	if err != nil || page.Scheme != "https" || page.Host != "pihu.nfks.co.in" {
		return "", fmt.Errorf("WhatsApp hosted auth page is not a trusted PIHU HTTPS URL")
	}
	page.Path = strings.TrimRight(page.Path, "/")
	query := page.Query()
	query.Set("session", session.ID)
	page.RawQuery = query.Encode()
	page.Fragment = "access=" + url.QueryEscape(session.ViewerToken)

	return page.String(), nil
}

func (relay *pairingRelay) reset() {
	relay.mu.Lock()
	relay.session = relaySession{}
	relay.opened = false
	relay.mu.Unlock()
}

func (relay *pairingRelay) publish(stage, qrPayload string, status map[string]any) error {
	if err := relay.begin(); err != nil {
		return err
	}
	payload := map[string]any{"stage": stage, "status": status}
	if stage == "pairing" && qrPayload != "" {
		code, err := qr.Encode(qrPayload, qr.M)
		if err != nil {
			return fmt.Errorf("encode WhatsApp pairing QR: %w", err)
		}
		payload["qr_svg"] = base64.StdEncoding.EncodeToString([]byte(qrSVG(code)))
		if err := relay.openPage(); err != nil {
			fmt.Printf("Could not open WhatsApp auth page: %v\n", err)
		}
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	relay.mu.Lock()
	session := relay.session
	relay.mu.Unlock()
	request, err := http.NewRequest(http.MethodPut, relay.apiURL+"/"+url.PathEscape(session.ID), bytes.NewReader(body))
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", "Bearer "+session.WriterToken)
	response, err := relay.client.Do(request)
	if err != nil {
		return fmt.Errorf("update WhatsApp pairing relay: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("update WhatsApp pairing relay returned HTTP %d", response.StatusCode)
	}
	return nil
}
