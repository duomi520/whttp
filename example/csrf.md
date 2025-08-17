# CSRF

- <https://zhuanlan.zhihu.com/p/1939417081423574014>

Go 1.25 引入了 CrossOriginProtection，用于保护 Web 应用免受 CSRF 攻击

CrossOriginProtection 会拒绝不安全的跨源浏览器请求。它的判断依据是：
Sec-Fetch-Site 请求头（2023 年起所有主流浏览器支持）
或 Origin 请求头与 Host 对比

默认允许的安全方法：GET、HEAD、OPTIONS

## 创建实例

```go
cop := http.NewCrossOriginProtection()
```

零值有效，无需初始化参数，支持并发调用

## 添加可信来源

```go
err := cop.AddTrustedOrigin("https://myapp.com")
if err != nil {
    log.Fatal(err)
}
```

允许来自特定 Origin 的请求，Origin 格式：scheme://host[:port]

## 添加不安全路径绕过

```go
cop.AddInsecureBypassPattern("/public/")
```

/public/ 路径将不进行跨源检查，规则与 http.ServeMux 相同，可并发调用

## Check 方法用法（手动检查单个请求）

Check 方法用于在业务逻辑中手动判断请求是否被允许。

```go
func apiHandler(w http.ResponseWriter, r *http.Request) {
    if err := cop.Check(r); err != nil {
        http.Error(w, err.Error(), http.StatusForbidden)
        return
    }
    // 请求合法，继续业务处理
    w.Write([]byte("API response"))
}
```

- 优点：精细控制，可根据条件选择性检查
- 场景：单个请求、复杂中间件或特殊逻辑处理

## 总结

| 方法                     | 用途                                    | 场景                       |
| ------------------------ | --------------------------------------- | -------------------------- |
| Handler                  | 自动检查跨源请求，失败返回 403 或自定义 | 整条 API 路径统一保护      |
| Check                    | 手动检查请求，返回 error                | 单个请求或复杂业务逻辑控制 |
| AddTrustedOrigin         | 添加可信 Origin                         | 允许特定来源请求           |
| AddInsecureBypassPattern | 添加不安全路径绕过                      | 静态资源或无需检查的路径   |
| SetDenyHandler           | 自定义拒绝逻辑                          | 替代默认 403 返回          |

## 完整示例

```go
package main

import (
"github.com/duomi520/whttp"
"log/slog"
"net/http"
)

func main() {
cop := http.NewCrossOriginProtection()
// 添加可信任的跨站来源
err := cop.AddTrustedOrigin("https://trusted.com")
if err != nil {
panic(err.Error())
}
// 添加无需检查的路径
cop.AddInsecureBypassPattern("/public")
route := whttp.NewRoute(nil)
srv := &http.Server{
Handler:        route.Mux,
MaxHeaderBytes: 1 << 20,
}
route.Static("/csrf", "csrf.html")
route.POST("/check", CSRFMiddleware(cop), func(c *whttp.HTTPContext) {
c.String(http.StatusOK, "check success")
})
if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
slog.Error(err.Error())
}
}

// CSRFMiddleware
func CSRFMiddleware(cop *http.CrossOriginProtection) func(*whttp.HTTPContext) {
return func(c *whttp.HTTPContext) {
if err := cop.Check(c.Request); err != nil {
c.String(http.StatusForbidden, "Forbidden")
return
}
c.Next()
}
}
```

```html
<html>
  <head>
    <link rel="icon" href="data:;base64,=qWNv" />
    <script>
      fetch(
        new Request("/check", {
          method: "POST",
          body: "param=value",
        })
      )
        .then((response) => {
          if (!response.ok) {
            throw new Error("Response was not ok");
          }
          return response.text();
        })
        .then((data) => {
          console.log(data);
        })
        .catch((error) => {
          console.error("There was a problem with the fetch operation:", error);
        });
    </script>
  </head>

  <body>
    <h2>CSRF</h2>
  </body>
</html>
```
