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
	"bytes"
	"encoding/json"
	"fmt"
	"time"

	"google.golang.org/protobuf/proto"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/karpenter/pkg/cloudprovider"
	pb "sigs.k8s.io/karpenter/pkg/controllers/orb/proto"

	v1 "k8s.io/api/core/v1"
)

type SchedulingInputDifferences struct {
	Added, Removed, Changed *SchedulingInput
}

type PodDifferences struct {
	Added, Removed, Changed []*v1.Pod
}

type SNPDifferences struct {
	Added, Removed, Changed []*StateNodeWithPods
}

type BindingDifferences struct {
	Added, Removed, Changed map[types.NamespacedName]string
}

type InstanceTypeDifferences struct {
	Added, Removed, Changed []*cloudprovider.InstanceType
}

func MarshalBatchedDifferences(batchedDifferences []*SchedulingInputDifferences) ([]byte, error) {
	protoAdded, protoRemoved, protoChanged := crossSection(batchedDifferences)
	protoDifferences := &pb.BatchedDifferences{
		Added:   protoSchedulingInputs(protoAdded),
		Removed: protoSchedulingInputs(protoRemoved),
		Changed: protoSchedulingInputs(protoChanged),
	}
	protoData, err := proto.Marshal(protoDifferences)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal Differences: %v", err)
	}
	return protoData, nil
}

func UnmarshalBatchedDifferences(differencesData []byte) ([]*SchedulingInputDifferences, error) {
	batchedDifferences := &pb.BatchedDifferences{}
	if err := proto.Unmarshal(differencesData, batchedDifferences); err != nil {
		return nil, fmt.Errorf("failed to unmarshal Differences: %v", err)
	}

	batchedAdded, err := reconstructSchedulingInputs(batchedDifferences.GetAdded())
	if err != nil {
		return nil, fmt.Errorf("failed to reconstruct Added: %v", err)
	}
	batchedRemoved, err := reconstructSchedulingInputs(batchedDifferences.GetRemoved())
	if err != nil {
		return nil, fmt.Errorf("failed to reconstruct Removed: %v", err)
	}
	batchedChanged, err := reconstructSchedulingInputs(batchedDifferences.GetChanged())
	if err != nil {
		return nil, fmt.Errorf("failed to reconstruct Changed: %v", err)
	}

	batchedSchedulingInputDifferences := []*SchedulingInputDifferences{}
	for i := 0; i < len(batchedAdded); i++ { // They should have the same dimensionality, even if some are empty
		batchedSchedulingInputDifferences = append(batchedSchedulingInputDifferences, &SchedulingInputDifferences{
			Added:   batchedAdded[i],
			Removed: batchedRemoved[i],
			Changed: batchedChanged[i],
		})
	}

	return batchedSchedulingInputDifferences, nil
}

func protoSchedulingInputs(si []*SchedulingInput) []*pb.SchedulingInput {
	protoSi := []*pb.SchedulingInput{}
	for _, schedulingInput := range si {
		protoSi = append(protoSi, protoSchedulingInput(schedulingInput))
	}
	return protoSi
}

func reconstructSchedulingInputs(pbsi []*pb.SchedulingInput) ([]*SchedulingInput, error) {
	reconstructedSi := []*SchedulingInput{}
	for _, schedulingInput := range pbsi {
		si, err := reconstructSchedulingInput(schedulingInput)
		if err != nil {
			return nil, err
		}
		reconstructedSi = append(reconstructedSi, si)
	}
	return reconstructedSi, nil
}

// Retrieves the time from any non-empty scheduling input in differences. The times are equivalent between them.
func (differences *SchedulingInputDifferences) GetTimestamp() time.Time {
	if differences.Added != nil && !differences.Added.Timestamp.IsZero() {
		return differences.Added.Timestamp
	} else if differences.Removed != nil && !differences.Removed.Timestamp.IsZero() {
		return differences.Removed.Timestamp
	} else if differences.Changed != nil && !differences.Changed.Timestamp.IsZero() {
		return differences.Changed.Timestamp
	}
	return time.Time{} // Return the zero value of time.Time if no timestamp is found
}

// Gets the time window (i.e. from start to end timestamp) from a slice of differences. It returns (start, end)
func GetTimeWindow(differences []*SchedulingInputDifferences) (time.Time, time.Time) {
	start := time.Now().UTC()
	end := time.Time{}.UTC()
	for _, diff := range differences {
		timestamp := diff.GetTimestamp()
		if timestamp.Before(start) {
			start = timestamp
		}
		if timestamp.After(end) {
			end = timestamp
		}
	}
	return start, end
}

// Pulls the cross-sectional slices of each Scheduling Input differences from a slice of Differences
func crossSection(differences []*SchedulingInputDifferences) ([]*SchedulingInput, []*SchedulingInput, []*SchedulingInput) {
	allAdded := []*SchedulingInput{}
	allRemoved := []*SchedulingInput{}
	allChanged := []*SchedulingInput{}

	for _, diff := range differences {
		if diff.Added != nil {
			allAdded = append(allAdded, diff.Added)
		}
		if diff.Removed != nil {
			allRemoved = append(allRemoved, diff.Removed)
		}
		if diff.Changed != nil {
			allChanged = append(allChanged, diff.Changed)
		}
	}
	return allAdded, allRemoved, allChanged
}

// Get the byte size of the cross-section of differences to compare for rebaselining
func (differences *SchedulingInputDifferences) getByteSize() int {
	protoAdded := protoSchedulingInput(differences.Added)
	protoRemoved := protoSchedulingInput(differences.Removed)
	protoChanged := protoSchedulingInput(differences.Changed)

	protoAddedData, err := proto.Marshal(protoAdded)
	if err != nil {
		protoAddedData = []byte{} // size 0
	}
	protoRemovedData, err := proto.Marshal(protoRemoved)
	if err != nil {
		protoRemovedData = []byte{} // size 0
	}
	protoChangedData, err := proto.Marshal(protoChanged)
	if err != nil {
		protoChangedData = []byte{} // size 0
	}

	return len(protoAddedData) + len(protoRemovedData) + len(protoChangedData)
}

// Functions to check the differences in all the fields of a SchedulingInput (except the timestamp)
func (si *SchedulingInput) Diff(oldSi *SchedulingInput) *SchedulingInputDifferences {
	// Determine the differences in each of the fields of ScheduleInput
	podDiff := diffPods(oldSi.PendingPods, si.PendingPods)
	snpDiff := diffStateNodes(oldSi.StateNodesWithPods, si.StateNodesWithPods)
	bindingsDiff := diffBindings(oldSi.Bindings, si.Bindings)
	itDiff := diffInstanceTypes(oldSi.InstanceTypes, si.InstanceTypes)

	diffAdded := &SchedulingInput{}
	diffRemoved := &SchedulingInput{}
	diffChanged := &SchedulingInput{}

	// If there are added differences, include them
	if len(podDiff.Added) > 0 || len(snpDiff.Added) > 0 || len(bindingsDiff.Added) > 0 || len(itDiff.Added) > 0 {
		diffAdded = NewReconstructedSchedulingInput(si.Timestamp, podDiff.Added, snpDiff.Added, bindingsDiff.Added, itDiff.Added)
		fmt.Println("Diff Scheduling Input added is... ", diffAdded.String()) // Test print, delete later
	}
	if len(podDiff.Removed) > 0 || len(snpDiff.Removed) > 0 || len(bindingsDiff.Removed) > 0 || len(itDiff.Removed) > 0 {
		diffRemoved = NewReconstructedSchedulingInput(si.Timestamp, podDiff.Removed, snpDiff.Removed, bindingsDiff.Removed, itDiff.Removed)
		fmt.Println("Diff Scheduling Input removed is... ", diffRemoved.String()) // Test print, delete later
	}
	if len(podDiff.Changed) > 0 || len(snpDiff.Changed) > 0 || len(bindingsDiff.Changed) > 0 || len(itDiff.Changed) > 0 {
		diffChanged = NewReconstructedSchedulingInput(si.Timestamp, podDiff.Changed, snpDiff.Changed, bindingsDiff.Changed, itDiff.Changed)
		fmt.Println("Diff Scheduling Input changed is... ", diffChanged.String()) // Test print, delete later
	}

	return &SchedulingInputDifferences{
		Added:   diffAdded,
		Removed: diffRemoved,
		Changed: diffChanged,
	}
}

// This is the diffPods function which gets the differences between pods
func diffPods(oldPods, newPods []*v1.Pod) PodDifferences {
	diff := PodDifferences{
		Added:   []*v1.Pod{},
		Removed: []*v1.Pod{},
		Changed: []*v1.Pod{},
	}

	// Cast pod slices to sets for unordered reference by their UID
	oldPodSet := map[string]*v1.Pod{}
	for _, pod := range oldPods {
		oldPodSet[string(pod.GetUID())] = pod
	}

	newPodSet := map[string]*v1.Pod{}
	for _, pod := range newPods {
		newPodSet[string(pod.GetUID())] = pod
	}

	// Find the added and changed pods
	for _, newPod := range newPods {
		oldPod, exists := oldPodSet[string(newPod.GetUID())]

		if !exists {
			diff.Added = append(diff.Added, newPod)
		} else if hasReducedPodChanged(oldPod, newPod) {
			// If pod has changed, add the whole changed pod
			// Simplification / Opportunity to optimize -- Only add sub-field.
			//    This requires more book-keeping on object reconstruction from logs later on.
			diff.Changed = append(diff.Changed, newPod)
		}
	}

	// Find the removed pods
	for _, oldPod := range oldPods {
		if _, exists := newPodSet[string(oldPod.GetUID())]; !exists {
			diff.Removed = append(diff.Removed, oldPod)
		}
	}

	return diff
}

// This is the diffStateNodes function which gets the differences between statenodes
func diffStateNodes(oldStateNodesWithPods, newStateNodesWithPods []*StateNodeWithPods) SNPDifferences {
	diff := SNPDifferences{
		Added:   []*StateNodeWithPods{},
		Removed: []*StateNodeWithPods{},
		Changed: []*StateNodeWithPods{},
	}

	// Cast StateNodesWithPods slices to sets for unordered reference by their name
	oldStateNodeSet := map[string]*StateNodeWithPods{}
	for _, stateNodeWithPods := range oldStateNodesWithPods {
		oldStateNodeSet[stateNodeWithPods.GetName()] = stateNodeWithPods
	}

	newStateNodeSet := map[string]*StateNodeWithPods{}
	for _, stateNodeWithPods := range newStateNodesWithPods {
		newStateNodeSet[stateNodeWithPods.GetName()] = stateNodeWithPods
	}

	// Find the added and changed StateNodesWithPods
	for _, newStateNodeWithPods := range newStateNodesWithPods {
		oldStateNodeWithPods, exists := oldStateNodeSet[newStateNodeWithPods.GetName()]

		if !exists { // Same opportunity for optimization as in pods
			diff.Added = append(diff.Added, newStateNodeWithPods)
		} else if hasStateNodeWithPodsChanged(oldStateNodeWithPods, newStateNodeWithPods) {
			diff.Changed = append(diff.Changed, newStateNodeWithPods)
		}
	}

	// Find the removed StateNodesWithPods
	for _, oldStateNodeWithPods := range oldStateNodesWithPods {
		if _, exists := newStateNodeSet[oldStateNodeWithPods.GetName()]; !exists {
			diff.Removed = append(diff.Removed, oldStateNodeWithPods)
		}
	}

	return diff
}

func diffBindings(old, new map[types.NamespacedName]string) BindingDifferences {
	diff := BindingDifferences{
		Added:   map[types.NamespacedName]string{},
		Removed: map[types.NamespacedName]string{},
		Changed: map[types.NamespacedName]string{},
	}

	// Find the changed or removed bindings
	for k, v := range old {
		if newVal, ok := new[k]; ok {
			if v != newVal {
				diff.Changed[k] = newVal
			}
		} else {
			diff.Removed[k] = v
		}
	}

	// Find the added bindings
	for k, v := range new {
		if _, ok := old[k]; !ok {
			diff.Added[k] = v
		}
	}

	return diff
}

// This is the diffInstanceTypes function which gets the differences between instance types
func diffInstanceTypes(oldTypes, newTypes []*cloudprovider.InstanceType) InstanceTypeDifferences {
	diff := InstanceTypeDifferences{
		Added:   []*cloudprovider.InstanceType{},
		Removed: []*cloudprovider.InstanceType{},
		Changed: []*cloudprovider.InstanceType{},
	}

	// Cast InstanceTypes slices to sets for unordered reference by their Name
	oldTypeSet := map[string]*cloudprovider.InstanceType{}
	for _, instanceType := range oldTypes {
		oldTypeSet[instanceType.Name] = instanceType
	}

	newTypeSet := map[string]*cloudprovider.InstanceType{}
	for _, instanceType := range newTypes {
		newTypeSet[instanceType.Name] = instanceType
	}

	// Find the added and changed instance types
	for _, newType := range newTypes {
		oldType, exists := oldTypeSet[newType.Name]

		if !exists {
			diff.Added = append(diff.Added, newType)
		} else if hasReducedInstanceTypeChanged(oldType, newType) {
			diff.Changed = append(diff.Changed, newType)
		}
	}

	// Find the removed instance types
	for _, oldType := range oldTypes {
		if _, exists := newTypeSet[oldType.Name]; !exists {
			diff.Removed = append(diff.Removed, oldType)
		}
	}

	return diff
}

func hasReducedPodChanged(oldPod, newPod *v1.Pod) bool {
	return !equality.Semantic.DeepEqual(oldPod.ObjectMeta, newPod.ObjectMeta) ||
		!equality.Semantic.DeepEqual(oldPod.Status, newPod.Status) ||
		!equality.Semantic.DeepEqual(oldPod.Spec, newPod.Spec)
}

func hasStateNodeWithPodsChanged(oldStateNodeWithPods, newStateNodeWithPods *StateNodeWithPods) bool {
	return !equality.Semantic.DeepEqual(oldStateNodeWithPods, newStateNodeWithPods)
}

func hasReducedInstanceTypeChanged(oldInstanceType, newInstanceType *cloudprovider.InstanceType) bool {
	return !equality.Semantic.DeepEqual(oldInstanceType.Name, newInstanceType.Name) ||
		!structEqualJSON(oldInstanceType.Offerings, newInstanceType.Offerings) || // Cannot deep equal these, they have unexported types
		!structEqualJSON(oldInstanceType.Requirements, newInstanceType.Requirements) // ^
}

// TODO: Likely inefficient equality checking for nested types Offerings and Requirements,
// but both have unexported types not compatible with DeepEqual
func structEqualJSON(a, b interface{}) bool {
	aBytes, _ := json.Marshal(a)
	bBytes, _ := json.Marshal(b)
	return bytes.Equal(aBytes, bBytes)
}
