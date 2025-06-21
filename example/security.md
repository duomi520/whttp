# 安全

Security 所提供的配置项是为了简化一些常见的 HTTP headers 的配置，如对配置项配置 HTTP headers 的作用感到困惑，可以自行在 [MDN Docs](https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers) 中进行查询它们的作用

```go
package main

import (
"github.com/duomi520/whttp"
"log/slog"
"net/http"
"time"
)

var security = map[string]string{
"frame-ancestors": "none",
"X-Frame-Optoins": "DENY",
"Content-Security-Policy": "default-src 'self'; connect-src *; font-src *; script-src-elem * 'unsafe-inline'; img-src * data:; style-src * 'unsafe-inline';",
"X-XSS-Protection": "1; mode=block",
"Strict-Transport-Security": "max-age=31536000; includeSubDomains; preload",
"Referrer-Policy": "strict-origin",
"X-Content-Type-Options": "nosniff",
"Permissions-Policy": "geolocation=(),midi=(),sync-xhr=(),microphone=(),camera=(),magnetometer=(),gyroscope=(),fullscreen=(self),payment=()",
}

func main() {
route := whttp.NewRoute(nil)
route.Use(whttp.HeaderMiddleware(security))
//配置服务
srv := &http.Server{
Handler: route.Mux,
MaxHeaderBytes: 1 << 20,
}
if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
slog.Error(err.Error())
}
}
```
