package webpresentation

import (
	"context"
	"strings"
	"ws-lan-chat/pkg/chatservice"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	recover2 "github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/teadove/teasutils/fiber_utils"
	"github.com/teadove/teasutils/service_utils/logger_utils"
)

type Presentation struct {
	chatService *chatservice.Service
}

func NewPresentation(chatService *chatservice.Service) *Presentation {
	return &Presentation{chatService: chatService}
}

func (r *Presentation) BuildApp() *fiber.App {
	app := fiber.New(fiber.Config{
		Immutable:    true,
		ErrorHandler: fiber_utils.ErrHandler(),
	})

	app.Use(recover2.New(recover2.Config{EnableStackTrace: true}))
	app.Use(fiber_utils.MiddlewareLogger())
	app.Use(cors.New(cors.ConfigDefault))

	app.Use("/ws", func(c *fiber.Ctx) error {
		// IsWebSocketUpgrade returns true if the client
		// requested upgrade to the WebSocket protocol.
		if websocket.IsWebSocketUpgrade(c) {
			c.Locals("allowed", true)
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})
	app.Get("/ws", websocket.New(func(c *websocket.Conn) {
		ctx := logger_utils.NewLoggedCtx()
		ctx = logger_utils.WithValue(ctx, "ip", c.IP())
		ctx = logger_utils.WithValue(ctx, "app_method", "/ws")
		ctx = logger_utils.WithValue(ctx, "user_agent", strings.Clone(c.Headers(fiber.HeaderUserAgent)))

		user := c.Headers("User", "anonimus")
		ctx = logger_utils.WithValue(ctx, "user", user)

		connID := uuid.New()
		ctx = logger_utils.WithValue(ctx, "conn_id", connID.String())

		zerolog.Ctx(ctx).Info().
			Msg("new.ws.connection")

		ctx, cancel := context.WithCancel(ctx)
		defer cancel()

		type ErrorMsg struct {
			Err string `json:"err"`
		}

		go func() {
			for {
				var msgRequest chatservice.MsgRequest
				err := c.ReadJSON(&msgRequest)
				if err != nil {
					zerolog.Ctx(ctx).Warn().Err(err).Msg("failed.to.read.message.from.client")
					innerErr := c.WriteJSON(ErrorMsg{err.Error()})
					if innerErr != nil {
						return
					}

					continue
				}

				err = r.chatService.SaveMessage(ctx, user, connID, &msgRequest)
				if err != nil {
					zerolog.Ctx(ctx).Error().Stack().Err(err).Msg("failed.to.save.message.from.client")
				}
			}
		}()

		defer r.chatService.Unsubscribe(connID)
		msgs, err := r.chatService.Subscribe(ctx, connID)
		if err != nil {
			_ = c.WriteJSON(ErrorMsg{err.Error()})
			return
		}

		for msg := range msgs {
			err = c.WriteJSON(msg)
			if err != nil {
				zerolog.Ctx(ctx).Error().Stack().Err(err).Msg("failed.to.write.msg.to.client")
				return
			}
		}
	}))

	return app
}
