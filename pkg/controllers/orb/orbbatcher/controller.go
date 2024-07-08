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

package orbbatcher

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/awslabs/operatorpkg/singleton"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	controllerruntime "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"sigs.k8s.io/karpenter/pkg/apis/v1beta1"
	"sigs.k8s.io/karpenter/pkg/cloudprovider"
	"sigs.k8s.io/karpenter/pkg/controllers/state"
	"sigs.k8s.io/karpenter/pkg/scheduling"
)

// Set global variable(s) for Mounted PV path
var mountPath = "/data"

// const (
// 	orbQueueBaseDelay = 100 * time.Millisecond
// 	orbQueueMaxDelay  = 10 * time.Second
// )

type PBQueue struct {
	//workqueue.RateLimitingInterface // TODO I saw this in another similar definition; what does it do, do I want/need it?
	mu   sync.Mutex
	data [][]byte
}

type TestQueue struct {
	mu  sync.Mutex
	Set sets.Set[string]
}

func NewPBQueue() *PBQueue {
	return &PBQueue{
		//RateLimitingInterface: workqueue.NewRateLimitingQueue(workqueue.NewItemExponentialFailureRateLimiter(orbQueueBaseDelay, orbQueueMaxDelay)),
		data: make([][]byte, 0),
	}
}

func NewTestQueue() *TestQueue {
	return &TestQueue{
		Set: sets.New[string](),
	}
}
func (q *PBQueue) PBEnqueue(msg []byte) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.data = append(q.data, msg)
}

func (q *TestQueue) TestEnqueue(str string) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.Set.Insert(str)
}

func (q *PBQueue) PBDequeue() ([]byte, bool) {
	q.mu.Lock()
	defer q.mu.Unlock()
	if len(q.data) == 0 {
		return nil, false
	}
	msg := q.data[0]
	q.data = q.data[1:]
	return msg, true
}

func (q *TestQueue) TestDequeue() (string, bool) {
	q.mu.Lock()
	defer q.mu.Unlock()
	if q.Set.Len() == 0 {
		return "", false
	}
	item := q.Set.UnsortedList()[0]
	q.Set.Delete(item)
	return item, true
}

// This will take the info we pass from Provisioner or Disruption to log
// This is the test print function TODO: make a pb one
func (q *TestQueue) LogLine(item string) {
	// Do some data validation?

	// Serialize it into the protobuffed structure binary?

	// Then insert it...?
	q.TestEnqueue(item) // Currently passing in a string, not a binary.
}

// // This is an initial protobuf serializing log. I'll need a deserializer too, to read.
// func LogEvent(timestamp, eventType, message string, tags []string) error {
// 	entry := &ORBLogEntry{
// 		Timestamp: timestamp,
// 		EventType: eventType,
// 		Message:   message,
// 		Tags:      tags,
// 	}
// 	data, err := proto.Marshal(entry)
// 	if err != nil {
// 		return err
// 	}
// 	// Write the serialized data to the log
// 	return WriteToLog(data)
// }

// Deserialize and JSON marshal cluster
// clusterJSON, err := json.Marshal(p.cluster)
// if err != nil {
// 	return nil, fmt.Errorf("marshaling cluster, %w", err)
// }
// fmt.Println(string(clusterJSON))

// Also only do this is pending pods has changed.
// fmt.Println("Pod 0:", pods[0])
// fmt.Print("Marshaled Pod 0: ")
// fmt.Println(pods[0].Marshal()) // I think this saves as protobuf intrinsically, using k8s api generated.pb.go
// k8s.io/api/core/v1/generated.proto

// This function serializes _ resource into protobuf

// This functions deserialized _ resource into protobuf

type Controller struct {
	queue *TestQueue
	// some sort of log store, or way to batch logs together before sending to PV
}

// TODO: add struct elements and their instantiations, when defined
func NewController(queue *TestQueue) *Controller {
	return &Controller{
		queue: queue,
	}
}

// This function batches together loglines into our Queue data structure
// This queue will be periodically dumped to the S3 Bucket
func (c *Controller) Reconcile(ctx context.Context) (reconcile.Result, error) {
	// TODO: what does this do / where does it reference to or need to reference to?
	// ctx = injection.WithControllerName(ctx, "orb.batcher")

	fmt.Println("Starting One Reconcile Print from ORB...")

	c.queue.TestEnqueue("Hello World from the ORB Batcher Reconciler")

	// While a queue is not empty (has items to dequeue), dequeue and print
	// TODO: There must be a prettier / more Go-like way to write this...
	for {
		item, nonempty := c.queue.TestDequeue()
		if !nonempty {
			break
		}
		fmt.Println(item)
	}

	fmt.Println("Ending One Reconcile Print from ORB...")
	fmt.Println()

	// // Save to the Persistent Volume
	// err := c.SaveToPV("helloworld.txt", "sample_logline")
	// if err != nil {
	// 	fmt.Println("Error saving to PV:", err)
	// 	return reconcile.Result{}, err
	// }

	return reconcile.Result{RequeueAfter: time.Second * 5}, nil
}

// TODO: What does this register function do? Is it needed for a controller to work?
func (c *Controller) Register(_ context.Context, m manager.Manager) error {
	return controllerruntime.NewControllerManagedBy(m).
		Named("orb.batcher").
		WatchesRawSource(singleton.Source()).
		Complete(singleton.AsReconciler(c))
}

/* The following functions are testing toString functions that will mirror what the serialization
   deserialization functions will do in protobuf. These are inefficient, but human-readable */

// TODO: Check if the fields exist before calling them.

// This function as a human readable test function for serializing desired pod data
// It takes in a v1.Pod and gets the string representations of all the fields we care about.
func PodToString(pod *v1.Pod) string {
	if pod == nil {
		return "<nil>"
	}
	return fmt.Sprintf("Name: %s, Namespace: %s, Phase: %s, NodeName: %s", pod.Name, pod.Namespace, pod.Status.Phase, pod.Spec.NodeName)
}

// Similar function for stateNode
func StateNodeToString(node *state.StateNode) string {
	if node == nil {
		return "<nil>"
	}
	return fmt.Sprintf("Node: %s, NodeClaim: %s", NodeToString(node.Node), NodeClaimToString(node.NodeClaim))
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

// Function for logging everything in the Provisioner Scheduler (i.e. pending pods, statenodes...)
func (q *TestQueue) TestLogProvisioningScheduler(pods []*v1.Pod, stateNodes []*state.StateNode, instanceTypes map[string][]*cloudprovider.InstanceType) {
	q.TestEnqueue("Testing from the Provisioner")

	//log.FromContext(ctx).Info("Context input to scheduler.NewScheduler", "ctx", ctx)

	//log.FromContext(ctx).Info("nodePools input to scheduler.NewScheduler", "nodePools", lo.ToSlicePtr(nodePoolList.Items))

	// log.Info("cluster input to scheduler.NewScheduler", "cluster", p.cluster)

	// Logs the pending pods
	// log.FromContext(ctx).Info("Pending pods", "pods", lo.ToSlicePtr(pods))
	q.TestLogPendingPods(pods)

	// Log the state nodes
	// log.FromContext(ctx).Info("State nodes", "stateNodes", lo.ToSlicePtr(stateNodes))
	q.TestLogStateNodes(stateNodes)

	// Log the instance types
	// log.FromContext(ctx).Info("Instance types", "instanceTypes", instanceTypes)
	q.TestLogInstanceTypes(instanceTypes)

	// Log the topology
	// log.FromContext(ctx).Info("Topology", "topology", topology)

	//log.Info("daemonSetPods input to scheduler.NewScheduler", "daemonSetPods", daemonSetPods)

	q.TestEnqueue("End Provisioner Test")
}

// Function for logging pending pods
func (q *TestQueue) TestLogPendingPods(pods []*v1.Pod) {
	// Test Log Pending Pods
	for _, pod := range pods {
		q.TestEnqueue(PodToString(pod))
	}
}

// Function for logging stateNodes
func (q *TestQueue) TestLogStateNodes(stateNodes []*state.StateNode) {
	// Test Log StateNodes
	for _, statenode := range stateNodes {
		q.TestEnqueue(StateNodeToString(statenode))
	}
}

// Function for logging instanceTypes
func (q *TestQueue) TestLogInstanceTypes(instanceTypes map[string][]*cloudprovider.InstanceType) {
	// Test Log InstanceTypes
	for _, it := range instanceTypes["default"] {
		q.TestEnqueue(InstanceTypeToString(it))
	}
}

// Similar for IT Capacity

// Security Issue Common Weakness Enumeration (CWE)-22,23 Path Traversal
// They highly recommend sanitizing inputs before accessing that path.
func (c *Controller) sanitizePath(path string) string {
	// Remove any leading or trailing slashes, "../" or "./"...
	path = strings.TrimPrefix(path, "/")
	path = strings.TrimSuffix(path, "/")
	path = regexp.MustCompile(`\.\.\/`).ReplaceAllString(path, "")
	path = regexp.MustCompile(`\.\/`).ReplaceAllString(path, "")
	path = strings.ReplaceAll(path, "../", "")

	return path
}

/* This function saves things to our Persistent Volume */
// Saves data to PV (S3 Bucket for AWS) via the mounted log path
// It takes a name of the log file as well as the logline to be logged.
// The function opens a file for writing, writes some data to the file, and then closes the file
func (c *Controller) SaveToPV(logname string, logline string) error {

	// Create the log file path and desired logline (example for now)
	sanitizedname := c.sanitizePath(logname)
	path := filepath.Join(mountPath, sanitizedname)
	//logline := fmt.Sprintf("Printing data (from %s) to the S3 bucket", logname)

	// Opens the mounted volume (S3 Bucket) file at that path
	file, err := os.Create(path)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return err
	}
	defer file.Close()

	// Writes data to the file
	_, err = fmt.Fprintln(file, logline)
	if err != nil {
		fmt.Println("Error writing to file:", err)
		return err
	}

	fmt.Println("Data written to S3 bucket successfully!")
	return nil
}
