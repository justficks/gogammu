package api

import (
	"github.com/gofiber/fiber/v2"
	"log/slog"
	"time"
)

func LoggerMiddleware(log *slog.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Записываем время начала обработки запроса
		start := time.Now()

		// Передаем обработку следующему в цепочке middleware или обработчику маршрута
		err := c.Next()

		// Логируем детали запроса после его обработки
		log.Info("HTTP request",
			"method", c.Method(),
			"path", c.Path(),
			"status", c.Response().StatusCode(),
			"duration", time.Since(start).String(),
		)

		return err
	}
}
