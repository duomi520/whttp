# pprof

```go
package main

import (
"github.com/duomi520/whttp"
"github.com/go-playground/validator"
"log/slog"
"net/http"
"net/http/pprof"
"time"
)

func main() {
route := whttp.NewRoute(validator.New(), nil)
route.Mux.HandleFunc("/debug/pprof/", pprof.Index)
route.Mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
route.Mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
route.Mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
route.Mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
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

// https://zhuanlan.zhihu.com/p/685823045
```