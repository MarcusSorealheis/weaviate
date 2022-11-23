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

package db

import (
	"context"
	"fmt"
	"sync"

	"github.com/go-openapi/strfmt"
	"github.com/google/uuid"
	"github.com/semi-technologies/weaviate/entities/storobj"
	"github.com/semi-technologies/weaviate/usecases/objects"
	"github.com/semi-technologies/weaviate/usecases/replica"
)

type replicaTask func(context.Context) interface{}

type pendingReplicaTasks struct {
	sync.Mutex
	Tasks map[string]replicaTask
}

func (p *pendingReplicaTasks) clear() {
	p.Lock()
	// TODO: can we postpone deletion until all pending replications are done
	p.Tasks = nil
	p.Unlock()
}

func (p *pendingReplicaTasks) get(requestID string) (replicaTask, bool) {
	p.Lock()
	defer p.Unlock()
	t, ok := p.Tasks[requestID]
	return t, ok
}

func (p *pendingReplicaTasks) set(requestID string, task replicaTask) {
	p.Lock()
	p.Tasks[requestID] = task
	p.Unlock()
}

func (p *pendingReplicaTasks) delete(requestID string) {
	p.Lock()
	delete(p.Tasks, requestID)
	p.Unlock()
}

func (s *Shard) commit(ctx context.Context, requestID string, backupReadLock *sync.RWMutex) interface{} {
	f, _ := s.replicationMap.get(requestID)
	defer s.replicationMap.delete(requestID)
	backupReadLock.RLock()
	defer backupReadLock.RUnlock()
	return f(ctx)
}

func (s *Shard) abort(ctx context.Context, requestID string) replica.SimpleResponse {
	s.replicationMap.delete(requestID)
	return replica.SimpleResponse{}
}

func (s *Shard) preparePutObject(ctx context.Context, requestID string, object *storobj.Object) replica.SimpleResponse {
	uuid, err := parseBytesUUID(object.ID())
	if err != nil {
		return replica.SimpleResponse{Errors: []replica.Error{*replica.NewError(replica.StatusPreconditionFailed, err.Error())}}
	}
	task := func(ctx context.Context) interface{} {
		resp := replica.SimpleResponse{}
		if err := s.putOne(ctx, uuid, object); err != nil {
			resp.Errors = []replica.Error{*replica.NewError(replica.StatusConflict, err.Error())}
		}
		return resp
	}
	s.replicationMap.set(requestID, task)
	return replica.SimpleResponse{}
}

func (s *Shard) prepareMergeObject(ctx context.Context, requestID string, doc *objects.MergeDocument) replica.SimpleResponse {
	uuid, err := parseBytesUUID(doc.ID)
	if err != nil {
		return replica.SimpleResponse{Errors: []replica.Error{*replica.NewError(replica.StatusPreconditionFailed, err.Error())}}
	}
	task := func(ctx context.Context) interface{} {
		resp := replica.SimpleResponse{}
		if err := s.merge(ctx, uuid, *doc); err != nil {
			resp.Errors = []replica.Error{*replica.NewError(replica.StatusConflict, err.Error())}
		}
		return resp
	}
	s.replicationMap.set(requestID, task)
	return replica.SimpleResponse{}
}

func (s *Shard) prepareDeleteObject(ctx context.Context, requestID string, uuid strfmt.UUID) replica.SimpleResponse {
	bucket, obj, idBytes, docID, err := s.canDeleteOne(ctx, uuid)
	if err != nil {
		return replica.SimpleResponse{Errors: []replica.Error{*replica.NewError(replica.StatusPreconditionFailed, err.Error())}}
	}
	task := func(ctx context.Context) interface{} {
		resp := replica.SimpleResponse{}
		if err := s.deleteOne(ctx, bucket, obj, idBytes, docID); err != nil {
			resp.Errors = []replica.Error{*replica.NewError(replica.StatusConflict, err.Error())}
		}
		return resp
	}
	s.replicationMap.set(requestID, task)
	return replica.SimpleResponse{}
}

func (s *Shard) preparePutObjects(ctx context.Context, requestID string, objects []*storobj.Object) replica.SimpleResponse {
	task := func(ctx context.Context) interface{} {
		rawErrs := s.putBatch(ctx, objects)
		resp := replica.SimpleResponse{Errors: make([]replica.Error, len(rawErrs))}
		for i, err := range rawErrs {
			if err != nil {
				resp.Errors[i] = *replica.NewError(replica.StatusConflict, err.Error())
			}
		}
		return resp
	}
	s.replicationMap.set(requestID, task)
	return replica.SimpleResponse{}
}

func (s *Shard) prepareDeleteObjects(ctx context.Context, requestID string, docIDs []uint64, dryRun bool) replica.SimpleResponse {
	task := func(ctx context.Context) interface{} {
		result := newDeleteObjectsBatcher(s).Delete(ctx, docIDs, dryRun)
		resp := replica.DeleteBatchResponse{
			Batch: make([]replica.UUID2Error, len(result)),
		}

		for i, r := range result {
			entry := replica.UUID2Error{UUID: string(r.UUID)}
			if err := r.Err; err != nil {
				entry.Error = *replica.NewError(replica.StatusConflict, err.Error())
			}
			resp.Batch[i] = entry
		}
		return resp
	}
	s.replicationMap.set(requestID, task)
	return replica.SimpleResponse{}
}

func (s *Shard) prepareAddReferences(ctx context.Context, requestID string, refs []objects.BatchReference) replica.SimpleResponse {
	task := func(ctx context.Context) interface{} {
		rawErrs := newReferencesBatcher(s).References(ctx, refs)
		resp := replica.SimpleResponse{Errors: make([]replica.Error, len(rawErrs))}
		for i, err := range rawErrs {
			if err != nil {
				resp.Errors[i] = *replica.NewError(replica.StatusConflict, err.Error())
			}
		}
		return resp
	}
	s.replicationMap.set(requestID, task)
	return replica.SimpleResponse{}
}

func parseBytesUUID(id strfmt.UUID) ([]byte, error) {
	uuid, err := uuid.Parse(string(id))
	if err != nil {
		return nil, fmt.Errorf("parse uuid %q: %w", id, err)
	}
	return uuid[:], nil
}
