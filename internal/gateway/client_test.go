package gateway

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestClientAcceptsFormalZeroExchangeAndAckResponses(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/overlay/v1/join-tokens/exchange":
			w.Header().Set("Cache-Control", "no-store")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"enrollment_token": "xenr_test", "token_type": "Bearer", "expires_at": now.Add(10 * time.Minute),
				"scope":             []string{"overlay:config:read", "overlay:config:ack", "overlay:device:revoke"},
				"device_credential": map[string]any{"credential_id": "credential-1", "credential": "xdc_test", "token_type": "Device", "issued_at": now, "expires_at": now.Add(24 * time.Hour), "scope": []string{"overlay:session:mint", "overlay:credential:rotate", "overlay:device:revoke"}},
				"device":            map[string]any{"id": "gw-test", "user_id": "11111111-2222-4333-8444-555555555555", "network_id": "net-test", "role": "gateway", "name": "gw-test", "platform": "linux", "hostname": "gateway", "wireguard_public_key": "public-key", "wireguard_address": "10.77.0.1/32", "created_at": now, "updated_at": now},
				"network":           map[string]any{"id": "net-test", "display_name": "Test", "cidr": "10.77.0.0/24", "gateway_id": "gw-test", "gateway_wireguard_public_key": "public-key", "gateway_endpoint_host": "gateway.example.test", "gateway_endpoint_port": 51820, "transport_server_name": "gateway.example.test", "transport_port": 443, "transport_auth_id": "11111111-1111-1111-1111-111111111111", "created_at": now, "updated_at": now},
				"signing_keys":      []any{},
			})
		case "/api/overlay/v1/enrollment/signed-config/2/ack":
			_ = json.NewEncoder(w).Encode(map[string]any{"acked": true, "duplicate": false, "ack": map[string]any{"device_id": "gw-test", "config_id": "cfg-test", "generation": 2, "applied_at": now, "received_at": now}})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatal(err)
	}
	client.http = server.Client()
	response, err := client.Exchange(context.Background(), "join-token", "gw-test", "gateway", "public-key")
	if err != nil || response.TokenType != "Bearer" || response.DeviceCredential.TokenType != "Device" {
		t.Fatalf("formal exchange response rejected: response=%#v err=%v", response, err)
	}
	if err := client.Ack(context.Background(), "enrollment-token", Config{ConfigID: "cfg-test", GatewayID: "gw-test", Generation: 2}); err != nil {
		t.Fatalf("formal ack response rejected: %v", err)
	}
}
