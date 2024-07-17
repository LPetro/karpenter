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
	"fmt"
	"os"
	"path/filepath"
	"time"

	proto "github.com/gogo/protobuf/proto"
	pb "sigs.k8s.io/karpenter/pkg/controllers/orb/proto"
	//"google.golang.org/protobuf/proto"
)

type SchedulingMetadata struct {
	Action    string
	Timestamp time.Time
}

type key int

const (
	schedulingMetadataKey key = iota
)

// WithSchedulingMetadata returns a new context with the provided scheduling metadata.
func WithSchedulingMetadata(ctx context.Context, action string, timestamp time.Time) context.Context {
	switch action { // Preserves that these actions being only valid values
	case "normal-provisioning", "single-node-consolidation", "multi-node-consolidation":
		metadata := SchedulingMetadata{
			Action:    action,
			Timestamp: timestamp,
		}
		return context.WithValue(ctx, schedulingMetadataKey, metadata)
	default:
		fmt.Println("Invalid scheduling action metadata:", action) //Testing, remove later
		return ctx
	}
}

// GetProvisioningMetadata retrieves the scheduling metadata from the context.
func GetSchedulingMetadata(ctx context.Context) (SchedulingMetadata, bool) {
	metadata, ok := ctx.Value(schedulingMetadataKey).(SchedulingMetadata)
	return metadata, ok
}

// Precondition, only called when heap isn't nil, and len > 0
func LogSchedulingMetadataHeapToPV(heap *SchedulingMetadataHeap) error {
	if heap == nil || heap.Len() == 0 {
		return fmt.Errorf("called with invalid heap or empty heap")
	}

	oldestStr := (*heap)[0].Timestamp.Format("2006-01-02_15-04-05")
	newestStr := (*heap)[len(*heap)-1].Timestamp.Format("2006-01-02_15-04-05")
	fileName := fmt.Sprintf("SchedulingMetadata_%s_to_%s.log", oldestStr, newestStr)
	path := filepath.Join("/data", fileName)

	file, err := os.Create(path)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return err
	}
	defer file.Close()

	// Pop each scheduling metadata off its heap (oldest first) and batch log to PV.
	mapping := &pb.SchedulingMetadataMapping{}
	for heap.Len() > 0 {
		metadata := heap.Pop().(SchedulingMetadata)

		entry := &pb.SchedulingMetadataMapping_MappingEntry{
			Action:    metadata.Action,
			Timestamp: metadata.Timestamp.Format("2006-01-02_15-04-05"),
		}

		mapping.Entries = append(mapping.Entries, entry)

	}

	mappingdata, err := proto.Marshal(mapping)
	if err != nil {
		fmt.Println("Error marshalling data:", err)
		return err
	}

	_, err = file.Write(mappingdata)
	if err != nil {
		fmt.Println("Error writing data to file:", err)
		return err
	}

	fmt.Println("Metadata written to S3 bucket successfully!")
	return nil
}

// Function to unmarshal the metadata

// // Reads in and parses all the scheduling metadata from the file in the PV.
// // TODO: grab data based on a range of time.Time timestamps from files, via the range in their name.
// func ReadSchedulingMetadataFromPV(timestampStr string) ([]*SchedulingMetadata, error) {
// 	fileName := fmt.Sprintf("SchedulingMetadata_%s.log", timestampStr)
// 	path := filepath.Join("/data", fileName)

// 	file, err := os.Open(path)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer file.Close()

// 	var existingmetadata []*SchedulingMetadata
// 	scanner := bufio.NewScanner(file)
// 	for scanner.Scan() {
// 		line := scanner.Text()
// 		parts := strings.Split(line, "\t")
// 		if len(parts) != 2 {
// 			return nil, fmt.Errorf("invalid line format: %s", line)
// 		}

// 		timestamp, err := time.Parse("2006-01-02_15-04-05", parts[1])
// 		if err != nil {
// 			return nil, fmt.Errorf("failed to parse timestamp: %v", err)
// 		}

// 		existingmetadata = append(existingmetadata, &SchedulingMetadata{
// 			Timestamp: timestamp,
// 			Action:    parts[0],
// 		})
// 	}

// 	if err := scanner.Err(); err != nil {
// 		return nil, err
// 	}

// 	return existingmetadata, nil
// }
