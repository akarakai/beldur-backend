package character

import (
	"beldur/internal/id"
	"beldur/pkg/httperr"
	"beldur/pkg/middleware"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type HttpHandler struct {
	createUC   *CreateUseCase
	errManager *httperr.Manager
}

func NewHttpHandler(createUC *CreateUseCase) *HttpHandler {
	return &HttpHandler{
		createUC:   createUC,
		errManager: NewCharacterApiErrorManager(),
	}
}

func (h *HttpHandler) HandleNpcCreation(c *fiber.Ctx) error {
	campaignInstr := c.Params("campaignId")
	if campaignInstr == "" {
		panic("wrong parameter naming")
	}
	campId, err := strconv.Atoi(campaignInstr)
	if err != nil {
		return c.SendStatus(fiber.StatusBadRequest)
	}

	req := c.Locals("body").(CreateCharacterRequest)

	p, ok := middleware.PrincipalFromCtx(c)
	if !ok {
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	resp, err := h.createUC.CreateNPC(c.Context(), req, id.CampaignId(campId), p.PlayerID)
	if err != nil {
		status, body := h.errManager.Map(err)
		return c.Status(status).JSON(body)
	}
	return c.Status(fiber.StatusCreated).JSON(resp)
}
