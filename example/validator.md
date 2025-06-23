# 验证

```go
package main

import (
"fmt"
"github.com/duomi520/whttp"
"github.com/go-playground/validator/v10"
"log/slog"
"net/http"
"time"
)
var Validator = validator.New()
// 自定义验证函数
func checkNameLen(fl validator.FieldLevel) bool {
n := utf8.RuneCountInString(fl.Field().String())
return n < 5
}
func main() {
err := Validator.RegisterValidation("checkNameLen", checkNameLen)
if err != nil {
panic(err)
}
route := whttp.NewRoute(nil)
//配置服务
srv := &http.Server{
Handler:        route.Mux,
ReadTimeout:    3600 * time.Second,
WriteTimeout:   3600 * time.Second,
MaxHeaderBytes: 1 << 20,
}
route.POST("/pathValue/{p}", func(c *whttp.HTTPContext) {
k := c.Request.PathValue("p")
if err := Validator.Var(k, "numeric"); err != nil {
c.String(http.StatusBadRequest, fmt.Sprintf("输入不为整数: %s", k))
return
}
c.String(http.StatusOK, k)
})
route.POST("/formValue", func(c *whttp.HTTPContext) {
k := c.Request.FormValue("f")
if err := Validator.Var(k, "numeric"); err != nil {
c.String(http.StatusBadRequest, fmt.Sprintf("输入不为整数: %s", k))
return
}
c.String(http.StatusOK, k)
})
route.POST("/struct", func(c *whttp.HTTPContext) {
type CarInfo struct {
Name  string `json:"name" validate:"checkNameLen"`
Level int    `json:"level" validate:"lt=50"`
}
var car CarInfo
err := c.BindJSON(&car)
if err != nil {
c.String(http.StatusBadRequest, err.Error())
return
}
err = Validator.Struct(&car)
if err != nil {
c.String(http.StatusBadRequest, err.Error())
return
}
c.JSON(http.StatusOK, car)
})
if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
slog.Error(err.Error())
}
}
// curl -X POST http://127.0.0.1/pathValue/123.5
// curl -X POST -d "f=125.5" http://127.0.0.1/formValue
// curl -X POST -H "Content-Type:application/json" -d "{\"name\":\"byd\",\"level\":70}" http://127.0.0.1/struct
```

## 验证规则

- required：字段必须设置，不能为默认值；
- omitempty：字段未设置，则忽略；
- ,: 把多个验证标记隔开。注意：隔开逗号之间不能有空格；
- -：跳过该字段，不检验；
- |：使用多个约束，只需要满足其中一个；例："rgb|rgba"
- email：验证字符串是 email 格式；例："email"
- ip：字段值是否包含有效的 IP 地址，例："ip"
- ipv4：字段值是否包含有效的 ipv4 地址，例："ipv4"
- ipv6：字段值是否包含有效的 ipv6 地址，例："ipv6"
- url：这将验证字符串值包含有效的网址；例："url"

## 范围验证

- max：字符串最大长度；例："max=20"
- min：字符串最小长度；例："min=6"
- excludesall：不能包含特殊字符；例："excludesall=0x2C", 注意这里用十六进制表示；
- len：字符长度必须等于 n，或者数组、切片、map 的 len 值为 n，即包含的项目数；例："len=6"
- eq：数字等于 n，或者或者数组、切片、map 的 len 值为 n，即包含的项目数；例："eq=6"
- ne：数字不等于 n，或者或者数组、切片、map 的 len 值不等于为 n，即包含的项目数不为 n，其和 eq 相反；例："ne=6"
- gt：数字大于 n，或者或者数组、切片、map 的 len 值大于 n，即包含的项目数大于 n；例："gt=6"
- gte：数字大于或等于 n，或者或者数组、切片、map 的 len 值大于或等于 n，即包含的项目数大于或等于 n；例："gte=6"
- lt：数字小于 n，或者或者数组、切片、map 的 len 值小于 n，即包含的项目数小于 n；例："lt=6"
- lte：数字小于或等于 n，或者或者数组、切片、map 的 len 值小于或等于 n，即包含的项目数小于或等于 n；例："lte=6"
- oneof：只能是列举出的值其中一个，这些值必须是数值或字符串，以空格分隔，如果字符串中有空格，将字符串用单引号包围；例："oneof=red green"

## 跨字段验证

- eqfield=Field: 必须等于 Field 的值；
- nefield=Field: 必须不等于 Field 的值；
- gtfield=Field: 必须大于 Field 的值；
- gtefield=Field: 必须大于等于 Field 的值；
- ltfield=Field: 必须小于 Field 的值；
- ltefield=Field: 必须小于等于 Field 的值；
- eqcsfield=Other.Field: 必须等于 struct Other 中 Field 的值；
- necsfield=Other.Field: 必须不等于 struct Other 中 Field 的值；
- gtcsfield=Other.Field: 必须大于 struct Other 中 Field 的值；
- gtecsfield=Other.Field: 必须大于等于 struct Other 中 Field 的值；
- ltcsfield=Other.Field: 必须小于 struct Other 中 Field 的值；
- ltecsfield=Other.Field: 必须小于等于 struct Other 中 Field 的值；

## 字符串约束

- excludesall：不包含参数中任意的 UNICODE 字符，例如 excludesall=ab；
- excludesrune：不包含参数表示的 rune 字符，excludesrune=asong；
- startswith：以参数子串为前缀，例如 startswith=hi；
- endswith：以参数子串为后缀，例如 endswith=bye。
- contains：包含参数子串，例如 contains=email；
- containsany：包含参数中任意的 UNICODE 字符，例如 containsany=ab；
- containsrune：包含参数表示的 rune 字符，例如`containsrune=asong；
- excludes：不包含参数子串，例如 excludes=email；

## 例子

- `validate:"required"` //非空
- `validate:"gte=0,lte=130"` // 0<=值<=130
- `validate:"required,email"` //非空，email 格式
- `validate:"numeric,len=11"` //数字类型，长度为 11
