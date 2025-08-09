package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/AlexxIT/SmartScaleConnect/internal"
	"gopkg.in/yaml.v3"
)

const Version = "0.2.0"

const usage = `Usage of scaleconnect:

  -c, --config  Path to config file
  -s, --sync    Sync name from scaleconnect.yaml
  -f, --from    From string
  -t, --to      To string
`

func main() {
	var confName, sync, from, to string

	flag.StringVar(&sync, "config", "", "")
	flag.StringVar(&sync, "c", "", "")
	flag.StringVar(&sync, "sync", "", "")
	flag.StringVar(&sync, "s", "", "")
	flag.StringVar(&from, "from", "", "")
	flag.StringVar(&from, "f", "", "")
	flag.StringVar(&to, "to", "", "")
	flag.StringVar(&to, "t", "", "")

	flag.Usage = func() { fmt.Print(usage) }
	flag.Parse()

	log.Printf("scaleconnect version %s\n", Version)

	type Config struct {
		From string `yaml:"from"`
		To   string `yaml:"to"`

		Expr map[string]string `yaml:"expr"`
	}

	var conf map[string]Config

	// if from and to not empty - no need to read config file
	if from != "" && to != "" {
		sync = "sync_cli"
		conf = map[string]Config{sync: {From: from, To: to}}
	} else {
		f, err := openConfig(confName)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()

		if err = yaml.NewDecoder(f).Decode(&conf); err != nil {
			log.Fatal(err)
		}
	}

	for name, v := range conf {
		if sync != "" && name != sync {
			continue
		}
		if from != "" {
			v.From = from
		}
		if to != "" {
			v.To = to
		}

		if v.From == "" || v.To == "" {
			continue
		}

		weights, err := internal.GetWeights(v.From)
		if err != nil {
			log.Printf("%s: load data error: %v\n", name, err)
			continue
		}

		if v.Expr != nil {
			if err = internal.Expr(v.Expr, weights); err != nil {
				log.Printf("%s: calc expr error: %v\n", name, err)
				continue
			}
		}

		if err = internal.SetWeights(v.To, weights); err != nil {
			log.Printf("%s: write data error: %v\n", name, err)
			continue
		}

		log.Printf("%s: OK\n", name)
	}
}

const configName = "scaleconnect.yaml"

func openConfig(name string) (*os.File, error) {
	if name != "" {
		return os.Open(name)
	}

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
