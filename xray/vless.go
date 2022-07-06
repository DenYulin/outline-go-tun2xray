package xray

type Rules struct {
	IP          []string `json:"ip,omitempty"`
	OutboundTag string   `json:"outboundTag"`
	Type        string   `json:"type"`
	Domain      []string `json:"domain,omitempty"`
}

type Users struct {
	ID         string `json:"id"`
	Encryption string `json:"encryption"`
	Flow       string `json:"flow"`
	Level      int    `json:"level"`
}

type Vnext struct {
	Address string  `json:"address"`
	Port    int     `json:"port"`
	Users   []Users `json:"users"`
}

type OutboundsSettings struct {
	Vnext          []Vnext `json:"vnext,omitempty"`
	DomainStrategy string  `json:"domainStrategy,omitempty"`
}

type QUICSettingsHeader struct {
	Type string `json:"type"`
}

type KCPSettingsHeader struct {
	Type string `json:"type"`
}

type TCPSettingsHeader struct {
	Type               string             `json:"type"`
	TCPSettingsRequest TCPSettingsRequest `json:"request"`
}

type TCPSettingsRequest struct {
	Version string      `json:"version"`
	Method  string      `json:"method"`
	Path    []string    `json:"path"`
	Headers HTTPHeaders `json:"headers"`
}

type HTTPHeaders struct {
	UserAgent      []string `json:"User-Agent"`
	AcceptEncoding []string `json:"Accept-Encoding"`
	Connection     string   `json:"Connection"`
	Pragma         string   `json:"Pragma"`
	Host           []string `json:"Host"`
}
