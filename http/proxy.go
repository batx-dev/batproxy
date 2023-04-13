package http

import (
	"net/http"

	"github.com/batx-dev/batproxy"
	restfulspec "github.com/emicklei/go-restful-openapi/v2"
	"github.com/emicklei/go-restful/v3"
)

func (s *Server) proxyService(ws *restful.WebService) {
	tags := []string{"proxies"}

	ws.Route(ws.POST("/proxies").To(s.createProxy).
		Doc("create a reverseProxy").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Reads(batproxy.Proxy{}).
		Writes(batproxy.Proxy{}).
		Returns(201, "Created", batproxy.Proxy{}).
		Returns(409, "Conflict", error.Error))

	ws.Route(ws.GET("/proxies").To(s.listProxies).
		Doc("list proxies").
		Param(ws.QueryParameter("proxy_id", "The uuid of reverseProxy").
			DataType("string")).
		Param(ws.QueryParameter("page_size", "Sets the maximum number of proxies to be returned").
			DataType("integer").DefaultValue("1000")).
		Param(ws.QueryParameter("page_token", "page_token may be filled in with the next_page_token from a previous list call").
			DataType("string")).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes(batproxy.ListProxiesPage{}).
		Returns(200, "OK", batproxy.ListProxiesPage{}))

	ws.Route(ws.DELETE("/proxies/{proxy_id}").To(s.deleteProxy).
		// docs
		Doc("delete a reverseProxy").
		Param(ws.PathParameter("proxy_id", "The id of the reverseProxy").
			DataType("string").Required(true)).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Returns(204, "NoContent", nil))
}

func (s *Server) createProxy(req *restful.Request, res *restful.Response) {
	opts := batproxy.CreateProxyOptions{}
	if err := decoder.Decode(&opts, req.Request.URL.Query()); err != nil {
		Error(res.ResponseWriter, req.Request, err)
		return
	}

	proxy := &batproxy.Proxy{}
	if err := req.ReadEntity(proxy); err != nil {
		Error(res.ResponseWriter, req.Request, err)
		return
	}

	if err := s.ProxyService.CreateProxy(
		req.Request.Context(),
		proxy,
		opts,
	); err != nil {
		Error(res.ResponseWriter, req.Request, err)
		return
	}

	err := res.WriteHeaderAndEntity(http.StatusCreated, proxy)
	if err != nil {
		s.logger.Error(err, "req", req.Request.URL)
	}
}

func (s *Server) listProxies(req *restful.Request, res *restful.Response) {
	opts := batproxy.ListProxiesOptions{}
	if err := decoder.Decode(&opts, req.Request.URL.Query()); err != nil {
		Error(res.ResponseWriter, req.Request, err)
		return
	}

	page, err := s.ProxyService.ListProxies(req.Request.Context(), opts)
	if err != nil {
		Error(res.ResponseWriter, req.Request, err)
		return
	}

	err = res.WriteEntity(page)
	if err != nil {
		s.logger.Error(err, "req", req.Request.URL)
	}
}

func (s *Server) updateProxy(req *restful.Request, res *restful.Response) {
	panic("implement me")
}

func (s *Server) deleteProxy(req *restful.Request, res *restful.Response) {
	if err := s.ProxyService.DeleteProxy(req.Request.Context(), req.PathParameter("proxy_id")); err != nil {
		Error(res.ResponseWriter, req.Request, err)
		return
	}

	res.WriteHeader(http.StatusNoContent)
}
