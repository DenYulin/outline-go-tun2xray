package xray

import (
	"encoding/json"
	"errors"
	"github.com/eycorsican/go-tun2socks/common/log"
	"github.com/xtls/xray-core/common/cmdarg"
	"github.com/xtls/xray-core/core"
	"os"
	"runtime"
	"runtime/debug"
)

const (
	JsonFormat  = "json"
	Dot         = "."
	EmptyString = ""
)

const (
	NoError                 = 0
	Unexpected              = 1
	ConfigFileNotExist      = 2
	UnsupportedConfigFormat = 3
	ClientStartError        = 4
	ClientCreateError       = 5
)

var XrayClient core.Server

func RegisterXray(xrayClient core.Server) {
	XrayClient = xrayClient
}

func StartXray(configFilePath string, checkXrayConfigFiles bool) error {
	//if isExists := dirExists(configFilePath); !isExists {
	//	log.Errorf("The config file of xray does not exist， configFilePath: %s", configFilePath)
	//	//return errors.New("the config file of xray does not exist")
	//	os.Exit(ConfigFileNotExist)
	//}

	//fileExt := filepath.Ext(configFilePath)
	//fileExt = strings.ReplaceAll(EmptyString, Dot, fileExt)
	//if fileExt != JsonFormat {
	//	log.Errorf("The config file of xray is not a json file, fileExt: %s, configFilePath: %s", fileExt, configFilePath)
	//	os.Exit(UnsupportedConfigFormat)
	//}

	xrayClient, err := CreateXrayClient(JsonFormat, configFilePath)
	if err != nil {
		log.Errorf("Failed to create xray client, configFilePath: %s, error: %s", configFilePath, err.Error())
		os.Exit(ClientCreateError)
	}

	if checkXrayConfigFiles {
		log.Infof("The xray config OK.")
		os.Exit(NoError)
	}

	if err = xrayClient.Start(); err != nil {
		log.Errorf("Failed to start xray client, error: %s", err.Error())
		os.Exit(ClientStartError)
	}
	defer xrayClient.Close()

	runtime.GC()
	debug.FreeOSMemory()
	return nil
}

func CreateXrayClient(configFileFormat, configFilePath string) (core.Server, error) {
	log.Infof("Start up xray client, configFileFormat: %s, configFilePath:% s", configFileFormat, configFilePath)

	configFiles := cmdarg.Arg{configFilePath}
	config, err := core.LoadConfig(configFileFormat, configFiles)
	if err != nil {
		log.Errorf("Load xray config error, error: %s", err.Error())
		return nil, errors.New("failed to load config files, configFiles: " + configFiles.String())
	}

	jsonBytes, _ := json.Marshal(config)
	log.Infof("Config Json: %s", string(jsonBytes))

	client, err := core.New(config)
	if err != nil {
		log.Infof("Failed to create xray client, error: %s", err.Error())
		return nil, errors.New("failed to create xray client, error: " + err.Error())
	}
	return client, nil
}

func init() {
	//serial.ReaderDecoderByFormat["json"] = serial.DecodeJSONConfig
	//serial.ReaderDecoderByFormat["yaml"] = serial.DecodeYAMLConfig
	//serial.ReaderDecoderByFormat["toml"] = serial.DecodeTOMLConfig

	//core.ConfigBuilderForFiles = serial.BuildConfig

	//confloader.EffectiveConfigFileLoader = external.ConfigLoader

	//proxyInboundManager := inbound.Manager{}
	//fmt.Println(proxyInboundManager)
	//proxyOutboundManager := outbound.Manager{}
	//fmt.Println(proxyOutboundManager)
}

func dirExists(file string) bool {
	if file == "" {
		return false
	}
	info, err := os.Stat(file)
	return err == nil && info.IsDir()
}
