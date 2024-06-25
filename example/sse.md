# Server-sent Events

```go
package main

import (
"log/slog"
"net/http"
"time"
"github.com/duomi520/whttp"
"github.com/go-playground/validator"
)
func main() {
route := whttp.NewRoute(validator.New(), nil)
//配置服务
srv := &http.Server{
Handler: route.Mux,
ReadTimeout: 3600 _ time.Second,
WriteTimeout: 3600 _ time.Second,
MaxHeaderBytes: 1 << 20,
}
route.GET("/SSEvent", func(c *whttp.HTTPContext) {
c.Writer.Header().Set("Content-Type", "text/event-stream")
c.Writer.Header().Set("Cache-Control", "no-cache")
c.Writer.Header().Set("Connection", "keep-alive")
flusher, err := c.Writer.(http.Flusher)
if !err {
c.String(http.StatusInternalServerError, "streaming unsupported!")
return
}
for i := 0; i < 10; i++ {
c.Writer.Write([]byte(time.Now().String() + "\n"))
flusher.Flush()
time.Sleep(time.Second)
}
})
if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
slog.Error(err.Error())
}
}
```
