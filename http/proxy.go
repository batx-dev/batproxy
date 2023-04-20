package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/batx-dev/batproxy"
	restfulspec "github.com/emicklei/go-restful-openapi/v2"
	"github.com/emicklei/go-restful/v3"
)

func (s *Server) proxyService(ws *restful.WebService) {
	tags := []string{"proxies"}

	ws.Route(ws.POST("/proxies").To(s.createProxy).
		Doc("create a reverse proxy rule").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Reads(batproxy.Proxy{}).
		Writes(batproxy.Proxy{}).
		Returns(201, "Created", batproxy.Proxy{}).
		Returns(409, "Conflict", batproxy.Error{}))

	ws.Route(ws.GET("/proxies").To(s.listProxies).
		Doc("list proxies").
		Param(ws.QueryParameter("proxy_id", "the proxy id (name) of reverse proxy").
			DataType("string")).
		Param(ws.QueryParameter("page_size", "sets the maximum number of proxies to be returned").
			DataType("integer").DefaultValue("1000")).
		Param(ws.QueryParameter("page_token", "page_token may be filled in with the next_page_token from a previous list call").
			DataType("string")).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes(batproxy.ListProxiesPage{}).
		Returns(200, "OK", batproxy.ListProxiesPage{}))

	ws.Route(ws.DELETE("/proxies/{proxy_id}").To(s.deleteProxy).
		// docs
		Doc("delete a reverse proxy").
		Param(ws.PathParameter("proxy_id", "the id of the reverse proxy").
			DataType("string").Required(true)).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Returns(204, "NoContent", nil))
}

func (s *Server) createProxy(req *restful.Request, res *restful.Response) {
	opts := batproxy.CreateProxyOptions{}
	decoder.IgnoreUnknownKeys(true)
	if err := decoder.Decode(&opts, req.Request.URL.Query()); err != nil {
		Error(res.ResponseWriter, req.Request, batproxy.Errorf(batproxy.EINVALID, "%v", err))
		return
	}

	proxy := &batproxy.Proxy{}
	if err := req.ReadEntity(proxy); err != nil {
		Error(res.ResponseWriter, req.Request, batproxy.Errorf(batproxy.EINVALID, "%v", err))
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
		s.logger.Error("proxy", "err", err, "req", req.Request.URL)
	}
}

func (s *Server) listProxies(req *restful.Request, res *restful.Response) {
	opts := batproxy.ListProxiesOptions{}
	decoder.IgnoreUnknownKeys(true)
	if err := decoder.Decode(&opts, req.Request.URL.Query()); err != nil {
		Error(res.ResponseWriter, req.Request, batproxy.Errorf(batproxy.EINVALID, "%v", err))
		return
	}

	page, err := s.ProxyService.ListProxies(req.Request.Context(), opts)
	if err != nil {
		Error(res.ResponseWriter, req.Request, err)
		return
	}

	err = res.WriteEntity(page)
	if err != nil {
		s.logger.Error("proxy", "err", err, "req", req.Request.URL)
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

type ProxyService struct {
	Client *Client
}

func NewProxyService(client *Client) *ProxyService {
	return &ProxyService{Client: client}
}

func (s *ProxyService) CreateProxy(ctx context.Context, proxy *batproxy.Proxy, opts batproxy.CreateProxyOptions) error {
	body, err := json.Marshal(proxy)
	if err != nil {
		return batproxy.Errorf(batproxy.EINVALID, "json encode: %v", err)
	}

	query := url.Values{}
	if err := encoder.Encode(opts, query); err != nil {
		return batproxy.Errorf(batproxy.EINVALID, "query encode: %v", err)
	}

	req, err := s.Client.newRequest(ctx, "POST",
		"/api/v1beta1/proxies?"+query.Encode(),
		bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("http new request: %v", err)
	}

	res, err := s.Client.Do(req)
	if err != nil {
		return err
	} else if res.StatusCode != http.StatusCreated {
		return parseResponseError(res)
	}
	defer res.Body.Close()

	if err := json.NewDecoder(res.Body).Decode(proxy); err != nil {
		return fmt.Errorf("json decode: %v", err)
	}

	return nil
}

func (s *ProxyService) ListProxies(ctx context.Context, opts batproxy.ListProxiesOptions) (*batproxy.ListProxiesPage, error) {
	query := url.Values{}
	err := encoder.Encode(opts, query)
	if err != nil {
		return nil, batproxy.Errorf(batproxy.EINVALID, "query encode: %v", err)
	}

	req, err := s.Client.newRequest(ctx, "GET",
		"/api/v1beta1/proxies?"+query.Encode(), nil)
	if err != nil {
		return nil, fmt.Errorf("http new request: %v", err)
	}

	res, err := s.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http do request: %v", err)
	} else if res.StatusCode != http.StatusOK {
		return nil, parseResponseError(res)
	}
	defer res.Body.Close()

	var page batproxy.ListProxiesPage
	if err := json.NewDecoder(res.Body).Decode(&page); err != nil {
		return nil, fmt.Errorf("json decode: %v", err)
	}

	return &page, nil
}

func (s *ProxyService) UpdateProxy(req *restful.Request, res *restful.Response) {
	panic("implement me")
}

func (s *ProxyService) DeleteProxy(ctx context.Context, proxyID string) error {
	req, err := s.Client.newRequest(ctx, "DELETE",
		"/api/v1beta1/proxies/"+proxyID+"?",
		nil)
	if err != nil {
		return fmt.Errorf("http new request: %v", err)
	}

	res, err := s.Client.Do(req)
	if err != nil {
		return err
	} else if res.StatusCode != http.StatusNoContent {
		return parseResponseError(res)
	}
	defer res.Body.Close()

	return nil
}
