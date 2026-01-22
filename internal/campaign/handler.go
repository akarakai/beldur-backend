package campaign

import (
	"beldur/internal/auth"
	"beldur/pkg/httperr"
	"log/slog"

	"github.com/gofiber/fiber/v2"
)

type HttpHandler struct {
	newCampaignUC *UseCase
	errManager    *httperr.Manager
}

func NewHttpHandler(newCampaignUC *UseCase) *HttpHandler {
	return &HttpHandler{
		newCampaignUC: newCampaignUC,
		errManager:    NewCampaignApiErrorManager(),
	}
}

func (h *HttpHandler) HandleCreateCampaign(c *fiber.Ctx) error {
	var req *CreationRequest
	if err := c.BodyParser(&req); err != nil {
		status, body := h.errManager.Map(err)
		return c.Status(status).JSON(body)
	}

	p, ok := auth.PrincipalFromCtx(c)
	if !ok {
		slog.Error("principal not found in context even if authentication is successful")
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	resp, err := h.newCampaignUC.CreateNewCampaign(c.Context(), *req, p.PlayerID)
	if err != nil {
		status, body := h.errManager.Map(err)
		return c.Status(status).JSON(body)
	}

	return c.Status(fiber.StatusCreated).JSON(resp)
}
