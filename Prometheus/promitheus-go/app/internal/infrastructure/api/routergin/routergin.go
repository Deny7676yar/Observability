package routergin

import (
	"fmt"
	"net/http"

	"github.com/Deny7676yar/observability/Prometheus/promitheus-go/app/internal/infrastructure/api/handler"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/render"
	"github.com/google/uuid"
)

type RouterGin struct {
	*gin.Engine
	hs *handler.Handlers
}

func NewRouterGin(hs *handler.Handlers) *RouterGin {
	r := gin.Default()
	ret := &RouterGin{
		hs: hs,
	}

	r.POST("/create", ret.CreateLink)
	r.GET("/read/:id", ret.ReadLinkRank)
	r.DELETE("/delete/:id", ret.DeleteLink)
	r.GET("/search/:q", ret.SearchLink)

	r.GET("/visitors", ret.GetLongURL)

	ret.Engine = r
	return ret
}

type Link handler.Link

func (rt *RouterGin) CreateLink(c *gin.Context) {
	ru := Link{}
	if err := c.ShouldBindJSON(&ru); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	l, err := rt.hs.CreateLink(c.Request.Context(), handler.Link(ru))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, l)
}

func (rt *RouterGin) ReadLinkRank(c *gin.Context) { //nolint
	sid := c.Param("id")

	uid, err := uuid.Parse(sid)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	l, err := rt.hs.ReadLinkRank(c.Request.Context(), uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, l)
}

func (rt *RouterGin) DeleteLink(c *gin.Context) { //nolint
	sid := c.Param("id")

	uid, err := uuid.Parse(sid)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	l, err := rt.hs.DeleteLink(c.Request.Context(), uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, l)
}

func (rt *RouterGin) SearchLink(c *gin.Context) {
	q := c.Param("id")
	w := c.Writer
	fmt.Fprintln(w, "[")
	comma := false
	err := rt.hs.SearchLink(c.Request.Context(), q, func(u handler.Link) error {
		if comma {
			fmt.Fprintln(w, ",")
		} else {
			comma = true
		}
		(render.JSON{Data: u}).Render(w) //nolint
		w.Flush()
		return nil
	})
	if err != nil {
		if comma {
			fmt.Fprint(w, ",")
		}
		(render.JSON{Data: err}).Render(w) //nolint
	}
	fmt.Fprintln(w, "]")
}

func (rt *RouterGin) GetLongURL(c *gin.Context) {
	s := c.Param("resultLink")

	if err := c.ShouldBind(s); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

	l, err := rt.hs.GetLongURL(c.Request.Context(), s)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, l)
}
