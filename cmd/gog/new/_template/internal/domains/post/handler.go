package post

import (
	"github.com/PROJECT_NAME/internal/config"
	"github.com/PROJECT_NAME/internal/domains/interfaces"
	"github.com/PROJECT_NAME/internal/domains/model"
	"github.com/PROJECT_NAME/internal/errors"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/nayla-finance/go-nayla/logger"
)

type (
	handlerDependencies interface {
		logger.Provider
		config.ConfigProvider
		interfaces.PostServiceProvider
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
// @Param			post	body	model.CreatePostDTO	true	"Post data"
// @Success		201		"Created"
// @Failure		400		{object}	errors.ErrorResponse
// @Failure		500		{object}	errors.ErrorResponse
// @Router			/posts [post]
func (h *Handler) CreatePost(c *fiber.Ctx) error {
	dto := &model.CreatePostDTO{}

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
// @Success		200	{object}	model.Post
// @Failure		400	{object}	errors.ErrorResponse
// @Failure		404	{object}	errors.ErrorResponse
// @Failure		500	{object}	errors.ErrorResponse
// @Router			/posts/{id} [get]
func (h *Handler) GetPost(c *fiber.Ctx) error {
	id := c.Params("id")

	post := &model.Post{}
	if err := h.d.PostService().GetPostByID(c.Context(), uuid.MustParse(id), post); err != nil {
		return h.d.NewError(errors.ErrInternal, err.Error())
	}

	return c.JSON(post)
}
