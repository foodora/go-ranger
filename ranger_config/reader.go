package ranger_config

import (
	"io/ioutil"
	"net/url"
	"strings"

	"github.com/foodora/go-ranger/ranger_http"
	"gopkg.in/yaml.v2"
)

const (
	defaultTimeout = 5
)

type Config struct {
	AppName           string `yaml:"app_name"`
	APIRequestTimeout int    `yaml:"api_request_timeout"`
	HTTPAddress       string `yaml:"http_address"`
	Version           string `yaml:"version"`
	LogstashAddress   string `yaml:"logstash_address"`
	LogstashProtocol  string `yaml:"logstash_protocol"`
}

// Reader is the interface for config readers
type Reader interface {
	ReadConfig() (interface{}, error)
	GetConfigPath() string
}

// apiConfigReader is the config reader implementation using an api as source
type remoteConfigReader struct {
	url    string
	client ranger_http.APIClientInterface
}

// localConfigReader reader
type localConfigReader struct {
	configPath string
	readFile   func(filename string) ([]byte, error)
}

// newAPIConfigReader is the factory for config readers.
func newRemoteConfigReader(apiClient ranger_http.APIClientInterface, configPath string) Reader {
	return &remoteConfigReader{
		url:    configPath,
		client: apiClient,
	}
}

// ReadConfig fetches the config for the app remotely
func (configReader *remoteConfigReader) ReadConfig() (interface{}, error) {
	// @todo define data structure and implement remoteConfigReader
	return nil, nil
}

// GetConfigPath ...
func (configReader *remoteConfigReader) GetConfigPath() string {
	return configReader.url
}

// newLocalConfigReader is the factory for config readers.
func newLocalConfigReader(configPath string) *localConfigReader {
	return &localConfigReader{
		configPath: configPath,
		readFile:   ioutil.ReadFile,
	}
}

// ReadConfig fetches the config for the app locally
func (configReader *localConfigReader) ReadConfig() (interface{}, error) {
	config := &Config{}
	fileContents, err := ioutil.ReadFile(configReader.configPath)

	if err != nil {
		return config, err
	}

	return config, yaml.Unmarshal(fileContents, &config)
}

// GetConfigPath ...
func (configReader *localConfigReader) GetConfigPath() string {
	return getLocalPath(configReader.configPath)
}

// GetConfigReader strategy
func GetConfigReader(path string) Reader {
	if isReadConfigurationLocal(path) {
		return newLocalConfigReader(getLocalPath(path))
	}

	return newRemoteConfigReader(ranger_http.NewAPIClient(defaultTimeout), path)
}

func isReadConfigurationLocal(path string) bool {
	u, err := url.Parse(path)
	if err != nil {
		panic(err)
	}
	return u.Scheme == "file"
}

func getLocalPath(path string) string {
	u, err := url.Parse(path)
	if err != nil {
		panic(err)
	}
	return strings.Replace(u.String(), "file://", "", -1)
}

// returns *Config instead of interface{}
func (configReader *localConfigReader) ReadConfigAsObject() (*Config, error) {
	interfaceConfig, err := configReader.ReadConfig()
	return interfaceConfig.(*Config), err
}
