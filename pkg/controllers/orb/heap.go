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
	"container/heap"
	"context"
	"time"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/karpenter/pkg/cloudprovider"
	pb "sigs.k8s.io/karpenter/pkg/controllers/orb/proto"
	scheduler "sigs.k8s.io/karpenter/pkg/controllers/provisioning/scheduling"
	"sigs.k8s.io/karpenter/pkg/controllers/state"
)

type TimestampedType interface {
	GetTime() time.Time
}

// MinTimeHeap is a generic min-heap implementation with the Timestamp field defined as the comparator
type MinTimeHeap[T TimestampedType] []T

func (h MinTimeHeap[T]) Len() int {
	return len(h)
}

// This compares timestamps for a min heap, so that oldest elements pop first.
func (h MinTimeHeap[T]) Less(i, j int) bool {
	return h[i].GetTime().Before(h[j].GetTime())
}

func (h MinTimeHeap[T]) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
}

func (h *MinTimeHeap[T]) Push(x interface{}) {
	*h = append(*h, x.(T))
}

func (h *MinTimeHeap[T]) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[:n-1]
	return x
}

func NewMinHeap[T TimestampedType]() *MinTimeHeap[T] {
	h := &MinTimeHeap[T]{}
	heap.Init(h)
	return h
}

// SchedulingInputHeap is a min-heap of SchedulingInputs
type SchedulingInputHeap = MinTimeHeap[SchedulingInput]

// Function for logging scheduling inputs to the Provisioner Scheduler. Batches via a min-heap, ordered by least recent.
func (h *MinTimeHeap[SchedulingInput]) LogSchedulingInput(ctx context.Context, kubeClient client.Client, scheduledTime time.Time, pods []*v1.Pod, stateNodes []*state.StateNode,
	bindings map[types.NamespacedName]string, instanceTypes map[string][]*cloudprovider.InstanceType, topology *scheduler.Topology, daemonSetPods []*v1.Pod) {
	si := NewSchedulingInput(ctx, kubeClient, scheduledTime, pods, stateNodes, bindings, instanceTypes, topology, daemonSetPods)
	si.Reduce()
	heap.Push(h, si)
}

// SchedulingMetadataHeap is a min-heap of SchedulingMetadata
type SchedulingMetadataHeap = MinTimeHeap[SchedulingMetadata]

// func NewSchedulingMetadataHeap() *SchedulingMetadataHeap {
// 	h := &SchedulingMetadataHeap{}
// 	heap.Init(h)
// 	return h
// }

func (h *MinTimeHeap[SchedulingMetadata]) LogSchedulingAction(ctx context.Context, schedulingTime time.Time) {
	metadata, ok := GetSchedulingMetadata(ctx)
	if !ok { // Provisioning metadata is not set in the context, set it to the default - normal provisioning action
		ctx = WithSchedulingMetadata(ctx, "normal-provisioning", schedulingTime)
		metadata, _ = GetSchedulingMetadata(ctx) // Get it again to update metadata
	}
	// This allows us to associate metadata with its respective scheduling action. It resolves the potential time
	// difference between the start of an action call (consolidation/drift) and its subsequent provisioning scheduling.
	metadata.Timestamp = schedulingTime

	heap.Push(h, metadata)
}

func protoSchedulingMetadataMap(heap *SchedulingMetadataHeap) *pb.SchedulingMetadataMap {
	mapping := &pb.SchedulingMetadataMap{}
	for heap.Len() > 0 {
		metadata := heap.Pop().(SchedulingMetadata)
		entry := protoSchedulingMetadata(metadata)
		mapping.Entries = append(mapping.Entries, entry)
	}
	return mapping
}

// Reconstructs scheduling metadata as a slice instead of back as a heap since each file will be heapified in aggregate.
func ReconstructAllSchedulingMetadata(mapping *pb.SchedulingMetadataMap) []*SchedulingMetadata {
	metadata := []*SchedulingMetadata{}
	for _, entry := range mapping.Entries {
		metadatum := reconstructSchedulingMetadata(entry)
		metadata = append(metadata, &metadatum)
	}
	return metadata
}
