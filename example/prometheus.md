# prometheus

```go
package main

import (
"log/slog"
"net/http"
"time"

"github.com/duomi520/whttp"
"github.com/prometheus/client_golang/prometheus"
"github.com/prometheus/client_golang/prometheus/promauto"
"github.com/prometheus/client_golang/prometheus/promhttp"
)

func recordMetrics() {
//每1秒，计数器增加1。
go func() {
for {
opsProcessed.Inc()
time.Sleep(time.Second)
}
}()
}

// 公开了 myapp_processed_ops_total 计数器
var (
opsProcessed = promauto.NewCounter(prometheus.CounterOpts{
Name: "myapp_processed_ops_total",
Help: "The total number of processed events",
})
)

func main() {
recordMetrics()
route := whttp.NewRoute(nil)
srv := &http.Server{
Handler:        route.Mux,
MaxHeaderBytes: 1 << 20,
}
route.Mux.Handle("/metrics", promhttp.Handler())
if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
slog.Error(err.Error())
}
}

```

- <https://zhuanlan.zhihu.com/p/267966193>
- <https://bbs.huaweicloud.com/blogs/308299>
