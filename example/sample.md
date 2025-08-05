# 范例

```go
package main

import (
 "context"
 "github.com/duomi520/whttp"
 "log/slog"
 "net/http"
 "os"
 "os/signal"
 "syscall"
 "time"
)
const (
 Version ="1.0.1"
)
func main() {
 var mf whttp.MemoryFile
 route := whttp.NewRoute(nil)
 //配置服务
 srv := &http.Server{
  Handler:        route.Mux,
  ReadTimeout:    3600 * time.Second,
  WriteTimeout:   3600 * time.Second,
  MaxHeaderBytes: 1 << 20,
 }
 //监听信号 ctrl+c kill
 exitChan := make(chan os.Signal, 16)
 signal.Notify(exitChan, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
 go func() {
  <-exitChan
  ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
  defer cancel()
  if err := srv.Shutdown(ctx); err != nil {
   slog.Error(err.Error())
  }
 }()
 //ping
 route.GET("/ping", func(c *whttp.HTTPContext) {
  c.String(http.StatusOK, "pong")
 })
 //version
 route.GET("/version", func(c *whttp.HTTPContext) {
  c.String(http.StatusOK, Version)
 })
 //启动服务
 if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
  slog.Error(err.Error())
 }
}
```
