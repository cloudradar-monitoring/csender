package main

import (
	"github.com/cloudradar-monitoring/csender"
	log "github.com/sirupsen/logrus"

	"flag"
	"fmt"
	"runtime"
	"strconv"
	"strings"
	"time"
)

var (
	// set on build:
	// go build -o csender -ldflags="-X main.VERSION=$(git --git-dir=src/github.com/cloudradar-monitoring/csender/.git describe --always --long --dirty --tag)" github.com/cloudradar-monitoring/csender/cmd/csender
	VERSION string
)

type flagslice []string

func (i *flagslice) String() string {
	return fmt.Sprintf("%d", *i)
}

func (i *flagslice) Set(value string) error {
	*i = append(*i, value)
	return nil
}

type intFlag struct {
	set   bool
	value int
}

func (sf *intFlag) Set(value string) error {
	i, err := strconv.Atoi(value)
	if err != nil {
		return err
	}
	sf.value = i
	sf.set = true
	return nil
}

func (sf *intFlag) String() string {
	return strconv.Itoa(sf.value)
}

func main() {
	cs := csender.New()
	cs.SetVersion(VERSION)

	var exitCodeFlag intFlag
	var successFlag intFlag
	var keyvalsFlag flagslice
	var messageFlag string
	cfgPathPtr := flag.String("c", csender.DefaultCfgPath, "config file path")

	keyPtr := flag.String("k", "", "measurements key (*required)")
	flag.Var(&exitCodeFlag, "e", "pass unix exit code")
	flag.Var(&successFlag, "s", "set success [0,1]")
	flag.Var(&keyvalsFlag, "o", "append key=value (up to 10)")
	flag.StringVar(&messageFlag, "m", "", "message (max 300 symbols)")

	logLevelPtr := flag.String("v", "", "log level – overrides the level in config file (values \"error\",\"info\",\"debug\")")
	versionPtr := flag.Bool("version", false, "show the csender version")

	flag.Parse()

	if *versionPtr {
		fmt.Printf("csender v%s released under MIT license. https://github.com/cloudradar-monitoring/csender/\n", VERSION)
		return
	}
	tfmt := log.TextFormatter{FullTimestamp: true, DisableLevelTruncation: true, DisableTimestamp: true}
	if runtime.GOOS == "windows" {
		tfmt.DisableColors = true
	}

	log.SetFormatter(&tfmt)

	if cfgPathPtr != nil {
		err := cs.ReadConfigFromFile(*cfgPathPtr, true)
		if err != nil {
			if strings.Contains(err.Error(), "cannot load TOML value of type int64 into a Go float") {
				log.Fatalf("Config load error: please use numbers with a decimal point for numerical values")
			} else {
				log.Fatalf("Config load error: %s", err.Error())
			}
		}
	}
	if *logLevelPtr == string(csender.LogLevelError) || *logLevelPtr == string(csender.LogLevelInfo) || *logLevelPtr == string(csender.LogLevelDebug) {
		cs.SetLogLevel(csender.LogLevel(*logLevelPtr))
	}

	if *keyPtr == "" {
		log.Fatalf("-k is required")
	}

	result := csender.Result{Timestamp: time.Now().Unix()}

	err := result.SetKey(*keyPtr)
	if err != nil {
		log.Fatal(err.Error())
	}

	for _, o := range keyvalsFlag {
		err = result.AddKeyvalue(o)
		if err != nil {
			log.Fatal(err.Error())
		}
	}

	if exitCodeFlag.set {
		err = result.SetExitCode(exitCodeFlag.value)
		if err != nil {
			log.Fatal(err.Error())
		}
	}

	if successFlag.set {
		err = result.SetSuccess(successFlag.value)
		if err != nil {
			log.Fatal(err.Error())
		}
	}

	if messageFlag != "" {
		err = result.SetMessage(messageFlag)
		if err != nil {
			log.Fatal(err.Error())
		}
	}

	err = cs.PostResultsToHub(result)

	if err != nil {
		log.Fatalf(err.Error())
	}
}
