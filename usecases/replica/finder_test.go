package replica

import (
	"context"
	"testing"

	"github.com/go-openapi/strfmt"
	"github.com/semi-technologies/weaviate/entities/additional"
	"github.com/semi-technologies/weaviate/entities/models"
	"github.com/semi-technologies/weaviate/entities/search"
	"github.com/semi-technologies/weaviate/entities/storobj"
	"github.com/stretchr/testify/assert"
)

func TestFinderReplicaNotFound(t *testing.T) {
	factory := newFakeFactory("C1", "S", []string{})
	f := factory.newFinder()
	_, err := f.Find(context.Background(), "ONE", "S", "id", nil, additional.Properties{})
	assert.ErrorIs(t, err, errReplicaNotFound)
}

func object(id strfmt.UUID, lastTime int64) *storobj.Object {
	return &storobj.Object{
		Object: models.Object{
			ID:                 id,
			LastUpdateTimeUnix: lastTime,
		},
	}
}

func TestFinderFind(t *testing.T) {
	var (
		id    = strfmt.UUID("123")
		cls   = "C1"
		shard = "SH1"
		nodes = []string{"A", "B", "C"}
		ctx   = context.Background()
		obj   = object(id, 3)
		adds  = additional.Properties{}
		proj  = search.SelectProperties{}
	)

	t.Run("All", func(t *testing.T) {
		f := newFakeFactory("C1", shard, nodes)
		finder := f.newFinder()
		for _, n := range nodes {
			f.Client.On("GetObject", anyVal, n, cls, shard, id, proj, adds).Return(obj, nil)
		}
		got, err := finder.Find(ctx, "ALL", shard, id, proj, adds)
		assert.Nil(t, err)
		assert.Equal(t, obj, got)
	})
	t.Run("AllButLastOne", func(t *testing.T) {
		f := newFakeFactory("C1", shard, nodes)
		finder := f.newFinder()
		for _, n := range nodes[:len(nodes)-1] {
			f.Client.On("GetObject", anyVal, n, cls, shard, id, proj, adds).Return(obj, nil)
		}
		f.Client.On("GetObject", anyVal, nodes[len(nodes)-1], cls, shard, id, proj, adds).Return(object(id, 1), nil)
		got, err := finder.Find(ctx, "ALL", shard, id, proj, adds)
		assert.NotNil(t, err)
		assert.Nil(t, got)
		assert.Contains(t, err.Error(), "A: 3, B: 3, C: 1")
	})
	t.Run("AllButFirstOne", func(t *testing.T) {
		f := newFakeFactory("C1", shard, nodes)
		finder := f.newFinder()
		for _, n := range nodes[1:] {
			f.Client.On("GetObject", anyVal, n, cls, shard, id, proj, adds).Return(obj, nil)
		}
		f.Client.On("GetObject", anyVal, nodes[0], cls, shard, id, proj, adds).Return(object(id, 1), nil)
		got, err := finder.Find(ctx, "ALL", shard, id, proj, adds)
		assert.NotNil(t, err)
		assert.Nil(t, got)
		assert.Contains(t, err.Error(), "A: 1, B: 3, C: 3")
	})
	t.Run("NoQuorum", func(t *testing.T) {
		f := newFakeFactory("C1", shard, nodes)
		finder := f.newFinder()
		for i, n := range nodes {
			obj := object(id, int64(i+1))
			f.Client.On("GetObject", anyVal, n, cls, shard, id, proj, adds).Return(obj, nil)
		}
		f.Client.On("GetObject", anyVal, nodes[0], cls, shard, id, proj, adds).Return(object(id, 1), nil)
		got, err := finder.Find(ctx, "QUORUM", shard, id, proj, adds)
		assert.Nil(t, got)
		assert.Contains(t, err.Error(), "A: 1, B: 2, C: 3")
	})
}
