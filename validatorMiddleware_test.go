package whttp

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/go-playground/validator/v10"
)

func TestValidatorMiddleware(t *testing.T) {
	fn := func(c *HTTPContext) {
		if !strings.EqualFold(c.Request.PathValue("a"), "777") {
			t.Fatal(c.Request.PathValue("a"))
		}
	}
	r := NewRoute(validator.New(), nil)
	r.Mux = http.NewServeMux()
	r.POST("/a/{a}", ValidatorMiddleware("a:numeric", "b:endswith=67"), fn)
	ts := httptest.NewServer(r.Mux)
	defer ts.Close()
	resp, err := http.Post(ts.URL+"/a/777", "application/x-www-form-urlencoded",
		strings.NewReader("b=667&d=hi"))
	if err != nil {
		t.Fatal(err)
	}
	data, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatal(string(data))
	}
}

type CarInfo struct {
	Name  string `validate:"checkNameLen"`
	Level int    `validate:"lt=50"`
}

// 自定义验证函数
func checkNameLen(fl validator.FieldLevel) bool {
	n := utf8.RuneCountInString(fl.Field().String())
	return n < 5
}

func TestValidatorStruct(t *testing.T) {
	validate := validator.New()
	err := validate.RegisterValidation("checkNameLen", checkNameLen)
	if err != nil {
		t.Fatal(err)
	}
	r := NewRoute(validate, nil)
	// func (c *HTTPContext) ValidatorStruct(a any) error
	err = r.validatorStruct(&CarInfo{Name: "宝马", Level: 70})
	fmt.Println(err.Error())
	err = r.validatorStruct(&CarInfo{Name: "奔驰", Level: 40})
	if err != nil {
		fmt.Println(err.Error())
	}
}

// Key: 'CarInfo.Level' Error:Field validation for 'Level' failed on the 'lt' tag
