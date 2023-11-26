package whttp

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/duomi520/utils"
	"github.com/julienschmidt/httprouter"
)

func TestLoggerMiddleware(t *testing.T) {
	logger, _ := utils.NewWLogger(utils.DebugLevel, "")
	fn := func(c *HTTPContext) {
	}
	r := &WRoute{router: httprouter.New(), logger: logger}
	r.POST("/", LoggerMiddleware(), fn)
	r.GET("/", LoggerMiddleware(), fn)
	ts := httptest.NewServer(r.router)
	defer ts.Close()
	_, err := http.Post(ts.URL, "application/x-www-form-urlencoded",
		strings.NewReader(""))
	if err != nil {
		t.Fatal(err)
	}
	_, err = http.Get(ts.URL)
	if err != nil {
		t.Fatal(err)
	}
}

/*
[Debug] 2023-05-01 13:59:46 |            0s | 127.0.0.1:61791 |     0 |    POST | / |
[Debug] 2023-05-01 13:59:46 |            0s | 127.0.0.1:61791 |     0 |     GET | / |
*/
