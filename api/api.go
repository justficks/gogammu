package api

import (
	"github.com/gofiber/fiber/v2"
	gammu "github.com/justficks/gogammu"
)

type Api struct {
	GammuFiber *fiber.App
}

func New(gammuInstance *gammu.Gammu) (api *Api) {
	app := fiber.New()

	h := &Handler{Gammu: gammuInstance}

	modem := app.Group("/modem/:num", h.ModemMiddleware)
	modem.Get("/", h.GetModem)
	modem.Post("/identify", h.IdentifyModem)
	modem.Get("/config", h.GetGammuConfig)
	modem.Post("/config", h.CreateGammuConfig)
	modem.Get("/pid", h.GetPid)
	modem.Get("/logs", h.GetModemLogs)
	modem.Post("/run", h.RunModemGammuSmsd)
	modem.Post("/stop", h.StopModemGammuSmsd)
	modem.Post("/reset", h.ResetModem)
	modem.Post("/reload", h.ReloadGammuSmsd)
	modem.Post("/monitor", h.GetModemMonitorInfo)
	modem.Post("/networkinfo", h.GetModemNetworkInfo)
	modem.Post("/sms/send/", h.SendSms)
	modem.Post("/ussd/send/", h.SendUSSD)

	app.Get("/modems", h.GetAllModems)
	app.Get("/modems/detect", h.DetectDevices)
	app.Get("/modems/identify", h.IdentifyAllModems)
	app.Get("/modems/networkinfo", h.NetworkInfoAll)
	app.Get("/modems/monitor", h.MonitorAll)
	app.Delete("/modems/monitor/:id", h.DeleteMonitor)
	app.Get("/modems/run", h.RunAll)
	app.Get("/modems/stop", h.StopAll)
	app.Get("/modems/pids", h.GetAllPids)

	return &Api{
		GammuFiber: app,
	}
}
