package cmd

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"sync"
	"time"

	"github.com/DKE-Data/agrirouter-sdk-go"
	"github.com/creack/pty"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

//go:embed ui/index.html
var uiFS embed.FS

const (
	serveAuthorizeURL   = "https://app-qa.agrirouter.com/en/authorize"
	serveRedirectURI    = "http://localhost:8080"
	serveAuthorizeScope = "endpoints:manage"
)

// Populated by serveCmd.RunE before the HTTP server starts; read by
// handleTerminalWS when spawning the repl so the subprocess can reference
// these via $ART_APPLICATION_ID / $ART_SOFTWARE_VERSION_ID.
var (
	replApplicationID     string
	replSoftwareVersionID string
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "serve a minimal web UI for the agrirouter authorization flow and validation",
	Long: `Serves a one-page web UI on http://localhost:8080 with three sections:
  1. A button that redirects the user to agrirouter to grant 'endpoints:manage'.
  2. After the user has been redirected back, a name field + button that asks
     the server side to PUT a test endpoint using the configured client
     credentials (validating that we have access).
  3. An xterm.js terminal wired over WebSocket to a local PTY (a shell), so
     the same art CLI can be used from the browser.

Reads client credentials from AGRIROUTER_OAUTH_CLIENT_ID and
AGRIROUTER_OAUTH_CLIENT_SECRET (same env vars as other art commands).`,
	RunE: func(cmd *cobra.Command, args []string) error {
		port, err := cmd.Flags().GetInt("port")
		if err != nil {
			return fmt.Errorf("failed to get port flag: %w", err)
		}

		applicationID, err := uuidFlagOrEnv(cmd, applicationIDOpt, "ART_APPLICATION_ID")
		if err != nil {
			return err
		}

		softwareVersionID, err := uuidFlagOrEnv(cmd, softwareVersionIDOpt, "ART_SOFTWARE_VERSION_ID")
		if err != nil {
			return err
		}
		replApplicationID = applicationID.String()
		replSoftwareVersionID = softwareVersionID.String()

		clientID := viper.GetString("AGRIROUTER_OAUTH_CLIENT_ID")
		if clientID == "" {
			return fmt.Errorf("AGRIROUTER_OAUTH_CLIENT_ID env var is required")
		}

		mux := http.NewServeMux()

		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/" {
				http.NotFound(w, r)
				return
			}
			data, err := uiFS.ReadFile("ui/index.html")
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			_, _ = w.Write(data)
		})

		mux.HandleFunc("/api/config", func(w http.ResponseWriter, r *http.Request) {
			authorize := fmt.Sprintf("%s?client_id=%s&redirect_uri=%s&scope=%s",
				serveAuthorizeURL, clientID, serveRedirectURI, serveAuthorizeScope)
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]string{
				"clientId":     clientID,
				"authorizeUrl": authorize,
			})
		})

		mux.HandleFunc("/api/validate", func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				return
			}
			var body struct {
				ExternalID string `json:"externalId"`
				Name       string `json:"name"`
				TenantID   string `json:"tenantId"`
			}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				http.Error(w, "invalid json: "+err.Error(), http.StatusBadRequest)
				return
			}
			externalID := body.ExternalID
			if externalID == "" {
				http.Error(w, "externalId is required", http.StatusBadRequest)
				return
			}
			if body.TenantID == "" {
				http.Error(w, "tenantId is required (expected as ?tenant_id=... query param after the authorize redirect)", http.StatusBadRequest)
				return
			}
			tenantID, err := uuid.Parse(body.TenantID)
			if err != nil {
				http.Error(w, fmt.Sprintf("invalid tenantId '%s': %v", body.TenantID, err), http.StatusBadRequest)
				return
			}

			normalizedName, err := agrirouter.NormalizeEndpointName(body.Name)
			if err != nil {
				http.Error(w, fmt.Sprintf("invalid name: %v", err), http.StatusBadRequest)
				return
			}

			ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
			defer cancel()

			client, err := getClient(ctx)
			if err != nil {
				slog.Error("serve: failed to create agrirouter client", "err", err)
				http.Error(w, err.Error(), http.StatusBadGateway)
				return
			}

			capabilities := []agrirouter.EndpointCapability{
				{MessageType: "iso:11783:-10:taskdata:zip", Direction: agrirouter.CapabilityDirectionSendReceive},
				{MessageType: "img:png", Direction: agrirouter.CapabilityDirectionSendReceive},
			}
			subscriptions := []agrirouter.EndpointSubscription{
				{MessageType: "iso:11783:-10:taskdata:zip"},
			}

			slog.Info("serve: putting endpoint",
				"externalID", externalID,
				"name", normalizedName,
				"tenantID", tenantID,
			)

			allowDeleteByUser := true

			epResult, err := client.PutEndpoint(ctx, externalID, &agrirouter.PutEndpointParams{
				XAgrirouterTenantId: tenantID,
			}, &agrirouter.PutEndpointRequest{
				Name:              &normalizedName,
				ApplicationId:     applicationID,
				SoftwareVersionId: softwareVersionID,
				EndpointType:      agrirouter.CloudSoftware,
				Capabilities:      capabilities,
				Subscriptions:     subscriptions,
				AllowDeleteByUser: &allowDeleteByUser,
			})
			if err != nil {
				slog.Error("serve: PutEndpoint failed", "err", err)
				http.Error(w, err.Error(), http.StatusBadGateway)
				return
			}

			slog.Info("serve: PutEndpoint succeeded", "result", fmt.Sprintf("%+v", epResult))
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"ok":         true,
				"externalId": externalID,
				"name":       normalizedName,
				"tenantId":   tenantID.String(),
				"result":     epResult,
			})
		})

		mux.HandleFunc("/ws/terminal", handleTerminalWS)

		addr := fmt.Sprintf(":%d", port)
		slog.Info("serve: listening", "addr", addr, "clientId", clientID)
		slog.Info("Open http://localhost:8080 in your browser to use the UI.")
		return http.ListenAndServe(addr, mux)
	},
}

var wsUpgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func handleTerminalWS(w http.ResponseWriter, r *http.Request) {
	conn, err := wsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Error("serve: ws upgrade failed", "err", err)
		return
	}
	defer conn.Close()

	bin, err := os.Executable()
	if err != nil {
		_ = conn.WriteMessage(websocket.TextMessage, []byte("failed to resolve art executable: "+err.Error()))
		return
	}
	subproc := exec.Command(bin, "repl")
	subproc.Env = append(os.Environ(),
		"TERM=xterm-256color",
		"ART_APPLICATION_ID="+replApplicationID,
		"ART_SOFTWARE_VERSION_ID="+replSoftwareVersionID,
	)

	ptmx, err := pty.Start(subproc)
	if err != nil {
		_ = conn.WriteMessage(websocket.TextMessage, []byte("failed to start pty: "+err.Error()))
		return
	}
	defer func() {
		_ = ptmx.Close()
		_ = subproc.Process.Kill()
		_, _ = subproc.Process.Wait()
	}()

	var writeMu sync.Mutex
	writeMsg := func(mt int, data []byte) error {
		writeMu.Lock()
		defer writeMu.Unlock()
		return conn.WriteMessage(mt, data)
	}

	// pty -> ws
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := ptmx.Read(buf)
			if n > 0 {
				chunk := make([]byte, n)
				copy(chunk, buf[:n])
				if err := writeMsg(websocket.BinaryMessage, chunk); err != nil {
					return
				}
			}
			if err != nil {
				return
			}
		}
	}()

	// ws -> pty
	for {
		mt, data, err := conn.ReadMessage()
		if err != nil {
			return
		}
		if mt != websocket.TextMessage {
			continue
		}
		var msg struct {
			Type string `json:"type"`
			Data string `json:"data"`
			Cols uint16 `json:"cols"`
			Rows uint16 `json:"rows"`
		}
		if err := json.Unmarshal(data, &msg); err != nil {
			continue
		}
		switch msg.Type {
		case "stdin":
			if _, err := ptmx.Write([]byte(msg.Data)); err != nil {
				return
			}
		case "resize":
			if msg.Cols > 0 && msg.Rows > 0 {
				_ = pty.Setsize(ptmx, &pty.Winsize{Cols: msg.Cols, Rows: msg.Rows})
			}
		}
	}
}

func init() {
	rootCmd.AddCommand(serveCmd)
	serveCmd.Flags().Int("port", 8080, "Port to listen on (must match the redirect_uri registered for the client)")
	serveCmd.Flags().String(applicationIDOpt, "", "The application ID to use when PUTting the test endpoint (default: $ART_APPLICATION_ID)")
	serveCmd.Flags().String(softwareVersionIDOpt, "", "The software version ID to use when PUTting the test endpoint (default: $ART_SOFTWARE_VERSION_ID)")
}
