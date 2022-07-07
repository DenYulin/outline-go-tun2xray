package features

type VLessOptions struct {
	UseIPv6       bool   `json:"useIPv6"`
	LogLevel      string `json:"logLevel"`
	RouteMode     int    `json:"routeMode"`
	DNS           string `json:"DNS"`
	AllowInsecure bool   `json:"allowInsecure"`
	Mux           int    `json:"mux"`
}

type VLess struct {
	Host     string
	Path     string
	TLS      string
	Address  string
	Port     int
	Net      string
	ID       string
	Flow     string
	Type     string // headerType
	Protocol string `json:"protocol"`
	VLessOptions
}

type Users struct {
	ID         string `json:"id"`
	Encryption int    `json:"encryption"`
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
