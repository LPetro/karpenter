/*
Copyright The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package orb

import (
	"context"
	"time"

	pb "sigs.k8s.io/karpenter/pkg/controllers/orb/proto"
)

type SchedulingMetadata struct {
	Action    string
	Timestamp time.Time
}

type key int

const (
	schedulingMetadataKey key = iota
)

func (sm SchedulingMetadata) GetTime() time.Time {
	return sm.Timestamp
}

// Returns a new context with the provided scheduling metadata.
func WithSchedulingMetadata(ctx context.Context, action string, timestamp time.Time) context.Context {
	switch action {
	case "normal-provisioning", "single-node-consolidation", "multi-node-consolidation", "drift":
		metadata := SchedulingMetadata{
			Action:    action,
			Timestamp: timestamp,
		}
		return context.WithValue(ctx, schedulingMetadataKey, metadata)
	default:
		return ctx
	}
}

// Retrieves the scheduling metadata from the context.
func GetSchedulingMetadata(ctx context.Context) (SchedulingMetadata, bool) {
	metadata, ok := ctx.Value(schedulingMetadataKey).(SchedulingMetadata)
	return metadata, ok
}

func protoSchedulingMetadata(metadata SchedulingMetadata) *pb.SchedulingMetadataMap_MappingEntry {
	return &pb.SchedulingMetadataMap_MappingEntry{
		Action:    metadata.Action,
		Timestamp: metadata.Timestamp.Format("2006-01-02_15-04-05"),
	}
}

func reconstructSchedulingMetadata(mappingEntry *pb.SchedulingMetadataMap_MappingEntry) SchedulingMetadata {
	timestamp, _ := time.Parse("2006-01-02_15-04-05", mappingEntry.Timestamp) // Pre-validated by proto definition
	return SchedulingMetadata{
		Action:    mappingEntry.Action,
		Timestamp: timestamp,
	}
}
