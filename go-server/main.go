package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
)


func main() {

	r := gin.Default()
	r.GET("/ok", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "ok",
		})
	})
	r.GET("/kill", func(c *gin.Context) {
		conn, _, err := c.Writer.Hijack()
		if err != nil {
			fmt.Printf("BAD QUERY ERROR: %s", err)
			return
		}
		if err := conn.Close(); err != nil {
			fmt.Printf("cannot close connection : %s", err)
			return
		}
	})
	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}