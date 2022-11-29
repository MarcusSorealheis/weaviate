//                           _       _
// __      _____  __ ___   ___  __ _| |_ ___
// \ \ /\ / / _ \/ _` \ \ / / |/ _` | __/ _ \
//  \ V  V /  __/ (_| |\ V /| | (_| | ||  __/
//   \_/\_/ \___|\__,_| \_/ |_|\__,_|\__\___|
//
//  Copyright © 2016 - 2022 SeMI Technologies B.V. All rights reserved.
//
//  CONTACT: hello@semi.technology
//

package replica

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/go-openapi/strfmt"
	"github.com/pkg/errors"
	"github.com/semi-technologies/weaviate/entities/additional"
	"github.com/semi-technologies/weaviate/entities/search"
	"github.com/semi-technologies/weaviate/entities/storobj"
	"golang.org/x/sync/errgroup"
)

var errReplicaNotFound = errors.New("no replica found")

// replicaFinder find nodes associated with a specific shard
type replicaFinder interface {
	FindReplicas(shardName string) []string
}

// readyOp asks a replica to be read to second phase commit
type readyOp func(ctx context.Context, host, requestID string) error

// readyOp asks a replica to execute the actual operation
type commitOp[T any] func(ctx context.Context, host, requestID string) (T, error)

type fetchOp[T any] func(_ context.Context, host string) (T, error)

// coordinator coordinates replication of write request
type coordinator[T any] struct {
	Client        // needed to commit and abort operation
	replicaFinder // host names of replicas
	class         string
	shard         string
	requestID     string
	// responses collect all responses of batch job
	responses []T
	nodes     []string
}

func newCoordinator[T any](r *Replicator, shard string) *coordinator[T] {
	return &coordinator[T]{
		Client: r.client,
		replicaFinder: &finder{
			schema:   r.stateGetter,
			resolver: r.resolver,
			class:    r.class,
		},
		class:     r.class,
		shard:     shard,
		requestID: time.Now().String(), // TODO: use a counter to build request id
	}
}

// broadcast sends write request to all replicas (first phase of a two-phase commit)
func (c *coordinator[T]) broadcast(ctx context.Context, replicas []string, op readyOp) error {
	errs := make([]error, len(replicas))
	var g errgroup.Group
	for i, replica := range replicas {
		i, replica := i, replica
		g.Go(func() error {
			errs[i] = op(ctx, replica, c.requestID)
			return nil
		})
	}
	g.Wait()
	var err error
	for _, err = range errs {
		if err != nil {
			break
		}
	}

	if err != nil {
		for _, node := range replicas {
			c.Abort(ctx, node, c.class, c.shard, c.requestID)
		}
	}

	return err
}

// commitAll tells replicas to commit pending updates related to a specific request
// (second phase of a two-phase commit)
func (c *coordinator[T]) commitAll(ctx context.Context, replicas []string, op commitOp[T]) error {
	var g errgroup.Group
	c.responses = make([]T, len(replicas))
	errs := make([]error, len(replicas))
	for i, replica := range replicas {
		i, replica := i, replica
		g.Go(func() error {
			resp, err := op(ctx, replica, c.requestID)
			c.responses[i], errs[i] = resp, err
			return nil
		})
	}
	g.Wait()
	var err error
	for _, err = range errs {
		if err != nil {
			return err
		}
	}

	return nil
}

// Replicate writes on all replicas of specific shard
func (c *coordinator[T]) Replicate(ctx context.Context, ask readyOp, com commitOp[T]) error {
	c.nodes = c.FindReplicas(c.shard)
	if len(c.nodes) == 0 {
		return fmt.Errorf("%w : class %q shard %q", errReplicaNotFound, c.class, c.shard)
	}
	if err := c.broadcast(ctx, c.nodes, ask); err != nil {
		return fmt.Errorf("broadcast: %w", err)
	}
	if err := c.commitAll(ctx, c.nodes, com); err != nil {
		return fmt.Errorf("commit: %w", err)
	}
	return nil
}

func (c *coordinator[T]) Fetch(ctx context.Context, replicas []string, cl int, index, shard string,
	id strfmt.UUID, props search.SelectProperties, additional additional.Properties,
) (*storobj.Object, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	responses := make(chan tuple, len(replicas))
	for i, host := range replicas {
		i, host := i, host
		go func() error {
			o, err := c.ReplicationClient.GetObject(ctx, host, index, shard, id, props, additional)
			responses <- tuple{o, i, err}
			return nil
		}()
	}
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
		}else{
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
