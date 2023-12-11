package api

import (
	"github.com/ansrivas/fiberprometheus/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	gammu "github.com/justficks/gogammu"
	slogfiber "github.com/samber/slog-fiber"
	"log/slog"
)

type Api struct {
	Fiber *fiber.App
}

func New(gammuInstance *gammu.Gammu, logger *slog.Logger) (api *Api) {
	app := fiber.New()

	prometheus := fiberprometheus.New("my-service-name")
	prometheus.RegisterAt(app, "/metrics")
	app.Use(prometheus.Middleware)

	app.Use(cors.New())
	app.Use(slogfiber.New(logger))

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

	app.Get("/globe/run", h.GlobeRun)
	app.Get("/reset-store", h.ResetStore)

	app.Get("/sms/inbox", h.GetInbox)
	app.Get("/sms/outbox", h.GetOutbox)
	app.Delete("/sms/inbox/:id", h.DeleteInboxSMS)
	app.Delete("/sms/outbox/:id", h.DeleteOutboxSMS)
	app.Get("/phones", h.GetPhones)
	app.Get("/phones-imsi", h.GetPhoneToIMSI)
	app.Patch("/phones-imsi/:id/phone", h.UpdatePhoneToIMSI)
	app.Post("/phones-imsi", h.AddPhoneToIMSI)
	app.Delete("/phones-imsi/:id", h.DeletePhoneToIMSI)

	return &Api{
		Fiber: app,
	}
}
