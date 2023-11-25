package api

import (
	"github.com/gofiber/fiber/v2"
	gammu "github.com/justficks/gogammu"
	"strconv"
)

func (h *Handler) GetAllModems(c *fiber.Ctx) error {
	modems := h.Gammu.Store.GetModems()
	res := make([]*gammu.Modem, 0, len(modems))
	for _, modem := range modems {
		res = append(res, modem)
	}
	return c.JSON(res)
}

func (h *Handler) DetectDevices(c *fiber.Ctx) error {
	gammurc, err := gammu.DetectDevices()
	if err != nil {
		return c.Status(500).SendString(err.Error())
	}
	devices := gammu.ExtractUSBDevices(gammurc)
	h.Gammu.Store.SetModemsDetect(devices)
	return c.JSON(devices)
}

func (h *Handler) IdentifyAllModems(c *fiber.Ctx) error {
	modems := gammu.IdentifyAll(h.Gammu.Store.GetModems())
	h.Gammu.Store.SetModemsIdentify(modems)
	return c.JSON(modems)
}

func (h *Handler) NetworkInfoAll(c *fiber.Ctx) error {
	modems := gammu.NetworkAll(h.Gammu.Store.GetModemsByStatus(gammu.Stop))
	h.Gammu.Store.SetModemsNetwork(modems)
	return c.SendStatus(200)
}

func (h *Handler) MonitorAll(c *fiber.Ctx) error {
	modems := h.Gammu.MonitorAll(h.Gammu.Store.GetModemsByStatus(gammu.Run))
	h.Gammu.Store.SetModemsMonitor(modems)
	return c.SendStatus(200)
}

func (h *Handler) DeleteMonitor(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return fiber.NewError(400, "Parameter :id must be a number.")
	}
	err = h.Gammu.DeleteMonitor(id)
	if err != nil {
		return c.Status(500).SendString(err.Error())
	}
	return c.SendStatus(200)
}

func (h *Handler) RunAll(c *fiber.Ctx) error {
	modems := h.Gammu.RunAll(h.Gammu.Store.GetModemsByStatus(gammu.Stop))
	h.Gammu.Store.SetModemsRun(modems)
	return c.SendStatus(200)
}

func (h *Handler) StopAll(c *fiber.Ctx) error {
	err := h.Gammu.KillSMSDs()
	if err != nil {
		return c.Status(500).SendString(err.Error())
	}
	return c.SendStatus(200)
}

func (h *Handler) GetAllPids(c *fiber.Ctx) error {
	modems := h.Gammu.Store.GetModemsByStatus(gammu.Run)
	for _, modem := range modems {
		pid, err := h.Gammu.GetPID(modem.IMSI)
		if err == nil && modem.Status == gammu.Run {
			h.Gammu.Store.SetModemPID(modem.Num, pid)
		}
	}
	return c.SendStatus(200)
}
