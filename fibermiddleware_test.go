package glogger

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/valyala/fasthttp"
	"gotest.tools/assert"
)

func TestFiberLogMiddleware(t *testing.T) {
	app := fiber.New()
	c := app.AcquireCtx(&fasthttp.RequestCtx{})
	defer app.ReleaseCtx(c)

	log, hook := test.NewNullLogger()

	app.Use(RequestFiberMiddlewareLogger(log, nil))
	called := false
	app.Get("/test", func(c *fiber.Ctx) error {
		called = true
		return nil
	})

	req := httptest.NewRequest("GET", "/test", nil)
	_, err := app.Test(req)
	assert.NilError(t, err)

	assert.Equal(t, called, true)
	t.Log(hook.AllEntries()[0])
}
