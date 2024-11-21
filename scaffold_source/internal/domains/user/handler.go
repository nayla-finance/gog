package user

import (
	"github.com/PROJECT_NAME/internal/config"
	"github.com/PROJECT_NAME/internal/errors"
	"github.com/PROJECT_NAME/internal/logger"
	"github.com/gofiber/fiber/v2"
)

type (
	handlerDependencies interface {
		config.ConfigProvider
		logger.LoggerProvider
		ServiceProvider
		errors.ErrorProvider
	}

	Handler struct {
		d handlerDependencies
	}
)

func NewHandler(d handlerDependencies) *Handler {
	return &Handler{
		d: d,
	}
}

func (h *Handler) RegisterRoutes(api fiber.Router) {
	api.Get("/users", h.getUsers)
	api.Post("/users", h.createUser)
	api.Get("/users/:id", h.getUser)
	api.Put("/users/:id", h.updateUser)
	api.Delete("/users/:id", h.deleteUser)
}

// @Summary		Get all users
// @Description	Get a list of all users
// @Tags			users
// @Accept			json
// @Produce		json
// @Success		200	{array}		User
// @Failure		500	{object}	errors.ErrorResponse
// @Router			/users [get]
func (h *Handler) getUsers(c *fiber.Ctx) error {
	users, err := h.d.UserService().GetUsers(c.Context())
	if err != nil {
		return h.d.NewError(errors.ErrInternal, err.Error())
	}

	return c.JSON(users)
}

// @Summary		Create a new user
// @Description	Create a new user with the provided data
// @Tags			users
// @Accept			json
// @Produce		json
// @Param			user	body	CreateUserDTO	true	"User data"
// @Success		201		"Created"
// @Failure		400		{object}	errors.ErrorResponse
// @Failure		500		{object}	errors.ErrorResponse
// @Router			/users [post]
func (h *Handler) createUser(c *fiber.Ctx) error {
	dto := &CreateUserDTO{}
	if err := c.BodyParser(dto); err != nil {
		return h.d.NewError(errors.ErrInternal, err.Error())
	}

	if err := dto.Validate(); err != nil {
		return h.d.NewError(errors.ErrBadRequest, err.Error())
	}

	if err := h.d.UserService().CreateUser(c.Context(), dto); err != nil {
		return h.d.NewError(errors.ErrInternal, err.Error())
	}

	return c.SendStatus(fiber.StatusCreated)
}

// @Summary		Get a user by ID
// @Description	Get a user's details by their ID
// @Tags			users
// @Accept			json
// @Produce		json
// @Param			id	path		string	true	"User ID"
// @Success		200	{object}	User
// @Failure		400	{object}	errors.ErrorResponse
// @Failure		404	{object}	errors.ErrorResponse
// @Failure		500	{object}	errors.ErrorResponse
// @Router			/users/{id} [get]
func (h *Handler) getUser(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return h.d.NewError(errors.ErrBadRequest, "missing user id")
	}

	user := &User{}
	if err := h.d.UserService().GetUserByID(c.Context(), id, user); err != nil {
		return h.d.NewError(errors.ErrInternal, err.Error())
	}

	if user == nil {
		return h.d.NewError(errors.ErrResourceNotFound, "user not found")
	}

	return c.JSON(user)
}

// @Summary		Update a user
// @Description	Update a user's details by their ID
// @Tags			users
// @Accept			json
// @Produce		json
// @Param			id		path	string			true	"User ID"
// @Param			user	body	UpdateUserDTO	true	"User data"
// @Success		204		"No Content"
// @Failure		400		{object}	errors.ErrorResponse
// @Failure		404		{object}	errors.ErrorResponse
// @Failure		500		{object}	errors.ErrorResponse
// @Router			/users/{id} [put]
func (h *Handler) updateUser(c *fiber.Ctx) error {
	// parser body

	// validate

	// ...

	return c.SendStatus(fiber.StatusNoContent)
}

// @Summary		Delete a user
// @Description	Delete a user by their ID
// @Tags			users
// @Accept			json
// @Produce		json
// @Param			id	path	string	true	"User ID"
// @Success		204	"No Content"
// @Failure		400	{object}	errors.ErrorResponse
// @Failure		404	{object}	errors.ErrorResponse
// @Failure		500	{object}	errors.ErrorResponse
// @Router			/users/{id} [delete]
func (h *Handler) deleteUser(c *fiber.Ctx) error {
	// delete user

	return c.SendStatus(fiber.StatusNoContent)
}
