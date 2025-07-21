package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func (app *application) routes() http.Handler {
	g := gin.Default()

	v1 := g.Group("/api/v1")
	{
		v1.GET("/subscription", app.listSubscription)
		v1.POST("/subscription", app.createSubscription)
		v1.GET("/subscription/:id", app.getSubscription)
		v1.PUT("/subscription/:id", app.updateSubscription)
		v1.DELETE("/subscription/:id", app.deleteSubscription)
		v1.GET("/subscription/period-price/:period", app.getPeriodPrice)
	}

	g.GET("/swagger/*any", func(c *gin.Context) {
		if c.Request.RequestURI == "/swagger/" {
			c.Redirect(302, "/swagger/index.html")
			return
		}
		ginSwagger.WrapHandler(swaggerFiles.Handler, ginSwagger.URL("http://localhost:8080/swagger/doc.json"))(c)
	})

	return g
}
