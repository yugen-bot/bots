package handlers

import (
	"github.com/disgoorg/snowflake/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/jurienhamaker/disgoplus"
	"github.com/sarulabs/di/v2"

	"jurien.dev/yugen/shared/static"
)

type UserResponse struct {
	ID        string `json:"id"`
	Username  string `json:"username"`
	AvatarURL string `json:"avatarUrl"`
}

type UserHandler struct {
	container *di.Container
}

func GetUserHandler(container *di.Container) *UserHandler {
	return &UserHandler{container: container}
}

func (h *UserHandler) AddRoutes(_ *fiber.App, router fiber.Router) {
	router.Get("/users/:userID", h.handleGetUser)
}

func (h *UserHandler) handleGetUser(c *fiber.Ctx) error {
	userID := c.Params("userID")

	id, err := snowflake.Parse(userID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Invalid user ID")
	}

	disgoBot := h.container.Get(static.DiBot).(*disgoplus.Bot)

	user, err := disgoBot.Client().Rest.GetUser(id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).SendString("User not found")
	}

	return c.JSON(UserResponse{
		ID:        user.ID.String(),
		Username:  user.Username,
		AvatarURL: user.EffectiveAvatarURL(),
	})
}
