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
