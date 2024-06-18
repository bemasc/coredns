package protocol_metadata

import (
	"context"

	"github.com/coredns/caddy"
	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"
	"github.com/coredns/coredns/plugin/metadata"
	"github.com/coredns/coredns/request"

	"github.com/miekg/dns"
)

// protocol_metadata is a demo that shows how to use the "metadata" plugin
// system to define custom selection policies for SELECT.

const pluginName = "protocol_metadata"

type handler struct {
	Next plugin.Handler
}

// ServeDNS implements the plugin.Handler interface.
func (h handler) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	return plugin.NextOrFailure(pluginName, h.Next, ctx, w, r)
}

// Metadata implements the metadata.Provider Interface in the metadata plugin, and is used to store
// the name of the protocol associated with each request.
func (h handler) Metadata(ctx context.Context, state request.Request) context.Context {
	metadata.SetValueFunc(ctx, pluginName+"/name", state.Proto)
	return ctx
}

// Name implements the Handler interface.
func (h handler) Name() string { return pluginName }

func setup(c *caddy.Controller) error {
	h := handler{}
	dnsserver.GetConfig(c).AddPlugin(func(next plugin.Handler) plugin.Handler {
		h.Next = next
		return h
	})

	return nil
}

func init() { plugin.Register(pluginName, setup) }
