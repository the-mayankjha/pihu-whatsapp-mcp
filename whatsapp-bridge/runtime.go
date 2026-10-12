package main

import (
	"encoding/base64"
	"fmt"
	"strings"
	"sync"

	"rsc.io/qr"
)

type whatsappRuntime struct {
	sync.RWMutex
	qrCode         string
	qrImage        string
	stage          string
	progress       int
	chatsSynced    int
	connected      bool
	loggedIn       bool
	pairingStarted bool
	jid            string
}

var runtimeStatus = &whatsappRuntime{stage: "waiting"}

func (state *whatsappRuntime) setPairingQR(payload string) error {
	code, err := qr.Encode(payload, qr.M)
	if err != nil {
		return err
	}
	image := qrSVG(code)
	state.Lock()
	state.qrCode = payload
	state.qrImage = "data:image/svg+xml;base64," + base64.StdEncoding.EncodeToString([]byte(image))
	state.stage = "pairing"
	state.progress = 0
	state.Unlock()
	return nil
}

func qrSVG(code *qr.Code) string {
	var path strings.Builder
	for y := 0; y < code.Size; y++ {
		for x := 0; x < code.Size; x++ {
			if code.Black(x, y) {
				fmt.Fprintf(&path, "M%d %dh1v1h-1z", x, y)
			}
		}
	}
	return fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="-4 -4 %d %d"><path fill="#fff" d="M-4-4h%dv%dh-%dz"/><path fill="#111" d="%s"/></svg>`, code.Size+8, code.Size+8, code.Size+8, code.Size+8, code.Size+8, path.String())
}

func (state *whatsappRuntime) clearQR() {
	state.Lock()
	state.qrCode = ""
	state.qrImage = ""
	state.Unlock()
}

func relayStatus() map[string]any {
	runtimeStatus.RLock()
	defer runtimeStatus.RUnlock()
	return map[string]any{
		"connected":    runtimeStatus.connected,
		"logged_in":    runtimeStatus.loggedIn,
		"progress":     runtimeStatus.progress,
		"chats_synced": runtimeStatus.chatsSynced,
	}
}
