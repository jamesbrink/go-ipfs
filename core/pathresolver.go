package core

import (
	"context"
	"errors"
	"strings"

	namesys "github.com/ipfs/go-ipfs/namesys"

	ipld "gx/ipfs/QmRL22E4paat7ky7vx9MLpR97JHHbFPrg3ytFQw6qp1y1s/go-ipld-format"
	path "gx/ipfs/QmWqh9oob7ZHQRwU5CdTqpnC8ip8BEkFNrwXRxeNo5Y7vA/go-path"
	resolver "gx/ipfs/QmWqh9oob7ZHQRwU5CdTqpnC8ip8BEkFNrwXRxeNo5Y7vA/go-path/resolver"
	logging "gx/ipfs/QmcuXC5cxs79ro2cUuHs4HQ2bkDLJUYokwL8aivcX6HW3C/go-log"
)

// ErrNoNamesys is an explicit error for when an IPFS node doesn't
// (yet) have a name system
var ErrNoNamesys = errors.New(
	"core/resolve: no Namesys on IpfsNode - can't resolve ipns entry")

// ResolveIPNS resolves /ipns paths
func ResolveIPNS(ctx context.Context, nsys namesys.NameSystem, p path.Path) (path.Path, error) {
	if strings.HasPrefix(p.String(), "/ipns/") {
		evt := log.EventBegin(ctx, "resolveIpnsPath")
		defer evt.Done()
		// resolve ipns paths

		// TODO(cryptix): we should be able to query the local cache for the path
		if nsys == nil {
			evt.Append(logging.LoggableMap{"error": ErrNoNamesys.Error()})
			return "", ErrNoNamesys
		}

		seg := p.Segments()

		if len(seg) < 2 || seg[1] == "" { // just "/<protocol/>" without further segments
			evt.Append(logging.LoggableMap{"error": path.ErrNoComponents.Error()})
			return "", path.ErrNoComponents
		}

		extensions := seg[2:]
		resolvable, err := path.FromSegments("/", seg[0], seg[1])
		if err != nil {
			evt.Append(logging.LoggableMap{"error": err.Error()})
			return "", err
		}

		respath, err := nsys.Resolve(ctx, resolvable.String())
		if err != nil {
			evt.Append(logging.LoggableMap{"error": err.Error()})
			return "", err
		}

		segments := append(respath.Segments(), extensions...)
		p, err = path.FromSegments("/", segments...)
		if err != nil {
			evt.Append(logging.LoggableMap{"error": err.Error()})
			return "", err
		}
	}
	return p, nil
}

// Resolve resolves the given path by parsing out protocol-specific
// entries (e.g. /ipns/<node-key>) and then going through the /ipfs/
// entries and returning the final node.
func Resolve(ctx context.Context, nsys namesys.NameSystem, r *resolver.Resolver, p path.Path) (ipld.Node, error) {
	p, err := ResolveIPNS(ctx, nsys, p)
	if err != nil {
		return nil, err
	}

	// ok, we have an IPFS path now (or what we'll treat as one)
	return r.ResolvePath(ctx, p)
}
