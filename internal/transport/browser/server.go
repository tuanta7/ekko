package browser

import (
	"bufio"
	"context"
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
	engine         *fiber.App
	handler        *handler.Handler
	ctx            context.Context
	cancel         context.CancelFunc
	transcriptChan chan string
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
		if s.ctx == nil || s.transcriptChan == nil {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"error": "no active session",
			})
		}

		c.Set("Content-Type", "text/event-stream")
		c.Set("Cache-Control", "no-cache")
		c.Set("Connection", "keep-alive")
		c.Set("Transfer-Encoding", "chunked")

		c.Status(fiber.StatusOK).RequestCtx().SetBodyStreamWriter(func(w *bufio.Writer) {
			_, _ = fmt.Fprintf(w, "data: {\"type\":\"connected\"}\n\n")
			_ = w.Flush()

			for {
				select {
				case <-s.ctx.Done():
					_, _ = fmt.Fprintf(w, "data: {\"type\":\"ended\"}\n\n")
					_ = w.Flush()
					return
				case text, ok := <-s.transcriptChan:
					if !ok { // Channel closed, session ended
						_, _ = fmt.Fprintf(w, "data: {\"type\":\"ended\"}\n\n")
						_ = w.Flush()
						return
					}

					_, _ = fmt.Fprintf(w, "data: {\"text\":\"%s\"}\n\n", text)
					_ = w.Flush()
				}
			}
		})

		return nil
	})

	s.engine.Get("/sources", func(c fiber.Ctx) error {
		sources, err := s.handler.ListSources(c.Context())
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		return c.JSON(sources)
	})

	s.engine.Post("/start", func(c fiber.Ctx) error {
		if s.ctx != nil || s.transcriptChan != nil {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"error": "session already active",
			})
		}

		var req struct {
			Source   string `json:"source" validate:"required"`
			Duration int    `json:"duration" validate:"gte=1,lte=300"`
		}
		if err := c.Bind().JSON(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "invalid request body",
				"cause": err.Error(),
			})
		}

		ctx, cancel := context.WithCancel(context.Background())
		s.ctx = ctx
		s.cancel = cancel
		s.transcriptChan = make(chan string, 100)

		go func() {
			duration := time.Duration(req.Duration) * time.Second
			_ = s.handler.StartRecord(ctx, duration, req.Source)
		}()

		go s.handler.StartTranscribe(ctx)

		go s.handler.CollectResults(ctx, func(text string) {
			select {
			case s.transcriptChan <- text:
			case <-ctx.Done():
			}
		})

		return c.JSON(fiber.Map{"status": "started"})
	})

	s.engine.Post("/stop", func(c fiber.Ctx) error {
		if s.cancel == nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "no active session",
			})
		}

		s.cancel()

		s.cancel = nil
		s.ctx = nil

		if s.transcriptChan != nil {
			close(s.transcriptChan)
			s.transcriptChan = nil
		}

		return c.JSON(fiber.Map{"status": "stopped"})
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
