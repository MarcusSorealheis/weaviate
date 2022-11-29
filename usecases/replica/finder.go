package replica

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-openapi/strfmt"
	"github.com/pkg/errors"
	"github.com/semi-technologies/weaviate/entities/additional"
	"github.com/semi-technologies/weaviate/entities/search"
	"github.com/semi-technologies/weaviate/entities/storobj"
	"golang.org/x/sync/errgroup"
)

type Finder struct {
	RClient       // needed to commit and abort operation
	replicaFinder // host names of replicas
	class         string
}

func NewFinder(className string,
	stateGetter shardingState, nodeResolver nodeResolver,
	client RClient,
) *Finder {
	return &Finder{
		class: className,
		replicaFinder: &rFinder{
			schema:   stateGetter,
			resolver: nodeResolver,
			class:    className,
		},
		RClient: client,
	}
}

func (f *Finder) Find(ctx context.Context, replicas []string, cl int, index, shard string,
	id strfmt.UUID, props search.SelectProperties, additional additional.Properties,
) (*storobj.Object, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	responses := make(chan tuple, len(replicas))
	var g errgroup.Group
	for i, host := range replicas {
		i, host := i, host
		g.Go(func() error {
			o, err := f.GetObject(ctx, host, index, shard, id, props, additional)
			responses <- tuple{o, i, err}
			return nil
		})
	}
	go func() { g.Wait(); close(responses) }()

	return extractObject(responses, cl, replicas)
}

func extractObject(responses chan tuple, cl int, replicas []string) (*storobj.Object, error) {
	counters := make([]tuple, len(replicas))
	nnf := 0
	for r := range responses {
		if r.err != nil {
			counters[r.i] = tuple{nil, 0, r.err}
			continue
		} else if r.o == nil {
			nnf++
			continue
		}
		counters[r.i] = tuple{r.o, 1, nil}
		max := 0
		for i, c := range counters {
			if c.o != nil && i != r.i && c.o.LastUpdateTimeUnix() == r.o.LastUpdateTimeUnix() {
				counters[i].i++
				// counters[rs.i].counter++
			}
			if max < c.i {
				max = c.i
			}
			if max >= cl {
				return c.o, nil
			}
		}
	}
	if nnf == len(replicas) { // object doesn't exist
		return nil, nil
	}

	var sb strings.Builder
	for i, c := range counters {
		if i != 0 {
			sb.WriteString(", ")
		}
		if c.err != nil {
			fmt.Fprintf(&sb, "%s: %s", replicas[i], c.err.Error())
		} else if c.o == nil {
			fmt.Fprintf(&sb, "%s: 0", replicas[i])
		} else {
			fmt.Fprintf(&sb, "%s: %d", replicas[i], c.o.LastUpdateTimeUnix())
		}
	}
	return nil, errors.New(sb.String())
}

type tuple struct {
	o   *storobj.Object
	i   int
	err error
}
