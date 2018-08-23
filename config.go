package csender

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"crypto/x509"
	"github.com/BurntSushi/toml"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/url"
	"strings"
)

var DefaultCfgPath string

type Csender struct {
	LogFile  string   `toml:"log"`
	LogLevel LogLevel `toml:"log_level"`

	HubURL           string `toml:"hub_url"`
	HubGzip          bool   `toml:"hub_gzip"` // enable gzip when sending results to the HUB
	HubUser          string `toml:"hub_user"`
	HubPassword      string `toml:"hub_password"`
	HubProxy         string `toml:"hub_proxy"`
	HubProxyUser     string `toml:"hub_proxy_user"`
	HubProxyPassword string `toml:"hub_proxy_password"`

	hubHttpClient *http.Client

	rootCAs *x509.CertPool
	version string
}

func New() *Csender {
	var defaultLogPath string
	var rootCertsPath string

	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	exPath := filepath.Dir(ex)

	switch runtime.GOOS {
	case "windows":
		DefaultCfgPath = filepath.Join(exPath, "./csender.conf")
		defaultLogPath = filepath.Join(exPath, "./csender.log")
	case "darwin":
		DefaultCfgPath = os.Getenv("HOME") + "/.csender/csender.conf"
		defaultLogPath = os.Getenv("HOME") + "/.csender/csender.log"
	default:
		rootCertsPath = "/etc/csender/cacert.pem"
		DefaultCfgPath = "/etc/csender/csender.conf"
		defaultLogPath = "/var/log/csender/csender.log"
	}

	cs := &Csender{
		LogFile: defaultLogPath,
	}

	if rootCertsPath != "" {
		if _, err := os.Stat(rootCertsPath); err == nil {
			certPool := x509.NewCertPool()

			b, err := ioutil.ReadFile(rootCertsPath)
			if err != nil {
				log.Error("Failed to read cacert.pem: ", err.Error())
			} else {
				ok := certPool.AppendCertsFromPEM(b)
				if ok {
					cs.rootCAs = certPool
				}
			}
		}
	}

	if cs.HubURL == "" && os.Getenv("CSENDER_HUB_URL") != "" {
		cs.HubURL = os.Getenv("CSENDER_HUB_URL")
	}

	cs.SetLogLevel(LogLevelInfo)
	return cs
}

func secToDuration(secs float64) time.Duration {
	return time.Duration(int64(float64(time.Second) * secs))
}

func (cs *Csender) SetVersion(version string) {
	cs.version = version
}

func (cs *Csender) userAgent() string {
	if cs.version == "" {
		cs.version = "{undefined}"
	}
	parts := strings.Split(cs.version, "-")

	return fmt.Sprintf("Csender v%s %s %s", parts[0], runtime.GOOS, runtime.GOARCH)
}

func (cs *Csender) ReadConfigFromFile(configFilePath string, createIfNotExists bool) error {
	dir := filepath.Dir(configFilePath)
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		log.WithError(err).Errorf("Failed to create the config dir: '%s'", dir)
	}

	if _, err := os.Stat(configFilePath); os.IsNotExist(err) {
		if !createIfNotExists {
			return fmt.Errorf("Config file not exists: %s", configFilePath)
		}
		f, err := os.OpenFile(configFilePath, os.O_WRONLY|os.O_CREATE, 0644)

		if err != nil {
			return fmt.Errorf("Failed to create the default config file: '%s'", configFilePath)
		}
		defer f.Close()
		enc := toml.NewEncoder(f)
		enc.Encode(cs)
	} else {
		_, err = os.Stat(configFilePath)
		if err != nil {
			return err
		}
		_, err = toml.DecodeFile(configFilePath, &cs)
		if err != nil {
			return err
		}
	}

	if cs.HubProxy != "" {
		if !strings.HasPrefix(cs.HubProxy, "http") {
			cs.HubProxy = "http://" + cs.HubProxy
		}
		_, err := url.Parse(cs.HubProxy)

		if err != nil {
			return fmt.Errorf("Failed to parse 'hub_proxy' URL")
		}
	}

	cs.SetLogLevel(cs.LogLevel)
	return addLogFileHook(cs.LogFile, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
}
