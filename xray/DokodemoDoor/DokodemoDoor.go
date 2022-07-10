package DokodemoDoor

type InboundsSettings struct {
	Address        string `json:"address"`
	Port           uint32 `json:"port"`
	Network        string `json:"network"`
	Timeout        int    `json:"timeout"`
	FollowRedirect bool   `json:"followRedirect"`
	UserLevel      int    `json:"userLevel"`
}

type InboundSniffingConfig struct {
}
