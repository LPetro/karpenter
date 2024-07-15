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
	"time"
	//"google.golang.org/protobuf/proto"
)

type SchedulingMetadata struct {
	Action    string
	Timestamp time.Time
}

type key int

const (
	provisioningMetadataKey key = iota
)

// case "normal-provisioning":
// case "consolidation-simulation":
// WithProvisioningMetadata returns a new context with the provided provisioning metadata.
func WithProvisioningMetadata(ctx context.Context, action string, timestamp time.Time) context.Context {
	metadata := SchedulingMetadata{
		Action:    action,
		Timestamp: timestamp,
	}
	return context.WithValue(ctx, provisioningMetadataKey, metadata)
}

// GetProvisioningMetadata retrieves the provisioning metadata from the context.
func GetProvisioningMetadata(ctx context.Context) (SchedulingMetadata, bool) {
	metadata, ok := ctx.Value(provisioningMetadataKey).(SchedulingMetadata)
	return metadata, ok
}
