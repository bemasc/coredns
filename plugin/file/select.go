package file

import (
	"context"
	"crypto/rand"
	"math/big"
	"net"
	"slices"
	"strconv"
	"strings"

	"github.com/coredns/coredns/plugin/metadata"
	"github.com/coredns/coredns/request"
	"github.com/miekg/dns"
)

// SelectorKey uniquely identifies an applicable selection policy.
type SelectorKey struct {
	Selector string
	Base     string
}

// NewSelectorKey creates a new SelectorKey from a SELECT record.
// Note that distinct SELECT records can have the same key, if they
// rely on the same Selector.
func NewSelectorKey(rr dns.SELECT) SelectorKey {
	return SelectorKey{
		Selector: rr.Selector,
		Base:     rr.Base,
	}
}

// SelectorCriteria provides the inputs that can be used by a Selector.
type SelectorCriteria struct {
	Ctx      context.Context
	Src      net.IP
	Protocol string
	Ecs      net.IPNet
}

// getSelectorCriteria extracts the SelectorCriteria for the given request.
func getSelectorCriteria(ctx context.Context, state request.Request) SelectorCriteria {
	var subnet net.IPNet
	edns := state.Req.IsEdns0()
	if edns != nil {
		for _, option := range edns.Option {
			if option.Option() == dns.EDNS0SUBNET {
				ecs := option.(*dns.EDNS0_SUBNET)
				subnet.IP = ecs.Address
				subnet.Mask = net.CIDRMask(int(ecs.SourceNetmask), 8*len(ecs.Address))
			}
		}
	}
	return SelectorCriteria{
		Ctx:      ctx,
		Src:      net.ParseIP(state.IP()),
		Protocol: state.Proto(),
		Ecs:      subnet,
	}
}

type SelectionResult struct {
	Option               string
	EcsScopePrefixLength uint8
}

// Selector is the abstract representation of a Selector.
type Selector interface {
	// Select returns one of the options, or "" to indicate the default.
	Select(SelectorCriteria) SelectionResult
}

type protocolSelector struct {
	options []string
}

type randomSelector struct {
	options []string
}

// NewRandomSelector returns a selector that chooses one of the options
// at random.
func NewRandomSelector(options []string) Selector {
	return randomSelector{options: options}
}

func (s randomSelector) Select(criteria SelectorCriteria) SelectionResult {
	if len(s.options) == 0 {
		return SelectionResult{}
	}
	i, err := rand.Int(rand.Reader, big.NewInt(int64(len(s.options))))
	if err != nil {
		return SelectionResult{}
	}
	return SelectionResult{Option: s.options[i.Int64()]}
}

type metadataSelector struct {
	key     string
	options []string
}

func NewMetadataSelector(key string, options []string) Selector {
	return metadataSelector{key: key, options: options}
}

func (s metadataSelector) Select(criteria SelectorCriteria) SelectionResult {
	var ret SelectionResult
	f := metadata.ValueFunc(criteria.Ctx, s.key)
	if f == nil {
		log.Infof("No metadata func for key %s", s.key)
		return ret
	}
	option := strings.ToLower(f())
	if option == "" {
		log.Infof("No metadata value for key '%s'", s.key)
		return ret
	}

	if slices.Contains(s.options, option) {
		ret.Option = option
	} else {
		log.Infof("Unrecognized option '%s' for key '%s'", option, s.key)
	}

	// CONVENTION: Adding "/_ecs-scope" to the metadata key creates a key
	// that indicates the ECS scope prefix length for the original value
	// (i.e. the range of IP addresses that should also receive the same value).
	ecsMetadataFunc := metadata.ValueFunc(criteria.Ctx, s.key+"/_ecs-scope")
	if ecsMetadataFunc != nil {
		log.Infof("Found ECS scope %s", ecsMetadataFunc())
		ecsScopePrefixLength, _ := strconv.Atoi(ecsMetadataFunc())
		ret.EcsScopePrefixLength = uint8(ecsScopePrefixLength)
	} else {
		log.Infof("No ECS scope for key '%s'", s.key)
	}

	return ret
}

func GetSelector(name string, options []string) Selector {
	switch name {
	case "@random":
		return NewRandomSelector(options)
	default:
		return NewMetadataSelector(name, options)
	}
}
