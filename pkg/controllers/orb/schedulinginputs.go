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
	"fmt"
	"time"

	"github.com/samber/lo"
	"sigs.k8s.io/karpenter/pkg/apis/v1beta1"
	"sigs.k8s.io/karpenter/pkg/cloudprovider"
	"sigs.k8s.io/karpenter/pkg/controllers/state"
	"sigs.k8s.io/karpenter/pkg/scheduling"

	//"google.golang.org/protobuf/proto"
	proto "github.com/gogo/protobuf/proto"
	v1 "k8s.io/api/core/v1"
)

// Timestamp, dynamic inputs (like pending pods, statenodes, etc.)
type SchedulingInput struct {
	Timestamp     time.Time
	PendingPods   []*v1.Pod
	StateNodes    []*state.StateNode
	InstanceTypes []*cloudprovider.InstanceType
	//all the other scheduling inputs...
}

func (si SchedulingInput) String() string {
	return fmt.Sprintf("Timestamp: %v\nPendingPods:\n%vStateNodes:\n%vInstanceTypes:\n%v",
		si.Timestamp.Format("2006-01-02_15-04-05"),
		PodsToString(si.PendingPods),
		StateNodesToString(si.StateNodes),
		InstanceTypesToString(si.InstanceTypes),
	)

}

// Function takes a slice of pod pointers and returns a string representation of the pods

// Function take a Scheduling Input to []byte, marshalled as a protobuf
// TODO: With a custom-defined .proto, this will look different.
func (si SchedulingInput) Marshal() ([]byte, error) {
	podList := &v1.PodList{
		Items: make([]v1.Pod, 0, len(si.PendingPods)),
	}

	for _, podPtr := range si.PendingPods {
		podList.Items = append(podList.Items, *podPtr)
	}
	return podList.Marshal()

	// // Create a slice to store the wire format data
	// podDataSlice := make([][]byte, 0, len(si.PendingPods))

	// // Iterate over the slice of Pods and marshal each one to its wire format
	// for _, pod := range si.PendingPods {
	// 	podData, err := proto.Marshal(pod)
	// 	if err != nil {
	// 		fmt.Println("Error marshaling pod:", err)
	// 		continue
	// 	}
	// 	podDataSlice = append(podDataSlice, podData)
	// }

	// // Create an ORBLogEntry message
	// entry := &ORBLogEntry{
	// 	Timestamp:      si.Timestamp.Format("2006-01-02_15-04-05"),
	// 	PendingpodData: podDataSlice,
	// }

	// return proto.Marshal(entry)
}

// func UnmarshalSchedulingInput(data []byte) (*SchedulingInput, error) {
// 	// Unmarshal the data into an ORBLogEntry struct
// 	entry := &ORBLogEntry{}
// 	if err := proto.Unmarshal(data, entry); err != nil {
// 		return nil, fmt.Errorf("failed to unmarshal ORBLogEntry: %v", err)
// 	}

// 	// Parse the timestamp
// 	timestamp, err := time.Parse("2006-01-02_15-04-05", entry.Timestamp)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to parse timestamp: %v", err)
// 	}

// 	// Unmarshal the PendingpodData into v1.Pod objects
// 	pendingPods := make([]*v1.Pod, 0, len(entry.PendingpodData))
// 	for _, podData := range entry.PendingpodData {
// 		var pod v1.Pod
// 		if err := proto.Unmarshal(podData, &pod); err != nil {
// 			return nil, fmt.Errorf("failed to unmarshal pod: %v", err)
// 		}
// 		pendingPods = append(pendingPods, &pod)
// 	}

// 	// Create a new SchedulingInput struct
// 	schedulingInput := &SchedulingInput{
// 		Timestamp:   timestamp,
// 		PendingPods: pendingPods,
// 	}

// 	return schedulingInput, nil
// }

// Function to do the reverse, take a scheduling input's []byte and unmarshal it back into a SchedulingInput
func PBToSchedulingInput(timestamp time.Time, data []byte) (SchedulingInput, error) {
	podList := &v1.PodList{}
	if err := proto.Unmarshal(data, podList); err != nil {
		return SchedulingInput{}, fmt.Errorf("unmarshaling pod list, %w", err)
	}
	pods := lo.ToSlicePtr(podList.Items)
	return ReconstructedSchedulingInput(timestamp, pods), nil
}

func NewSchedulingInput(pendingPods []*v1.Pod, stateNodes []*state.StateNode, instanceTypes []*cloudprovider.InstanceType) SchedulingInput {
	return SchedulingInput{
		Timestamp:     time.Now(),
		PendingPods:   pendingPods,
		StateNodes:    stateNodes,
		InstanceTypes: instanceTypes,
	}
}

// Reconstruct a scheduling input (presumably from a file)
func ReconstructedSchedulingInput(timestamp time.Time, pendingPods []*v1.Pod) SchedulingInput {
	return SchedulingInput{
		Timestamp:   timestamp,
		PendingPods: pendingPods,
	}
}

/* The following functions are testing toString functions that will mirror what the serialization
   deserialization functions will do in protobuf. These are inefficient, but human-readable */

// TODO: This eventually will be "as simple" as reconstructing the data structures from
// the log data and using K8S and/or Karpenter representation to present as JSON or YAML or something

// This function as a human readable test function for serializing desired pod data
// It takes in a v1.Pod and gets the string representations of all the fields we care about.
func PodToString(pod *v1.Pod) string {
	if pod == nil {
		return "<nil>"
	}
	return fmt.Sprintf("Name: %s, Namespace: %s, Phase: %s, NodeName: %s", pod.Name, pod.Namespace, pod.Status.Phase, pod.Spec.NodeName)
}

func PodsToString(pods []*v1.Pod) string {
	if pods == nil {
		return "<nil>"
	}
	var buf bytes.Buffer
	for _, pod := range pods {
		buf.WriteString(PodToString(pod) + "\n") // TODO: Can replace with pod.String() if I want/need
	}
	return buf.String()
}

// Similar function for stateNode
func StateNodeToString(node *state.StateNode) string {
	if node == nil {
		return "<nil>"
	}
	return fmt.Sprintf("Node: %s, NodeClaim: %s", NodeToString(node.Node), NodeClaimToString(node.NodeClaim))
}

func StateNodesToString(nodes []*state.StateNode) string {
	if nodes == nil {
		return "<nil>"
	}
	var buf bytes.Buffer
	for _, node := range nodes {
		buf.WriteString(StateNodeToString(node) + "\n")
	}
	return buf.String()
}

// Similar function for human-readable string serialization of a v1.Node
func NodeToString(node *v1.Node) string {
	if node == nil {
		return "<nil>"
	}
	return fmt.Sprintf("Name: %s, Status: %s, NodeName: %s", node.Name, node.Status.Phase, node.Status.NodeInfo.SystemUUID)
}

// Similar function for NodeClaim
func NodeClaimToString(nodeClaim *v1beta1.NodeClaim) string {
	if nodeClaim == nil {
		return "<nil>"
	}
	return fmt.Sprintf("NodeClaimName: %s", nodeClaim.Name)
}

// Similar for instanceTypes (name, requirements, offerings, capacity, overhead
func InstanceTypeToString(instanceType *cloudprovider.InstanceType) string {
	if instanceType == nil {
		return "<nil>"
	}
	// TODO: String print the sub-types, like Offerings, too, all of them
	return fmt.Sprintf("Name: %s, Requirements: %s, Offerings: %s", instanceType.Name,
		RequirementsToString(&instanceType.Requirements), OfferingToString(&instanceType.Offerings[0]))
}

func InstanceTypesToString(instanceTypes []*cloudprovider.InstanceType) string {
	if instanceTypes == nil {
		return "<nil>"
	}
	var buf bytes.Buffer
	for _, instanceType := range instanceTypes {
		buf.WriteString(InstanceTypeToString(instanceType) + "\n")
	}
	return buf.String()
}

// Similar for IT Requirements
func RequirementsToString(requirements *scheduling.Requirements) string {
	if requirements == nil {
		return "<nil>"
	}
	return fmt.Sprintf("Requirements: %s", requirements)
}

// Similar for IT Offerings (Requirements, Price, Availability)
func OfferingToString(offering *cloudprovider.Offering) string {
	if offering == nil {
		return "<nil>"
	}
	return fmt.Sprintf("Offering Requirements: %s, Price: %f, Available: %t",
		RequirementsToString(&offering.Requirements), offering.Price, offering.Available)
}
