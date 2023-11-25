package api

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	gammu "github.com/justficks/gogammu"
	"os"
	"path/filepath"
	"strconv"
)

func (h *Handler) ModemMiddleware(c *fiber.Ctx) error {
	num, err := strconv.Atoi(c.Params("num"))
	if err != nil {
		return fiber.NewError(400, "Parameter :num must be a number.")
	}

	modem, ok := h.Gammu.Store.GetModem(num)
	if !ok {
		return fiber.NewError(400, "Modem not found")
	}

	c.Locals("modem", modem)

	return c.Next()
}

func (h *Handler) GetModem(c *fiber.Ctx) error {
	modem := c.Locals("modem").(*gammu.Modem)
	return c.JSON(modem)
}

func (h *Handler) IdentifyModem(c *fiber.Ctx) error {
	modem := c.Locals("modem").(*gammu.Modem)

	modemIdentify, err := gammu.Identify(modem.Num)
	if err != nil {
		return fiber.NewError(400, fmt.Sprintf("Identify modem %d error: %s", modem.Num, err))
	}

	h.Gammu.Store.SetModemIdentify(modemIdentify)

	return c.JSON(modemIdentify)
}

func (h *Handler) CreateGammuConfig(c *fiber.Ctx) error {
	modem := c.Locals("modem").(*gammu.Modem)

	_, cfgPath, err := h.Gammu.CreateConfig(modem)
	if err != nil {
		return fiber.NewError(400, fmt.Sprintf("Create config error: %s", err))
	}

	return c.JSON(cfgPath)
}

func (h *Handler) RunModemGammuSmsd(c *fiber.Ctx) error {
	modem := c.Locals("modem").(*gammu.Modem)

	modemIdentify, err := gammu.Identify(modem.Num)
	if err != nil {
		return fiber.NewError(400, fmt.Sprintf("Identify modem %d error: %s", modem.Num, err))
	}

	h.Gammu.Store.SetModemIdentify(modemIdentify)

	err = h.Gammu.Run(modem)
	if err != nil {
		h.Gammu.Store.SetModemStatus(modem.Num, gammu.Error)
		h.Gammu.Store.SetModemError(modem.Num, err.Error())
		return fiber.NewError(400, fmt.Sprintf("Run gammu-smsd error: %s", err))
	}

	h.Gammu.Store.SetModemStatus(modem.Num, gammu.Run)

	return c.JSON(modem)
}

func (h *Handler) StopModemGammuSmsd(c *fiber.Ctx) error {
	modem := c.Locals("modem").(*gammu.Modem)

	err := h.Gammu.Stop(modem.IMSI)
	if err != nil {
		return fiber.NewError(400, fmt.Sprintf("Stop gammu-smsd error: %s", err))
	}

	h.Gammu.Store.SetModemStatus(modem.Num, gammu.Stop)

	return c.SendString("OK")
}

func (h *Handler) ReloadGammuSmsd(c *fiber.Ctx) error {
	modem := c.Locals("modem").(*gammu.Modem)

	err := h.Gammu.Reload(modem.IMSI)
	if err != nil {
		return fiber.NewError(400, fmt.Sprintf("Reload gammu-smsd error: %s", err))
	}

	return c.SendString("OK")
}

func (h *Handler) GetModemLogs(c *fiber.Ctx) error {
	modem := c.Locals("modem").(*gammu.Modem)

	path := filepath.Join(h.Gammu.LogDir, modem.IMSI)
	out, err := os.ReadFile(path)
	if err != nil {
		return fiber.NewError(400, fmt.Sprintf("Ошибка чтения файла логов %s по причине: %s", path, err))
	}

	return c.Send(out)
}

func (h *Handler) GetGammuConfig(c *fiber.Ctx) error {
	modem := c.Locals("modem").(*gammu.Modem)

	path := filepath.Join(h.Gammu.CfgDir, modem.IMSI)
	out, err := os.ReadFile(path)
	if err != nil {
		return fiber.NewError(400, fmt.Sprintf("Ошибка чтения файла конфигурации %s по причине: %s", path, err))
	}

	return c.Send(out)
}

func (h *Handler) GetPid(c *fiber.Ctx) error {
	modem := c.Locals("modem").(*gammu.Modem)

	pid, err := h.Gammu.GetPID(modem.IMSI)
	if err != nil {
		return fiber.NewError(400, fmt.Sprintf("Ошибка получения PID процесса gammu-smsd %s по причине: %s", modem.IMSI, err))
	}

	if modem.Status == gammu.Run {
		h.Gammu.Store.SetModemPID(modem.Num, pid)
	}

	return c.SendString(pid)
}

func (h *Handler) SendSms(c *fiber.Ctx) error {
	modem := c.Locals("modem").(*gammu.Modem)

	body := c.Body()

	phone := c.Query("phone")
	text := string(body)

	err := h.Gammu.SendSMS(modem.IMSI, phone, text)
	if err != nil {
		return fiber.NewError(400, fmt.Sprintf("Ошибка отправки СМС с модема %d на телефон %s с текстом %s по причине: %s", modem.Num, phone, text, err))
	}

	return c.SendStatus(200)
}

func (h *Handler) SendUSSD(c *fiber.Ctx) error {
	modem := c.Locals("modem").(*gammu.Modem)

	body := c.Body()
	ussd := string(body)

	response, err := h.Gammu.SendUSSD(modem.Num, ussd)
	if err != nil {
		return fiber.NewError(400, fmt.Sprintf("Ошибка отправки USSD %s с модема %d по причине: %s", ussd, modem.Num, err))
	}

	return c.Send([]byte(response))
}

func (h *Handler) GetModemMonitorInfo(c *fiber.Ctx) error {
	modem := c.Locals("modem").(*gammu.Modem)

	monitorInfo, err := h.Gammu.Monitor(modem.IMSI)
	if err != nil {
		h.Gammu.Store.SetModemStatus(modem.Num, gammu.Error)
		h.Gammu.Store.SetModemError(modem.Num, err.Error())
		return fiber.NewError(400, fmt.Sprintf("Ошибка получения информации о модеме %d по причине: %s", modem.Num, err))
	}

	h.Gammu.Store.SetModemMonitor(modem.Num, monitorInfo)

	return c.SendStatus(200)
}

func (h *Handler) GetModemNetworkInfo(c *fiber.Ctx) error {
	modem := c.Locals("modem").(*gammu.Modem)

	networkInfo, err := gammu.Network(modem.Num)
	if err != nil {
		h.Gammu.Store.SetModemStatus(modem.Num, gammu.Error)
		h.Gammu.Store.SetModemError(modem.Num, err.Error())
		return fiber.NewError(400, fmt.Sprintf("Ошибка получения информации о модеме %d по причине: %s", modem.Num, err))
	}

	h.Gammu.Store.SetModemNetwork(modem.Num, networkInfo)

	return c.JSON(networkInfo)
}

func (h *Handler) ResetModem(c *fiber.Ctx) error {
	modem := c.Locals("modem").(*gammu.Modem)

	err := gammu.ResetModem(modem.Num)
	if err != nil {
		return fiber.NewError(400, fmt.Sprintf("Ошибка сброса настроек модема: %s", err))
	}

	return c.SendStatus(200)
}
