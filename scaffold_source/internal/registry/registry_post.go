package registry

import (
	"github.com/PROJECT_NAME/internal/domains/post"
)

func (r *Registry) PostRepository() post.Repository {
	if r.postRepository == nil {
		r.postRepository = post.NewRepository(r)
	}

	return r.postRepository
}

func (r *Registry) PostService() post.Service {
	if r.postService == nil {
		r.postService = post.NewService(r)
	}

	return r.postService
}
