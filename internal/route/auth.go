package routes

import (
	"encoding/json"

	"github.com/valyala/fasthttp"
)

type authDefaultImpl struct {
	r *Router
}

func registerAuth(r *Router) {
	impl := authDefaultImpl{r}

	r.rtr.POST("/api/v1/auth", fasthttp.CompressHandler(noCache(impl.auth)))
	r.rtr.POST("/api/v1/register", fasthttp.CompressHandler(noCache(impl.register)))
	r.rtr.GET("/api/v1/file", fasthttp.CompressHandler(impl.getFile))
}

func (impl *authDefaultImpl) auth(ctx *fasthttp.RequestCtx) {
	type requestBody struct {
		Login string `json:"login"`
		Pass  string `json:"pass"`
	}

	var request requestBody
	err := json.Unmarshal(ctx.PostBody(), &request)
	if err != nil {
		ctx.SetStatusCode(400)
		ctx.SetBody([]byte(err.Error()))
		return
	}

	userId, err := impl.r.storage.Auth(request.Login, request.Pass)
	if err != nil {
		if err.Error() == "user not found" {
			ctx.SetStatusCode(404)
			return
		} else {
			ctx.SetStatusCode(500)
			return
		}
	}

	token, err := impl.r.messager.GetTokenForUser(userId)
	if err != nil {
		if err.Error() == "login alrady exit" {
			ctx.SetStatusCode(409)
			return
		} else {
			ctx.SetStatusCode(500)
			return
		}
	}

	ctx.SetStatusCode(200)
	ctx.SetBody([]byte(token))
}

func (impl *authDefaultImpl) register(ctx *fasthttp.RequestCtx) {
	type requestBody struct {
		Login string `json:"login"`
		Pass  string `json:"pass"`
	}

	var request requestBody
	err := json.Unmarshal(ctx.PostBody(), &request)
	if err != nil {
		ctx.SetStatusCode(400)
		ctx.SetBody([]byte(err.Error()))
		return
	}

	userId, err := impl.r.storage.Register(request.Login, request.Pass)
	if err != nil {
		if err.Error() == "login alrady exit" {
			ctx.SetStatusCode(409)
			return
		} else {
			ctx.SetStatusCode(500)
			return
		}
	}

	token, err := impl.r.messager.GetTokenForUser(userId)
	if err != nil {
		ctx.SetStatusCode(500)
		return
	}

	ctx.SetStatusCode(200)
	ctx.SetBody([]byte(token))
}

func (impl *authDefaultImpl) getFile(ctx *fasthttp.RequestCtx) {

}
