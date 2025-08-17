# 流式 HTTP

chunked encoding 传统意义上的 流式 HTTP（Streaming HTTP）

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
Handler:        route.Mux,
MaxHeaderBytes: 1 << 20,
}
route.Static("/now", "now.html")
route.GET("/stream", func(c *whttp.HTTPContext) {
flusher, ok := c.Writer.(http.Flusher)
if !ok {
c.String(http.StatusInternalServerError, "expected http.ResponseWriter to be an http.Flusher")
return
}
c.Writer.Header().Set("Transfer-Encoding", "chunked")
c.Writer.Header().Set("Connection", "Keep-Alive")
c.Writer.Header().Set("X-Content-Type-Options", "nosniff")
for range 10 {
_, err := c.Writer.Write([]byte(time.Now().String() + "\n"))
if err != nil {
slog.Error(err.Error())
return
}
flusher.Flush()
time.Sleep(time.Second)
}
})
if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
slog.Error(err.Error())
}
}
```

```html
<html>
  <head>
    <link rel="icon" href="data:;base64,=qWNv" />
    <script>
      fetch("/stream")
        .then((response) => {
          const reader = response.body.getReader();
          const decoder = new TextDecoder("utf-8");
          function read() {
            return reader.read().then(({ done, value }) => {
              if (done) {
                console.log("stream complete");
                return;
              }
              const chunk = decoder.decode(value);
              console.log("chunk:", chunk);
              return read();
            });
          }
          return read();
        })
        .catch((error) => {
          console.error("Error:", error);
        });
    </script>
  </head>

  <body>
    <h2>Welcome</h2>
  </body>
</html>
```
