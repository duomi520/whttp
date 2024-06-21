# session

```go
package main

import (
"fmt"
"log/slog"
"net/http"
"time"

"github.com/duomi520/whttp"
"github.com/go-playground/validator"
"github.com/gorilla/securecookie"
"github.com/gorilla/sessions"
)

var (
store = sessions.NewFilesystemStore("./", securecookie.GenerateRandomKey(32),securecookie.GenerateRandomKey(32))
)

func set(c *whttp.HTTPContext) {
session, _ := store.Get(c.Request, "user")
session.Values["name"] = "dj"
session.Values["age"] = 18
err := sessions.Save(c.Request, c.Writer)
if err != nil {
c.String(http.StatusInternalServerError, err.Error())
return
}
c.String(http.StatusOK, "Hello World")
}

func read(c *whttp.HTTPContext) {
session, _ := store.Get(c.Request, "user")
c.String(http.StatusOK, fmt.Sprintf("name:%s age:%d\n", session.Values["name"], session.Values["age"]))
}

func main() {
route := whttp.NewRoute(validator.New(), nil)
//配置服务
srv := &http.Server{
Handler:        route.Mux,
ReadTimeout:    3600 * time.Second,
WriteTimeout:   3600 * time.Second,
MaxHeaderBytes: 1 << 20,
}
route.GET("/set", set)
route.GET("/read", read)
if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
slog.Error(err.Error())
}
}
// https://www.cnblogs.com/luckzack/p/17737280.html
```
