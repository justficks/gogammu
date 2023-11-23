package gammu

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
)

func (g *Gammu) SendSMS(modemId, phone, text string) error {
	if phone == "" || text == "" {
		return fmt.Errorf("phone or text is empty")
	}

	cfgPath := filepath.Join(g.CfgDir, modemId)

	if _, err := os.Stat(cfgPath); errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("cfgPath %s not exists", cfgPath)
	}

	out, err := exec.Command("gammu-smsd-inject", "-c", cfgPath, "TEXT", phone, "-text", text).CombinedOutput()
	if err != nil {
		return fmt.Errorf("gammu-smsd-inject error: %s, output: %s, config: %s", err, string(out), cfgPath)
	}

	return nil
}

func (g *Gammu) SendUSSD(modemNum int, ussd string) (string, error) {
	if ussd == "" {
		return "", fmt.Errorf("ussd is empty")
	}

	out, err := exec.Command("timeout", "30", "gammu", "-s", strconv.Itoa(modemNum), "getussd", ussd).CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("gammu -s %d getussd %s error: %s, output: %s", modemNum, ussd, err, string(out))
	}

	return string(out), nil
}
