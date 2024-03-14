package main

import (
	"encoding/json"
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/justficks/gogammu/internal/config"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"

	gammu "github.com/justficks/gogammu"
)

// + Получаем с основной программы данные из Store через http запрос
// + Останавливаем все модемы через основную программу
// + GetUSSD по каждой симке
// - Запускаем GlobeRun через основную программу

func MustSetupLogger(cfg *config.Config) *slog.Logger {
	file, err := os.OpenFile(filepath.Join(cfg.LogFilePath, cfg.LogFileName), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		panic(fmt.Sprintf("Unable to open log file '%s/%s', because: %s", cfg.LogFilePath, cfg.LogFileName, err.Error()))
	}

	return slog.New(slog.NewJSONHandler(file, &slog.HandlerOptions{Level: slog.LevelDebug}))
}

var appUrl string
var appLog *slog.Logger
var botToken string
var botChatId string

func main() {
	cfg := config.MustLoad()
	appLog = MustSetupLogger(cfg)

	appUrl = cfg.AppUrl
	botToken = cfg.BotToken
	botChatId = cfg.BotChatId

	appLog.Debug("ЗАПУСК ПРОГРАММЫ")

	appLog.Debug("Отправляем запрос на получение модемов")
	modems, err := getModems()
	if err != nil {
		sendTgMsg("Ошибка получения информации о модемах после запуска telebank-checker")
		appLog.Error(err.Error())
		panic(err.Error())
	}
	appLog.Debug(fmt.Sprintf("Получено %d объектов", len(modems)))

	appLog.Debug("Отправляем запрос остановку всех модемов")
	err = stopAllModems()
	if err != nil {
		sendTgMsg("Ошибка после запроса на остановку всех модемах в telebank-checker")
		appLog.Error(err.Error())
		panic(err.Error())
	}

	for _, modem := range modems {
		if modem.Status == "no-sim" {
			appLog.Debug("Пропускаем проверку модема без SIM карты", slog.Int("Num", modem.Num))
			continue
		}

		respBody, err := sendNetworkinfo(modem.Num)
		if err != nil {
			appLog.Debug("Запрос на отправку /modem/:num/identify завершился с ошибкой", slog.Int("Num", modem.Num), slog.String("Error", err.Error()), slog.Any("respBody", respBody))
			msg := fmt.Sprintf("Кажется SIM карта (IMSI: %s) в модеме № %d не работает", modem.IMSI, modem.Num)
			sendTgMsg(msg)
			continue
		}

		if respBody.NetworkState == "registration to network denied" {
			msg := fmt.Sprintf("В Модем № %d с IMSI: %s SIM-карта не может зарегестрироваться в сети", modem.Num, modem.IMSI)
			appLog.Warn(msg, slog.Any("respBody", respBody))
			sendTgMsg(msg)
		}
	}

	appLog.Debug("Отправляем запрос на запуск всех модемов")
	err = runAll()
	if err != nil {
		appLog.Error(err.Error())
		sendTgMsg(fmt.Sprintf("После проверки SIM карт произошла ошибка: %s", err.Error()))
	}

	appLog.Debug("ВЫПОЛНЕНИЕ ПРОГРАММЫ ЗАВЕРШЕНО")
}

func sendTgMsg(text string) {
	client := resty.New()
	resp, err := client.R().
		SetBody(map[string]interface{}{"chat_id": botChatId, "text": text}).
		Post("https://api.telegram.org/bot" + botToken + "/sendMessage")

	if err != nil {
		appLog.Error("При отправке сообщения в телеграм произошла ошибка", slog.String("Error", err.Error()))
		return
	}
	if resp.StatusCode() >= 300 {
		appLog.Error(fmt.Sprintf("Запрос на отправку телеграм сообщения вернулся с ответом: %x и body: %s", resp.StatusCode(), resp.Body()))
		return
	}
}

func getModems() ([]gammu.Modem, error) {
	client := resty.New()

	resp, err := client.R().
		Get(appUrl + "/modems")
	if err != nil {
		return nil, fmt.Errorf("неудалось отправить запрос на /modems. Ошибка: %s", err.Error())
	}

	respBody := resp.Body()

	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("запрос на /modems вернулся с ответом: %d и body: %s", resp.StatusCode(), string(respBody))
	}

	var modems []gammu.Modem
	err = json.Unmarshal(respBody, &modems)
	if err != nil {
		return nil, fmt.Errorf("ошибка при десериализации JSON: %s", err.Error())
	}

	return modems, nil
}

func stopAllModems() error {
	client := resty.New()
	resp, err := client.R().
		Get(appUrl + "/modems/stop")

	if err != nil {
		return fmt.Errorf("неудалось отправить запрос на /modems/stop. Ошибка: %s", err.Error())
	}
	if resp.StatusCode() != 200 {
		return fmt.Errorf("запрос на /modems/stop вернулся с ответом: %d и body: %s", resp.StatusCode(), resp.Body())
	}

	return nil
}

type ModemNetwork struct {
	ModemNumber        int    `json:"modem_number"`
	Error              string `json:"error"`
	NetworkState       string `json:"network_state"`
	Network            string `json:"network"`
	NameInPhone        string `json:"name_in_phone"`
	PacketNetworkState string `json:"packet_network_state"`
	PacketNetwork      string `json:"packet_network"`
	GPRS               string `json:"gprs"`
}

func sendNetworkinfo(Num int) (ModemNetwork, error) {
	client := resty.New()

	resp, err := client.R().
		SetHeader("Content-Type", "text/plain").
		SetBody("*111*0887#").
		Post(appUrl + "/modem/" + strconv.Itoa(Num) + "/networkinfo")
	if err != nil {
		return ModemNetwork{}, fmt.Errorf("неудалось отправить запрос на /modem/:num/networkinfo. Ошибка: %s", err.Error())
	}

	respBody := string(resp.Body())
	if resp.StatusCode() != 200 {
		return ModemNetwork{}, fmt.Errorf("запрос на /modem/:num/networkinfo ответил со статусом %d и body: %s", resp.StatusCode(), respBody)
	}

	var modem ModemNetwork

	// Десериализация JSON-строки в структуру ModemNetwork
	err = json.Unmarshal([]byte(respBody), &modem)
	if err != nil {
		fmt.Println("Error:", err)
		return ModemNetwork{}, fmt.Errorf("ошибка десериализаци JSON-строки. respBody: %s", respBody)
	}

	return modem, nil
}

func runAll() error {
	client := resty.New()

	resp, err := client.R().
		Get(appUrl + "/globe/run")
	if err != nil {
		return fmt.Errorf("неудалось отправить запрос на /globe/run. Ошибка: %s", err.Error())
	}

	respBody := string(resp.Body())
	if resp.StatusCode() != 200 {
		return fmt.Errorf("запрос на /globe/run ответил со статусом %d и body: %s", resp.StatusCode(), respBody)
	}

	return nil
}
