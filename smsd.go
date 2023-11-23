package gammu

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

type ModemRun struct {
	ModemNumber int
	Run         bool
	Error       string
}

func (g *Gammu) Run(modem *Modem) error {
	_, cfgPath, err := g.CreateConfig(modem)
	if err != nil {
		return fmt.Errorf("create config error: %s", err)
	}

	pidPath := filepath.Join(g.PidDir, modem.IMSI)

	out, err := exec.Command("gammu-smsd", "-c", cfgPath, "-p", pidPath, "-d").CombinedOutput()
	if err != nil {
		return fmt.Errorf("run gammu-smsd error: %s, output: %s, config: %s", err, string(out), cfgPath)
	}

	return nil
}

func (g *Gammu) RunAll(modems map[int]*Modem) []ModemRun {
	var wg sync.WaitGroup
	ch := make(chan ModemRun)

	for _, modem := range modems {
		wg.Add(1)
		i := modem
		go func() {
			defer wg.Done()
			err := g.Run(i)
			if err == nil {
				ch <- ModemRun{ModemNumber: i.Num, Run: true}
			} else {
				ch <- ModemRun{ModemNumber: i.Num, Run: false, Error: err.Error()}
			}
		}()
	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	var results []ModemRun
	for channelResult := range ch {
		results = append(results, channelResult)
	}

	return results
}

func (g *Gammu) Stop(modemID string) error {
	pid, err := g.GetPID(modemID)
	if err != nil {
		return err
	}
	out, err := exec.Command("kill", "-SIGINT", pid).CombinedOutput()
	if err != nil {
		return fmt.Errorf("stop gammu-smsd (modemID: %s) error: %s, output: %s", modemID, err, string(out))
	}
	return nil
}

func (g *Gammu) Reload(modemID string) error {
	pid, err := g.GetPID(modemID)
	if err != nil {
		return err
	}
	out, err := exec.Command("kill", "-SIGHUP", pid).CombinedOutput()
	if err != nil {
		return fmt.Errorf("reload gammu-smsd (modemID: %s) error: %s, Output: %s", modemID, err, string(out))
	}
	return nil
}

func (g *Gammu) GetPID(modemID string) (string, error) {
	path := filepath.Join(g.PidDir, modemID)
	out, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read %s error: %s, output: %s", path, err, string(out))
	}
	return strings.TrimSpace(string(out)), nil
}
