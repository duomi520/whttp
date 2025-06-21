# 更换默认 json

```go
package main
import (
"github.com/bytedance/sonic"
"github.com/duomi520/whttp"
"log/slog"
"net/http"
"time"
)
func main() {
whttp.DefaultMarshal = sonic.Marshal
whttp.DefaultUnmarshal = sonic.Unmarshal
route := whttp.NewRoute(nil)
//配置服务
srv := &http.Server{
Handler: route.Mux,
MaxHeaderBytes: 1 << 20,
}
route.GET("/", func(c *whttp.HTTPContext) {
c.JSON(http.StatusOK, whttp.H{"id": 1, "name": "wang"})
})
if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
slog.Error(err.Error())
}
}
```
