package account

import (
	"beldur/internal/auth"
	"beldur/pkg/httperr"
	"errors"
	"time"

	"github.com/gofiber/fiber/v2"
)

type HttpHandler struct {
	registrationUC *Registration
	loginUC        *UsernamePasswordLogin
	manageUC       *Management
	errManager     *httperr.Manager
}

func NewHttpHandler(registrationUC *Registration, loginUC *UsernamePasswordLogin, accountManagement *Management) *HttpHandler {
	return &HttpHandler{
		registrationUC: registrationUC,
		loginUC:        loginUC,
		manageUC:       accountManagement,
		errManager:     NewAccountApiErrorManager(),
	}
}

// Register creates a new account + new player as a side effect
func (h *HttpHandler) Register(c *fiber.Ctx) error {
	var req CreateAccountRequest
	if err := c.BodyParser(&req); err != nil {
		status, body := h.errManager.Map(err)
		return c.Status(status).JSON(body)
	}

	response, token, err := h.registrationUC.RegisterAccount(c.Context(), req)
	if err != nil {
		status, body := h.errManager.Map(err)
		return c.Status(status).JSON(body)
	}

	if err := h.attachTokenToCookie(c, token); err != nil {

	}

	return c.Status(fiber.StatusCreated).JSON(response)
}

// Login logins the user and gives a jwtauth (httperr only) cookie
func (h *HttpHandler) Login(c *fiber.Ctx) error {
	var req UsernamePasswordLoginRequest
	if err := c.BodyParser(&req); err != nil {
		status, body := h.errManager.Map(err)
		return c.Status(status).JSON(body)
	}

	jwt, err := h.loginUC.Login(c.Context(), req)
	if err != nil {
		status, body := h.errManager.Map(err)
		return c.Status(status).JSON(body)
	}

	if err := h.attachTokenToCookie(c, jwt); err != nil {
		status, body := h.errManager.Map(err)
		return c.Status(status).JSON(body)
	}

	return c.SendStatus(fiber.StatusOK)
}

func (h *HttpHandler) UpdateAccount(c *fiber.Ctx) error {
	var req UpdateAccountRequest
	if err := c.BodyParser(&req); err != nil {
		return c.SendStatus(fiber.StatusBadRequest)
	}
	p, ok := auth.PrincipalFromCtx(c)
	if !ok {
		return c.SendStatus(fiber.StatusUnauthorized)
	}
	resp, err := h.manageUC.UpdateAccount(c.Context(), req, p.AccountID)
	if err != nil {
		status, body := h.errManager.Map(err)
		return c.Status(status).JSON(body)
	}
	return c.Status(fiber.StatusOK).JSON(resp)
}

func (h *HttpHandler) attachTokenToCookie(c *fiber.Ctx, token string) error {
	if token == "" {
		return errors.New("empty token")
	}

	// TODO CHANGE
	exp := time.Now().Add(24 * time.Hour)

	c.Cookie(&fiber.Cookie{
		Name:     "jwt",
		Value:    token,
		Path:     "/",
		HTTPOnly: true,
		Secure:   false,
		SameSite: "Lax",
		Expires:  exp,
	})

	return nil
}
