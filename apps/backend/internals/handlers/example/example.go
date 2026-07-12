package example

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func ExampleHandler(c *gin.Context) {
	name := c.PostForm("name")
	c.JSON(http.StatusCreated, gin.H{"user": name})
}
