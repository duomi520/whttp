package whttp

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/duomi520/utils"
	"github.com/go-playground/validator"
)

func TestLimitMiddleware(t *testing.T) {
	limiter := utils.NewTokenBucketLimiter(2, 5, 100*time.Millisecond)
	defer limiter.Close()
	go limiter.Run()
	time.Sleep(10 * time.Millisecond)
	fn := func(c *HTTPContext) {
		c.String(200, "Hi")
	}
	r := NewRoute(validator.New(), nil)
	r.Mux = http.NewServeMux()
	r.GET("/", LimitMiddleware(limiter), fn)
	ts := httptest.NewServer(r.Mux)
	defer ts.Close()
	for i := 0; i < 20; i++ {
		resp, err := http.Get(ts.URL)
		if err != nil {
			t.Fatal(err)
		}
		data, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			t.Fatal(err)
		}
		fmt.Println(strconv.Itoa(i), resp.StatusCode, string(data))
		time.Sleep(10 * time.Millisecond)
	}
}

/*
0 200 Hi
1 200 Hi
2 200 Hi
3 200 Hi
4 200 Hi
5 429 rate limit
6 200 Hi
7 200 Hi
8 429 rate limit
9 429 rate limit
10 429 rate limit
11 429 rate limit
12 429 rate limit
13 200 Hi
14 200 Hi
15 429 rate limit
16 429 rate limit
17 429 rate limit
18 429 rate limit
19 429 rate limit
*/
