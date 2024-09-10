package gammu

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"sync"

	"github.com/go-pg/pg/v10"
)

type Gammu struct {
	AppDir  string
	AppPort int

	CfgDir string
	PidDir string
	LogDir string

	DbHost string
	DbUser string
	DbPass string

	RunOnMsgScript  string
	RunOnErrScript  string
	RunOnCallScript string

	OnMsgCallback   func(msg *NewMsg) error
	OnErrorCallback func(text string) error

	Store *Store
	DB    *pg.DB

	log *slog.Logger
}

type ConfigNewGammu struct {
	AppDir  string
	AppPort string

	DbHost string
	DbUser string
	DbPass string
	DbName string

	OnMsgCallback   func(msg *NewMsg) error
	OnErrorCallback func(text string) error

	Logger *slog.Logger
}

func NewGammu(cfg ConfigNewGammu) (*Gammu, error) {
	cfgDir := filepath.Join(cfg.AppDir, "configs")
	pidDir := filepath.Join(cfg.AppDir, "pids")
	logDir := filepath.Join(cfg.AppDir, "logs")

	dirs := []string{
		cfg.AppDir,
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

	runOnScripts := map[string]string{
		"msg":  filepath.Join(cfg.AppDir, "runOnMsg.sh"),
		"err":  filepath.Join(cfg.AppDir, "runOnErr.sh"),
		"call": filepath.Join(cfg.AppDir, "runOnCall.sh"),
	}
	for eventType, scriptPath := range runOnScripts {
		err := CreateRunOnScript(scriptPath, cfg.AppPort, eventType)
		if err != nil {
			return nil, fmt.Errorf("create %s error %s", scriptPath, err)
		}
	}

	DbConnection := pg.Connect(&pg.Options{
		Addr:     cfg.DbHost,
		User:     cfg.DbUser,
		Password: cfg.DbPass,
		Database: cfg.DbName,
	})

	// Проверка соединения с БД
	_, err := DbConnection.Exec("SELECT 1")
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %v", err)
	}

	return &Gammu{
		AppDir: cfg.AppDir,
		CfgDir: cfgDir,
		PidDir: pidDir,
		LogDir: logDir,

		// Required for gammu configs
		DbHost: cfg.DbHost,
		DbUser: cfg.DbUser,
		DbPass: cfg.DbPass,

		RunOnMsgScript:  runOnScripts["msg"],
		RunOnErrScript:  runOnScripts["err"],
		RunOnCallScript: runOnScripts["call"],

		OnMsgCallback:   cfg.OnMsgCallback,
		OnErrorCallback: cfg.OnErrorCallback,

		Store: GetStore(),
		DB:    DbConnection,

		log: cfg.Logger,
	}, nil
}

type ModemStatus string

const (
	Run   ModemStatus = "run"
	Stop  ModemStatus = "stop"
	Error ModemStatus = "error"
	NoSIM ModemStatus = "no-sim"
	Init  ModemStatus = "init"

	NetworkError ModemStatus = "network-error"
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
		go func() {
			defer wg.Done()
			_ = g.Run(modem)
		}()
	}

	wg.Wait()

	return nil
}

func (g *Gammu) Run(m *Modem) error {
	err := g.DeleteMonitor(m.IMSI)
	if err != nil {
		g.log.Error("Gammu -> Run. Error remove row from DB before start", slog.Any("err", err))
	}

	g.Store.SetModemStatus(m.Num, Init)
	identify, err := Identify(m.Num)
	if err != nil {
		g.Store.SetModemStatus(m.Num, Error)
		g.Store.SetModemError(m.Num, err.Error())
		return err
	}
	if identify.ErrorCode == 114 {
		g.Store.SetModemStatus(m.Num, NoSIM)
		return fmt.Errorf("no sim card in modem %d", m.Num)
	}
	g.Store.SetModemIdentify(identify)
	network, err := Network(m.Num)
	if err != nil {
		g.Store.SetModemStatus(m.Num, Error)
		g.Store.SetModemError(m.Num, err.Error())
		return err
	}
	if network.NetworkState == "not logged into network" || network.NetworkState == "request to network denied" || network.NetworkState == "registration to network denied" {
		g.Store.SetModemStatus(m.Num, NetworkError)
		g.Store.SetModemError(m.Num, "Registration to network error")
		return fmt.Errorf("registration to network error")
	}
	g.Store.SetModemNetwork(m.Num, network)
	err = g.RunSMSD(m)
	if err != nil {
		g.Store.SetModemStatus(m.Num, Error)
		g.Store.SetModemError(m.Num, err.Error())
		return err
	}
	g.Store.SetModemStatus(m.Num, Run)
	return nil
}
