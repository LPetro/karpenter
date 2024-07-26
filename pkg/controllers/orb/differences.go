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

type NodePoolsToInstanceTypesDifferences struct {
	Added, Removed, Changed map[string][]string
}

type TopologyDifferences struct {
	Added, Removed, Changed *scheduler.Topology
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

// Pulls a cross-section maps of each Scheduling Input differences mapped by their timestamp,
// returned alongside the corresponding sorted slice of times for that batch, from oldest to most recent.
func crossSectionByTimestamp(differences []*SchedulingInputDifferences) (map[time.Time]*SchedulingInput, map[time.Time]*SchedulingInput, map[time.Time]*SchedulingInput, []time.Time) {
	allAdded := map[time.Time]*SchedulingInput{}
	allRemoved := map[time.Time]*SchedulingInput{}
	allChanged := map[time.Time]*SchedulingInput{}
	batchedTimes := []time.Time{}

	for _, diff := range differences {
		if diff.Added != nil && !diff.Added.Timestamp.IsZero() {
			allAdded[diff.Added.Timestamp] = diff.Added
		}
		if diff.Removed != nil && !diff.Removed.Timestamp.IsZero() {
			allRemoved[diff.Removed.Timestamp] = diff.Removed
		}
		if diff.Changed != nil && !diff.Changed.Timestamp.IsZero() {
			allChanged[diff.Changed.Timestamp] = diff.Changed
		}
		batchedTimes = append(batchedTimes, diff.GetTimestamp())
	}

	sort.Slice(batchedTimes, func(i, j int) bool { return batchedTimes[i].Before(batchedTimes[j]) })

	return allAdded, allRemoved, allChanged, batchedTimes
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

// Function to merge differences read (back from the PV) to the baseline. They'll be read back and saved as a []*SchedulingInputDifferences
// This function will need to extract the cross sections, take in a baseline *SchedulingInput and add the consolidatedAdded, conslidatedDelete and consolidatedChanges
// which are just three *SchedulingInput's themselves. You can assume and put in placeholder functions for consolidateDifferences which pulls out those three consolidations.
func MergeDifferences(baseline *SchedulingInput, batchedDifferences []*SchedulingInputDifferences, reconstructTime time.Time) *SchedulingInput {
	// Extract the cross sections mapped by their timestamps
	batchedAdded, batchedRemoved, batchedChanged, sortedBatchedTimes := crossSectionByTimestamp(batchedDifferences)

	mergedInputs := &SchedulingInput{
		Timestamp:             reconstructTime,
		PendingPods:           baseline.PendingPods,
		StateNodesWithPods:    baseline.StateNodesWithPods,
		Bindings:              baseline.Bindings,
		AllInstanceTypes:      baseline.AllInstanceTypes,
		NodePoolInstanceTypes: baseline.NodePoolInstanceTypes,
		Topology:              baseline.Topology,
	}

	// Iterate over the baseline, making each time's set of changes
	for _, differencesTime := range sortedBatchedTimes {
		// Stop if we've gotten to differences occuring after the time for when we're reconstructing.
		if differencesTime.After(reconstructTime) {
			break
		}

		// Merge the cross sections with the baseline
		mergeSchedulingInputs(mergedInputs, &SchedulingInputDifferences{
			Added:   batchedAdded[differencesTime],
			Removed: batchedRemoved[differencesTime],
			Changed: batchedChanged[differencesTime],
		})
	}
	return mergedInputs
}

// Merge SchedulingInputs function takes in those consolidations and the baseline and returns the merged schedulingInput reconstructed
// Merging is adding, removing and changing each field as appropriate. Since a given time cross-section's changes are mutually exclusive
// with each other (you won't have a removal for the same pod you just added), we can apply these as is.
func mergeSchedulingInputs(iteratingInput *SchedulingInput, differences *SchedulingInputDifferences) {
	iteratingInput.PendingPods = mergePods(iteratingInput.PendingPods, differences)
	iteratingInput.StateNodesWithPods = mergeStateNodesWithPods(iteratingInput.StateNodesWithPods, differences)
	mergeBindings(iteratingInput.Bindings, differences) // Already a map, so can merge in place
	iteratingInput.AllInstanceTypes = mergeInstanceTypes(iteratingInput.AllInstanceTypes, differences)
	mergeNodePoolInstanceTypes(iteratingInput.NodePoolInstanceTypes, differences) // Already a map, so can merge in place
	mergeTopology(iteratingInput.Topology, differences)
}

// TODO: Generalize the functions below to some interface mapping or "lo"-based helper.

// Merge one time's set of differences over the baseline input or its merging iterant.
func mergePods(iteratingPods []*v1.Pod, differences *SchedulingInputDifferences) []*v1.Pod {
	iteratingPodMap := mapPodsByUID(iteratingPods)

	// Add, remove and change pods from the iterating pods
	if differences.Added != nil && !differences.Added.isEmpty() {
		for _, addingPod := range differences.Added.PendingPods {
			iteratingPodMap[addingPod.GetUID()] = addingPod
		}
	}
	if differences.Removed != nil && !differences.Removed.isEmpty() {
		for _, removingPod := range differences.Removed.PendingPods {
			delete(iteratingPodMap, removingPod.GetUID())
		}
	}
	if differences.Changed != nil && !differences.Changed.isEmpty() {
		for _, changingPod := range differences.Changed.PendingPods {
			iteratingPodMap[changingPod.GetUID()] = changingPod
		}
	}

	// Rebuild the iteratingPods slice from the podMap
	mergedPods := []*v1.Pod{}
	for _, pod := range iteratingPodMap {
		mergedPods = append(mergedPods, pod)
	}
	return mergedPods
}

// Helper function to merge the baseline stateNodesWithPods with the set of differences.
func mergeStateNodesWithPods(iteratingStateNodesWithPods []*StateNodeWithPods, differences *SchedulingInputDifferences) []*StateNodeWithPods {
	iteratingStateNodesWithPodsMap := mapStateNodesWithPodsByName(iteratingStateNodesWithPods)

	// Add, remove and change stateNodesWithPods from the iterating stateNodesWithPods
	if differences.Added != nil && !differences.Added.isEmpty() {
		for _, addingStateNodeWithPods := range differences.Added.StateNodesWithPods {
			iteratingStateNodesWithPodsMap[addingStateNodeWithPods.GetName()] = addingStateNodeWithPods
		}
	}
	if differences.Removed != nil && !differences.Removed.isEmpty() {
		for _, removingStateNodeWithPods := range differences.Removed.StateNodesWithPods {
			delete(iteratingStateNodesWithPodsMap, removingStateNodeWithPods.GetName())
		}
	}
	if differences.Changed != nil && !differences.Changed.isEmpty() {
		for _, changingStateNodeWithPods := range differences.Changed.StateNodesWithPods {
			iteratingStateNodesWithPodsMap[changingStateNodeWithPods.GetName()] = changingStateNodeWithPods
		}
	}

	// Rebuild the iteratingStateNodesWithPods slice from the stateNodesWithPodsMap
	mergedStateNodesWithPods := []*StateNodeWithPods{}
	for _, stateNodeWithPods := range iteratingStateNodesWithPodsMap {
		mergedStateNodesWithPods = append(mergedStateNodesWithPods, stateNodeWithPods)
	}
	return mergedStateNodesWithPods
}

// Merge the baseline/iterating bindings with the set of differences.
func mergeBindings(iteratingBindings map[types.NamespacedName]string, differences *SchedulingInputDifferences) {
	// Add, remove and change bindings from the iterating bindings
	if differences.Added != nil && !differences.Added.isEmpty() {
		for name, addingBinding := range differences.Added.Bindings {
			iteratingBindings[name] = addingBinding
		}
	}
	if differences.Removed != nil && !differences.Removed.isEmpty() {
		for name := range differences.Removed.Bindings {
			delete(iteratingBindings, name)
		}
	}
	if differences.Changed != nil && !differences.Changed.isEmpty() {
		for name, changingBinding := range differences.Changed.Bindings {
			iteratingBindings[name] = changingBinding
		}
	}
}

// Helper function to merge the baseline instanceTypes with the set of differences.
func mergeInstanceTypes(iteratingInstanceTypes []*cloudprovider.InstanceType, differences *SchedulingInputDifferences) []*cloudprovider.InstanceType {
	iteratingInstanceTypesMap := MapInstanceTypesByName(iteratingInstanceTypes)

	// Add, remove and change instanceTypes from the iterating instanceTypes
	if differences.Added != nil && !differences.Added.isEmpty() {
		for _, addingInstanceType := range differences.Added.AllInstanceTypes {
			iteratingInstanceTypesMap[addingInstanceType.Name] = addingInstanceType
		}
	}
	if differences.Removed != nil && !differences.Removed.isEmpty() {
		for _, removingInstanceType := range differences.Removed.AllInstanceTypes {
			delete(iteratingInstanceTypesMap, removingInstanceType.Name)
		}
	}
	if differences.Changed != nil && !differences.Changed.isEmpty() {
		for _, changingInstanceType := range differences.Changed.AllInstanceTypes {
			iteratingInstanceTypesMap[changingInstanceType.Name] = changingInstanceType
		}
	}

	// Rebuild the iteratingInstanceTypes slice from the instanceTypesMap
	mergedInstanceTypes := []*cloudprovider.InstanceType{}
	for _, instanceType := range iteratingInstanceTypesMap {
		mergedInstanceTypes = append(mergedInstanceTypes, instanceType)
	}
	return mergedInstanceTypes
}

func mergeNodePoolInstanceTypes(iteratingNodePoolInstanceTypes map[string][]string, differences *SchedulingInputDifferences) {
	// Add, remove and change nodePoolInstanceTypes from the iterating nodePoolInstanceTypes
	if differences.Added != nil && !differences.Added.isEmpty() {
		for nodePoolName, addingInstanceTypes := range differences.Added.NodePoolInstanceTypes {
			iteratingNodePoolInstanceTypes[nodePoolName] = addingInstanceTypes
		}
	}
	if differences.Removed != nil && !differences.Removed.isEmpty() {
		for nodePoolName := range differences.Removed.NodePoolInstanceTypes {
			delete(iteratingNodePoolInstanceTypes, nodePoolName)
		}
	}
	if differences.Changed != nil && !differences.Changed.isEmpty() {
		for nodePoolName, changingInstanceTypes := range differences.Changed.NodePoolInstanceTypes {
			iteratingNodePoolInstanceTypes[nodePoolName] = changingInstanceTypes
		}
	}
}

func mergeTopology(iteratingTopology *scheduler.Topology, differences *SchedulingInputDifferences) {
	if differences.Changed != nil && !differences.Changed.isEmpty() {
		if differences.Changed.Topology != nil {
			iteratingTopology = differences.Changed.Topology
		}
	}
}

// Functions to check the differences in all the fields of a SchedulingInput (except the timestamp)
func (si *SchedulingInput) Diff(oldSi *SchedulingInput) *SchedulingInputDifferences {
	// Determine the differences in each of the fields of ScheduleInput
	podDiff := diffPods(oldSi.PendingPods, si.PendingPods)
	snpDiff := diffStateNodes(oldSi.StateNodesWithPods, si.StateNodesWithPods)
	bindingsDiff := diffBindings(oldSi.Bindings, si.Bindings)
	itDiff := diffInstanceTypes(oldSi.AllInstanceTypes, si.AllInstanceTypes)
	npitDiff := diffNodePoolsToInstanceTypes(oldSi.NodePoolInstanceTypes, si.NodePoolInstanceTypes)
	topologyDiff := diffTopology(oldSi.Topology, si.Topology)

	diffAdded := &SchedulingInput{}
	diffRemoved := &SchedulingInput{}
	diffChanged := &SchedulingInput{}

	// If there are added differences, include them
	if len(podDiff.Added) > 0 || len(snpDiff.Added) > 0 || len(bindingsDiff.Added) > 0 || len(itDiff.Added) > 0 || len(npitDiff.Added) > 0 {
		diffAdded = NewReconstructedSchedulingInput(si.Timestamp, podDiff.Added, snpDiff.Added, bindingsDiff.Added, itDiff.Added, npitDiff.Added, topologyDiff.Added)
		// fmt.Println("Diff Scheduling Input added is... ", diffAdded.String()) // Test print, delete later
	}
	if len(podDiff.Removed) > 0 || len(snpDiff.Removed) > 0 || len(bindingsDiff.Removed) > 0 || len(itDiff.Removed) > 0 || len(npitDiff.Removed) > 0 {
		diffRemoved = NewReconstructedSchedulingInput(si.Timestamp, podDiff.Removed, snpDiff.Removed, bindingsDiff.Removed, itDiff.Removed, npitDiff.Removed, topologyDiff.Removed)
		// fmt.Println("Diff Scheduling Input removed is... ", diffRemoved.String()) // Test print, delete later
	}
	if len(podDiff.Changed) > 0 || len(snpDiff.Changed) > 0 || len(bindingsDiff.Changed) > 0 || len(itDiff.Changed) > 0 || len(npitDiff.Changed) > 0 || (topologyDiff.Changed != nil) {
		diffChanged = NewReconstructedSchedulingInput(si.Timestamp, podDiff.Changed, snpDiff.Changed, bindingsDiff.Changed, itDiff.Changed, npitDiff.Changed, topologyDiff.Changed)
		// fmt.Println("Diff Scheduling Input changed is... ", diffChanged.String()) // Test print, delete later
	}

	return &SchedulingInputDifferences{
		Added:   diffAdded,
		Removed: diffRemoved,
		Changed: diffChanged,
	}
}

// TODO: Generalize Mapping by defining a GetKey function for each resource.
// TODO: Generalize Diff-ing by generalizing the diff functions for generic types and defining change functions generically.

// Converts pod slice to a map from its UID.
func mapPodsByUID(pods []*v1.Pod) map[types.UID]*v1.Pod {
	podMap := map[types.UID]*v1.Pod{}
	for _, pod := range pods {
		podMap[pod.GetUID()] = pod
	}
	return podMap
}

// This is the diffPods function which gets the differences between pods
func diffPods(oldPods, newPods []*v1.Pod) PodDifferences {
	diff := PodDifferences{
		Added:   []*v1.Pod{},
		Removed: []*v1.Pod{},
		Changed: []*v1.Pod{},
	}

	oldPodMap := mapPodsByUID(oldPods)
	oldPodSet := sets.KeySet(oldPodMap)
	newPodMap := mapPodsByUID(newPods)
	newPodSet := sets.KeySet(newPodMap)

	// Add the new pods to Pod Differences
	for addedUID := range newPodSet.Difference(oldPodSet) {
		diff.Added = append(diff.Added, newPodMap[addedUID])
	}
	// Add the removed pods to Pod Differences
	for removedUID := range oldPodSet.Difference(newPodSet) {
		diff.Removed = append(diff.Removed, oldPodMap[removedUID])
	}
	// Add the changed pods to Pod Differences, only after checking if they've changed.
	// Simplification / Opportunity to optimize -- Only add sub-field.
	//    This requires more book-keeping on object reconstruction from logs later on.
	for commonUID := range newPodSet.Intersection(oldPodSet) {
		if hasReducedPodChanged(oldPodMap[commonUID], newPodMap[commonUID]) {
			diff.Changed = append(diff.Changed, newPodMap[commonUID])
		}
	}
	return diff
}

// Function from StateNodeWithPods slice to Map by name
func mapStateNodesWithPodsByName(stateNodesWithPods []*StateNodeWithPods) map[string]*StateNodeWithPods {
	snpMap := map[string]*StateNodeWithPods{}
	for _, snp := range stateNodesWithPods {
		snpMap[snp.GetName()] = snp
	}
	return snpMap
}

// This is the diffStateNodes function which gets the differences between statenodes
func diffStateNodes(oldStateNodesWithPods, newStateNodesWithPods []*StateNodeWithPods) SNPDifferences {
	diff := SNPDifferences{
		Added:   []*StateNodeWithPods{},
		Removed: []*StateNodeWithPods{},
		Changed: []*StateNodeWithPods{},
	}

	// Cast StateNodesWithPods slices to sets for unordered reference by their name
	oldStateNodeMap := mapStateNodesWithPodsByName(oldStateNodesWithPods)
	oldStateNodeSet := sets.KeySet(oldStateNodeMap)
	newStateNodeMap := mapStateNodesWithPodsByName(newStateNodesWithPods)
	newStateNodeSet := sets.KeySet(newStateNodeMap)

	// Find the added, removed and changed stateNodes
	for addedName := range newStateNodeSet.Difference(oldStateNodeSet) {
		diff.Added = append(diff.Added, newStateNodeMap[addedName])
	}
	for removedName := range oldStateNodeSet.Difference(newStateNodeSet) {
		diff.Removed = append(diff.Removed, oldStateNodeMap[removedName])
	}
	for commonName := range newStateNodeSet.Intersection(oldStateNodeSet) {
		if hasStateNodeWithPodsChanged(oldStateNodeMap[commonName], newStateNodeMap[commonName]) {
			diff.Changed = append(diff.Changed, newStateNodeMap[commonName])
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
	for namespacedName, binding := range old {
		if newBinding, ok := new[namespacedName]; ok {
			if binding != newBinding {
				diff.Changed[namespacedName] = newBinding
			}
		} else {
			diff.Removed[namespacedName] = binding
		}
	}
	// Find the added bindings
	for namespacedName, binding := range new {
		if _, ok := old[namespacedName]; !ok {
			diff.Added[namespacedName] = binding
		}
	}
	return diff
}

// Take differenceTypes to map like above, by name
func MapInstanceTypesByName(instanceTypes []*cloudprovider.InstanceType) map[string]*cloudprovider.InstanceType {
	itMap := map[string]*cloudprovider.InstanceType{}
	for _, it := range instanceTypes {
		itMap[it.Name] = it
	}
	return itMap
}

// This is the diffInstanceTypes function which gets the differences between instance types
func diffInstanceTypes(oldTypes, newTypes []*cloudprovider.InstanceType) InstanceTypeDifferences {
	diff := InstanceTypeDifferences{
		Added:   []*cloudprovider.InstanceType{},
		Removed: []*cloudprovider.InstanceType{},
		Changed: []*cloudprovider.InstanceType{},
	}

	oldTypeMap := MapInstanceTypesByName(oldTypes)
	oldTypeSet := sets.KeySet(oldTypeMap)
	newTypeMap := MapInstanceTypesByName(newTypes)
	newTypeSet := sets.KeySet(newTypeMap)

	// Find the added, removed and changed instanceTypes
	for addedName := range newTypeSet.Difference(oldTypeSet) {
		diff.Added = append(diff.Added, newTypeMap[addedName])
	}
	for removedName := range oldTypeSet.Difference(newTypeSet) {
		diff.Removed = append(diff.Removed, oldTypeMap[removedName])
	}
	for commonName := range newTypeSet.Intersection(oldTypeSet) {
		if hasReducedInstanceTypeChanged(oldTypeMap[commonName], newTypeMap[commonName]) {
			diff.Changed = append(diff.Changed, newTypeMap[commonName])
		}
	}
	return diff
}

// This function is already a mapping, so it will mirror the diff for Bindings in the map checking
func diffNodePoolsToInstanceTypes(old, new map[string][]string) NodePoolsToInstanceTypesDifferences {
	diff := NodePoolsToInstanceTypesDifferences{
		Added:   map[string][]string{},
		Removed: map[string][]string{},
		Changed: map[string][]string{},
	}

	// Find the changed or removed node pools
	for nodepool, instancetypes := range old {
		if newInstanceTypes, ok := new[nodepool]; ok {
			if !equality.Semantic.DeepEqual(sets.NewString(instancetypes...), sets.NewString(newInstanceTypes...)) {
				diff.Changed[nodepool] = newInstanceTypes
			}
		} else {
			diff.Removed[nodepool] = instancetypes
		}
	}
	// Find the added node pools
	for nodepool, instancetypes := range new {
		if _, ok := old[nodepool]; !ok {
			diff.Added[nodepool] = instancetypes
		}
	}
	return diff
}

func diffTopology(oldTopology, newTopology *scheduler.Topology) TopologyDifferences {
	diff := TopologyDifferences{
		Added:   nil, // Empty for Topology by construct (only coarsely checking) data structure equality, not for each recursive internal differences
		Removed: nil, // ^
		Changed: nil, // Only these matter, but keeping the "Differences" construct so I can (TODO) make a general interface{} later and simplify
	}

	if !structEqualJSON(oldTopology, newTopology) {
		diff.Changed = newTopology
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
