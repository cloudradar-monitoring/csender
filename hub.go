package csender

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"

	"net/url"
	"strings"

	"compress/gzip"

	"crypto/tls"
	"fmt"
	log "github.com/sirupsen/logrus"
)

func (cs *Csender) initHubHttpClient() {
	if cs.hubHttpClient == nil {
		tr := *(http.DefaultTransport.(*http.Transport))
		if cs.rootCAs != nil {
			tr.TLSClientConfig = &tls.Config{RootCAs: cs.rootCAs}
		}
		if cs.HubProxy != "" {
			if !strings.HasPrefix(cs.HubProxy, "http://") {
				cs.HubProxy = "http://" + cs.HubProxy
			}

			u, err := url.Parse(cs.HubProxy)

			if err != nil {
				log.Errorf("Failed to parse 'hub_proxy' URL")
			} else {
				if cs.HubProxyUser != "" {
					u.User = url.UserPassword(cs.HubProxyUser, cs.HubProxyPassword)
				}
				tr.Proxy = func(_ *http.Request) (*url.URL, error) {
					return u, nil
				}
			}
		} else {
			tr.Proxy = http.ProxyFromEnvironment
		}

		cs.hubHttpClient = &http.Client{
			Timeout:   time.Second * 30,
			Transport: &tr,
		}
	}
}

func (cs *Csender) PostResultsToHub(result Result) error {
	cs.initHubHttpClient()

	if cs.HubURL == "" {
		return fmt.Errorf("Both 'hub_url' in config and 'CSENDER_HUB_URL' env variable are empty")
	}

	if _, err := url.Parse(cs.HubURL); err != nil {
		return fmt.Errorf("Can't parse Hub URL: %s", err.Error())
	}

	b, err := json.Marshal(result)
	if err != nil {
		return err
	}

	var req *http.Request

	if cs.HubGzip {
		var buffer bytes.Buffer
		zw := gzip.NewWriter(&buffer)
		zw.Write(b)
		zw.Close()
		req, err = http.NewRequest("POST", cs.HubURL, &buffer)
		req.Header.Set("Content-Encoding", "gzip")
	} else {
		req, err = http.NewRequest("POST", cs.HubURL, bytes.NewBuffer(b))
	}

	if err != nil {
		return err
	}

	req.Header.Add("User-Agent", cs.userAgent())

	if cs.HubUser != "" {
		req.SetBasicAuth(cs.HubUser, cs.HubPassword)
	}

	resp, err := cs.hubHttpClient.Do(req)

	if err != nil {
		return err
	}

	log.Debugf("Sent to HUB.. Status %d", resp.StatusCode)

	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("Hub responded with a bad HTTP status: %d(%s)", resp.StatusCode, resp.Status)
	}

	return nil
}
