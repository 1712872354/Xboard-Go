package grpc

import "encoding/json"

// =============================================================================
// Proto message types for node ↔ panel communication.
// Serialized over gRPC using a JSON codec (see codec.go).
// =============================================================================

// --- Interfaces ---

// NodePayload marks message types sent from node to panel.
type NodePayload interface{ isNodePayload() }

// PanelPayload marks message types sent from panel to node.
type PanelPayload interface{ isPanelPayload() }

// --- Handshake ---

// HandshakeRequest is sent by the node when it first connects.
type HandshakeRequest struct {
	Token     string `json:"token"`
	NodeID    uint32 `json:"node_id"`
	MachineID uint32 `json:"machine_id"`
	Version   string `json:"version"`
	Os        string `json:"os"`
	Hostname  string `json:"hostname"`
}

// HandshakeResponse is returned by the panel after successful authentication.
type HandshakeResponse struct {
	Success       bool        `json:"success"`
	Message       string      `json:"message,omitempty"`
	PushInterval  int32       `json:"push_interval"`
	PullInterval  int32       `json:"pull_interval"`
	TrackInterval int32       `json:"track_interval"`
	Config        *NodeConfig `json:"config,omitempty"`
	Users         []*User     `json:"users,omitempty"`
}

// --- Config ---

// ConfigRequest is a lightweight request for GetConfig.
type ConfigRequest struct {
	NodeID uint32 `json:"node_id"`
}

// NodeConfig holds the proxy configuration for a single node.
type NodeConfig struct {
	ID         uint32            `json:"id"`
	Name       string            `json:"name"`
	Protocol   string            `json:"protocol"`
	Address    string            `json:"address"`
	Port       int32             `json:"port"`
	ServerInfo string            `json:"server_info"`
	Rate       float32           `json:"rate"`
	GroupID    uint32            `json:"group_id"`
	ParentID   uint32            `json:"parent_id"`
	Status     int32             `json:"status"`
	Tags       map[string]string `json:"tags,omitempty"`
	Routes     []*Route          `json:"routes,omitempty"`
}

// Route represents a routing rule.
type Route struct {
	Match  []string `json:"match,omitempty"`
	Action string   `json:"action"`
	Target string   `json:"target,omitempty"`
}

// --- Users ---

// UserListRequest queries users for a node.
type UserListRequest struct {
	NodeID uint32 `json:"node_id"`
}

// UserList is the full user list response.
type UserList struct {
	Users []*User `json:"users"`
}

// User represents a single proxy user for a node.
type User struct {
	ID           uint32 `json:"id"`
	UUID         string `json:"uuid"`
	Email        string `json:"email"`
	Passwd       string `json:"passwd,omitempty"`
	SpeedLimit   int32  `json:"speed_limit"`
	DeviceLimit  int32  `json:"device_limit"`
	TrafficLimit int64  `json:"traffic_limit"`
	UsedTraffic  int64  `json:"used_traffic"`
	ExpiredAt    int64  `json:"expired_at"`
	Status       int32  `json:"status"`
}

// --- Node → Panel report types ---

// TrafficReport carries per-user traffic deltas.
type TrafficReport struct {
	Deltas []*TrafficDelta `json:"deltas"`
}

// TrafficDelta is a single user's upload/download increment.
type TrafficDelta struct {
	UserID   uint32 `json:"user_id"`
	Upload   int64  `json:"upload"`
	Download int64  `json:"download"`
}

// AliveReport carries the list of currently active user IPs.
type AliveReport struct {
	Alive map[uint32]*AliveIPs `json:"alive"`
}

// AliveIPs holds the IPs for a single user.
type AliveIPs struct {
	IPs []string `json:"ips"`
}

// StatusReport carries node system-level status (CPU, memory, etc.).
type StatusReport struct {
	CPU         float32 `json:"cpu"`
	MemTotal    uint64  `json:"mem_total"`
	MemUsed     uint64  `json:"mem_used"`
	DiskTotal   uint64  `json:"disk_total"`
	DiskUsed    uint64  `json:"disk_used"`
	Uptime      uint64  `json:"uptime"`
	Goroutines  int32   `json:"goroutines"`
	ActiveConns int32   `json:"active_conns"`
	ActiveUsers int32   `json:"active_users"`
}

// DeviceReport carries connected device information.
type DeviceReport struct {
	Devices map[uint32]*DeviceIPs `json:"devices"`
}

// DeviceIPs holds the IPs for a single user's devices.
type DeviceIPs struct {
	IPs []string `json:"ips"`
}

// --- Panel → Node update types ---

// ConfigUpdate pushes a new configuration to the node.
type ConfigUpdate struct {
	Config *NodeConfig `json:"config"`
}

// UserListUpdate pushes a complete user list to the node.
type UserListUpdate struct {
	Users []*User `json:"users"`
}

// UserDelta pushes incremental user additions/removals to the node.
type UserDelta struct {
	Added   []*User  `json:"added,omitempty"`
	Removed []uint32 `json:"removed,omitempty"`
}

// DeviceState represents a mapping of user devices (panel→node).
type DeviceState struct {
	Devices map[uint32]*DeviceIPs `json:"devices"`
}

// --- Heartbeat ---

// Ping is sent by the panel to indicate liveness check.
type Ping struct {
	Timestamp int64 `json:"timestamp"`
}

// Pong is the node's acknowledgement of a Ping.
type Pong struct {
	Timestamp int64 `json:"timestamp"`
}

// =============================================================================
// Payload interface implementations
// =============================================================================

// --- Node Payload implementations ---
func (*TrafficReport) isNodePayload() {}
func (*AliveReport) isNodePayload()   {}
func (*StatusReport) isNodePayload()  {}
func (*DeviceReport) isNodePayload()  {}
func (*Pong) isNodePayload()          {}

// --- Panel Payload implementations ---
func (*ConfigUpdate) isPanelPayload()   {}
func (*UserListUpdate) isPanelPayload() {}
func (*UserDelta) isPanelPayload()      {}
func (*DeviceState) isPanelPayload()    {}
func (*Ping) isPanelPayload()           {}
func (*Pong) isPanelPayload()           {}

// =============================================================================
// Stream envelope messages with Payload interface
// =============================================================================

// NodeMessage wraps a payload from node to panel.
type NodeMessage struct {
	Payload NodePayload `json:"payload"`
}

// PanelMessage wraps a payload from panel to node.
type PanelMessage struct {
	Payload PanelPayload `json:"payload"`
}

// =============================================================================
// Custom JSON serialization for NodeMessage
// =============================================================================

func (m NodeMessage) MarshalJSON() ([]byte, error) {
	switch p := m.Payload.(type) {
	case *TrafficReport:
		return json.Marshal(map[string]any{"payload": map[string]any{"traffic_report": p}})
	case *AliveReport:
		return json.Marshal(map[string]any{"payload": map[string]any{"alive_report": p}})
	case *StatusReport:
		return json.Marshal(map[string]any{"payload": map[string]any{"status_report": p}})
	case *DeviceReport:
		return json.Marshal(map[string]any{"payload": map[string]any{"device_report": p}})
	case *Pong:
		return json.Marshal(map[string]any{"payload": map[string]any{"pong": p}})
	default:
		return json.Marshal(map[string]any{"payload": nil})
	}
}

func (m *NodeMessage) UnmarshalJSON(data []byte) error {
	var raw struct {
		Payload json.RawMessage `json:"payload"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if len(raw.Payload) == 0 || string(raw.Payload) == "null" {
		return nil
	}
	var probe struct {
		TrafficReport *json.RawMessage `json:"traffic_report"`
		AliveReport   *json.RawMessage `json:"alive_report"`
		StatusReport  *json.RawMessage `json:"status_report"`
		DeviceReport  *json.RawMessage `json:"device_report"`
		Pong          *json.RawMessage `json:"pong"`
	}
	if err := json.Unmarshal(raw.Payload, &probe); err != nil {
		return err
	}
	switch {
	case probe.TrafficReport != nil:
		m.Payload = new(TrafficReport)
		return json.Unmarshal(*probe.TrafficReport, m.Payload)
	case probe.AliveReport != nil:
		m.Payload = new(AliveReport)
		return json.Unmarshal(*probe.AliveReport, m.Payload)
	case probe.StatusReport != nil:
		m.Payload = new(StatusReport)
		return json.Unmarshal(*probe.StatusReport, m.Payload)
	case probe.DeviceReport != nil:
		m.Payload = new(DeviceReport)
		return json.Unmarshal(*probe.DeviceReport, m.Payload)
	case probe.Pong != nil:
		m.Payload = new(Pong)
		return json.Unmarshal(*probe.Pong, m.Payload)
	}
	return nil
}

// =============================================================================
// Custom JSON serialization for PanelMessage
// =============================================================================

func (m PanelMessage) MarshalJSON() ([]byte, error) {
	switch p := m.Payload.(type) {
	case *ConfigUpdate:
		return json.Marshal(map[string]any{"payload": map[string]any{"config_update": p}})
	case *UserListUpdate:
		return json.Marshal(map[string]any{"payload": map[string]any{"user_list_update": p}})
	case *UserDelta:
		return json.Marshal(map[string]any{"payload": map[string]any{"user_delta": p}})
	case *DeviceState:
		return json.Marshal(map[string]any{"payload": map[string]any{"device_state": p}})
	case *Ping:
		return json.Marshal(map[string]any{"payload": map[string]any{"ping": p}})
	case *Pong:
		return json.Marshal(map[string]any{"payload": map[string]any{"pong": p}})
	default:
		return json.Marshal(map[string]any{"payload": nil})
	}
}

func (m *PanelMessage) UnmarshalJSON(data []byte) error {
	var raw struct {
		Payload json.RawMessage `json:"payload"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if len(raw.Payload) == 0 || string(raw.Payload) == "null" {
		return nil
	}
	var probe struct {
		ConfigUpdate   *json.RawMessage `json:"config_update"`
		UserListUpdate *json.RawMessage `json:"user_list_update"`
		UserDelta      *json.RawMessage `json:"user_delta"`
		DeviceState    *json.RawMessage `json:"device_state"`
		Ping           *json.RawMessage `json:"ping"`
		Pong           *json.RawMessage `json:"pong"`
	}
	if err := json.Unmarshal(raw.Payload, &probe); err != nil {
		return err
	}
	switch {
	case probe.ConfigUpdate != nil:
		m.Payload = new(ConfigUpdate)
		return json.Unmarshal(*probe.ConfigUpdate, m.Payload)
	case probe.UserListUpdate != nil:
		m.Payload = new(UserListUpdate)
		return json.Unmarshal(*probe.UserListUpdate, m.Payload)
	case probe.UserDelta != nil:
		m.Payload = new(UserDelta)
		return json.Unmarshal(*probe.UserDelta, m.Payload)
	case probe.DeviceState != nil:
		m.Payload = new(DeviceState)
		return json.Unmarshal(*probe.DeviceState, m.Payload)
	case probe.Ping != nil:
		m.Payload = new(Ping)
		return json.Unmarshal(*probe.Ping, m.Payload)
	case probe.Pong != nil:
		m.Payload = new(Pong)
		return json.Unmarshal(*probe.Pong, m.Payload)
	}
	return nil
}
