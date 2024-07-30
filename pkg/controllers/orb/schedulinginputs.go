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
	"encoding/json"
	"fmt"
	"time"

	proto "google.golang.org/protobuf/proto"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/karpenter/pkg/apis/v1beta1"
	"sigs.k8s.io/karpenter/pkg/cloudprovider"
	pb "sigs.k8s.io/karpenter/pkg/controllers/orb/proto"
	scheduler "sigs.k8s.io/karpenter/pkg/controllers/provisioning/scheduling"
	"sigs.k8s.io/karpenter/pkg/controllers/state"
	"sigs.k8s.io/karpenter/pkg/scheduling"
	"sigs.k8s.io/karpenter/pkg/utils/pretty"
)

type BindingsMap map[types.NamespacedName]string // Alias to allow JSON Marshal definition
// type Topology *scheduler.Topology                // Alias for JSON Marshaling

// These are the inputs to the scheduling function (scheduler.NewSchedule) which change more dynamically
type SchedulingInput struct {
	Timestamp             time.Time
	PendingPods           []*v1.Pod
	StateNodesWithPods    []*StateNodeWithPods
	Bindings              BindingsMap
	AllInstanceTypes      []*cloudprovider.InstanceType
	NodePoolInstanceTypes map[string][]string
	Topology              *scheduler.Topology
	DaemonSetPods         []*v1.Pod
	PVList                *v1.PersistentVolumeList
	PVCList               *v1.PersistentVolumeClaimList
	ScheduledPodList      *v1.PodList
}

// A stateNode with the Pods it has on it.
type StateNodeWithPods struct {
	Node      *v1.Node
	NodeClaim *v1beta1.NodeClaim
	Pods      []*v1.Pod
}

// // These are the fields of topology not associated with the kubeClient or Cluster.
// type IsolatedTopology struct {
// 	// Both the topologies and inverseTopologies are maps of the hash from TopologyGroup.Hash() to the topology group
// 	// itself. This is used to allow us to store one topology group that tracks the topology of many pods instead of
// 	// having a 1<->1 mapping between topology groups and pods owned/selected by that group.
// 	topologies map[uint64]*scheduler.TopologyGroup
// 	// Anti-affinity works both ways (if a zone has a pod foo with anti-affinity to a pod bar, we can't schedule bar to
// 	// that zone, even though bar has no anti affinity terms on it. For this to work, we need to separately track the
// 	// topologies of pods with anti-affinity terms, so we can prevent scheduling the pods they have anti-affinity to
// 	// in some cases.
// 	inverseTopologies map[uint64]*scheduler.TopologyGroup
// 	// The universe of domains by topology key
// 	domains map[string]sets.Set[string]
// 	// excludedPods are the pod UIDs of pods that are excluded from counting.  This is used so we can simulate
// 	// moving pods to prevent them from being double counted.
// 	excludedPods sets.Set[string]
// }

// Construct and reduce the Scheduling Input down to what's minimally required for re-simulation
func NewSchedulingInput(ctx context.Context, kubeClient client.Client, scheduledTime time.Time, pendingPods []*v1.Pod, stateNodes []*state.StateNode,
	bindings map[types.NamespacedName]string, instanceTypes map[string][]*cloudprovider.InstanceType, topology *scheduler.Topology, daemonSetPods []*v1.Pod) SchedulingInput {
	allInstanceTypes, nodePoolInstanceTypes := getAllInstanceTypesAndNodePoolMapping(instanceTypes)
	// Get all PVs and PVCs from kubeClient
	pvcList := &v1.PersistentVolumeClaimList{}
	err := kubeClient.List(ctx, pvcList)
	if err != nil {
		fmt.Println("PVC List error in Scheduling Input logging: ", err)
		return SchedulingInput{}
	}
	pvList := &v1.PersistentVolumeList{}
	err = kubeClient.List(ctx, pvList)
	if err != nil {
		fmt.Println("PV List error in Scheduling Input logging: ", err)
		return SchedulingInput{}
	}
	scheduledPodList := &v1.PodList{}
	// Get all pods from kubeClient
	err = kubeClient.List(ctx, scheduledPodList)
	if err != nil {
		fmt.Println("Pod List error in Scheduling Input logging: ", err)
		return SchedulingInput{}
	}

	return SchedulingInput{
		Timestamp:             scheduledTime,
		PendingPods:           pendingPods,
		StateNodesWithPods:    newStateNodesWithPods(ctx, kubeClient, stateNodes),
		Bindings:              bindings,
		AllInstanceTypes:      allInstanceTypes,
		NodePoolInstanceTypes: nodePoolInstanceTypes,
		Topology:              topology,
		DaemonSetPods:         daemonSetPods,
		PVList:                pvList,
		PVCList:               pvcList,
		ScheduledPodList:      scheduledPodList,
	}
}

func NewReconstructedSchedulingInput(timestamp time.Time, pendingPods []*v1.Pod, stateNodesWithPods []*StateNodeWithPods, bindings map[types.NamespacedName]string,
	instanceTypes []*cloudprovider.InstanceType, nodePoolInstanceTypes map[string][]string, topology *scheduler.Topology, daemonSetPods []*v1.Pod,
	pvList *v1.PersistentVolumeList, pvcList *v1.PersistentVolumeClaimList, scheduledPodList *v1.PodList) *SchedulingInput {
	return &SchedulingInput{
		Timestamp:             timestamp,
		PendingPods:           pendingPods,
		StateNodesWithPods:    stateNodesWithPods,
		Bindings:              bindings,
		AllInstanceTypes:      instanceTypes,
		NodePoolInstanceTypes: nodePoolInstanceTypes,
		Topology:              topology,
		DaemonSetPods:         daemonSetPods,
		PVList:                pvList,
		PVCList:               pvcList,
		ScheduledPodList:      scheduledPodList,
	}
}

// Mirrors StateNode's GetName
func (snp StateNodeWithPods) GetName() string {
	if snp.Node == nil {
		return snp.NodeClaim.GetName()
	}
	return snp.Node.GetName()
}

func (si *SchedulingInput) Reduce() {
	si.PendingPods = ReducePods(si.PendingPods)
	// si.AllInstanceTypes = ReduceInstanceTypes(si.AllInstanceTypes)
}

func (si SchedulingInput) String() string {
	return protoSchedulingInput(&si).String()
}

// TODO: I don't think this is marshalling Requirements appropriately. It has Key, but no values. String() gathers that information, so it definitely exists.
func (si SchedulingInput) Json() string {
	return pretty.Concise(si)
}

func (m BindingsMap) MarshalJSON() ([]byte, error) {
	temp := map[string]interface{}{}
	for k, v := range m {
		temp[k.String()] = v
	}
	return json.Marshal(temp)
}

func (si *SchedulingInput) isEmpty() bool {
	return len(si.PendingPods) == 0 &&
		len(si.StateNodesWithPods) == 0 &&
		len(si.Bindings) == 0 &&
		len(si.AllInstanceTypes) == 0
}

func newStateNodesWithPods(ctx context.Context, kubeClient client.Client, stateNodes []*state.StateNode) []*StateNodeWithPods {
	stateNodesWithPods := []*StateNodeWithPods{}
	for _, stateNode := range reduceStateNodes(stateNodes) {
		pods, err := stateNode.Pods(ctx, kubeClient)
		if err != nil {
			pods = nil
		}

		stateNodesWithPods = append(stateNodesWithPods, &StateNodeWithPods{
			Node:      stateNode.Node,
			NodeClaim: stateNode.NodeClaim,
			Pods:      ReducePods(pods),
		})
	}
	return stateNodesWithPods
}

// Gets the superset of all InstanceTypes from the mapping. This is to simplify saving the NodePool -> InstanceType map
// by allowing us to save all of them by their unique name, and then associating the NodePool name with its corresponding instancetype names
func getAllInstanceTypesAndNodePoolMapping(instanceTypes map[string][]*cloudprovider.InstanceType) ([]*cloudprovider.InstanceType, map[string][]string) {
	allInstanceTypesNameMap := map[string]*cloudprovider.InstanceType{}
	nodePoolToInstanceTypes := map[string][]string{}
	for nodePool, instanceTypeSlice := range instanceTypes {
		instanceTypeSliceNameMap := MapInstanceTypesByName(instanceTypeSlice)
		for instanceTypeName, instanceType := range instanceTypeSliceNameMap {
			allInstanceTypesNameMap[instanceTypeName] = instanceType
		}
		nodePoolToInstanceTypes[nodePool] = sets.KeySet(instanceTypeSliceNameMap).UnsortedList()
	}
	uniqueInstanceTypeNames := sets.KeySet(allInstanceTypesNameMap).UnsortedList()
	uniqueInstanceTypes := []*cloudprovider.InstanceType{}
	for _, instanceTypeName := range uniqueInstanceTypeNames {
		uniqueInstanceTypes = append(uniqueInstanceTypes, allInstanceTypesNameMap[instanceTypeName])
	}
	return uniqueInstanceTypes, nodePoolToInstanceTypes
}

/* Functions to reduce resources in Scheduling Inputs to the constituent parts we care to log / introspect */

func ReducePods(pods []*v1.Pod) []*v1.Pod {
	reducedPods := []*v1.Pod{}
	for _, pod := range pods {
		reducedPod := &v1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      pod.Name,
				Namespace: pod.Namespace,
				UID:       pod.GetUID(),
			},
			Status: v1.PodStatus{
				Phase:      pod.Status.Phase,
				Conditions: reducePodConditions(pod.Status.Conditions),
			},
		}
		reducedPods = append(reducedPods, reducedPod)
	}
	return reducedPods
}

func reducePodConditions(conditions []v1.PodCondition) []v1.PodCondition {
	reducedConditions := []v1.PodCondition{}
	for _, condition := range conditions {
		reducedCondition := v1.PodCondition{
			Type:    condition.Type,
			Status:  condition.Status,
			Reason:  condition.Reason,
			Message: condition.Message,
		}
		reducedConditions = append(reducedConditions, reducedCondition)
	}
	return reducedConditions
}

func reduceStateNodes(nodes []*state.StateNode) []*state.StateNode {
	reducedStateNodes := []*state.StateNode{}
	for _, node := range nodes {
		if node != nil {
			reducedStateNode := &state.StateNode{}
			reducedStateNode.Node = node.Node
			reducedStateNode.NodeClaim = node.NodeClaim
			if reducedStateNode.Node != nil || reducedStateNode.NodeClaim != nil {
				reducedStateNodes = append(reducedStateNodes, reducedStateNode)
			}
		}
	}
	return reducedStateNodes
}

// func reduceOfferings(offerings cloudprovider.Offerings) cloudprovider.Offerings {
// 	strippedOfferings := cloudprovider.Offerings{}
// 	for _, offering := range offerings {
// 		strippedOffering := &cloudprovider.Offering{
// 			Requirements: reduceRequirements(offering.Requirements), // TODO: should I be reducing these?
// 			Price:        offering.Price,
// 			Available:    offering.Available,
// 		}
// 		strippedOfferings = append(strippedOfferings, *strippedOffering)
// 	}
// 	return strippedOfferings
// }

// // Reduce Requirements returns Requirements of these keys: karpenter.sh/capacity-type, topology.k8s.aws/zone-id and topology.kubernetes.io/zone
// // TODO Should these keys be called more generically? i.e. via v1beta1.CapacityTypeLabelKey, v1.LabelTopologyZone or something?
// func reduceRequirements(requirements scheduling.Requirements) scheduling.Requirements {
// 	reducedRequirements := scheduling.Requirements{}
// 	for key, value := range requirements {
// 		switch key {
// 		case "karpenter.sh/capacity-type", "topology.k8s.aws/zone-id", "topology.kubernetes.io/zone":
// 			reducedRequirements[key] = value
// 		}
// 	}
// 	return reducedRequirements
// }

// func ReduceInstanceTypes(its []*cloudprovider.InstanceType) []*cloudprovider.InstanceType {
// 	reducedInstanceTypes := []*cloudprovider.InstanceType{}
// 	for _, it := range its {
// 		reducedInstanceType := &cloudprovider.InstanceType{
// 			Name:         it.Name,
// 			Requirements: reduceRequirements(it.Requirements),
// 			Offerings:    reduceOfferings(it.Offerings),
// 		}
// 		reducedInstanceTypes = append(reducedInstanceTypes, reducedInstanceType)
// 	}
// 	return reducedInstanceTypes
// }

/* Functions to convert between SchedulingInputs and the proto-defined version
   Via pairs: Marshal <--> Unmarshal and proto <--> reconstruct */

func MarshalSchedulingInput(si *SchedulingInput) ([]byte, error) {
	return proto.Marshal(protoSchedulingInput(si))
}

func UnmarshalSchedulingInput(schedulingInputData []byte) (*SchedulingInput, error) {
	entry := &pb.SchedulingInput{}

	if err := proto.Unmarshal(schedulingInputData, entry); err != nil {
		return nil, fmt.Errorf("failed to unmarshal SchedulingInput: %v", err)
	}

	si, err := reconstructSchedulingInput(entry)
	if err != nil {
		return nil, fmt.Errorf("failed to reconstruct SchedulingInput: %v", err)
	}
	return si, nil
}

func protoSchedulingInput(si *SchedulingInput) *pb.SchedulingInput {
	return &pb.SchedulingInput{
		Timestamp:                    si.Timestamp.Format("2006-01-02_15-04-05"),
		PendingpodData:               protoPods(si.PendingPods),
		BindingsData:                 protoBindings(si.Bindings),
		StatenodesData:               protoStateNodesWithPods(si.StateNodesWithPods),
		InstancetypesData:            protoInstanceTypes(si.AllInstanceTypes),
		NodepoolstoinstancetypesData: protoNodePoolInstanceTypes(si.NodePoolInstanceTypes),
		TopologyData:                 protoTopology(si.Topology),
		DaemonsetpodsData:            protoDaemonSetPods(si.DaemonSetPods),
		PvlistData:                   protoPVList(si.PVList),
		PvclistData:                  protoPVCList(si.PVCList),
		ScheduledpodlistData:         protoScheduledPodList(si.ScheduledPodList),
	}
}

func reconstructSchedulingInput(pbsi *pb.SchedulingInput) (*SchedulingInput, error) {
	timestamp, err := time.Parse("2006-01-02_15-04-05", pbsi.GetTimestamp())
	if err != nil {
		return nil, fmt.Errorf("failed to parse timestamp: %v", err)
	}

	return NewReconstructedSchedulingInput(
		timestamp,
		reconstructPods(pbsi.GetPendingpodData()),
		reconstructStateNodesWithPods(pbsi.GetStatenodesData()),
		reconstructBindings(pbsi.GetBindingsData()),
		reconstructInstanceTypes(pbsi.GetInstancetypesData()),
		reconstructNodePoolInstanceTypes(pbsi.GetNodepoolstoinstancetypesData()),
		reconstructTopology(pbsi.GetTopologyData()),
		reconstructDaemonSetPods(pbsi.GetDaemonsetpodsData()),
		reconstructPVList(pbsi.GetPvlistData()),
		reconstructPVCList(pbsi.GetPvclistData()),
		reconstructScheduledPodList(pbsi.GetScheduledpodlistData()),
	), nil
}

func protoPods(pods []*v1.Pod) []*pb.ReducedPod {
	reducedPods := []*pb.ReducedPod{}
	for _, pod := range pods {
		reducedPod := &pb.ReducedPod{
			Name:       pod.Name,
			Namespace:  pod.Namespace,
			Uid:        string(pod.GetUID()),
			Phase:      string(pod.Status.Phase),
			Conditions: protoPodConditions(pod.Status.Conditions),
		}
		reducedPods = append(reducedPods, reducedPod)
	}
	return reducedPods
}

func reconstructPods(reducedPods []*pb.ReducedPod) []*v1.Pod {
	pods := []*v1.Pod{}
	for _, reducedPod := range reducedPods {
		pod := &v1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      reducedPod.Name,
				Namespace: reducedPod.Namespace,
				UID:       types.UID(reducedPod.Uid),
			},
			Status: v1.PodStatus{
				Phase:      v1.PodPhase(reducedPod.Phase),
				Conditions: reconstructPodConditions(reducedPod.Conditions),
			},
		}
		pods = append(pods, pod)
	}
	return pods
}

func protoPodConditions(conditions []v1.PodCondition) []*pb.ReducedPod_PodCondition {
	reducedPodConditions := []*pb.ReducedPod_PodCondition{}
	for _, condition := range conditions {
		reducedPodCondition := &pb.ReducedPod_PodCondition{
			Type:    string(condition.Type),
			Status:  string(condition.Status),
			Reason:  condition.Reason,
			Message: condition.Message,
		}
		reducedPodConditions = append(reducedPodConditions, reducedPodCondition)
	}
	return reducedPodConditions
}

func reconstructPodConditions(reducedPodConditions []*pb.ReducedPod_PodCondition) []v1.PodCondition {
	podConditions := []v1.PodCondition{}
	for _, reducedPodCondition := range reducedPodConditions {
		podCondition := v1.PodCondition{
			Type:    v1.PodConditionType(reducedPodCondition.Type),
			Status:  v1.ConditionStatus(reducedPodCondition.Status),
			Reason:  reducedPodCondition.Reason,
			Message: reducedPodCondition.Message,
		}
		podConditions = append(podConditions, podCondition)
	}
	return podConditions
}

func protoStateNodesWithPods(stateNodesWithPods []*StateNodeWithPods) []*pb.StateNodeWithPods {
	snpData := []*pb.StateNodeWithPods{}
	for _, snp := range stateNodesWithPods {
		nodeData := []byte{}
		err := error(nil)
		if snp.Node != nil {
			nodeData, err = snp.Node.Marshal()
			if err != nil {
				continue // There is no Node, maybe there's a nodeclaim
			}
		}
		nodeClaimData := []byte{}
		if snp.NodeClaim != nil {
			nodeClaimData, err = json.Marshal(snp.NodeClaim)
			if err != nil {
				continue // There is no NodeClaim
			}
		}
		snpData = append(snpData, &pb.StateNodeWithPods{
			Node:      nodeData,
			NodeClaim: nodeClaimData,
			Pods:      protoPods(snp.Pods),
		})
	}
	return snpData
}

func reconstructStateNodesWithPods(snpData []*pb.StateNodeWithPods) []*StateNodeWithPods {
	stateNodesWithPods := []*StateNodeWithPods{}
	for _, snpData := range snpData {
		node := &v1.Node{}
		node.Unmarshal(snpData.Node)
		nodeClaim := &v1beta1.NodeClaim{}
		json.Unmarshal(snpData.NodeClaim, nodeClaim)

		stateNodesWithPods = append(stateNodesWithPods, &StateNodeWithPods{
			Node:      node,
			NodeClaim: nodeClaim,
			Pods:      reconstructPods(snpData.Pods),
		})
	}
	return stateNodesWithPods
}

func protoInstanceTypes(instanceTypes []*cloudprovider.InstanceType) []*pb.InstanceType {
	itData := []*pb.InstanceType{}
	for _, it := range instanceTypes {
		itData = append(itData, &pb.InstanceType{
			Name:         it.Name,
			Requirements: protoRequirements(it.Requirements),
			Offerings:    protoOfferings(it.Offerings),
			Capacity:     protoCapacity(it.Capacity),
			Overhead:     protoOverhead(it.Overhead),
		})
	}
	return itData
}

func reconstructInstanceTypes(itData []*pb.InstanceType) []*cloudprovider.InstanceType {
	instanceTypes := []*cloudprovider.InstanceType{}
	for _, it := range itData {
		instanceTypes = append(instanceTypes, &cloudprovider.InstanceType{
			Name:         it.Name,
			Requirements: reconstructRequirements(it.Requirements),
			Offerings:    reconstructOfferings(it.Offerings),
			Capacity:     reconstructResourceList(it.Capacity),
			Overhead:     reconstructOverhead(it.Overhead),
		})
	}
	return instanceTypes
}

func protoRequirements(requirements scheduling.Requirements) []*pb.InstanceType_Requirement {
	requirementsData := []*pb.InstanceType_Requirement{}
	for _, requirement := range requirements {
		requirementsData = append(requirementsData, &pb.InstanceType_Requirement{
			Key:                  requirement.Key,
			Nodeselectoroperator: string(requirement.Operator()),
			Values:               requirement.Values(),
		})
	}
	return requirementsData
}

func reconstructRequirements(requirementsData []*pb.InstanceType_Requirement) scheduling.Requirements {
	requirements := scheduling.Requirements{}
	for _, requirementData := range requirementsData {
		requirements.Add(scheduling.NewRequirement(
			requirementData.Key,
			v1.NodeSelectorOperator(requirementData.Nodeselectoroperator),
			requirementData.Values...,
		))
	}
	return requirements
}

func protoOfferings(offerings cloudprovider.Offerings) []*pb.InstanceType_Offering {
	offeringsData := []*pb.InstanceType_Offering{}
	for _, offering := range offerings {
		offeringsData = append(offeringsData, &pb.InstanceType_Offering{
			Requirements: protoRequirements(offering.Requirements),
			Price:        offering.Price,
			Available:    offering.Available,
		})
	}
	return offeringsData
}

func reconstructOfferings(offeringsData []*pb.InstanceType_Offering) cloudprovider.Offerings {
	offerings := cloudprovider.Offerings{}
	for _, offeringData := range offeringsData {
		offerings = append(offerings, cloudprovider.Offering{
			Requirements: reconstructRequirements(offeringData.Requirements),
			Price:        offeringData.Price,
			Available:    offeringData.Available,
		})
	}
	return offerings
}

func protoResourceList(resourceList v1.ResourceList) *pb.InstanceType_ResourceList {
	protoResourceList := &pb.InstanceType_ResourceList{}
	for resourceName, quantity := range resourceList {
		protoQuantity, err := quantity.Marshal()
		if err != nil {
			fmt.Println("cannot marshal quantity in protoResourceList")
		}

		protoResourceList.Resources = append(protoResourceList.Resources, &pb.InstanceType_ResourceQuantity{
			ResourceName: string(resourceName),
			Quantity:     protoQuantity,
		})
	}
	return protoResourceList
}

func protoCapacity(capacity v1.ResourceList) *pb.InstanceType_ResourceList {
	return protoResourceList(capacity)
}

func reconstructResourceList(protoCapacity *pb.InstanceType_ResourceList) v1.ResourceList {
	capacity := v1.ResourceList{}
	for _, resourceQuantity := range protoCapacity.Resources {
		quantity := &resource.Quantity{}
		quantity.Unmarshal(resourceQuantity.Quantity)
		capacity[v1.ResourceName(resourceQuantity.ResourceName)] = *quantity
	}
	return capacity
}

func protoOverhead(overhead *cloudprovider.InstanceTypeOverhead) *pb.InstanceType_Overhead {
	return &pb.InstanceType_Overhead{
		Kubereserved:      protoResourceList(overhead.KubeReserved),
		Systemreserved:    protoResourceList(overhead.SystemReserved),
		Evictionthreshold: protoResourceList(overhead.EvictionThreshold),
	}
}

func reconstructOverhead(protoOverhead *pb.InstanceType_Overhead) *cloudprovider.InstanceTypeOverhead {
	return &cloudprovider.InstanceTypeOverhead{
		KubeReserved:      reconstructResourceList(protoOverhead.Kubereserved),
		SystemReserved:    reconstructResourceList(protoOverhead.Systemreserved),
		EvictionThreshold: reconstructResourceList(protoOverhead.Evictionthreshold),
	}
}

func protoBindings(bindings map[types.NamespacedName]string) *pb.Bindings {
	bindingsProto := &pb.Bindings{}
	for podNamespacedName, nodeName := range bindings {
		binding := &pb.Bindings_Binding{
			PodNamespacedName: &pb.Bindings_Binding_NamespacedName{
				Namespace: podNamespacedName.Namespace,
				Name:      podNamespacedName.Name,
			},
			NodeName: nodeName,
		}
		bindingsProto.Binding = append(bindingsProto.Binding, binding)
	}
	return bindingsProto
}

func reconstructBindings(bindingsProto *pb.Bindings) map[types.NamespacedName]string {
	bindings := map[types.NamespacedName]string{}
	for _, binding := range bindingsProto.Binding {
		podNamespacedName := types.NamespacedName{
			Namespace: binding.PodNamespacedName.Namespace,
			Name:      binding.PodNamespacedName.Name,
		}
		bindings[podNamespacedName] = binding.NodeName
	}
	return bindings
}

func protoNodePoolInstanceTypes(nodePoolInstanceTypes map[string][]string) *pb.NodePoolsToInstanceTypes {
	npitProto := &pb.NodePoolsToInstanceTypes{}
	for nodePool, instanceTypeNames := range nodePoolInstanceTypes {
		npitProto.Nodepoolstoinstancetypes = append(npitProto.Nodepoolstoinstancetypes, &pb.NodePoolsToInstanceTypes_NodePoolToInstanceTypes{
			Nodepool:         nodePool,
			InstancetypeName: instanceTypeNames,
		})
	}
	return npitProto
}

func reconstructNodePoolInstanceTypes(npitProto *pb.NodePoolsToInstanceTypes) map[string][]string {
	nodePoolInstanceTypes := map[string][]string{}
	for _, npit := range npitProto.Nodepoolstoinstancetypes {
		nodePoolInstanceTypes[npit.Nodepool] = npit.InstancetypeName
	}
	return nodePoolInstanceTypes
}

func protoTopology(topology *scheduler.Topology) []byte {
	if topology == nil {
		return []byte{}
	}
	topologyData, err := json.Marshal(topology)
	if err != nil {
		fmt.Println("Error marshaling topology to JSON", err)
		return []byte{}
	}
	return topologyData
}

func reconstructTopology(topologyData []byte) *scheduler.Topology {
	topology := &scheduler.Topology{}
	json.Unmarshal(topologyData, topology)
	return topology
}

func protoDaemonSetPods(daemonSetPods []*v1.Pod) []byte {
	podList := &v1.PodList{
		Items: []v1.Pod{},
	}
	for _, pod := range daemonSetPods {
		podList.Items = append(podList.Items, *pod)
	}

	dspData, err := podList.Marshal()
	if err != nil {
		fmt.Println("Error marshaling DaemonSetPods to JSON", err)
		return []byte{}
	}
	return dspData
}

func reconstructDaemonSetPods(dspData []byte) []*v1.Pod {
	podList := &v1.PodList{}
	podList.Unmarshal(dspData)

	daemonSetPods := []*v1.Pod{}

	for _, pod := range podList.Items {
		daemonSetPods = append(daemonSetPods, &pod)
	}
	return daemonSetPods
}

func protoPVList(pvList *v1.PersistentVolumeList) []byte {
	if pvList == nil {
		return []byte{}
	}
	pvListData, err := pvList.Marshal()
	if err != nil {
		fmt.Println("Error marshaling PVList to JSON", err)
		return []byte{}
	}
	return pvListData
}

func reconstructPVList(pvListData []byte) *v1.PersistentVolumeList {
	pvList := &v1.PersistentVolumeList{}
	pvList.Unmarshal(pvListData)
	return pvList
}

func protoPVCList(pvcList *v1.PersistentVolumeClaimList) []byte {
	if pvcList == nil {
		return []byte{}
	}
	pvcListData, err := pvcList.Marshal()
	if err != nil {
		fmt.Println("Error marshaling PVCList to JSON", err)
		return []byte{}
	}
	return pvcListData
}

func reconstructPVCList(pvcListData []byte) *v1.PersistentVolumeClaimList {
	pvcList := &v1.PersistentVolumeClaimList{}
	pvcList.Unmarshal(pvcListData)
	return pvcList
}

func protoScheduledPodList(scheduledPodList *v1.PodList) []byte {
	if scheduledPodList == nil {
		return []byte{}
	}
	scheduledPodListData, err := scheduledPodList.Marshal()
	if err != nil {
		fmt.Println("Error marshaling ScheduledPodList to JSON", err)
		return []byte{}
	}
	return scheduledPodListData
}

func reconstructScheduledPodList(scheduledPodListData []byte) *v1.PodList {
	scheduledPodList := &v1.PodList{}
	scheduledPodList.Unmarshal(scheduledPodListData)
	return scheduledPodList
}
