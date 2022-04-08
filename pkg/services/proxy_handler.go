package services

import (
	"fmt"
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

		connectResolver := newConsulConnectResolver(&config)
		connectTransport := newConsulConnectTransport(&config, service)

		return &connectProxyHandler{
			resolver:   connectResolver,
			intentions: client.Connect(),
			service:    service,
			handler:    proxy.NewHandlerFuncWithTransport(config.FaaS.GetReadTimeout(), connectTransport, connectResolver, logger),
		}, nil
	} else {
		consulResolver, err := resolver.NewConsulResolver(&config, logger)
		if err != nil {
			return nil, err
		}

		transport := newConsulCatalogTransport(&config)

		return &consulCatalogProxyHandler{
			resolver: consulResolver,
			handler:  proxy.NewHandlerFuncWithTransport(config.FaaS.GetReadTimeout(), transport, consulResolver, logger),
		}, err
	}
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

func newConsulConnectTransport(config *types.ProviderConfig, svc *consulconnect.Service) *http.Transport {
	transport := &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		DialTLS:               svc.HTTPDialTLS,
		MaxIdleConns:          config.FaaS.GetMaxIdleConns(),
		MaxIdleConnsPerHost:   config.FaaS.GetMaxIdleConnsPerHost(),
		IdleConnTimeout:       120 * time.Millisecond,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1500 * time.Millisecond,
	}
	_ = http2.ConfigureTransport(transport)
	return transport
}

func newConsulConnectResolver(config *types.ProviderConfig) *consulConnectResolver {
	return &consulConnectResolver{
		prefix:    config.Scheduling.JobPrefix,
		namespace: config.Scheduling.Namespace,
	}
}

type consulConnectResolver struct {
	prefix    string
	namespace string
}

func (ccr *consulConnectResolver) Resolve(functionName string) (url.URL, error) {
	urlAsString := fmt.Sprintf("https://%s.service.consul", fmt.Sprintf("%s%s", ccr.prefix, strings.TrimSuffix(functionName, "."+ccr.namespace)))
	parsedUrl, err := url.Parse(urlAsString)
	if err != nil {
		return url.URL{}, err
	}
	return *parsedUrl, nil
}

func (ccr *consulConnectResolver) ResolveAll(functionName string) ([]url.URL, error) {
	addr, err := ccr.Resolve(functionName)
	if err != nil {
		return nil, err
	}
	return []url.URL{addr}, nil
}

func newConsulCatalogTransport(config *types.ProviderConfig) *http.Transport {
	return &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   config.FaaS.GetReadTimeout(),
			KeepAlive: 1 * time.Second,
			DualStack: true,
		}).DialContext,
		MaxIdleConns:          config.FaaS.GetMaxIdleConns(),
		MaxIdleConnsPerHost:   config.FaaS.GetMaxIdleConnsPerHost(),
		IdleConnTimeout:       120 * time.Millisecond,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1500 * time.Millisecond,
	}
}
