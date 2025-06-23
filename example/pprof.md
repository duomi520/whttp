# pprof

```go
package main

import (
"github.com/duomi520/whttp"
"log/slog"
"net/http"
"net/http/pprof"
"time"
)

func main() {
route := whttp.NewRoute(nil)
route.Mux.HandleFunc("/debug/pprof/", pprof.Index)
route.Mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
route.Mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
route.Mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
route.Mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
srv := &http.Server{
Handler: route.Mux,
MaxHeaderBytes: 1 << 20,
}
route.GET("/", func(c *whttp.HTTPContext) {
c.String(http.StatusOK, "Hi")
})
if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
slog.Error(err.Error())
}
}
```

- https://zhuanlan.zhihu.com/p/685823045
- https://www.jb51.net/jiaoben/319016j3j.htm
