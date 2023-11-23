package gammu

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"sync"
)

type ModemIdentify struct {
	ModemNumber int
	Error       string
	ErrorCode   int
	// Fields after execution cmd command "gammu identify"
	Device       string
	Manufacturer string
	Model        string
	Firmware     string
	IMEI         string
	IMSI         string
}

func Identify(n int) (*ModemIdentify, error) {
	out, err := exec.Command("timeout", "60", "gammu", "-s", strconv.Itoa(n), "identify").CombinedOutput()
	if err != nil {
		if err.Error() == "exit status 114" {
			return &ModemIdentify{ModemNumber: n, Error: "No SIM", ErrorCode: 114}, nil
		}
		return nil, fmt.Errorf("gammu-identify error: %s, modemN: %d, output: %s", err, n, string(out))
	}
	modem := ParseIdentify(string(out))
	modem.ModemNumber = n
	return modem, err
}

func IdentifyAll(modems map[int]*Modem) []ModemIdentify {
	var wg sync.WaitGroup
	ch := make(chan ModemIdentify)

	for _, modem := range modems {
		wg.Add(1)
		i := modem.Num
		go func() {
			defer wg.Done()
			info, err := Identify(i)
			if err == nil {
				ch <- *info
			} else {
				ch <- ModemIdentify{ModemNumber: i, Error: err.Error()}
			}
		}()
	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	var results []ModemIdentify
	for channelResult := range ch {
		results = append(results, channelResult)
	}

	return results
}

func ParseIdentify(s string) *ModemIdentify {
	result := &ModemIdentify{}
	lines := strings.Split(s, "\n")
	for _, line := range lines {
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		switch key {
		case "Device":
			result.Device = value
		case "Manufacturer":
			result.Manufacturer = value
		case "Model":
			result.Model = value
		case "Firmware":
			result.Firmware = value
		case "IMEI":
			result.IMEI = value
		case "SIM IMSI":
			result.IMSI = value
		}
	}
	return result
}
