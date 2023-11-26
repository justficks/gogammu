package gammu

import (
	"fmt"
	"github.com/go-pg/pg/v10"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"sync"
)

type Gammu struct {
	AppDir string
	CfgDir string
	PidDir string
	LogDir string
	Script string // path to notify.sh

	DbHost string
	DbUser string
	DbPass string

	Store *Store
	DB    *pg.DB
}

func executeSQLFromFile(db *pg.DB, filename string) error {
	content, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	sql := string(content)
	_, err = db.Exec(sql)
	return err
}

func NewGammu(appDir, appHttpPort, dbAddr, dbUser, dbPass, dbName string) (*Gammu, error) {
	cfgDir := filepath.Join(appDir, "configs")
	pidDir := filepath.Join(appDir, "pids")
	logDir := filepath.Join(appDir, "logs")

	dirs := []string{
		appDir,
		cfgDir,
		pidDir,
		logDir,
	}

	for _, dir := range dirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			err = os.MkdirAll(dir, os.ModePerm)
			if err != nil {
				return nil, fmt.Errorf("create %s error %s", dir, err)
			}
		}
	}

	scriptPath := filepath.Join(appDir, "notify.sh")
	err := CreateNotifyScript(scriptPath, appHttpPort)
	if err != nil {
		return nil, fmt.Errorf("create %s error %s", scriptPath, err)
	}

	DbConnection := pg.Connect(&pg.Options{
		Addr:     dbAddr,
		User:     dbUser,
		Password: dbPass,
		Database: dbName,
	})

	// Проверка соединения
	_, err = DbConnection.Exec("SELECT 1")
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %v", err)
	}

	// Выполнение SQL-кода из файла
	err = executeSQLFromFile(DbConnection, "gammu-pg.sql")
	if err != nil {
		return nil, fmt.Errorf("failed to execute SQL file: %v", err)
	}

	return &Gammu{
		AppDir: appDir,
		CfgDir: filepath.Join(appDir, "./configs"),
		PidDir: filepath.Join(appDir, "./pids"),
		LogDir: filepath.Join(appDir, "./logs"),
		Script: scriptPath,

		// Required for gammu configs
		DbHost: dbAddr,
		DbUser: dbUser,
		DbPass: dbPass,

		Store: GetStore(),
		DB:    DbConnection,
	}, nil
}

type ModemStatus string

const (
	Run   ModemStatus = "run"
	Stop  ModemStatus = "stop"
	Error ModemStatus = "error"
	NoSIM ModemStatus = "no-sim"
	Init  ModemStatus = "init"
)

type Modem struct {
	Num int // Порядковый номер модема в ферме

	Status ModemStatus // Статус модема = start | run | stop | error
	Error  string      // Описание ошибки, если есть
	PID    string      // PID процесса gammu-smsd

	// common
	IMEI string // Идентификатор модема
	IMSI string // Идентификатор СИМ-карты = имя файла конфигурации и файла с его текущим pid

	// identify
	Device       string
	Manufacturer string
	Model        string
	Firmware     string

	// monitor
	PhoneID       string
	Sent          string
	Received      string
	Failed        string
	BatterPercent string
	NetworkSignal string

	// networkinfo
	NetworkState       string
	Network            string
	NameInPhone        string
	PacketNetworkState string
	PacketNetwork      string
	GPRS               string
}

func ResetModem(n int) error {
	out, err := exec.Command("timeout", "30", "gammu", "-s", strconv.Itoa(n), "reset", "SOFT").CombinedOutput()
	if err != nil {
		return fmt.Errorf("reset modem: %d error: %s, Output: %s", n, err, string(out))
	}
	return nil
}

func (g *Gammu) KillSMSDs() error {
	out, err := exec.Command("pkill", "gammu-smsd").CombinedOutput()
	if err != nil {
		return fmt.Errorf("pkill gammu-smsd error: %s, output: %s", err, string(out))
	}
	g.Store.SetModemsStop()
	return nil
}

func (g *Gammu) GlobeRun() error {
	var wg sync.WaitGroup

	gammurc, err := DetectDevices()
	if err != nil {
		return err
	}
	devices := ExtractUSBDevices(gammurc)
	g.Store.SetModemsDetect(devices)

	for _, m := range g.Store.modems {
		wg.Add(1)
		modem := m
		modemNum := m.Num
		go func(id int) {
			defer wg.Done()
			g.Store.SetModemStatus(modemNum, Init)
			identify, err := Identify(modemNum)
			if err != nil {
				g.Store.SetModemStatus(modemNum, Error)
				g.Store.SetModemError(modemNum, err.Error())
				return
			}
			if identify.ErrorCode == 114 {
				g.Store.SetModemStatus(modemNum, NoSIM)
				return
			}
			g.Store.SetModemIdentify(identify)
			network, err := Network(modemNum)
			if err != nil {
				g.Store.SetModemStatus(modemNum, Error)
				g.Store.SetModemError(modemNum, err.Error())
				return
			}
			if network.NetworkState == "not logged into network" {
				g.Store.SetModemStatus(modemNum, Error)
				g.Store.SetModemError(modemNum, "not logged into network")
				return
			}
			g.Store.SetModemNetwork(modemNum, network)
			err = g.Run(modem)
			if err != nil {
				g.Store.SetModemStatus(modemNum, Error)
			}
			g.Store.SetModemStatus(modemNum, Run)
		}(modemNum)
	}

	wg.Wait()

	return nil
}
