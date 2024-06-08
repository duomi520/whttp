# websocket

```go
package main
import (
"log/slog"
"net/http"
"time"
"github.com/duomi520/whttp"
"github.com/go-playground/validator"
"github.com/gorilla/websocket"
)
var upgrader = websocket.Upgrader{}
func echo(ctx *whttp.HTTPContext) {
c, err := upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
if err != nil {
slog.Error("upgrade:", err)
return
}
defer c.Close()
for {
mt, message, err := c.ReadMessage()
if err != nil {
slog.Error("read:", err)
break
}
slog.Info("recv:" + string(message))
err = c.WriteMessage(mt, message)
if err != nil {
slog.Error("write:", err)
break
}
}
}
func main() {
r := whttp.NewRoute(validator.New(), nil)
r.Static("/", "index.html")
r.GET("/echo", echo)
//启动服务
srv := &http.Server{
Handler:        r.Mux,
ReadTimeout:    3600 * time.Second,
WriteTimeout:   3600 * time.Second,
MaxHeaderBytes: 1 << 20,
}
if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
slog.Error(err.Error())
}
}
```

```html
<!DOCTYPE html>
<html>
  <head>
    <meta charset="utf-8" />
    <script>
      window.addEventListener("load", function (evt) {
        var output = document.getElementById("output");
        var input = document.getElementById("input");
        var ws;
        var print = function (message) {
          var d = document.createElement("div");
          d.innerHTML = message;
          output.appendChild(d);
        };
        document.getElementById("open").onclick = function (evt) {
          if (ws) {
            return false;
          }
          ws = new WebSocket("ws://127.0.0.1/echo");
          ws.onopen = function (evt) {
            print("OPEN");
          };
          ws.onclose = function (evt) {
            print("CLOSE");
            ws = null;
          };
          ws.onmessage = function (evt) {
            print("RESPONSE: " + evt.data);
          };
          ws.onerror = function (evt) {
            print("ERROR: " + evt.data);
          };
          return false;
        };
        document.getElementById("send").onclick = function (evt) {
          if (!ws) {
            return false;
          }
          print("SEND: " + input.value);
          ws.send(input.value);
          return false;
        };
        document.getElementById("close").onclick = function (evt) {
          if (!ws) {
            return false;
          }
          ws.close();
          return false;
        };
      });
    </script>
  </head>

  <body>
    <table>
      <tr>
        <td valign="top" width="50%">
          <p>
            Click "Open" to create a connection to the server, "Send" to send a
            message to the server and "Close" to close the connection. You can
            change the message and send multiple times.
          </p>

          <p></p>
          <form>
            <button id="open">Open</button>
            <button id="close">Close</button>
            <p>
              <input id="input" type="text" value="Hello world!" />
              <button id="send">Send</button>
            </p>
          </form>
        </td>
        <td valign="top" width="50%">
          <div id="output"></div>
        </td>
      </tr>
    </table>
  </body>
</html>
```
