package gammu

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

type ModemMonitor struct {
	ModemNumber int
	Error       string
	// From Exec
	PhoneID       string
	IMEI          string
	IMSI          string
	Sent          string
	Received      string
	Failed        string
	BatterPercent string
	NetworkSignal string
}

func (g *Gammu) Monitor(id string) (*ModemMonitor, error) {
	cfgPath := filepath.Join(g.CfgDir, id)
	out, err := exec.Command("gammu-smsd-monitor", "-c", cfgPath, "-n", "1", "-d", "0").CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("run gammu-smsd-monitor error: %s, output: %s, config: %s", err, string(out), cfgPath)
	}
	data := ParseMonitor(string(out))
	if data.IMSI == "" {
		return nil, fmt.Errorf("parse gammu-smsd-monitor error: %s", string(out))
	}
	return data, nil
}

func (g *Gammu) MonitorAll(modems map[int]*Modem) []ModemMonitor {
	var wg sync.WaitGroup
	ch := make(chan ModemMonitor)

	for _, modem := range modems {
		wg.Add(1)
		i := modem.IMSI
		go func() {
			defer wg.Done()
			info, err := g.Monitor(i)
			if err == nil {
				ch <- *info
			} else {
				ch <- ModemMonitor{Error: err.Error()}
			}
		}()
	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	var results []ModemMonitor
	for channelResult := range ch {
		results = append(results, channelResult)
	}

	return results
}

func ParseMonitor(s string) *ModemMonitor {
	result := &ModemMonitor{}
	lines := strings.Split(s, "\n")
	for _, line := range lines {
		data := strings.Split(line, ":")
		if len(data) != 2 {
			continue
		}
		key := strings.ReplaceAll(data[0], " ", "")
		value := strings.ReplaceAll(data[1], " ", "")
		if strings.Contains(key, "PhoneID") {
			result.PhoneID = value
		} else if strings.Contains(key, "IMEI") {
			result.IMEI = value
		} else if strings.Contains(key, "IMSI") {
			result.IMSI = value
		} else if strings.Contains(key, "Sent") {
			result.Sent = value
		} else if strings.Contains(key, "Received") {
			result.Received = value
		} else if strings.Contains(key, "Failed") {
			result.Failed = value
		} else if strings.Contains(key, "BatterPercent") {
			result.BatterPercent = value
		} else if strings.Contains(key, "NetworkSignal") {
			result.NetworkSignal = value
		} else {
			continue
		}
	}
	return result
}
