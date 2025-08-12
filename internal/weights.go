package internal

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/AlexxIT/SmartScaleConnect/pkg/core"
	"github.com/AlexxIT/SmartScaleConnect/pkg/csv"
	"github.com/AlexxIT/SmartScaleConnect/pkg/fitbit"
)

func GetWeights(from any) ([]*core.Weight, error) {
	switch from.(type) {
	case string:
		return getWeights(from.(string))

	case map[string]any:
		data, err := json.Marshal(from)
		if err != nil {
			return nil, err
		}

		var weight *core.Weight
		if err = json.Unmarshal(data, &weight); err != nil {
			return nil, err
		}
		if weight.Date.IsZero() {
			weight.Date = time.Now()
		}
		return []*core.Weight{weight}, nil

	case []any:
		data, err := json.Marshal(from)
		if err != nil {
			return nil, err
		}

		var weights []*core.Weight
		if err = json.Unmarshal(data, &weights); err != nil {
			return nil, err
		}
		return weights, nil
	}

	return nil, fmt.Errorf("wrong from format: %v", from)
}

func getWeights(config string) ([]*core.Weight, error) {
	switch fields := strings.Fields(config); fields[0] {
	case "csv":
		rd, err := openFile(fields[1])
		if err != nil {
			return nil, err
		}
		defer rd.Close()

		return csv.Read(rd)

	case "json":
		rd, err := openFile(fields[1])
		if err != nil {
			return nil, err
		}
		defer rd.Close()

		var weights []*core.Weight
		if err = json.NewDecoder(rd).Decode(&weights); err != nil {
			return nil, err
		}
		return weights, nil

	case "fitbit":
		return fitbit.Read(fields[1])

	case "garmin", "tanita":
		acc, err := GetAccount(fields)
		if err != nil {
			return nil, err
		}
		return acc.GetAllWeights()

	case "picooc", "xiaomi", "zepp/xiaomi":
		acc, err := GetAccount(fields)
		if err != nil {
			return nil, err
		}

		if len(fields) < 4 {
			return acc.GetAllWeights()
		}

		return acc.(core.AccountWithFilter).GetFilterWeights(fields[3])

	default:
		return nil, errors.New("unsupported type: " + fields[0])
	}
}

func SetWeights(config string, src []*core.Weight) error {
	switch fields := strings.Fields(config); fields[0] {
	case "csv", "json":
		return writeFile(config, src)

	case "garmin", "zepp/xiaomi":
		return appendAccount(config, src)

	case "json/latest":
		return postLatest(config, src)

	default:
		return errors.New("unsupported type: " + fields[0])
	}
}

func openFile(path string) (io.ReadCloser, error) {
	if strings.Contains(path, "://") {
		res, err := http.Get(path)
		if err != nil {
			return nil, err
		}
		return res.Body, nil
	} else {
		f, err := os.Open(path)
		if err != nil {
			return nil, err
		}
		return f, nil
	}
}

func writeFile(config string, src []*core.Weight) error {
	fields := strings.Fields(config)
	format := fields[0]
	filename := fields[1]

	if strings.Contains(filename, "://") {
		return postFile(format, filename, src)
	}

	if filename == "stdout" {
		return writeToStdout(format, src)
	}

	// important read file before os.Create
	// empty dst file is OK
	dst, _ := GetWeights(config)
	dst = appendFile(dst, src)

	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	if format == "csv" {
		return csv.Write(f, dst)
	} else {
		return json.NewEncoder(f).Encode(dst)
	}
}

func appendFile(dst, src []*core.Weight) []*core.Weight {
	for _, s := range src {
		i := slices.IndexFunc(dst, func(d *core.Weight) bool {
			return s.Date.Unix() == d.Date.Unix()
		})

		if i >= 0 {
			if s.Weight == 0 {
				dst = append(dst[:i], dst[i+1:]...) // remove
			} else if !core.Equal(s, dst[i]) {
				dst[i] = s // replace
			} else {
				// skip
			}
		} else {
			if s.Weight > 0 {
				dst = append(dst, s) // add
			} else {
				// skip
			}
		}
	}

	slices.SortFunc(dst, func(a, b *core.Weight) int {
		return a.Date.Compare(b.Date)
	})

	return dst
}

func appendAccount(config string, src []*core.Weight) error {
	dst, err := GetWeights(config)
	if err != nil {
		return err
	}

	acc, err := GetAccount(strings.Fields(config))
	if err != nil {
		return err
	}

	client := acc.(core.AccountWithAddWeights)

	var add []*core.Weight

	for _, s := range src {
		i := slices.IndexFunc(dst, func(d *core.Weight) bool {
			return s.Date.Unix() == d.Date.Unix()
		})

		if i >= 0 {
			d := dst[i]
			if s.Weight == 0 {
				// remove
				if err = client.DeleteWeight(d); err != nil {
					return err
				}
			} else if !client.Equal(s, d) {
				// replace
				if err = client.DeleteWeight(d); err != nil {
					return err
				}
				add = append(add, s)
			} else {
				// skip
			}
		} else {
			if s.Weight > 0 {
				add = append(add, s) // add
			} else {
				// skip
			}
		}
	}

	if len(add) == 0 {
		return nil
	}

	return client.AddWeights(add)
}

func prepareFile(src []*core.Weight) []*core.Weight {
	// skip zero weights
	dst := make([]*core.Weight, 0, len(src))
	for _, weight := range src {
		if weight.Weight != 0 {
			dst = append(dst, weight)
		}
	}

	// sort weights (latest last)
	slices.SortFunc(dst, func(a, b *core.Weight) int {
		return a.Date.Compare(b.Date)
	})

	return dst
}

func postFile(format, url string, src []*core.Weight) (err error) {
	body := bytes.NewBuffer(nil)
	dst := prepareFile(src)

	if format == "csv" {
		if err = csv.Write(body, dst); err != nil {
			return err
		}
		_, err = http.Post(url, "text/csv", body)
	} else {
		if err = json.NewEncoder(body).Encode(dst); err != nil {
			return err
		}
		_, err = http.Post(url, "application/json", body)
	}

	return
}

func postLatest(config string, src []*core.Weight) error {
	slices.SortFunc(src, func(a, b *core.Weight) int {
		return b.Date.Compare(a.Date) // latest first
	})

	for _, weight := range src {
		if weight.Weight == 0 {
			continue
		}

		data, err := json.Marshal(weight)
		if err != nil {
			return err
		}

		fields := strings.Fields(config)

		res, err := http.Post(fields[1], "application/json", bytes.NewBuffer(data))
		if err != nil {
			return err
		}
		defer res.Body.Close()

		break
	}

	return nil
}

func writeToStdout(format string, src []*core.Weight) error {
	dst := prepareFile(src)

	if format == "csv" {
		return csv.Write(os.Stdout, dst)
	} else {
		return json.NewEncoder(os.Stdout).Encode(dst)
	}
}
