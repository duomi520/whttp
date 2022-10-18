package whttp

import (
	"github.com/duomi520/utils"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestLoggerMiddleware(t *testing.T) {
	logger, _ := utils.NewWLogger(utils.DebugLevel, "")
	fn := func(c *HTTPContext) {
	}
	group := HTTPMiddleware(LoggerMiddleware())
	r := &WRoute{router: mux.NewRouter(), logger: logger}
	ts := httptest.NewServer(http.HandlerFunc(r.Warp(group, fn)))
	defer ts.Close()
	_, err := http.Post(ts.URL, "application/x-www-form-urlencoded",
		strings.NewReader(""))
	if err != nil {
		log.Fatal(err)
	}
	_, err = http.Get(ts.URL)
	if err != nil {
		log.Fatal(err)
	}
}

/*
[Debug] 2022-10-18 18:33:11 |            0s | 127.0.0.1:60511 |     0 |    POST | / |
[Debug] 2022-10-18 18:33:11 |            0s | 127.0.0.1:60511 |     0 |     GET | / |
*/
