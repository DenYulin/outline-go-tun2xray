package utils

import (
	"encoding/json"
	"github.com/eycorsican/go-tun2socks/common/log"
)

func ToJsonString(any interface{}) string {
	jsonBytes, err := json.Marshal(any)
	if err != nil {
		log.Errorf("Json serialization failed, error: %s", err.Error())
		return ""
	}
	return string(jsonBytes)
}
