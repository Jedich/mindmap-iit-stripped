package router

import (
	"fmt"
	"github.com/ansrivas/fiberprometheus/v2"
	"github.com/fluent/fluent-logger-golang/fluent"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	jwtware "github.com/gofiber/jwt/v3"
	"mindmap-go/utils"
)

type Router struct {
	App        fiber.Router
	UserRouter *UserRouter
	MapRouter  *MapRouter
	CardRouter *CardRouter
}

func NewRouter(fiber *fiber.App, userRouter *UserRouter, mapRouter *MapRouter, cardRouter *CardRouter) *Router {
	return &Router{
		App:        fiber,
		UserRouter: userRouter,
		MapRouter:  mapRouter,
		CardRouter: cardRouter,
	}
}

// Register routes
func (r *Router) Register() {
	r.App.Use(logger.New())
	log := NewFluentLogger()
	r.App.Use(fluentdMiddleware(log))

	prometheus := fiberprometheus.New("mindmap")
	prometheus.RegisterAt(r.App, "/metrics")
	r.App.Use(prometheus.Middleware)

	// Routes, unrestricted access
	r.App.Get("/ping", func(c *fiber.Ctx) error {
		return c.SendString("Pong! 👋")
	})
	r.App.Static("/img", "./resources")

	// Register auth routes
	r.UserRouter.RegisterAuthRoutes()

	// JWT Middleware
	r.App.Use(jwtware.New(jwtware.Config{
		SigningKey: []byte(utils.ReadEnv("JWT_SECRET")),
		ErrorHandler: func(ctx *fiber.Ctx, err error) error {
			return &utils.UnauthorizedEntryError{Message: "No secret key provided"}
		},
	}))

	// Register routes of modules, restricted access
	r.UserRouter.RegisterUserRoutes()
	r.MapRouter.RegisterMapRoutes()
	r.CardRouter.RegisterCardRoutes()
}

func fluentdMiddleware(logger *fluent.Fluent) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Log request details
		logger.Post("http_request", map[string]string{
			"method": c.Method(),
			"path":   c.Path(),
			"ip":     c.IP(),
		})

		// Continue processing the request
		return c.Next()
	}
}

func NewFluentLogger() *fluent.Fluent {
	logger, err := fluent.New(fluent.Config{
		FluentHost: "fluentd", // Fluentd server host
		FluentPort: 24224,     // Fluentd server port
		TagPrefix:  "mindmap", // Fluentd tag prefix
		Async:      false,     // Whether to send logs asynchronously (true) or synchronously (false)
	})
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("Connected to fluentd")
	return logger
}
