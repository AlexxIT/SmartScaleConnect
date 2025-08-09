package main

import (
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/AlexxIT/SmartScaleConnect/internal"
	"github.com/AlexxIT/SmartScaleConnect/pkg/core"
	"gopkg.in/yaml.v3"
)

const Version = "0.2.0"

func main() {
	log.Printf("SmartScaleConnect version %s %s/%s\n", Version, runtime.GOOS, runtime.GOARCH)

	f, err := openConfig()
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	type Config struct {
		From string `yaml:"from"`
		To   string `yaml:"to"`

		Expr map[string]string `yaml:"expr"`
	}

	var conf map[string]Config
	if err = yaml.NewDecoder(f).Decode(&conf); err != nil {
		log.Fatal(err)
	}

	for name, v := range conf {
		if !strings.HasPrefix(name, "sync") {
			continue
		}

		var weights []*core.Weight
		if weights, err = internal.GetWeights(v.From); err != nil {
			log.Printf("%s: GET ERROR: %v\n", name, err)
			continue
		}

		if v.Expr != nil {
			if err = internal.Expr(v.Expr, weights); err != nil {
				log.Printf("%s: EXPR ERROR: %v\n", name, err)
				continue
			}
		}

		if err = internal.SetWeights(v.To, weights); err != nil {
			log.Printf("%s: SET ERROR: %v\n", name, err)
			continue
		}

		log.Printf("%s: OK\n", name)
	}
}

const configName = "scaleconnect.yaml"

func openConfig() (*os.File, error) {
	// check config file in CWD
	if f, err := os.Open(configName); err == nil {
		return f, nil
	}

	ex, err := os.Executable()
	if err != nil {
		return nil, err
	}

	// set CWD to app dir
	if err = os.Chdir(filepath.Dir(ex)); err != nil {
		return nil, err
	}

	// check config file again
	return os.Open(configName)
}
