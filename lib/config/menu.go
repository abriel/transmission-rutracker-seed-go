package config

import (
  "html/template"
  "net"
  "net/http"
  "github.com/gin-gonic/gin"
  "fmt"
  "math/rand/v2"
  "os"
  "embed"
  "io/fs"
)

var myconfig = new(Config)

//go:embed "*.html"
var htmlFS embed.FS

//go:embed "html_assets"
var staticAssets embed.FS

func Run() {
  myconfig, _ = New("skip_validation")

  listener := func() net.Listener {
    for {
      port := rand.IntN(65535 - 1024) + 1024
      l, err := net.Listen("tcp", fmt.Sprint("127.0.0.1:", port))
      if err == nil {
        fmt.Print("To access configuration page, open http://127.0.0.1:", port, "\n")
        return l
      }
    }
  }()
  defer listener.Close()

  if _, present := os.LookupEnv("DEBUG"); present {
    gin.SetMode(gin.DebugMode)
  } else {
    gin.SetMode(gin.ReleaseMode)
  }

  gin_engine := gin.Default()

  if t, err := template.ParseFS(htmlFS, "*") ; err != nil {
    fmt.Println("Fail to load the view.html template:", err)
    return
  } else {
    gin_engine.SetHTMLTemplate(t)
  }

  gin_engine.GET("/", index)

  if sub, err := fs.Sub(staticAssets, "html_assets") ; err != nil {
    fmt.Println("Cannot find embeded html static files:", err)
    return
  } else {
    gin_engine.StaticFS("/static", http.FS(sub))
  }

  gin_engine.POST("/save/transmission", save_transmission)
  gin_engine.POST("/save/browser", save_browser)

  if err := gin_engine.RunListener(listener); err != nil {
    fmt.Println("Error occurred on attaching Gin to a listener:", err)
  }

  return
}
