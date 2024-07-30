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

package pkg

import (
	"context"
	"fmt"
	"io"
	"os"

	"sigs.k8s.io/yaml"

	"k8s.io/utils/clock"
	_ "knative.dev/pkg/system/testing"
	"sigs.k8s.io/karpenter/pkg/apis/v1beta1"
	"sigs.k8s.io/karpenter/pkg/cloudprovider"
	"sigs.k8s.io/karpenter/pkg/controllers/orb"
	scheduler "sigs.k8s.io/karpenter/pkg/controllers/provisioning/scheduling"
	"sigs.k8s.io/karpenter/pkg/controllers/state"
)

// Resimulate a scheduling run using the scheduling input reconstructed from the logs.
func Resimulate(reconstructedSchedulingInput *orb.SchedulingInput, nodepoolYamlFilepath string) (scheduler.Results, error) {
	ctx := context.Background()

	nodePool, err := unmarshalNodePoolFromUser(nodepoolYamlFilepath)
	if err != nil {
		fmt.Println("Error unmarshalling node pools:", err)
		return scheduler.Results{}, err
	}

	// TODO: unmarshal directly into a NodePoolList
	nodePools := []*v1beta1.NodePool{nodePool}

	// Reconstruct the dynamic fields from my logged inputs
	// TODO?: Currently this is via the saved .log files, not directly from a provided PV, maybe that's ok
	stateNodes := getStateNodesFromSchedulingInput(reconstructedSchedulingInput)
	instanceTypes := getInstanceTypesFromSchedulingInput(reconstructedSchedulingInput)
	pendingPods := reconstructedSchedulingInput.PendingPods
	topology := reconstructedSchedulingInput.Topology
	daemonSetPods := reconstructedSchedulingInput.DaemonSetPods
	pvList := reconstructedSchedulingInput.PVList
	pvcList := reconstructedSchedulingInput.PVCList
	allPods := reconstructedSchedulingInput.ScheduledPodList

	dataClient := New(pvList, pvcList)
	cluster := state.NewCluster(clock.RealClock{}, dataClient)
	cluster.ReconstructCluster(ctx, reconstructedSchedulingInput.Bindings, stateNodes, daemonSetPods, allPods)

	s := scheduler.NewScheduler(dataClient, nodePools, cluster, cluster.Nodes(), topology, instanceTypes, daemonSetPods, nil)
	return s.Solve(ctx, pendingPods), nil
}

// TODO: Functionality not yet tested, needs to return []*NodePool
func unmarshalNodePoolFromUser(nodepoolYamlFilepath string) (*v1beta1.NodePool, error) {
	yamlFile, err := os.Open(nodepoolYamlFilepath)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return nil, err
	}
	defer yamlFile.Close()

	yamlData, err := io.ReadAll(yamlFile)
	if err != nil {
		fmt.Println("Error reading yaml file:", err)
		return nil, err
	}

	nodePool := &v1beta1.NodePool{}
	err = yaml.Unmarshal(yamlData, nodePool)
	if err != nil {
		return nil, err
	}

	// "k8s.io/apimachinery/pkg/runtime"
	// "k8s.io/apimachinery/pkg/runtime/serializer"

	// codecs := serializer.NewCodecFactory(runtime.NewScheme())
	// deserializer := codecs.UniversalDeserializer()

	// obj, _, err := deserializer.Decode(yamlData, nil, nil)
	// if err != nil {
	// 	return nil, err
	// }

	// nodePools, ok := obj.(*karpenterv1beta1.NodePoolList)
	// if !ok {
	// 	return nil, fmt.Errorf("unexpected object type: %T", obj)
	// }

	// return nodePools.Items, nil

	return nodePool, nil
}

// Extracts the statenodes from the scheduling input.
func getStateNodesFromSchedulingInput(schedulingInput *orb.SchedulingInput) []*state.StateNode {
	stateNodes := []*state.StateNode{}
	for _, stateNodeWithPods := range schedulingInput.StateNodesWithPods {
		stateNodes = append(stateNodes, &state.StateNode{
			Node:      stateNodeWithPods.Node,
			NodeClaim: stateNodeWithPods.NodeClaim,
		})
	}
	return stateNodes
}

// Extract and reassociate nodepools to the instance types from the scheduling input.
func getInstanceTypesFromSchedulingInput(schedulingInput *orb.SchedulingInput) map[string][]*cloudprovider.InstanceType {
	instanceTypes := map[string][]*cloudprovider.InstanceType{}
	allInstanceTypesMap := orb.MapInstanceTypesByName(schedulingInput.AllInstanceTypes)
	for nodepoolName, instanceTypeNames := range schedulingInput.NodePoolInstanceTypes {
		instanceTypes[nodepoolName] = []*cloudprovider.InstanceType{}
		for _, instanceTypeName := range instanceTypeNames {
			instanceTypes[nodepoolName] = append(instanceTypes[nodepoolName], allInstanceTypesMap[instanceTypeName])
		}
	}
	return instanceTypes
}
