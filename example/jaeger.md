# jaeger

```go
package main
import (
"io"
"log/slog"
"net/http"
"github.com/duomi520/whttp"
"github.com/opentracing/opentracing-go"
"github.com/opentracing/opentracing-go/ext"
"github.com/uber/jaeger-client-go"
jaegercfg "github.com/uber/jaeger-client-go/config"
jaegerlog "github.com/uber/jaeger-client-go/log"
)
func main() {
route := whttp.NewRoute(nil)
route.Use(UseOpenTracing("http://127.0.0.1"))
srv := &http.Server{
Handler: route.Mux,
MaxHeaderBytes: 1 << 20,
}
route.GET("/hi", func(c *whttp.HTTPContext) {
c.String(http.StatusOK, "Hi")
})
if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
slog.Error(err.Error())
}
}
// 从上下文中解析并创建一个新的 trace，获得传播的 上下文(SpanContext)
func CreateTracer(serviceName string, url string, header http.Header) (opentracing.Tracer, opentracing.SpanContext, io.Closer, error) {
var cfg = jaegercfg.Configuration{
ServiceName: serviceName,
Sampler: &jaegercfg.SamplerConfig{
Type: jaeger.SamplerTypeConst,
Param: 1,
},
Reporter: &jaegercfg.ReporterConfig{
LogSpans: true,
CollectorEndpoint: url + ":14268/api/traces",
},
}
jLogger := jaegerlog.StdLogger
tracer, closer, err := cfg.NewTracer(
jaegercfg.Logger(jLogger),
)
// 继承别的进程传递过来的上下文
spanContext, _ := tracer.Extract(opentracing.HTTPHeaders,
opentracing.HTTPHeadersCarrier(header))
return tracer, spanContext, closer, err
}

func UseOpenTracing(url string) func(*whttp.HTTPContext) {
return func(c *whttp.HTTPContext) {
// 使用 opentracing.GlobalTracer() 获取全局 Tracer
tracer, spanContext, closer, _ := CreateTracer("Test whttp", url, c.Request.Header)
defer closer.Close()
// 生成依赖关系，并新建一个 span、
// 这里很重要，因为生成了 References []SpanReference 依赖关系
startSpan := tracer.StartSpan(c.Request.URL.Path, ext.RPCServerOption(spanContext))
defer startSpan.Finish()
// 记录 tag
// 记录请求 Url
ext.HTTPUrl.Set(startSpan, c.Request.URL.Path)
// Http Method
ext.HTTPMethod.Set(startSpan, c.Request.Method)
// 记录组件名称
ext.Component.Set(startSpan, "whttp")
// 在 header 中加上当前进程的上下文信息
c.Request = c.Request.WithContext(opentracing.ContextWithSpan(c.Request.Context(), startSpan))
// 传递给下一个中间件
c.Next()
// 继续设置 tag
// ext.HTTPStatusCode.Set(startSpan, uint16(c.Writer.Status()))
}
}
```

- <https://www.cnblogs.com/whuanle/p/14598049.html>
- <https://www.cnblogs.com/whuanle/p/14598049.html>
