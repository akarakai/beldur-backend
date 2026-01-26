package campaign

import (
	"beldur/internal/id"
	"beldur/pkg/httperr"
	"beldur/pkg/middleware"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type HttpHandler struct {
	campaignUC *UseCase
	errManager *httperr.Manager
}

func NewHttpHandler(campaignUC *UseCase) *HttpHandler {
	return &HttpHandler{
		campaignUC: campaignUC,
		errManager: NewCampaignApiErrorManager(),
	}
}

func (h *HttpHandler) HandleCreateCampaign(c *fiber.Ctx) error {
	req := c.Locals("body").(CreationRequest)

	p, ok := middleware.PrincipalFromCtx(c)
	if !ok {
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	resp, err := h.campaignUC.CreateNewCampaign(c.Context(), req, p.PlayerID)
	if err != nil {
		status, body := h.errManager.Map(err)
		return c.Status(status).JSON(body)
	}

	return c.Status(fiber.StatusCreated).JSON(resp)
}

func (h *HttpHandler) HandleJoinCampaign(c *fiber.Ctx) error {
	req := c.Locals("body").(JoinRequest)

	campaignInstr := c.Params("campaignId")
	if campaignInstr == "" {
		panic("wrong parameter naming")
	}
	campaignId, err := strconv.Atoi(campaignInstr)
	if err != nil {
		return c.SendStatus(fiber.StatusBadRequest)
	}

	p, ok := middleware.PrincipalFromCtx(c)
	if !ok {
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	resp, err := h.campaignUC.JoinCampaign(c.Context(), req, id.CampaignId(campaignId), p.PlayerID)
	if err != nil {
		status, body := h.errManager.Map(err)
		return c.Status(status).JSON(body)
	}
	return c.Status(fiber.StatusCreated).JSON(resp)
}

// HandleGetCampaign require no authentication
func (h *HttpHandler) HandleGetCampaign(c *fiber.Ctx) error {
	resp, err := h.campaignUC.SearchCampaign(c.Context())
	if err != nil {
		status, body := h.errManager.Map(err)
		return c.Status(status).JSON(body)
	}
	return c.Status(fiber.StatusOK).JSON(resp)
}
