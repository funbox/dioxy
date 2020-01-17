package daemon

import (
	"pkg.re/essentialkaos/ek.v10/log"

	"github.com/valyala/fasthttp"
)

// ////////////////////////////////////////////////////////////////////////////////// //

// startHTTPServer start HTTP server
func startHTTPServer(ip, port string) error {
	addr := ip + ":" + port

	log.Info("HTTP server is started on %s", addr)

	server := fasthttp.Server{
		Handler: fastHTTPHandler,
		Name:    APP + "/" + VER,
	}

	return server.ListenAndServe(addr)
}

// fastHTTPHandler handler for fast http requests
func fastHTTPHandler(ctx *fasthttp.RequestCtx) {
	defer requestRecover(ctx)

	path := string(ctx.Path())

	if path != "/" {
		ctx.SetStatusCode(404)
		return
	}

	statusHandler(ctx)
}

// requestRecover recover panic in request
func requestRecover(ctx *fasthttp.RequestCtx) {
	r := recover()

	if r != nil {
		log.Error("Recovered internal error in HTTP request handler: %v", r)
		ctx.SetStatusCode(501)
	}
}

// ////////////////////////////////////////////////////////////////////////////////// //

// statusHandler is status request handler
func statusHandler(ctx *fasthttp.RequestCtx) {
	ctx.Response.Header.Set("Content-Type", "application/json")
	ctx.Response.Header.Set("Cache-Control", "no-cache, no-store, must-revalidate")
	ctx.Response.Header.Set("Pragma", "no-cache")
	ctx.Response.Header.Set("Expires", "0")

	body, err := datastore.MarshalJSON()

	if err != nil {
		log.Warn("Internal Server Error while processing user request: %s", err.Error())
		ctx.SetStatusCode(500)
		return
	}

	log.Info("Respond with status 200")
	ctx.Write(body)
	ctx.SetStatusCode(200)
}
