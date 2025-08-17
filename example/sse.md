# Server-sent Events

```go
package main

import (
"log/slog"
"net/http"
"time"
"github.com/duomi520/whttp"
)
func main() {
route := whttp.NewRoute(nil)
srv := &http.Server{
Handler: route.Mux,
MaxHeaderBytes: 1 << 20,
}
route.GET("/SSEvent", func(c *whttp.HTTPContext) {
flusher, err := c.Writer.(http.Flusher)
if !err {
c.String(http.StatusInternalServerError, "streaming unsupported!")
return
}
c.Writer.Header().Set("Content-Type", "text/event-stream")
c.Writer.Header().Set("Cache-Control", "no-cache")
c.Writer.Header().Set("Connection", "keep-alive")
c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
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

```javascript
const eventSource = new EventSource("/SSEvent");
eventSource.onmessage = function (e) {
  console.log("Message", e.data);
};
```
