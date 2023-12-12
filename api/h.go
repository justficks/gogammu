package api

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	gammu "github.com/justficks/gogammu"
	"log/slog"
	"time"
)

type Handler struct {
	Gammu *gammu.Gammu
	log   *slog.Logger
}

func (h *Handler) GlobeRun(c *fiber.Ctx) error {
	err := h.Gammu.GlobeRun()
	if err != nil {
		return c.Status(500).SendString(err.Error())
	}
	return c.SendStatus(200)
}

func (h *Handler) ResetStore(c *fiber.Ctx) error {
	h.Gammu.Store.Clear()
	return c.SendStatus(200)
}

func (h *Handler) RunOnError(c *fiber.Ctx) error {
	const op = "RunOnError"
	log := h.log.With(
		slog.String("op", op),
	)

	body := string(c.Body()) // "123456789012345 INIT"

	log.Info("Income body", body)

	notify, err := gammu.ParseRunOnErrBody(body)
	if err != nil {
		log.Error("Parse body error", err)
		return c.Status(400).SendString(err.Error())
	}

	modem, isExist := h.Gammu.Store.GetModemByIMSI(notify.PhoneID)
	if !isExist {
		log.Error("Modem not found", notify.PhoneID)
		return c.Status(400).SendString(fmt.Sprintf("Modem %s not found", notify.PhoneID))
	}

	err = h.Gammu.Stop(modem.IMSI)
	if err != nil {
		log.Error("Stop gammu-smsd error", err)
		return c.Status(400).SendString(err.Error())
	}

	err = h.Gammu.Run(modem)
	if err != nil {
		log.Error("Run gammu-smsd error", err)
		return c.Status(400).SendString(err.Error())
	}

	return c.SendStatus(200)
}

func (h *Handler) RunOnMessage(c *fiber.Ctx) error {
	const op = "RunOnMessage"
	log := h.log.With(
		slog.String("op", op),
	)

	body := string(c.Body()) // "123456789012345 msgId1 msgId2 ... msgIdN"

	log.Info("Income body", body)

	_, err := gammu.ParseRunOnMsgBody(body)
	if err != nil {
		log.Error("Parse body error", err)
		return c.Status(400).SendString(err.Error())
	}

	//newMsg, err := h.Gammu.ConcatSMS(notify)
	//if err != nil {
	//	return fiber.NewError(400, err.Error())
	//}

	err = h.Gammu.OnMsgCallback(&gammu.NewMsg{Phone: "201290129", Text: "test test", From: "some gay", Date: time.Now()})
	if err != nil {
		return fiber.NewError(400, err.Error())
	}

	return c.SendStatus(200)
}
