# Clickjacking

ClickjackingMiddleware 点击劫持 是指攻击者使用多个透明或不透明层来诱使用户在打算点击顶层页面时点击另一个页面上的按钮或链接。
因此攻击者正在劫持针对其页面的点击，并将它们路由到另一个页面，该页面很可能是另一个应用程序或域。

使用内容安全策略（CSP）frame-ancestors 指令进行防御
此设置阻止任何域使用框架对页面进行引用
"frame-ancestors": "none"
使用 X-Frame-Options HTTP 响应标头进行防御
此设置阻止任何域使用框架对页面进行引用
"X-Frame-Optoins": "DENY"

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
route.Use(whttp.HeaderMiddleware(map[string]string{"frame-ancestors": "none", "X-Frame-Optoins": "DENY"}))
//配置服务
srv := &http.Server{
Handler:        route.Mux,
ReadTimeout:    3600 * time.Second,
WriteTimeout:   3600 * time.Second,
MaxHeaderBytes: 1 << 20,
}
if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
slog.Error(err.Error())
}
}
```
