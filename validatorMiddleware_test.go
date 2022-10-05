package whttp

import (
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-playground/validator"
	"github.com/gorilla/mux"
)

func TestValidatorMiddleware(t *testing.T) {
	v := validator.New()
	fn := func(c *HTTPContext) {
		if !strings.EqualFold(c.vars["number"], "777") {
			log.Fatal(c.vars["number"])
		}
	}
	group := HTTPMiddleware(ValidatorMiddleware("number:numeric"))
	r := &WRoute{router: mux.NewRouter(), validatorVar: v.Var, validatorStruct: v.Struct}
	mr := mux.NewRouter()
	ts := httptest.NewServer(mr)
	mr.HandleFunc("/number/{number}", r.Warp(group, fn)).Methods("POST")
	defer ts.Close()
	resp, err := http.Post(ts.URL+"/number/777", "application/x-www-form-urlencoded",
		strings.NewReader(""))
	if err != nil {
		log.Fatal(err)
	}
	data, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		log.Fatal(err)
	}
	if resp.StatusCode != 200 {
		log.Fatal(string(data))
	}
}
