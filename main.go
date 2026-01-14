package main

import (
	"ws-lan-chat/pkg/chatservice"
	"ws-lan-chat/pkg/msgrepo"
	"ws-lan-chat/pkg/settings"
	"ws-lan-chat/pkg/webpresentation"

	"github.com/gofiber/fiber/v2"
	"github.com/pkg/errors"
	"github.com/teadove/teasutils/service_utils/db_utils"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

func build() (*fiber.App, error) {
	db, err := gorm.Open(sqlite.Open(settings.Settings.DB),
		&gorm.Config{
			NamingStrategy: schema.NamingStrategy{SingularTable: true},
			TranslateError: true,
			Logger:         db_utils.ZerologAdapter{},
		},
	)
	if err != nil {
		return nil, errors.Wrap(err, "open gorm db")
	}

	err = db.AutoMigrate(new(msgrepo.Message))
	if err != nil {
		return nil, errors.Wrap(err, "auto migrate")
	}

	msgRepo := msgrepo.New(db)

	chatService := chatservice.NewService(msgRepo)

	presentation := webpresentation.NewPresentation(chatService)

	return presentation.BuildApp(), nil
}

func main() {
	app, err := build()
	if err != nil {
		panic(err)
	}

	err = app.Listen(":8080")
	if err != nil {
		panic(err)
	}
}
