package api

import (
	"github.com/gofiber/fiber/v2"
	gammu "github.com/justficks/gogammu"
	"strconv"
)

type PaginatedInboxResponse struct {
	Items        []gammu.Inbox `json:"items"`
	TotalRecords int           `json:"totalRecords"`
	TotalPages   int           `json:"totalPages"`
	CurrentPage  int           `json:"currentPage"`
	PageSize     int           `json:"pageSize"`
}

var PageSize = 30

func (h *Handler) GetInbox(c *fiber.Ctx) error {
	page := c.QueryInt("page", 1)
	pageSize := c.QueryInt("pageSize", PageSize)

	items, err := h.Gammu.GetInbox(page, pageSize)
	if err != nil {
		return c.Status(500).SendString(err.Error())
	}

	totalRecords, err := h.Gammu.GetInboxCount()
	if err != nil {
		return c.Status(500).SendString(err.Error())
	}

	totalPages := (totalRecords + pageSize - 1) / pageSize

	return c.JSON(PaginatedInboxResponse{
		Items:        items,
		TotalRecords: totalRecords,
		TotalPages:   totalPages,
		CurrentPage:  page,
		PageSize:     PageSize,
	})
}

func (h *Handler) DeleteInboxSMS(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return fiber.NewError(400, "Parameter :id must be a number.")
	}
	err = h.Gammu.DeleteInbox(id)
	if err != nil {
		return c.Status(500).SendString(err.Error())
	}
	return c.SendStatus(200)
}

func (h *Handler) GetOutbox(c *fiber.Ctx) error {
	items, err := h.Gammu.GetOutbox()
	if err != nil {
		return c.Status(500).SendString(err.Error())
	}
	return c.JSON(items)
}

func (h *Handler) DeleteOutboxSMS(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return fiber.NewError(400, "Parameter :id must be a number.")
	}
	err = h.Gammu.DeleteOutbox(id)
	if err != nil {
		return c.Status(500).SendString(err.Error())
	}
	return c.SendStatus(200)
}

func (h *Handler) GetPhones(c *fiber.Ctx) error {
	items, err := h.Gammu.GetPhones()
	if err != nil {
		return c.Status(500).SendString(err.Error())
	}
	return c.JSON(items)
}

func (h *Handler) GetPhoneToIMSI(c *fiber.Ctx) error {
	items, err := h.Gammu.GetPhonesIMSI()
	if err != nil {
		return c.Status(500).SendString(err.Error())
	}
	return c.JSON(items)
}

func (h *Handler) AddPhoneToIMSI(c *fiber.Ctx) error {
	var input gammu.PhonesIMSI
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).SendString("Failed to parse request body")
	}
	err := h.Gammu.AddPhoneIMSI(input)
	if err != nil {
		return c.Status(500).SendString(err.Error())
	}
	return c.SendStatus(201)
}

func (h *Handler) UpdatePhoneToIMSI(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return fiber.NewError(400, "Parameter :id must be a number.")
	}
	newPhone := string(c.Body())
	err = h.Gammu.UpdatePhoneIMSI(id, newPhone)
	if err != nil {
		return c.Status(500).SendString(err.Error())
	}
	return c.SendStatus(200)
}

func (h *Handler) DeletePhoneToIMSI(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return fiber.NewError(400, "Parameter :id must be a number.")
	}
	err = h.Gammu.DeletePhoneIMSI(id)
	if err != nil {
		return c.Status(500).SendString(err.Error())
	}
	return c.SendStatus(200)
}
