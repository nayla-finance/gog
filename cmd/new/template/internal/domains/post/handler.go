package post

import (
	"github.com/gofiber/fiber/v2"
	"github.com/project-name/internal/config"
	"github.com/project-name/internal/domains/user"
	"github.com/project-name/internal/errors"
	"github.com/project-name/internal/logger"
)

type (
	HandlerProvider interface {
		PostHandler() *Handler
	}

	handlerDependencies interface {
		logger.LoggerProvider
		config.ConfigProvider
		user.ServiceProvider
		ServiceProvider
		errors.ErrorProvider
	}

	Handler struct {
		d handlerDependencies
	}
)

func NewHandler(d handlerDependencies) *Handler {
	return &Handler{d: d}
}

func (h *Handler) RegisterRoutes(r fiber.Router) {
	// specific middlewares for post domain
	r.Use(NewSpecificPostMiddleware(h.d).Handle)

	r.Post("/posts", h.CreatePost)
	r.Get("/posts/:id", h.GetPost)
	// ... other routes go here
}

// @Summary		Create a new post
// @Description	Create a new post with the provided data
// @Tags			posts
// @Accept			json
// @Produce		json
// @Param			post	body	post.CreatePostDTO	true	"Post data"
// @Success		201		"Created"
// @Failure		400		{object}	errors.ErrorResponse
// @Failure		500		{object}	errors.ErrorResponse
// @Router			/posts [post]
func (h *Handler) CreatePost(c *fiber.Ctx) error {
	dto := &CreatePostDTO{}

	if err := c.BodyParser(dto); err != nil {
		return h.d.NewError(errors.ErrInternal, err.Error())
	}

	if err := dto.Validate(); err != nil {
		return h.d.NewError(errors.ErrBadRequest, err.Error())
	}

	return h.d.PostService().CreatePost(c.Context(), dto)
}

// @Summary		Get a post by ID
// @Description	Get a post's details by its ID
// @Tags			posts
// @Accept			json
// @Produce		json
// @Param			id	path		string	true	"Post ID"
// @Success		200	{object}	Post
// @Failure		400	{object}	errors.ErrorResponse
// @Failure		404	{object}	errors.ErrorResponse
// @Failure		500	{object}	errors.ErrorResponse
// @Router			/posts/{id} [get]
func (h *Handler) GetPost(c *fiber.Ctx) error {
	id := c.Params("id")

	post, err := h.d.PostService().GetPostByID(c.Context(), id)
	if err != nil {
		return h.d.NewError(errors.ErrInternal, err.Error())
	}

	return c.JSON(post)
}
