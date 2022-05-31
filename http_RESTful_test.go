package utils

import (
	"bytes"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/duomi520/utils"
	"github.com/go-playground/validator"
	"github.com/gorilla/mux"
)

func TestWarp(t *testing.T) {
	r := &WRoute{router: mux.NewRouter()}
	fn := func(c *HTTPContext) {
		if (strings.Compare(c.Params("name"), "linda") == 0) && (strings.Compare(c.Params("mobile"), "xxxxxxxx") == 0) {
			c.String(http.StatusOK, "OK")
		} else {
			c.String(http.StatusOK, "NG")
		}
	}
	ts := httptest.NewServer(http.HandlerFunc(r.Warp(nil, fn)))
	defer ts.Close()
	res, err := http.Post(ts.URL, "application/x-www-form-urlencoded",
		strings.NewReader("name=linda&mobile=xxxxxxxx"))
	if err != nil {
		log.Fatal(err)
	}
	greeting, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		log.Fatal(err)
	}
	if !bytes.Equal(greeting, []byte("OK")) {
		t.Errorf("got %s | expected ok", string(greeting))
	}
}

func TestMiddleware(t *testing.T) {
	signature := ""
	r := &WRoute{router: mux.NewRouter()}
	MiddlewareA := func() func(*HTTPContext) {
		return func(c *HTTPContext) {
			signature += "A1"
			c.Next()
			signature += "A2"
		}
	}
	MiddlewareB := func() func(*HTTPContext) {
		return func(c *HTTPContext) {
			signature += "B1"
			c.Next()
			signature += "B2"
		}
	}
	group := HTTPMiddleware(MiddlewareA(), MiddlewareB())
	fn := func(c *HTTPContext) {
		signature += "<->"
	}
	ts := httptest.NewServer(http.HandlerFunc(r.Warp(group, fn)))
	defer ts.Close()
	_, err := http.Post(ts.URL, "application/x-www-form-urlencoded",
		strings.NewReader(""))
	if err != nil {
		log.Fatal(err)
	}
	if !strings.EqualFold(signature, "A1B1<->B2A2") {
		t.Errorf("got %s | expected A1B1<->B2A2", signature)
	}
}

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
[Debug] 2022-05-04 23:32:00 |             0 | 127.0.0.1:62286 |     0 |    POST | / |
[Debug] 2022-05-04 23:32:00 |             0 | 127.0.0.1:62286 |     0 |     GET | / |
*/

func TestValidatorMiddleware(t *testing.T) {
	v := validator.New()
	fn := func(c *HTTPContext) {}
	group := HTTPMiddleware(ValidatorMiddleware("number:numeric"))
	r := &WRoute{router: mux.NewRouter(), validatorVar: v.Var, validatorStruct: v.Struct}
	mr := mux.NewRouter()
	ts := httptest.NewServer(mr)
	mr.HandleFunc("/number/{number}", r.Warp(group, fn)).Methods("POST")
	defer ts.Close()
	resp, err := http.Post(ts.URL+"/number/77", "application/x-www-form-urlencoded",
		strings.NewReader(""))
	if err != nil {
		log.Fatal(err)
	}
	data, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		log.Fatal(err)
	}
	if resp.StatusCode != 200 {
		log.Fatal(string(data))
	}
}
