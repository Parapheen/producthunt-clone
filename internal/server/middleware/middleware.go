package middleware

type Middleware struct {
	UserService UserService
}

func NewMiddleware(userService UserService) *Middleware {
	return &Middleware{
		UserService: userService,
	}
}
