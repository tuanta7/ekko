package browser

import (
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/tuanta7/ekko/internal/handler"
)

type Server struct {
	engine  *fiber.App
	handler *handler.Handler
}

func NewServer(handler *handler.Handler) *Server {
	return &Server{
		engine: fiber.New(fiber.Config{
			ReadTimeout: 10 * time.Second,
		}),
		handler: handler,
	}
}

func (s *Server) Run(addr string) {
	go func() {
		err := s.engine.Listen(addr)
		if err != nil {
			panic(err)
		}
	}()
}

func (s *Server) Close(timeout time.Duration) error {
	return s.engine.ShutdownWithTimeout(timeout)
}
