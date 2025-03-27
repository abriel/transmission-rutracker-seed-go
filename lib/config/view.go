package config

import (
  "net/http"
  "github.com/gin-gonic/gin"
  "fmt"
)

func index(c *gin.Context) {
  c.HTML(http.StatusOK, "view.html", myconfig)
}

func save_transmission(c *gin.Context) {
  var username_block string

  u := c.PostForm("username")
  p := c.PostForm("password")
  if u != "" && p != "" {
    username_block = fmt.Sprintf("%s:%s@", u, p)
  }

  myconfig.BtUriRaw = fmt.Sprintf(
    "%s://%s%s:%s%s",
    c.PostForm("protocol"),
    username_block,
    c.PostForm("host"),
    c.PostForm("port"),
    c.PostForm("path"),
  )
  save_and_index(c)
}

func save_browser(c *gin.Context) {
  myconfig.UserAgent = c.PostForm("user_agent")
  save_and_index(c)
}

func save_and_index(c *gin.Context) {
  if err := myconfig.Save() ; err != nil {
    c.String(http.StatusBadRequest, "Error during saving the configuration: %v", err)
  } else {
    c.Redirect(http.StatusFound, "/")
  }  
}
