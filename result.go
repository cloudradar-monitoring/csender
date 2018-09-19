package csender

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
)

const MaxMessageLength = 300
const MaxKeyLength = 100
const MaxValueLength = 20

type MeasurementsMap map[string]interface{}

type Result struct {
	Timestamp    int64           `json:"timestamp"`
	Measurements MeasurementsMap `json:"measurements"`
	Message      interface{}     `json:"message"`
	key          string
}

var keyRE = regexp.MustCompile(`^[a-z.]+$`)

func validateKey(key string) error {
	if !keyRE.MatchString(key) {
		return errors.New("can contains only \"a-z\" and \".\"")
	}

	if len(key) > MaxKeyLength {
		return fmt.Errorf("length is longer than maximum %d", MaxKeyLength)
	}

	if key[0:1] == "." || key[len(key)-1:] == "." {
		return errors.New("starts or ends with a dot")
	}
	if strings.Index(key, "..") >= 0 {
		return errors.New("has more than 1 dot in a row")
	}

	return nil
}

func (res *Result) SetKey(key string) error {
	if err := validateKey(key); err != nil {
		return fmt.Errorf("invalid key: %s, got \"%s\"", err.Error(), key)
	}

	res.key = key
	res.Measurements = make(MeasurementsMap)
	return nil
}

func (res *Result) AddKeyvalue(s string) error {
	parts := strings.Split(s, "=")
	if len(parts) < 2 {
		return fmt.Errorf("failed to parse key=value: %s", s)
	}
	key := strings.TrimSpace(parts[0])

	// will check the concat'ed key to validate the maximum key size
	if err := validateKey(res.key + key); err != nil {
		return fmt.Errorf("invalid key: %s, got \"%s\"", err.Error(), key)
	}

	value := strings.TrimSpace(parts[1])

	if len(value) > MaxValueLength {
		return fmt.Errorf("invalid value: length is longer than maximum %d", MaxValueLength)
	}

	valueParsed, err := strconv.ParseFloat(value, 64)

	if err == nil {
		res.Measurements[res.key+"."+key] = valueParsed
	} else {
		res.Measurements[res.key+"."+key] = value
	}

	return nil
}

func (res *Result) SetSuccess(s int) error {
	// will check the concat'ed key to validate the maximum key size
	key := res.key + ".success"
	if len(key) > MaxKeyLength {
		return fmt.Errorf("invalid key: longer than maximum %d, got %s", MaxKeyLength, res.key+".success")
	}

	if s != 0 && s != 1 {
		return fmt.Errorf("success status can be either 0 or 1, got: %d", s)
	}

	res.Measurements[key] = s
	return nil
}

func (res *Result) SetMessage(s string) error {
	if len(s) > MaxMessageLength {
		s = s[0:MaxMessageLength]
		log.Errorf("message was truncated to %d symbols", MaxMessageLength)
	}
	res.Message = s
	return nil
}

func (res *Result) SetExitCode(s int) error {
	if s == 0 {
		return res.SetSuccess(1)
	}
	return res.SetSuccess(0)
}
