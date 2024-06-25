# 安全

```go
package main

import (
"github.com/duomi520/whttp"
"github.com/go-playground/validator"
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
route := whttp.NewRoute(validator.New(), nil)
route.Use(whttp.HeaderMiddleware(security))
//配置服务
srv := &http.Server{
Handler: route.Mux,
ReadTimeout: 3600 _ time.Second,
WriteTimeout: 3600 _ time.Second,
MaxHeaderBytes: 1 << 20,
}
if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
slog.Error(err.Error())
}
}
```
