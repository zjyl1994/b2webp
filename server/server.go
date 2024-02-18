package server

import (
	"errors"
	"net/http"

	"github.com/Masterminds/sprig/v3"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/favicon"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"github.com/gofiber/template/html/v2"
	"github.com/inhies/go-bytesize"
	"github.com/zjyl1994/b2webp/assets"
	"github.com/zjyl1994/b2webp/common/vars"
)

func Run(listen string) error {
	engine := html.NewFileSystem(assets.GetFS("template"), ".html")
	engine.Reload(vars.DebugMode)
	engine.AddFuncMap(sprig.FuncMap())

	app := fiber.New(fiber.Config{
		AppName:               vars.APP_NAME,
		ServerHeader:          vars.APP_NAME,
		DisableStartupMessage: true,
		Views:                 engine,
		ViewsLayout:           "layout",
		PassLocalsToViews:     true,
		BodyLimit:             int(5 * bytesize.MB),
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			var code int
			var errorMessage string
			var e *fiber.Error
			if errors.As(err, &e) {
				code = e.Code
				errorMessage = e.Message
			} else {
				code = fiber.StatusInternalServerError
				errorMessage = err.Error()
			}
			return c.Render("error", fiber.Map{
				"status_code":   code,
				"status_text":   http.StatusText(code),
				"error_message": errorMessage,
				"site_name":     vars.Getenv("B2WEBP_SITE_NAME"),
			})
		},
	})

	app.Use(favicon.New(favicon.Config{Data: assets.Favicon}))
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("site_name", vars.Getenv("B2WEBP_SITE_NAME"))
		c.Locals("total_count", vars.TotalImageCount)
		c.Locals("total_size", vars.TotalImageSize)
		c.Locals("need_password", len(vars.UploadPassword) > 0)
		return c.Next()
	})

	app.Use("/static", filesystem.New(filesystem.Config{Root: assets.GetFS("static")}))

	app.Get("/test", testHandler)

	app.Get("/", UploadImagePage)
	app.Get("/upload", UploadImagePage)
	app.Post("/upload", UploadImageHandler)
	app.Get("/delete/:hashid/:delcode", DeleteImageHandler)

	app.Get("/:hashid", GetImageHandler)
	app.Get("/info/:hashid", GetImagePage)

	return app.Listen(listen)
}

func testHandler(c *fiber.Ctx) error {
	return fiber.ErrNotImplemented
}
