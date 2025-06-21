# WHTTP

WHTTP 是一个用[Go](https://go.dev/) 开发的 web 脚手架，基于标准库简单封装。

**特性:**

- 中间件支持
- 验证
- 缓存文件

## 开始

### 必要条件

需要 [Go](https://go.dev/) 版本 [1.22](https://go.dev/doc/devel/release#go1.22.0) 以上

### 获取

下载

```sh
go get github.com/duomi520/whttp
```

将 WHTTP 引入到代码中

```sh
import "github.com/duomi520/whttp"
```

### 运行

简单例子:

```go
package main

import (
  "github.com/duomi520/whttp"
  "log/slog"
  "net/http"
)

func main() {
  route := whttp.NewRoute(nil)
  // 配置服务
  srv := &http.Server{
    Handler:        route.Mux,
    MaxHeaderBytes: 1 << 20,
  }
  route.GET("/", func(c *whttp.HTTPContext) {
    c.String(http.StatusOK, "Hi")
  })
  // 监听并在 0.0.0.0 上启动服务 (windows "localhost")
  if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
    slog.Error(err.Error())
  }
}
```

使用`go run` 命令运行：

```sh
go run main.go
```

在您的浏览器查看结果 [`0.0.0.0`](http://0.0.0.0) !

## api 的用法

### 路由定义

- GET 查询数据，对应 get 请求
- POST 创建数据，对应 post 请求
- PUT 更新数据，对应 put 请求
- DELETE 删除数据，对应 delete 请求

### 控制器函数

控制器函数只接受一个 whttp.HTTPContext 上下文参数

func HandlerFunc(c \*whttp.HTTPContext)

(whttp.HTTPContext).Writer 为原始的 http.ResponseWriter

(whttp.HTTPContext).Request 为原始的 \*http.Request

### 路由参数

```go
//匹配 /user/linda/mobile/xxxxxxxx
route.GET("/user/{name}/mobile/{mobile}", func(c *HTTPContext) {
  user := c.Request.PathValue("name")
  mobile := c.Request.PathValue("mobile")
  c.String(http.StatusOK, user+mobile)
})
```

### 查询字符串参数

```go
//示例 URL： /name=linda&mobile=xxxxxxxx
route.POST("/", func(c *HTTPContext) {
  user := c.Request.FormValue("name")
  mobile := c.Request.FormValue("mobile")
  c.String(http.StatusOK, user+mobile)
})
```

| 操作          | 解析                        | 读取 URL | 读取 Body（表单） | 支持文本 | 支持二进制 |
| ------------- | --------------------------- | -------- | ----------------- | -------- | ---------- |
| Form          | ParseForm                   | √        | √                 | √        |            |
| PostForm      | ParseForm                   |          | √                 | √        |            |
| FormValue     | 自动调用 ParseForm          | √        | √                 | √        |            |
| PostFormValue | 自动调用 ParseForm          |          | √                 | √        |            |
| MultipartForm | ParseMultipartForm          |          | √                 | √        | √          |
| FormFile      | 自动调用 ParseMultipartForm |          | √                 |          | √          |

### 响应方式

status 为 http 状态码，响应方式会在中间件执行完毕后执行

- 返回字符串 func (c \*HTTPContext) String(status int, msg string)
- 返回 JSON func (c \*HTTPContext) JSON(status int, v any)
- 返回二进制 func (c \*HTTPContext) Blob(status int, contentType string, data []byte)
- 返回文件 func (c \*HTTPContext) File(filepath string)

自定义

```go
c.Flush = func() (int, error) {
  return c.Writer.Write(d)
}
```

### 模板

模版文件 file.tmpl 内容

```sh
{{.}}
```

渲染模版文件

```go
//解析模板文件
tl, err := template.ParseFiles("file.tmpl")
if err != nil {
  c.String(http.StatusInternalServerError, err.Error())
}
//注册模板
route.SetRenderer(tl)
route.GET("/", func(c *HTTPContext) {
  //渲染模板
  c.Render(http.StatusOK, "file.tmpl", "6月7日")
})
```

渲染结果为：6 月 7 日

### 静态文件服务

group 为中间件函数

func (r *WRoute) Static(relativePath, root string, group ...func(*HTTPContext))

```go
route.Static("/", "welcome.txt")
```

func (r *WRoute) StaticFS(dir string, group ...func(*HTTPContext))

```go
route.StaticFS("txt")
```

### 中间件

中间件指的是可以拦截 http 请求-响应生命周期的特殊函数，在请求-响应生命周期中可以注册多个中间件，每个中间件执行不同的功能，一个中间执行完再轮到下一个中间件执行

```go
Middleware := func() func(*HTTPContext) {
return func(c *HTTPContext) {
  //处理拦截请求的逻辑
  fmt.Println("前处理")
  //执行下一个中间件或者执行控制器函数
  c.Next()
  //后处理
  fmt.Println("后处理")
}
}
```

为每个路由添加任意数量的中间件

```go
route.GET("/some", MiddlewareA(),MiddlewareB(), Endpoint)
```

为中间件建一切片，简化重复输入

```go
g := []func(*HTTPContext){MiddlewareA(), MiddlewareB(), MiddlewareC()}
route.GET("/some", append(g, Endpoint)...)
```

全局中间件，需在初始化后立即加载。

```go
route.Use(LoggerMiddleware())
```

### 自定义日志

日志使用 "log/slog" ,NewRoute()初始化路由时加载自定义日志

```go
func(c *HTTPContext) {
  c.Debug("This is Debug Level")
  c.Info("This is Info Level")
  c.Warn("This is Warn Level")
  c.Error("This is Error Level")
}
```

## 进阶的用法

[examples](https://github.com/duomi520/whttp/tree/master/example) 样例库

## 与众不同的模块

待续...
