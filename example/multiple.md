# 多个服务

```go
package main

import (
"context"
"log/slog"
"net/http"
"os"
"os/signal"
"syscall"
"time"
"github.com/duomi520/whttp"
)

func service1(srv *http.Server) {
route := whttp.NewRoute(nil)
srv.Addr = ":8080"
srv.Handler = route.Mux
route.GET("/", func(c *whttp.HTTPContext) {
c.String(http.StatusOK, "service1")
})
if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
slog.Error(err.Error())
}
}
func service2(srv *http.Server) {
route := whttp.NewRoute(nil)
srv.Addr = ":8090"
srv.Handler = route.Mux
route.GET("/", func(c *whttp.HTTPContext) {
c.String(http.StatusOK, "service2")
})
if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
slog.Error(err.Error())
}
}

func main() {
srv1 := &http.Server{}
srv2 := &http.Server{}
go service1(srv1)
go service2(srv2)
exitChan := make(chan os.Signal, 16)
signal.Notify(exitChan, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
<-exitChan
ctx, cancel := context.WithTimeout(context.Background(), time.Second)
defer cancel()
if err := srv1.Shutdown(ctx); err != nil {
slog.Error(err.Error())
}
if err := srv2.Shutdown(ctx); err != nil {
slog.Error(err.Error())
}
time.Sleep(5 * time.Second)
}
```
