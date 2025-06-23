# ratelimitMiddleware

```go
package main

import (
"log/slog"
"net/http"
"time"

"github.com/duomi520/whttp"
"github.com/go-kratos/aegis/ratelimit"
"github.com/go-kratos/aegis/ratelimit/bbr"

)

func LimiterMiddleware(limiter *bbr.BBR) func(*whttp.HTTPContext) {
return func(c *whttp.HTTPContext) {
done, err := limiter.Allow()
if err != nil {
c.String(http.StatusTooManyRequests, err.Error())
return
}
c.Next()
done(ratelimit.DoneInfo{})
}
}

func main() {
limiter := bbr.NewLimiter(
bbr.WithWindow(time.Second),
bbr.WithBucket(50),
bbr.WithCPUThreshold(100))
route := whttp.NewRoute(nil)
srv := &http.Server{
Handler: route.Mux,
MaxHeaderBytes: 1 << 20,
}
route.GET("/Hi", LimiterMiddleware(limiter), func(c *whttp.HTTPContext) {
c.String(http.StatusOK, "Hi")
})
if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
slog.Error(err.Error())
}
}
```

- <https://www.jianshu.com/p/ec43d44ff59b>
- <https://www.cnblogs.com/daemon365/p/15227815.html>
