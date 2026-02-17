package browser

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/static"
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

func (s *Server) Run(addr string) error {
	s.engine.Get("/sse", func(c fiber.Ctx) error {
		c.Set("Content-Type", "text/event-stream")
		c.Set("Cache-Control", "no-cache")
		c.Set("Connection", "keep-alive")
		c.Set("Transfer-Encoding", "chunked")
		c.Status(fiber.StatusOK)

		return nil
	})
	s.engine.Get("/*", static.New("./web", static.Config{
		Browse: true,
	}))

	errCh := make(chan error, 1)
	go func() {
		if err := s.engine.Listen(addr); err != nil {
			errCh <- err
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-quit:
		fmt.Println("Shutting down server gracefully...")
		return s.engine.ShutdownWithTimeout(30 * time.Second)
	case err := <-errCh:
		return err
	}
}
