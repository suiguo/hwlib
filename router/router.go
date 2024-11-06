package router

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type RouterGroup string

type MethodName string

const (
	Any    MethodName = "Any"
	GET    MethodName = "GET"
	POST   MethodName = "POST"
	DELETE MethodName = "DELETE"
	PATCH  MethodName = "PATCH"
	PUT    MethodName = "PUT"
)

type Router struct {
	router map[RouterGroup][]EndPointGroup
}

type HandlerFunc func(c *gin.Context) any

type Url struct {
	Path   string
	Method MethodName
}
type EndPointGroup interface {
	Urls() []Url
	Router(string) HandlerFunc
}

func NewRouter() *Router {
	return &Router{router: make(map[RouterGroup][]EndPointGroup)}
}
func (r *Router) RegisterRouter(group RouterGroup, param EndPointGroup) *Router {
	if param == nil {
		return r
	}
	if r.router == nil {
		r.router = make(map[RouterGroup][]EndPointGroup)
	}
	_, ok := r.router[group]
	if !ok {
		r.router[group] = make([]EndPointGroup, 0)
	}
	r.router[group] = append(r.router[group], param)
	return r
}

func (r *Router) Package(g *gin.Engine) *gin.Engine {
	for key, roues := range r.router {
		group := g.Group(string(key))
		for idx := range roues {
			h := roues[idx]
			urls := h.Urls()
			for _, val := range urls {
				funcHandler := h.Router(val.Path)
				if funcHandler != nil {
					do := func(ctx *gin.Context) {
						data := funcHandler(ctx)
						switch d := data.(type) {
						case string:
							ctx.String(http.StatusOK, d)
						default:
							ctx.SecureJSON(200, d)
						}
					}
					switch val.Method {
					case Any:
						group.Any(val.Path, do)
					case GET:
						group.GET(val.Path, do)
					case POST:
						group.POST(val.Path, do)
					case DELETE:
						group.DELETE(val.Path, do)
					case PATCH:
						group.PATCH(val.Path, do)
					case PUT:
						group.PUT(val.Path, do)
					}
				}
			}
		}
	}
	return g
}
