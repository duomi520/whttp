# https

```go
package main

import (
 "github.com/duomi520/whttp"
 "log/slog"
 "net/http"
 "time"
)

func main() {
route := whttp.NewRoute(nil)
srv := &http.Server{
Handler:        route.Mux,
MaxHeaderBytes: 1 << 20,
}
route.GET("/ping", func(c *whttp.HTTPContext) {
c.String(http.StatusOK, "pong")
})
if err := srv.ListenAndServeTLS("server.crt", "server.key"); err != nil && err != http.ErrServerClosed {
slog.Error(err.Error())
}
}
// https://www.zhihu.com/question/305961226
```
