package handlers

import (
	"github.com/gofiber/fiber/v3"
	"github.com/sirini/goapi/internal/services"
	"github.com/sirini/goapi/pkg/models"
	"github.com/sirini/goapi/pkg/utils"
)

type PushHandler interface {
	RegisterDeviceHandler(c fiber.Ctx) error
	UnregisterDeviceHandler(c fiber.Ctx) error
}

type NuboPushHandler struct {
	service *services.Service
}

func NewNuboPushHandler(service *services.Service) *NuboPushHandler {
	return &NuboPushHandler{service: service}
}

func (h *NuboPushHandler) RegisterDeviceHandler(c fiber.Ctx) error {
	param := models.PushDeviceParam{}
	if err := c.Bind().Body(&param); err != nil {
		return utils.Err(c, "invalid push device", models.CODE_INVALID_PARAMETER)
	}
	userUid := utils.ExtractUserUid(c.Get(models.AUTH_KEY))
	if err := h.service.Push.RegisterDevice(uint(userUid), param); err != nil {
		return utils.Err(c, err.Error(), models.CODE_INVALID_PARAMETER)
	}
	return utils.Ok(c, nil)
}

func (h *NuboPushHandler) UnregisterDeviceHandler(c fiber.Ctx) error {
	param := models.PushDeviceParam{}
	if err := c.Bind().Body(&param); err != nil {
		return utils.Err(c, "invalid push device", models.CODE_INVALID_PARAMETER)
	}
	userUid := utils.ExtractUserUid(c.Get(models.AUTH_KEY))
	if err := h.service.Push.UnregisterDevice(uint(userUid), param); err != nil {
		return utils.Err(c, err.Error(), models.CODE_INVALID_PARAMETER)
	}
	return utils.Ok(c, nil)
}
