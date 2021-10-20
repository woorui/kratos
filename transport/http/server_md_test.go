package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/internal/host"
	"github.com/go-kratos/kratos/v2/metadata"
	mmd "github.com/go-kratos/kratos/v2/middleware/metadata"
	"github.com/stretchr/testify/assert"
)

var testmd = map[string]string{"abc": "abc"}

func TestAppMd(t *testing.T) {
	hs := NewServer(
		Middleware(mmd.Server()), // use metadata middleware
	)
	route := hs.Route("/v1")
	route.GET("/metadata", func(ctx Context) error {
		md, _ := metadata.FromServerContext(ctx)
		appInfo, _ := kratos.FromContext(ctx)
		return ctx.Result(200, map[string]interface{}{
			"md":    md,
			"appmd": appInfo.Metadata(),
		})
	})

	// inject testmd and hs
	app := kratos.New(kratos.Server(hs), kratos.Metadata(testmd))

	go func() {
		if err := app.Run(); err != nil {
			panic(err)
		}
	}()

	time.Sleep(time.Second)
	testAppMd(t, hs)
	_ = app.Stop()
}

func testAppMd(t *testing.T, hs *Server) {
	port, ok := host.Port(hs.lis)
	if !ok {
		t.Fatalf("extract port error: %v", hs.lis)
	}
	base := fmt.Sprintf("http://127.0.0.1:%d/v1", port)

	resp, err := http.Get(base + "/metadata")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	var data map[string]map[string]string
	if err = json.NewDecoder(resp.Body).Decode(&data); err != nil {
		t.Fatal(err)
	}

	// metadata != app.Metadata()
	assert.Equal(t, data["md"], testmd)    // not passed
	assert.Equal(t, data["appmd"], testmd) // passed
}
