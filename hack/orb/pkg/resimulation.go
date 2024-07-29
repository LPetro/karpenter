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

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/klog/v2"
	"sigs.k8s.io/yaml"

	"k8s.io/utils/clock"
	_ "knative.dev/pkg/system/testing"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/karpenter/pkg/apis/v1beta1"
	"sigs.k8s.io/karpenter/pkg/cloudprovider"
	"sigs.k8s.io/karpenter/pkg/controllers/orb"
	scheduler "sigs.k8s.io/karpenter/pkg/controllers/provisioning/scheduling"
	"sigs.k8s.io/karpenter/pkg/controllers/state"
	"sigs.k8s.io/karpenter/pkg/scheduling"
)

// Resimulate from this scheduling input
func Resimulate(reconstructedSchedulingInput *orb.SchedulingInput, nodepoolYamlFilepath string) (scheduler.Results, error) {
	ctx := context.Background()
	// clock := op.Clock
	// kubeClient := op.GetClient() // ??
	// cluster := state.NewCluster(clock, kubeClient)

	nodePool, err := unmarshalNodePoolFromUser(nodepoolYamlFilepath)
	if err != nil {
		fmt.Println("Error unmarshalling node pools:", err)
		return scheduler.Results{}, err
	}

	nodePools := []*v1beta1.NodePool{nodePool}

	// Reconstruct the dynamic fields from my logged inputs
	// TODO: Currently this is via the saved .log files, not directly from a provided PV
	stateNodes := getStateNodesFromSchedulingInput(reconstructedSchedulingInput)
	instanceTypes := getInstanceTypesFromSchedulingInput(reconstructedSchedulingInput)
	pods := reconstructedSchedulingInput.PendingPods
	topology := reconstructedSchedulingInput.Topology
	daemonSetPods := reconstructedSchedulingInput.DaemonSetPods
	pvList := reconstructedSchedulingInput.PVList
	pvcList := reconstructedSchedulingInput.PVCList

	dataClient := New(pvList, pvcList)
	cluster := state.NewCluster(clock.RealClock{}, dataClient)

	// How do I put bindings back into cluster? I just made this
	// TODO: Check if this even works the way I want it to.
	cluster.ReconstructCluster(ctx, reconstructedSchedulingInput.Bindings, stateNodes, daemonSetPods)

	// TODO: Figure out why this makes Karpenter call the command-line arguments I've passed in; try to clear them or get around this
	s := scheduler.NewScheduler(dataClient, nodePools, cluster, stateNodes, topology, instanceTypes, daemonSetPods, nil)
	return s.Solve(ctx, pods), nil
	// fmt.Println("Resimulating from this scheduling input:", reconstructedSchedulingInput) // Delete
	// return scheduler.Results{}, nil
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

// Calculate the Topology, copies the structure from provisioner.go
func getTopologyFromSchedulingInput(ctx context.Context, kubeClient client.Client, cluster *state.Cluster, nodePools []*v1beta1.NodePool,
	instanceTypes map[string][]*cloudprovider.InstanceType, pods []*v1.Pod) (*scheduler.Topology, []*v1.Pod, error) {
	domains := map[string]sets.Set[string]{}
	for _, nodePool := range nodePools {
		instanceTypeOptions := instanceTypes[nodePool.Name]
		// Construct Topology Domains
		for _, instanceType := range instanceTypeOptions {
			// We need to intersect the instance type requirements with the current nodePool requirements.  This
			// ensures that something like zones from an instance type don't expand the universe of valid domains.
			requirements := scheduling.NewNodeSelectorRequirementsWithMinValues(nodePool.Spec.Template.Spec.Requirements...)
			requirements.Add(scheduling.NewLabelRequirements(nodePool.Spec.Template.Labels).Values()...)
			requirements.Add(instanceType.Requirements.Values()...)

			for key, requirement := range requirements {
				// This code used to execute a Union between domains[key] and requirement.Values().
				// The downside of this is that Union is immutable and takes a copy of the set it is executed upon.
				// This resulted in a lot of memory pressure on the heap and poor performance
				// https://github.com/aws/karpenter/issues/3565
				if domains[key] == nil {
					domains[key] = sets.New(requirement.Values()...)
				} else {
					domains[key].Insert(requirement.Values()...)
				}
			}
		}

		requirements := scheduling.NewNodeSelectorRequirementsWithMinValues(nodePool.Spec.Template.Spec.Requirements...)
		requirements.Add(scheduling.NewLabelRequirements(nodePool.Spec.Template.Labels).Values()...)
		for key, requirement := range requirements {
			if requirement.Operator() == v1.NodeSelectorOpIn {
				// The following is a performance optimisation, for the explanation see the comment above
				if domains[key] == nil {
					domains[key] = sets.New(requirement.Values()...)
				} else {
					domains[key].Insert(requirement.Values()...)
				}
			}
		}
	}
	// inject topology constraints
	var schedulablePods []*v1.Pod
	volumeTopology := scheduler.NewVolumeTopology(kubeClient)
	for _, pod := range pods {
		if err := volumeTopology.Inject(ctx, pod); err != nil {
			log.FromContext(ctx).WithValues("Pod", klog.KRef(pod.Namespace, pod.Name)).Error(err, "failed getting volume topology requirements")
		} else {
			schedulablePods = append(schedulablePods, pod)
		}
	}
	topology, err := scheduler.NewTopology(ctx, kubeClient, cluster, domains, schedulablePods)
	return topology, schedulablePods, err
}
