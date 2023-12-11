package api

import (
	"github.com/gofiber/fiber/v2"
	gammu "github.com/justficks/gogammu"
)

type Handler struct {
	Gammu *gammu.Gammu
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
