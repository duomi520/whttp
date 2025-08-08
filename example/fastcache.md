# fastcache

- <https://github.com/VictoriaMetrics/fastcache>

```go
package main

import (
"github.com/VictoriaMetrics/fastcache"
"github.com/duomi520/whttp"
"log/slog"
"net/http"
)

type Cache struct {
c *fastcache.Cache
}
func (c *Cache) Del(key string) {
c.c.Del([]byte(key))
}
func (c *Cache) Set(key string, value []byte) {
c.c.Set([]byte(key), value)
}
func (c *Cache) HasGet(dst []byte, key string) ([]byte, bool) {
return c.c.HasGet(dst, []byte(key))
}
func main() {
cache := &Cache{}
cache.c = fastcache.New(32 * 1024 * 1024)
route := whttp.NewRoute(nil)
srv := &http.Server{
Handler:        route.Mux,
MaxHeaderBytes: 1 << 20,
}
route.GET("/hi", whttp.CacheMiddleware(cache, nil), func(c *whttp.HTTPContext) {
c.String(http.StatusOK, "Hi")
})
if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
slog.Error(err.Error())
}
}
```
