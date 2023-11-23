package gammu

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"sync"
)

type ModemNetwork struct {
	ModemNumber int
	Error       string
	// Fields after execution cmd command "gammu networkinfo"
	NetworkState       string
	Network            string
	NameInPhone        string
	PacketNetworkState string
	PacketNetwork      string
	GPRS               string
}

func Network(n int) (*ModemNetwork, error) {
	out, err := exec.Command("timeout", "60", "gammu", "-s", strconv.Itoa(n), "networkinfo").CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("run gammu -s %d networkinfo error: %v, output: %s", n, err, string(out))
	}
	modem := ParseNetwork(string(out))
	modem.ModemNumber = n
	return modem, err
}

func NetworkAll(modems map[int]*Modem) []ModemNetwork {
	var wg sync.WaitGroup
	ch := make(chan ModemNetwork)

	for _, modem := range modems {
		wg.Add(1)
		i := modem.Num
		go func() {
			defer wg.Done()
			info, err := Network(i)
			// Нужно обработать ошибку
			if err == nil {
				ch <- *info
			} else {
				ch <- ModemNetwork{ModemNumber: i, Error: err.Error()}
			}
		}()
	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	var results []ModemNetwork
	for channelResult := range ch {
		results = append(results, channelResult)
	}

	return results
}

func ParseNetwork(s string) *ModemNetwork {
	ni := &ModemNetwork{}
	lines := strings.Split(s, "\n")
	for _, line := range lines {
		parts := strings.SplitN(line, ":", 2) // разделение строки на 2 части по двоеточию
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])   // удаляем пробелы вокруг ключа
		value := strings.TrimSpace(parts[1]) // удаляем пробелы вокруг значения

		switch key {
		case "Network state":
			ni.NetworkState = value
		case "Network":
			ni.Network = value
		case "Name in phone":
			ni.NameInPhone = value
		case "Packet network state":
			ni.PacketNetworkState = value
		case "Packet network":
			ni.PacketNetwork = value
		case "GPRS":
			ni.GPRS = value
		}
	}
	return ni
}
