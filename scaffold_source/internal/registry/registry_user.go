package registry

import "github.com/PROJECT_NAME/internal/domains/user"

func (r *Registry) UserRepository() user.Repository {
	if r.userRepository == nil {
		r.userRepository = user.NewRepository(r)
	}

	return r.userRepository
}

func (r *Registry) UserService() user.Service {
	if r.userService == nil {
		r.userService = user.NewService(r)
	}

	return r.userService
}
