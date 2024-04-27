package whttp

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/duomi520/utils"
)

func TestLoggerMiddleware(t *testing.T) {
	logger, _ := utils.NewWLogger(utils.DebugLevel, "")
	fn := func(c *HTTPContext) {}
	r := &WRoute{mux: http.NewServeMux(), logger: logger}
	r.POST("/p", LoggerMiddleware(), fn)
	r.GET("/g", LoggerMiddleware(), fn)
	ts := httptest.NewServer(r.mux)
	defer ts.Close()
	_, err := http.Post(ts.URL+"/p", "application/x-www-form-urlencoded",
		strings.NewReader(""))
	if err != nil {
		t.Fatal(err)
	}
	_, err = http.Get(ts.URL + "/g")
	if err != nil {
		t.Fatal(err)
	}
}

/*
[Debug] 2024-04-27 13:40:57 |            0s | 127.0.0.1:54783 |     0 |    POST | /p |
[Debug] 2024-04-27 13:40:57 |            0s | 127.0.0.1:54783 |     0 |     GET | /g |
*/
