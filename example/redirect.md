# 重定向

```go
package main

import (
"github.com/duomi520/whttp"
"github.com/go-playground/validator"
"log/slog"
"net/http"
"time"
)

func main() {
route := whttp.NewRoute(validator.New(), nil)
srv := &http.Server{
Handler:        route.Mux,
MaxHeaderBytes: 1 << 20,
}
// http.StatusMovedPermanently 永久性重定向
route.GET("/baidu", func(c *whttp.HTTPContext) {
http.Redirect(c.Writer, c.Request, "https://www.baidu.com/", http.StatusMovedPermanently)
})
// http.StatusFound 临时重定向
route.GET("/bing", func(c *whttp.HTTPContext) {
http.Redirect(c.Writer, c.Request, "https://cn.bing.com/", http.StatusFound)
})
// 路由重定向
route.GET("/test", func(c *whttp.HTTPContext) {
http.Redirect(c.Writer, c.Request, "lower", http.StatusFound)
})
route.GET("/lower", func(c *whttp.HTTPContext) {
c.String(http.StatusOK, "Hi")
})
if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
slog.Error(err.Error())
}
}
```
