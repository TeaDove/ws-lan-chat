package webpresentation

import (
	"context"
	"github.com/pkg/errors"
	"strings"
	"sync"
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

		user := c.Query("user", "anonimus")
		ctx = logger_utils.WithValue(ctx, "user", user)

		connID := uuid.New()
		ctx = logger_utils.WithValue(ctx, "conn_id", connID.String())

		zerolog.Ctx(ctx).Info().
			Msg("new.ws.connection")

		ctx, cancel := context.WithCancel(ctx)
		defer cancel()

		var wg sync.WaitGroup
		wg.Go(func() {
			err := r.publish(ctx, c, user, connID)
			if err != nil {
				closeOnErr(ctx, c, err)
			}
		})
		wg.Go(func() {
			err := r.subscribe(ctx, c, connID)
			if err != nil {
				closeOnErr(ctx, c, err)
			}
		})

		wg.Wait()
	}))

	return app
}

func closeOnErr(ctx context.Context, c *websocket.Conn, err error) {
	zerolog.Ctx(ctx).
		Error().
		Err(err).
		Msg("failed.to.process.ws")
	err = c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseInternalServerErr, err.Error()))
	if err != nil {
		zerolog.Ctx(ctx).
			Error().
			Err(err).
			Msg("failed.to.close")
	}
}

func (r *Presentation) publish(ctx context.Context, c *websocket.Conn, user string, connID uuid.UUID) error {
	for {
		var msgRequest chatservice.MsgRequest
		err := c.ReadJSON(&msgRequest)
		if err != nil {
			return errors.Wrap(err, "read json")
		}

		err = r.chatService.SaveMessage(ctx, user, connID, &msgRequest)
		if err != nil {
			return errors.Wrap(err, "save message")
		}
	}
}

func (r *Presentation) subscribe(ctx context.Context, c *websocket.Conn, connID uuid.UUID) error {
	defer r.chatService.Unsubscribe(connID)
	msgs, err := r.chatService.Subscribe(ctx, connID)
	if err != nil {
		return errors.Wrap(err, "subscribe")
	}

	for msg := range msgs {
		err = c.WriteJSON(msg)
		if err != nil {
			return errors.Wrap(err, "write json to client")
		}
	}

	return nil
}
