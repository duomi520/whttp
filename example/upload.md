# 上传文件

```go
package main

import (
"io"
"log/slog"
"net/http"
"os"
"strings"
"text/template"
"time"
"github.com/duomi520/whttp"
)
func main() {
route := whttp.NewRoute(nil)
 //配置服务
srv := &http.Server{
Handler:        route.Mux,
MaxHeaderBytes: 1 << 20,
}
route.Static("/", "upload.html")
route.POST("/upload", func(c *whttp.HTTPContext) {
//在使用r.MultipartForm前必须先调用ParseMultipartForm方法，参数为最大缓存
c.Request.ParseMultipartForm(32 << 20)
if c.Request.MultipartForm != nil && c.Request.MultipartForm.File != nil {
//获取所有上传文件信息
fhs := c.Request.MultipartForm.File["userfile"]
//循环对每个文件进行处理
for _, fheader := range fhs {
str := fheader.Filename
//替换"/"
str = strings.Replace(str, "/", "", -1)
//替换"\"
str = strings.Replace(str, "\\", "", -1)
//避免XSS
str = template.HTMLEscapeString(str)
//设置文件名
newFileName := "./upload/" + str
//打开上传文件
uploadFile, err := fheader.Open()
defer uploadFile.Close()
if err != nil {
c.Error(err.Error())
return
}
//保存文件
saveFile, err := os.OpenFile(newFileName, os.O_WRONLY|os.O_CREATE, 0666)
defer saveFile.Close()
if err != nil {
c.Error(err.Error())
return
}
io.Copy(saveFile, uploadFile)
}
}
})
if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
slog.Error(err.Error())
}
}

```

```html
<html>
  <head>
    <meta charset="utf-8" />
  </head>

  <body>
    <form
      id="uploadForm"
      name="uploadForm"
      role="form"
      enctype="multipart/form-data"
      action="/upload"
      method="POST"
    >
      <input type="file" id="userfile" name="userfile" multiple />
      <div>
        <button type="button" id="upload" onclick="UpLoadFile()">上传</button>
      </div>
    </form>
  </body>

  <script>
    var lock = false;
    function UpLoadFile() {
      if (lock) return;
      var userfile = document.getElementById("userfile").value;
      if (userfile.length == 0) {
        return;
      }
      // FormData 对象
      var form = new FormData(document.forms.namedItem("uploadForm"));
      // XMLHttpRequest 对象
      var xhr = new XMLHttpRequest();
      xhr.open("post", "/upload", true);
      // xhr.upload 储存上传过程中的信息
      xhr.upload.onprogress = function (ev) {
        if (ev.lengthComputable) {
        }
      };
      xhr.onload = function (oEvent) {
        if (xhr.status == 200) {
        }
        if (xhr.status == 500) {
        }
      };
      xhr.onerror = function (oEvent) {};
      xhr.onreadystatechange = function () {
        switch (xhr.readyState) {
          case 4:
            lock = false;
            break;
        }
      };
      lock = true;
      //发送数据
      xhr.send(form);
    }
  </script>
</html>
```
