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
	"reflect"
	"sort"
	"time"

	"google.golang.org/protobuf/proto"

	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"
	"sigs.k8s.io/karpenter/pkg/cloudprovider"
	pb "sigs.k8s.io/karpenter/pkg/controllers/orb/proto"
	scheduler "sigs.k8s.io/karpenter/pkg/controllers/provisioning/scheduling"

	v1 "k8s.io/api/core/v1"
)

// A resource's Differences can be generalized three versions of that same resource: what has been Added, Removed or Changed.
type Differences[T any] struct {
	Added   T
	Removed T
	Changed T
}

// Aliases for Differences definitions
type SchedulingInputDifferences Differences[*SchedulingInput]
type PodDifferences Differences[[]*v1.Pod]
type SNPDifferences Differences[[]*StateNodeWithPods]
type BindingDifferences Differences[map[types.NamespacedName]string]
type InstanceTypeDifferences Differences[[]*cloudprovider.InstanceType]
type NodePoolsToInstanceTypesDifferences Differences[map[string][]string]
type TopologyDifferences Differences[*scheduler.Topology]
type PVListDifferences Differences[*v1.PersistentVolumeList]
type PVCListDifferences Differences[*v1.PersistentVolumeClaimList]
type ScheduledPodDifferences Differences[*v1.PodList]

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

// Get the byte size of the cross-section of differences to compare in rebaselining logic
func (differences *SchedulingInputDifferences) getByteSize() int {
	differencesData, err := MarshalBatchedDifferences([]*SchedulingInputDifferences{differences})
	if err != nil {
		differencesData = []byte{} // size 0
	}
	return len(differencesData)
}

// Gets the time window (i.e. from start to end timestamp) from a slice of differences. It returns (start, end)
func GetTimeWindow(differences []*SchedulingInputDifferences) (time.Time, time.Time) {
	start := time.Time{}
	end := time.Time{}
	for _, diff := range differences {
		timestamp := diff.GetTimestamp()
		if start.IsZero() || timestamp.Before(start) {
			start = timestamp
		}
		if end.IsZero() || timestamp.After(end) {
			end = timestamp
		}
	}
	return start, end
}

// Pulls the cross-sectional slices (added, removed, changed) of each Scheduling Input Differences
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

// Pulls out cross-sectional maps of each Scheduling Input Difference mapped by their timestamp,
// returned alongside the corresponding sorted slice of times for that batch, from oldest to most recent.
func crossSectionByTimestamp(differences []*SchedulingInputDifferences) (map[time.Time]*SchedulingInput, map[time.Time]*SchedulingInput, map[time.Time]*SchedulingInput, []time.Time) {
	allAdded := map[time.Time]*SchedulingInput{}
	allRemoved := map[time.Time]*SchedulingInput{}
	allChanged := map[time.Time]*SchedulingInput{}
	batchedTimes := []time.Time{}

	for _, diff := range differences {
		// Map the Timestamp to its corresponding Difference
		if diff.Added != nil && !diff.Added.Timestamp.IsZero() {
			allAdded[diff.Added.Timestamp] = diff.Added
		}
		if diff.Removed != nil && !diff.Removed.Timestamp.IsZero() {
			allRemoved[diff.Removed.Timestamp] = diff.Removed
		}
		if diff.Changed != nil && !diff.Changed.Timestamp.IsZero() {
			allChanged[diff.Changed.Timestamp] = diff.Changed
		}
		// Batch all Differences' timestamps together
		batchedTimes = append(batchedTimes, diff.GetTimestamp())
	}

	sort.Slice(batchedTimes, func(i, j int) bool { return batchedTimes[i].Before(batchedTimes[j]) })
	return allAdded, allRemoved, allChanged, batchedTimes
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
	for i := 0; i < len(batchedAdded); i++ { // They will have the same dimensionality, even if some are empty
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

func CreateMapFromSlice[T any, K comparable](slice []*T, getKey func(*T) K) map[K]*T {
	result := map[K]*T{}
	for _, item := range slice {
		key := getKey(item)
		result[key] = item
	}
	return result
}

// Aliases for the getKey() functions passed into merge
func getPodKey(pod *v1.Pod) string                             { return string(pod.GetUID()) }
func getStateNodeWithPodsKey(snwp *StateNodeWithPods) string   { return snwp.GetName() }
func GetInstanceTypeKey(it *cloudprovider.InstanceType) string { return it.Name }

// Generalized function to merge in the Differences into each different Scheduling Input field.
func merge[T any](items []*T, added []*T, removed []*T, changed []*T, getKey func(item *T) string) []*T {
	// Create the initial map from items
	mergedMap := CreateMapFromSlice(items, getKey)

	for _, addedItem := range added {
		mergedMap[getKey(addedItem)] = addedItem
	}
	for _, removedItem := range removed {
		delete(mergedMap, getKey(removedItem))
	}
	for _, changedItem := range changed {
		mergedMap[getKey(changedItem)] = changedItem
	}

	mergedItems := []*T{}
	for _, mergedItem := range mergedMap {
		mergedItems = append(mergedItems, mergedItem)
	}
	return mergedItems
}

// Generalized function to merge Differences in place in a map.
func mergeMap[K comparable, V any](m map[K]V, added, removed, changed map[K]V) {
	for k, v := range added {
		m[k] = v
	}
	for k := range removed {
		delete(m, k)
	}
	for k, v := range changed {
		m[k] = v
	}
}

// Some fields of Scheduling Inputs won't have anything added or removed from them because they aren't slices or maps, and are only defined for changes.
// This function handles those cases by merging the changed field into the iterating field if the changed field is non-empty.
func mergeOnlyChanged[T any](original *T, difference *T) *T {
	if difference != nil && !reflect.DeepEqual(*difference, reflect.Zero(reflect.TypeOf(*difference)).Interface()) {
		return difference
	}
	return original
}

// Reconstructs a Scheduling Input at a desired reconstruct time by iteratively adding back all logged Differences onto the baseline Scheduling Input up until that time.
func MergeDifferences(baseline *SchedulingInput, batchedDifferences []*SchedulingInputDifferences, reconstructTime time.Time) *SchedulingInput {
	batchedAdded, batchedRemoved, batchedChanged, sortedBatchedTimes := crossSectionByTimestamp(batchedDifferences)

	mergingInputs := &SchedulingInput{
		Timestamp:             reconstructTime,
		PendingPods:           baseline.PendingPods,
		StateNodesWithPods:    baseline.StateNodesWithPods,
		Bindings:              baseline.Bindings,
		AllInstanceTypes:      baseline.AllInstanceTypes,
		NodePoolInstanceTypes: baseline.NodePoolInstanceTypes,
		Topology:              baseline.Topology,
		DaemonSetPods:         baseline.DaemonSetPods,
		PVList:                baseline.PVList,
		PVCList:               baseline.PVCList,
		ScheduledPodList:      baseline.ScheduledPodList,
	}

	for _, differencesTime := range sortedBatchedTimes {
		if differencesTime.After(reconstructTime) {
			break
		}
		mergeSchedulingInputs(mergingInputs, &SchedulingInputDifferences{
			Added:   batchedAdded[differencesTime],
			Removed: batchedRemoved[differencesTime],
			Changed: batchedChanged[differencesTime],
		})
	}
	return mergingInputs
}

// Merges one time's set of Differences on the baseline/merging Inputs; adding, removing and changing each field as appropriate.
func mergeSchedulingInputs(iteratingInput *SchedulingInput, differences *SchedulingInputDifferences) {
	iteratingInput.PendingPods = merge(iteratingInput.PendingPods, differences.Added.PendingPods, differences.Removed.PendingPods, differences.Changed.PendingPods, getPodKey)
	iteratingInput.StateNodesWithPods = merge(iteratingInput.StateNodesWithPods, differences.Added.StateNodesWithPods, differences.Removed.StateNodesWithPods, differences.Changed.StateNodesWithPods, getStateNodeWithPodsKey)
	mergeMap(iteratingInput.Bindings, differences.Added.Bindings, differences.Removed.Bindings, differences.Changed.Bindings)
	iteratingInput.AllInstanceTypes = merge(iteratingInput.AllInstanceTypes, differences.Added.AllInstanceTypes, differences.Removed.AllInstanceTypes, differences.Changed.AllInstanceTypes, GetInstanceTypeKey)
	mergeMap(iteratingInput.NodePoolInstanceTypes, differences.Added.NodePoolInstanceTypes, differences.Removed.NodePoolInstanceTypes, differences.Changed.NodePoolInstanceTypes)
	iteratingInput.Topology = mergeOnlyChanged(iteratingInput.Topology, differences.Changed.Topology)
	iteratingInput.DaemonSetPods = merge(iteratingInput.DaemonSetPods, differences.Added.DaemonSetPods, differences.Removed.DaemonSetPods, differences.Changed.DaemonSetPods, getPodKey)
	iteratingInput.PVList = mergeOnlyChanged(iteratingInput.PVList, differences.Changed.PVList)
	iteratingInput.PVCList = mergeOnlyChanged(iteratingInput.PVCList, differences.Changed.PVCList)
	iteratingInput.ScheduledPodList = mergeOnlyChanged(iteratingInput.ScheduledPodList, differences.Changed.ScheduledPodList)
}

// Functions to check the differences in all the fields of a SchedulingInput (except the timestamp)
func (oldSi *SchedulingInput) Diff(si *SchedulingInput) *SchedulingInputDifferences {
	// Determine the differences in each of the fields of ScheduleInput
	podDiff := diffSlice(oldSi.PendingPods, si.PendingPods, getPodKey, hasPodChanged)
	// podDiff := diffPods(oldSi.PendingPods, si.PendingPods)
	snpDiff := diffSlice(oldSi.StateNodesWithPods, si.StateNodesWithPods, getStateNodeWithPodsKey, hasStateNodeWithPodsChanged)
	// snpDiff := diffStateNodes(oldSi.StateNodesWithPods, si.StateNodesWithPods)
	bindingsDiff := diffMap(oldSi.Bindings, si.Bindings, hasBindingChanged)
	// bindingsDiff := diffBindings(oldSi.Bindings, si.Bindings)
	itDiff := diffSlice(oldSi.AllInstanceTypes, si.AllInstanceTypes, GetInstanceTypeKey, hasInstanceTypeChanged)
	// itDiff := diffInstanceTypes(oldSi.AllInstanceTypes, si.AllInstanceTypes)
	npitDiff := diffMap(oldSi.NodePoolInstanceTypes, si.NodePoolInstanceTypes, hasNodePoolInstanceTypeChanged)
	// npitDiff := diffNodePoolsToInstanceTypes(oldSi.NodePoolInstanceTypes, si.NodePoolInstanceTypes)
	topologyDiff := diffOnlyChanges(oldSi.Topology, si.Topology)
	dspDiff := diffSlice(oldSi.DaemonSetPods, si.DaemonSetPods, getPodKey, hasPodChanged)
	pvListDiff := diffOnlyChanges(oldSi.PVList, si.PVList)
	pvcListDiff := diffOnlyChanges(oldSi.PVCList, si.PVCList)
	scheduledPodListDiff := diffOnlyChanges(oldSi.ScheduledPodList, si.ScheduledPodList)
	// topologyDiff := diffTopology(oldSi.Topology, si.Topology)
	// dspDiff := diffDaemonSetPods(oldSi.DaemonSetPods, si.DaemonSetPods)
	// pvListDiff := diffPVList(oldSi.PVList, si.PVList)
	// pvcListDiff := diffPVCList(oldSi.PVCList, si.PVCList)
	// scheduledPodListDiff := diffScheduledPods(oldSi.ScheduledPodList, si.ScheduledPodList)

	diffAdded := &SchedulingInput{}
	diffRemoved := &SchedulingInput{}
	diffChanged := &SchedulingInput{}

	// If there are added differences, include them
	if len(podDiff.Added) > 0 || len(snpDiff.Added) > 0 || len(bindingsDiff.Added) > 0 || len(itDiff.Added) > 0 || len(npitDiff.Added) > 0 || len(dspDiff.Added) > 0 {
		diffAdded = &SchedulingInput{si.Timestamp, podDiff.Added, snpDiff.Added, bindingsDiff.Added, itDiff.Added, npitDiff.Added, nil, dspDiff.Added, nil, nil, nil}
	}
	if len(podDiff.Removed) > 0 || len(snpDiff.Removed) > 0 || len(bindingsDiff.Removed) > 0 || len(itDiff.Removed) > 0 || len(npitDiff.Removed) > 0 || len(dspDiff.Removed) > 0 {
		diffRemoved = &SchedulingInput{si.Timestamp, podDiff.Removed, snpDiff.Removed, bindingsDiff.Removed, itDiff.Removed, npitDiff.Removed, nil, dspDiff.Removed, nil, nil, nil}
	}
	if len(podDiff.Changed) > 0 || len(snpDiff.Changed) > 0 || len(bindingsDiff.Changed) > 0 || len(itDiff.Changed) > 0 || len(npitDiff.Changed) > 0 || (topologyDiff.Changed != nil) || len(dspDiff.Changed) > 0 || pvListDiff.Changed != nil || pvcListDiff.Changed != nil || scheduledPodListDiff.Changed != nil {
		diffChanged = &SchedulingInput{si.Timestamp, podDiff.Changed, snpDiff.Changed, bindingsDiff.Changed, itDiff.Changed, npitDiff.Changed, topologyDiff.Changed, dspDiff.Changed, pvListDiff.Changed, pvcListDiff.Changed, scheduledPodListDiff.Changed}
	}

	return &SchedulingInputDifferences{
		Added:   diffAdded,
		Removed: diffRemoved,
		Changed: diffChanged,
	}
}

func diffSlice[T any](oldresource, newresource []*T, getKey func(*T) string, hasChanged func(*T, *T) bool) Differences[[]*T] {
	added := []*T{}
	removed := []*T{}
	changed := []*T{}

	oldResourceMap := CreateMapFromSlice(oldresource, getKey)
	newResourceMap := CreateMapFromSlice(newresource, getKey)
	oldResourceSet := sets.KeySet(oldResourceMap)
	newResourceSet := sets.KeySet(newResourceMap)

	for addedKey := range newResourceSet.Difference(oldResourceSet) {
		added = append(added, newResourceMap[addedKey])
	}

	for removedKey := range oldResourceSet.Difference(newResourceSet) {
		removed = append(removed, oldResourceMap[removedKey])
	}

	// Add the changed pods to Pod Differences, only after checking if they've changed.
	// TODO: Simplification / Opportunity to optimize -- Diff/Merge by sub-field.
	//    This requires more book-keeping on object reconstruction from logs later on.
	for commonKey := range oldResourceSet.Intersection(newResourceSet) {
		if hasChanged(oldResourceMap[commonKey], newResourceMap[commonKey]) {
			changed = append(changed, newResourceMap[commonKey])
		}
	}

	return Differences[[]*T]{Added: added, Removed: removed, Changed: changed}
}

func diffMap[K comparable, V any](oldResourceMap, newResourceMap map[K]V, hasChanged func(V, V) bool) Differences[map[K]V] {
	added := map[K]V{}
	removed := map[K]V{}
	changed := map[K]V{}

	for key, resource := range oldResourceMap {
		if newValue, ok := newResourceMap[key]; ok {
			if hasChanged(resource, newValue) {
				changed[key] = newValue
			}
		} else {
			removed[key] = resource
		}
	}

	for key, resource := range newResourceMap {
		if _, ok := oldResourceMap[key]; !ok {
			added[key] = resource
		}
	}

	return Differences[map[K]V]{Added: added, Removed: removed, Changed: changed}
}

func diffOnlyChanges[T any](oldresource, newresource *T) Differences[*T] {
	if !structEqualJSON(oldresource, newresource) {
		return Differences[*T]{nil, nil, newresource}
	}
	return Differences[*T]{nil, nil, oldresource}
}

// // Converts pod slice to a map from its UID.
// func mapPodsByUID(pods []*v1.Pod) map[types.UID]*v1.Pod {
// 	podMap := map[types.UID]*v1.Pod{}
// 	for _, pod := range pods {
// 		podMap[pod.GetUID()] = pod
// 	}
// 	return podMap
// }

// // This is the diffPods function which gets the differences between pods
// func diffPods(oldPods, newPods []*v1.Pod) PodDifferences {
// 	diff := PodDifferences{
// 		Added:   []*v1.Pod{},
// 		Removed: []*v1.Pod{},
// 		Changed: []*v1.Pod{},
// 	}

// 	oldPodMap := mapPodsByUID(oldPods)
// 	oldPodSet := sets.KeySet(oldPodMap)
// 	newPodMap := mapPodsByUID(newPods)
// 	newPodSet := sets.KeySet(newPodMap)

// 	// Add the new pods to Pod Differences
// 	for addedUID := range newPodSet.Difference(oldPodSet) {
// 		diff.Added = append(diff.Added, newPodMap[addedUID])
// 	}
// 	// Add the removed pods to Pod Differences
// 	for removedUID := range oldPodSet.Difference(newPodSet) {
// 		diff.Removed = append(diff.Removed, oldPodMap[removedUID])
// 	}
// 	// Add the changed pods to Pod Differences, only after checking if they've changed.
// 	// Simplification / Opportunity to optimize -- Only add sub-field.
// 	//    This requires more book-keeping on object reconstruction from logs later on.
// 	for commonUID := range newPodSet.Intersection(oldPodSet) {
// 		if hasReducedPodChanged(oldPodMap[commonUID], newPodMap[commonUID]) {
// 			diff.Changed = append(diff.Changed, newPodMap[commonUID])
// 		}
// 	}
// 	return diff
// }

// // Function from StateNodeWithPods slice to Map by name
// func mapStateNodesWithPodsByName(stateNodesWithPods []*StateNodeWithPods) map[string]*StateNodeWithPods {
// 	snpMap := map[string]*StateNodeWithPods{}
// 	for _, snp := range stateNodesWithPods {
// 		snpMap[snp.GetName()] = snp
// 	}
// 	return snpMap
// }

// // This is the diffStateNodes function which gets the differences between statenodes
// func diffStateNodes(oldStateNodesWithPods, newStateNodesWithPods []*StateNodeWithPods) SNPDifferences {
// 	diff := SNPDifferences{
// 		Added:   []*StateNodeWithPods{},
// 		Removed: []*StateNodeWithPods{},
// 		Changed: []*StateNodeWithPods{},
// 	}

// 	// Cast StateNodesWithPods slices to sets for unordered reference by their name
// 	oldStateNodeMap := mapStateNodesWithPodsByName(oldStateNodesWithPods)
// 	oldStateNodeSet := sets.KeySet(oldStateNodeMap)
// 	newStateNodeMap := mapStateNodesWithPodsByName(newStateNodesWithPods)
// 	newStateNodeSet := sets.KeySet(newStateNodeMap)

// 	// Find the added, removed and changed stateNodes
// 	for addedName := range newStateNodeSet.Difference(oldStateNodeSet) {
// 		diff.Added = append(diff.Added, newStateNodeMap[addedName])
// 	}
// 	for removedName := range oldStateNodeSet.Difference(newStateNodeSet) {
// 		diff.Removed = append(diff.Removed, oldStateNodeMap[removedName])
// 	}
// 	for commonName := range newStateNodeSet.Intersection(oldStateNodeSet) {
// 		if hasStateNodeWithPodsChanged(oldStateNodeMap[commonName], newStateNodeMap[commonName]) {
// 			diff.Changed = append(diff.Changed, newStateNodeMap[commonName])
// 		}
// 	}
// 	return diff
// }

// func diffBindings(old, new map[types.NamespacedName]string) BindingDifferences {
// 	diff := BindingDifferences{
// 		Added:   map[types.NamespacedName]string{},
// 		Removed: map[types.NamespacedName]string{},
// 		Changed: map[types.NamespacedName]string{},
// 	}

// 	// Find the changed or removed bindings
// 	for namespacedName, binding := range old {
// 		if newBinding, ok := new[namespacedName]; ok {
// 			if binding != newBinding {
// 				diff.Changed[namespacedName] = newBinding
// 			}
// 		} else {
// 			diff.Removed[namespacedName] = binding
// 		}
// 	}
// 	// Find the added bindings
// 	for namespacedName, binding := range new {
// 		if _, ok := old[namespacedName]; !ok {
// 			diff.Added[namespacedName] = binding
// 		}
// 	}
// 	return diff
// }

// // Take differenceTypes to map like above, by name
// func MapInstanceTypesByName(instanceTypes []*cloudprovider.InstanceType) map[string]*cloudprovider.InstanceType {
// 	itMap := map[string]*cloudprovider.InstanceType{}
// 	for _, it := range instanceTypes {
// 		itMap[it.Name] = it
// 	}
// 	return itMap
// }

// // This is the diffInstanceTypes function which gets the differences between instance types
// func diffInstanceTypes(oldTypes, newTypes []*cloudprovider.InstanceType) InstanceTypeDifferences {
// 	diff := InstanceTypeDifferences{
// 		Added:   []*cloudprovider.InstanceType{},
// 		Removed: []*cloudprovider.InstanceType{},
// 		Changed: []*cloudprovider.InstanceType{},
// 	}

// 	oldTypeMap := MapInstanceTypesByName(oldTypes)
// 	oldTypeSet := sets.KeySet(oldTypeMap)
// 	newTypeMap := MapInstanceTypesByName(newTypes)
// 	newTypeSet := sets.KeySet(newTypeMap)

// 	// Find the added, removed and changed instanceTypes
// 	for addedName := range newTypeSet.Difference(oldTypeSet) {
// 		diff.Added = append(diff.Added, newTypeMap[addedName])
// 	}
// 	for removedName := range oldTypeSet.Difference(newTypeSet) {
// 		diff.Removed = append(diff.Removed, oldTypeMap[removedName])
// 	}
// 	for commonName := range newTypeSet.Intersection(oldTypeSet) {
// 		if hasInstanceTypeChanged(oldTypeMap[commonName], newTypeMap[commonName]) {
// 			diff.Changed = append(diff.Changed, newTypeMap[commonName])
// 		}
// 	}
// 	return diff
// }

// // This function is already a mapping, so it will mirror the diff for Bindings in the map checking
// func diffNodePoolsToInstanceTypes(old, new map[string][]string) NodePoolsToInstanceTypesDifferences {
// 	diff := NodePoolsToInstanceTypesDifferences{
// 		Added:   map[string][]string{},
// 		Removed: map[string][]string{},
// 		Changed: map[string][]string{},
// 	}

// 	// Find the changed or removed node pools
// 	for nodepool, instancetypes := range old {
// 		if newInstanceTypes, ok := new[nodepool]; ok {
// 			if !equality.Semantic.DeepEqual(sets.NewString(instancetypes...), sets.NewString(newInstanceTypes...)) {
// 				diff.Changed[nodepool] = newInstanceTypes
// 			}
// 		} else {
// 			diff.Removed[nodepool] = instancetypes
// 		}
// 	}
// 	// Find the added node pools
// 	for nodepool, instancetypes := range new {
// 		if _, ok := old[nodepool]; !ok {
// 			diff.Added[nodepool] = instancetypes
// 		}
// 	}
// 	return diff
// }

// func diffTopology(oldTopology, newTopology *scheduler.Topology) TopologyDifferences {
// 	diff := TopologyDifferences{
// 		Added:   nil, // Empty for Topology by construct (only coarsely checking) data structure equality, not for each recursive internal differences
// 		Removed: nil, // ^
// 		Changed: nil, // Only these matter, but keeping the "Differences" construct so I can (TODO) make a general interface{} later and simplify
// 	}

// 	if !structEqualJSON(oldTopology, newTopology) {
// 		diff.Changed = newTopology
// 	}
// 	return diff
// }

// func diffDaemonSetPods(oldPods, newPods []*v1.Pod) PodDifferences {
// 	diff := PodDifferences{
// 		Added:   []*v1.Pod{},
// 		Removed: []*v1.Pod{},
// 		Changed: []*v1.Pod{},
// 	}

// 	oldPodMap := mapPodsByUID(oldPods)
// 	oldPodSet := sets.KeySet(oldPodMap)
// 	newPodMap := mapPodsByUID(newPods)
// 	newPodSet := sets.KeySet(newPodMap)

// 	// Add the new pods to Pod Differences
// 	for addedUID := range newPodSet.Difference(oldPodSet) {
// 		diff.Added = append(diff.Added, newPodMap[addedUID])
// 	}
// 	// Add the removed pods to Pod Differences
// 	for removedUID := range oldPodSet.Difference(newPodSet) {
// 		diff.Removed = append(diff.Removed, oldPodMap[removedUID])
// 	}
// 	// Add the changed pods to Pod Differences, only after checking if they've changed.
// 	// Simplification / Opportunity to optimize -- Only add sub-field.
// 	//    This requires more book-keeping on object reconstruction from logs later on.
// 	for commonUID := range newPodSet.Intersection(oldPodSet) {
// 		if structEqualJSON(oldPodMap[commonUID], newPodMap[commonUID]) {
// 			diff.Changed = append(diff.Changed, newPodMap[commonUID])
// 		}
// 	}
// 	return diff
// }

// // This will mirror Topology's diff, being nil and only accounting for changes
// func diffPVList(oldpvList, newpvList *v1.PersistentVolumeList) PVListDifferences {
// 	diff := PVListDifferences{
// 		Added:   nil,
// 		Removed: nil,
// 		Changed: nil,
// 	}

// 	// Only check if the PVs have changed
// 	if !structEqualJSON(oldpvList, newpvList) {
// 		diff.Changed = newpvList
// 	}
// 	return diff
// }

// func diffPVCList(oldpvcList, newpvcList *v1.PersistentVolumeClaimList) PVCListDifferences {
// 	diff := PVCListDifferences{
// 		Added:   nil,
// 		Removed: nil,
// 		Changed: nil,
// 	}

// 	if !structEqualJSON(oldpvcList, newpvcList) {
// 		diff.Changed = newpvcList
// 	}
// 	return diff
// }

// func diffScheduledPods(oldScheduledPods, newScheduledPods *v1.PodList) ScheduledPodDifferences {
// 	diff := ScheduledPodDifferences{
// 		Added:   nil,
// 		Removed: nil,
// 		Changed: nil,
// 	}

// 	if !structEqualJSON(oldScheduledPods, newScheduledPods) {
// 		diff.Changed = newScheduledPods
// 	}
// 	return diff
// }

// Equality functions for hasChanged lambda functions
func hasPodChanged(oldPod, newPod *v1.Pod) bool {
	return !equality.Semantic.DeepEqual(oldPod.ObjectMeta, newPod.ObjectMeta) ||
		!equality.Semantic.DeepEqual(oldPod.Status, newPod.Status) ||
		!equality.Semantic.DeepEqual(oldPod.Spec, newPod.Spec)
}

func hasStateNodeWithPodsChanged(oldStateNodeWithPods, newStateNodeWithPods *StateNodeWithPods) bool {
	return !equality.Semantic.DeepEqual(oldStateNodeWithPods, newStateNodeWithPods)
}

func hasBindingChanged(oldBinding, newBinding string) bool {
	return oldBinding != newBinding
}

func hasInstanceTypeChanged(oldInstanceType, newInstanceType *cloudprovider.InstanceType) bool {
	return !equality.Semantic.DeepEqual(oldInstanceType.Name, newInstanceType.Name) ||
		!structEqualJSON(oldInstanceType.Offerings, newInstanceType.Offerings) ||
		!structEqualJSON(oldInstanceType.Requirements, newInstanceType.Requirements) ||
		!structEqualJSON(oldInstanceType.Capacity, newInstanceType.Capacity) ||
		!structEqualJSON(oldInstanceType.Overhead, newInstanceType.Overhead)
}

func hasNodePoolInstanceTypeChanged(instancetypes, newInstanceTypes []string) bool {
	return !equality.Semantic.DeepEqual(sets.NewString(instancetypes...), sets.NewString(newInstanceTypes...))
}

// Used when fields contain unexported types, which would cause DeepEqual to panic.
// TODO: Likely inefficient equality checking for nested types like Offerings and Requirements,
// but both have unexported types not compatible with DeepEqual
func structEqualJSON(a, b interface{}) bool {
	aBytes, _ := json.Marshal(a)
	bBytes, _ := json.Marshal(b)
	return bytes.Equal(aBytes, bBytes)
}
