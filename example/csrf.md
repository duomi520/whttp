# CSRF

```go
package main
import (
"log/slog"
"net/http"
"time"
"github.com/duomi520/whttp"
"github.com/gorilla/csrf"
)

func getCSRF(c *whttp.HTTPContext) {
// 获取 token 值
token := csrf.Token(c.Request)
// 将 token 写入到 header 中
c.Writer.Header().Set("X-CSRF-Token", token)
// 业务代码
str1 := `<html><form method="post" name="myForm" action="http://127.0.0.1/p">
<input type="hidden" name="gorilla.csrf.Token" value="`
str2 := `"><button type="submit">Send</button></form></html>`
c.String(http.StatusOK, str1+token+str2)
}
func postCSRF(c *whttp.HTTPContext) {
c.String(http.StatusOK, "hello")
}
func main() {
route := whttp.NewRoute(nil)
csrfMiddleware := csrf.Protect([]byte("32-byte-long-auth-key"))
route.GET("/g", getCSRF)
route.POST("/p", postCSRF)
srv := &http.Server{
//csrfMiddleware 默认只对 POST 验证
Handler: csrfMiddleware(route.Mux),
MaxHeaderBytes: 1 << 20,
}
if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
slog.Error(err.Error())
}
}
```
