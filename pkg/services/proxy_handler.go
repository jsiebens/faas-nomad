package services

import (
	"context"
	"fmt"
	"github.com/hashicorp/consul/agent/connect"
	consulapi "github.com/hashicorp/consul/api"
	consulconnect "github.com/hashicorp/consul/connect"
	"github.com/hashicorp/go-hclog"
	"github.com/jsiebens/faas-nomad/pkg/proxy"
	"github.com/jsiebens/faas-nomad/pkg/resolver"
	"github.com/jsiebens/faas-nomad/pkg/types"
	"golang.org/x/net/http2"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type ProxyHandler interface {
	Handler() (handler http.HandlerFunc)
	Resolver() resolver.ServiceResolver
	Close()
}

func NewProxyHandler(config types.ProviderConfig, logger hclog.Logger) (ProxyHandler, error) {
	consulResolver, err := resolver.NewConsulResolver(&config, logger)
	if err != nil {
		return nil, err
	}

	if config.Consul.ConnectAware {
		c := consulapi.DefaultConfig()

		c.Address = config.Consul.Addr
		c.Token = config.Consul.ACLToken
		if strings.HasPrefix(config.Consul.Addr, "https") {
			c.TLSConfig = consulapi.TLSConfig{
				CAFile:             config.Consul.CACert,
				CertFile:           config.Consul.ClientCert,
				KeyFile:            config.Consul.ClientKey,
				InsecureSkipVerify: !config.Consul.TLSSkipVerify,
			}
		}

		client, err := consulapi.NewClient(c)

		if err != nil {
			return nil, err
		}

		service, err := consulconnect.NewServiceWithLogger("faas-provider", client, logger)

		if err != nil {
			return nil, err
		}

		connectTransport := newConsulConnectTransport(&config, service, consulResolver)

		return &connectProxyHandler{
			resolver:   consulResolver,
			intentions: client.Connect(),
			service:    service,
			handler:    proxy.NewHandlerFuncWithTransport(config.FaaS.GetReadTimeout(), connectTransport, &baseUrlResolver{protocol: "https"}, logger),
		}, nil
	} else {
		transport := newConsulCatalogTransport(&config, consulResolver)

		return &consulCatalogProxyHandler{
			resolver: consulResolver,
			handler:  proxy.NewHandlerFuncWithTransport(config.FaaS.GetReadTimeout(), transport, &baseUrlResolver{protocol: "http"}, logger),
		}, err
	}
}

type baseUrlResolver struct {
	protocol string
}

func (b *baseUrlResolver) Resolve(functionName string) (url.URL, error) {
	parse, err := url.Parse(fmt.Sprintf("%s://%s", b.protocol, functionName))
	return *parse, err
}

type consulCatalogProxyHandler struct {
	resolver resolver.ServiceResolver
	handler  http.HandlerFunc
}

func (cc *consulCatalogProxyHandler) Resolver() resolver.ServiceResolver {
	return cc.resolver
}

func (cc *consulCatalogProxyHandler) Handler() http.HandlerFunc {
	return cc.handler
}

func (cc *consulCatalogProxyHandler) Close() {
}

type connectProxyHandler struct {
	intentions *consulapi.Connect
	resolver   resolver.ServiceResolver
	service    *consulconnect.Service
	handler    http.HandlerFunc
}

func (c *connectProxyHandler) Resolver() resolver.ServiceResolver {
	return c.resolver
}

func (c *connectProxyHandler) Handler() http.HandlerFunc {
	return c.handler
}

func (c *connectProxyHandler) Close() {
	_ = c.service.Close()
}

func newConsulConnectTransport(config *types.ProviderConfig, svc *consulconnect.Service, serviceResolver resolver.ServiceResolver) *http.Transport {
	dc := func(ctx context.Context, network, addr string) (net.Conn, error) {
		host := stripPort(addr)
		return svc.Dial(ctx, &internalResolver{functionName: host, resolver: serviceResolver})
	}

	transport := &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		DialTLSContext:        dc,
		MaxIdleConns:          config.FaaS.GetMaxIdleConns(),
		MaxIdleConnsPerHost:   config.FaaS.GetMaxIdleConnsPerHost(),
		IdleConnTimeout:       120 * time.Millisecond,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1500 * time.Millisecond,
	}
	_ = http2.ConfigureTransport(transport)
	return transport
}

type internalResolver struct {
	functionName string
	resolver     resolver.ServiceResolver
}

func (r *internalResolver) Resolve(ctx context.Context) (addr string, certURI connect.CertURI, err error) {
	return r.resolver.Resolve(r.functionName)
}

func newConsulCatalogTransport(config *types.ProviderConfig, serviceResolver resolver.ServiceResolver) *http.Transport {
	dialer := &net.Dialer{
		Timeout:   config.FaaS.GetReadTimeout(),
		KeepAlive: 1 * time.Second,
		DualStack: true,
	}

	dc := func(ctx context.Context, network, addr string) (net.Conn, error) {
		host := stripPort(addr)
		resolvedAddr, _, err := serviceResolver.Resolve(host)
		if err != nil {
			return nil, err
		}
		return dialer.DialContext(ctx, network, resolvedAddr)
	}

	return &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		DialContext:           dc,
		MaxIdleConns:          config.FaaS.GetMaxIdleConns(),
		MaxIdleConnsPerHost:   config.FaaS.GetMaxIdleConnsPerHost(),
		IdleConnTimeout:       120 * time.Millisecond,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1500 * time.Millisecond,
	}
}

func stripPort(hostport string) string {
	colon := strings.IndexByte(hostport, ':')
	if colon == -1 {
		return hostport
	}
	if i := strings.IndexByte(hostport, ']'); i != -1 {
		return strings.TrimPrefix(hostport[:i], "[")
	}
	return hostport[:colon]
}
