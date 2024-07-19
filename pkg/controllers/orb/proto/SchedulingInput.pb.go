// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.34.2
// 	protoc        v5.27.1
// source: SchedulingInput.proto

package proto

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type Differences struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Added   *SchedulingInput `protobuf:"bytes,1,opt,name=added,proto3,oneof" json:"added,omitempty"`
	Removed *SchedulingInput `protobuf:"bytes,2,opt,name=removed,proto3,oneof" json:"removed,omitempty"`
	Changed *SchedulingInput `protobuf:"bytes,3,opt,name=changed,proto3,oneof" json:"changed,omitempty"`
}

func (x *Differences) Reset() {
	*x = Differences{}
	if protoimpl.UnsafeEnabled {
		mi := &file_SchedulingInput_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Differences) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Differences) ProtoMessage() {}

func (x *Differences) ProtoReflect() protoreflect.Message {
	mi := &file_SchedulingInput_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Differences.ProtoReflect.Descriptor instead.
func (*Differences) Descriptor() ([]byte, []int) {
	return file_SchedulingInput_proto_rawDescGZIP(), []int{0}
}

func (x *Differences) GetAdded() *SchedulingInput {
	if x != nil {
		return x.Added
	}
	return nil
}

func (x *Differences) GetRemoved() *SchedulingInput {
	if x != nil {
		return x.Removed
	}
	return nil
}

func (x *Differences) GetChanged() *SchedulingInput {
	if x != nil {
		return x.Changed
	}
	return nil
}

// TODO: will need to be updated as I continue to update SchedulingInput
// Anything as bytes are the self-describing wire format provided by K8s
type SchedulingInput struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Timestamp         string                 `protobuf:"bytes,1,opt,name=timestamp,proto3" json:"timestamp,omitempty"`
	PendingpodData    []*ReducedPod          `protobuf:"bytes,2,rep,name=pendingpod_data,json=pendingpodData,proto3" json:"pendingpod_data,omitempty"`
	StatenodesData    []*StateNodeWithPods   `protobuf:"bytes,3,rep,name=statenodes_data,json=statenodesData,proto3" json:"statenodes_data,omitempty"`
	InstancetypesData []*ReducedInstanceType `protobuf:"bytes,4,rep,name=instancetypes_data,json=instancetypesData,proto3" json:"instancetypes_data,omitempty"`
}

func (x *SchedulingInput) Reset() {
	*x = SchedulingInput{}
	if protoimpl.UnsafeEnabled {
		mi := &file_SchedulingInput_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SchedulingInput) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SchedulingInput) ProtoMessage() {}

func (x *SchedulingInput) ProtoReflect() protoreflect.Message {
	mi := &file_SchedulingInput_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SchedulingInput.ProtoReflect.Descriptor instead.
func (*SchedulingInput) Descriptor() ([]byte, []int) {
	return file_SchedulingInput_proto_rawDescGZIP(), []int{1}
}

func (x *SchedulingInput) GetTimestamp() string {
	if x != nil {
		return x.Timestamp
	}
	return ""
}

func (x *SchedulingInput) GetPendingpodData() []*ReducedPod {
	if x != nil {
		return x.PendingpodData
	}
	return nil
}

func (x *SchedulingInput) GetStatenodesData() []*StateNodeWithPods {
	if x != nil {
		return x.StatenodesData
	}
	return nil
}

func (x *SchedulingInput) GetInstancetypesData() []*ReducedInstanceType {
	if x != nil {
		return x.InstancetypesData
	}
	return nil
}

type ReducedPod struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Name      string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	Namespace string `protobuf:"bytes,2,opt,name=namespace,proto3" json:"namespace,omitempty"`
	Phase     string `protobuf:"bytes,3,opt,name=phase,proto3" json:"phase,omitempty"`
}

func (x *ReducedPod) Reset() {
	*x = ReducedPod{}
	if protoimpl.UnsafeEnabled {
		mi := &file_SchedulingInput_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ReducedPod) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ReducedPod) ProtoMessage() {}

func (x *ReducedPod) ProtoReflect() protoreflect.Message {
	mi := &file_SchedulingInput_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ReducedPod.ProtoReflect.Descriptor instead.
func (*ReducedPod) Descriptor() ([]byte, []int) {
	return file_SchedulingInput_proto_rawDescGZIP(), []int{2}
}

func (x *ReducedPod) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *ReducedPod) GetNamespace() string {
	if x != nil {
		return x.Namespace
	}
	return ""
}

func (x *ReducedPod) GetPhase() string {
	if x != nil {
		return x.Phase
	}
	return ""
}

// A stateNode with the Pods it has on it.
type StateNodeWithPods struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Node      *StateNodeWithPods_ReducedNode      `protobuf:"bytes,1,opt,name=node,proto3,oneof" json:"node,omitempty"`
	NodeClaim *StateNodeWithPods_ReducedNodeClaim `protobuf:"bytes,2,opt,name=nodeClaim,proto3,oneof" json:"nodeClaim,omitempty"`
	Pods      []*ReducedPod                       `protobuf:"bytes,3,rep,name=pods,proto3" json:"pods,omitempty"`
}

func (x *StateNodeWithPods) Reset() {
	*x = StateNodeWithPods{}
	if protoimpl.UnsafeEnabled {
		mi := &file_SchedulingInput_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *StateNodeWithPods) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*StateNodeWithPods) ProtoMessage() {}

func (x *StateNodeWithPods) ProtoReflect() protoreflect.Message {
	mi := &file_SchedulingInput_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use StateNodeWithPods.ProtoReflect.Descriptor instead.
func (*StateNodeWithPods) Descriptor() ([]byte, []int) {
	return file_SchedulingInput_proto_rawDescGZIP(), []int{3}
}

func (x *StateNodeWithPods) GetNode() *StateNodeWithPods_ReducedNode {
	if x != nil {
		return x.Node
	}
	return nil
}

func (x *StateNodeWithPods) GetNodeClaim() *StateNodeWithPods_ReducedNodeClaim {
	if x != nil {
		return x.NodeClaim
	}
	return nil
}

func (x *StateNodeWithPods) GetPods() []*ReducedPod {
	if x != nil {
		return x.Pods
	}
	return nil
}

type ReducedInstanceType struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Name         string                                    `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	Requirements []*ReducedInstanceType_ReducedRequirement `protobuf:"bytes,2,rep,name=requirements,proto3" json:"requirements,omitempty"`
	Offerings    []*ReducedInstanceType_ReducedOffering    `protobuf:"bytes,3,rep,name=offerings,proto3" json:"offerings,omitempty"`
}

func (x *ReducedInstanceType) Reset() {
	*x = ReducedInstanceType{}
	if protoimpl.UnsafeEnabled {
		mi := &file_SchedulingInput_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ReducedInstanceType) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ReducedInstanceType) ProtoMessage() {}

func (x *ReducedInstanceType) ProtoReflect() protoreflect.Message {
	mi := &file_SchedulingInput_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ReducedInstanceType.ProtoReflect.Descriptor instead.
func (*ReducedInstanceType) Descriptor() ([]byte, []int) {
	return file_SchedulingInput_proto_rawDescGZIP(), []int{4}
}

func (x *ReducedInstanceType) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *ReducedInstanceType) GetRequirements() []*ReducedInstanceType_ReducedRequirement {
	if x != nil {
		return x.Requirements
	}
	return nil
}

func (x *ReducedInstanceType) GetOfferings() []*ReducedInstanceType_ReducedOffering {
	if x != nil {
		return x.Offerings
	}
	return nil
}

type SchedulingMetadataMap struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Entries []*SchedulingMetadataMap_MappingEntry `protobuf:"bytes,1,rep,name=entries,proto3" json:"entries,omitempty"`
}

func (x *SchedulingMetadataMap) Reset() {
	*x = SchedulingMetadataMap{}
	if protoimpl.UnsafeEnabled {
		mi := &file_SchedulingInput_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SchedulingMetadataMap) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SchedulingMetadataMap) ProtoMessage() {}

func (x *SchedulingMetadataMap) ProtoReflect() protoreflect.Message {
	mi := &file_SchedulingInput_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SchedulingMetadataMap.ProtoReflect.Descriptor instead.
func (*SchedulingMetadataMap) Descriptor() ([]byte, []int) {
	return file_SchedulingInput_proto_rawDescGZIP(), []int{5}
}

func (x *SchedulingMetadataMap) GetEntries() []*SchedulingMetadataMap_MappingEntry {
	if x != nil {
		return x.Entries
	}
	return nil
}

type StateNodeWithPods_ReducedNode struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Name       string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	Nodestatus []byte `protobuf:"bytes,2,opt,name=nodestatus,proto3" json:"nodestatus,omitempty"`
}

func (x *StateNodeWithPods_ReducedNode) Reset() {
	*x = StateNodeWithPods_ReducedNode{}
	if protoimpl.UnsafeEnabled {
		mi := &file_SchedulingInput_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *StateNodeWithPods_ReducedNode) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*StateNodeWithPods_ReducedNode) ProtoMessage() {}

func (x *StateNodeWithPods_ReducedNode) ProtoReflect() protoreflect.Message {
	mi := &file_SchedulingInput_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use StateNodeWithPods_ReducedNode.ProtoReflect.Descriptor instead.
func (*StateNodeWithPods_ReducedNode) Descriptor() ([]byte, []int) {
	return file_SchedulingInput_proto_rawDescGZIP(), []int{3, 0}
}

func (x *StateNodeWithPods_ReducedNode) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *StateNodeWithPods_ReducedNode) GetNodestatus() []byte {
	if x != nil {
		return x.Nodestatus
	}
	return nil
}

type StateNodeWithPods_ReducedNodeClaim struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Name string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
}

func (x *StateNodeWithPods_ReducedNodeClaim) Reset() {
	*x = StateNodeWithPods_ReducedNodeClaim{}
	if protoimpl.UnsafeEnabled {
		mi := &file_SchedulingInput_proto_msgTypes[7]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *StateNodeWithPods_ReducedNodeClaim) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*StateNodeWithPods_ReducedNodeClaim) ProtoMessage() {}

func (x *StateNodeWithPods_ReducedNodeClaim) ProtoReflect() protoreflect.Message {
	mi := &file_SchedulingInput_proto_msgTypes[7]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use StateNodeWithPods_ReducedNodeClaim.ProtoReflect.Descriptor instead.
func (*StateNodeWithPods_ReducedNodeClaim) Descriptor() ([]byte, []int) {
	return file_SchedulingInput_proto_rawDescGZIP(), []int{3, 1}
}

func (x *StateNodeWithPods_ReducedNodeClaim) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

type ReducedInstanceType_ReducedRequirement struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Key                  string   `protobuf:"bytes,1,opt,name=key,proto3" json:"key,omitempty"`
	Nodeselectoroperator string   `protobuf:"bytes,2,opt,name=nodeselectoroperator,proto3" json:"nodeselectoroperator,omitempty"`
	Values               []string `protobuf:"bytes,3,rep,name=values,proto3" json:"values,omitempty"`
}

func (x *ReducedInstanceType_ReducedRequirement) Reset() {
	*x = ReducedInstanceType_ReducedRequirement{}
	if protoimpl.UnsafeEnabled {
		mi := &file_SchedulingInput_proto_msgTypes[8]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ReducedInstanceType_ReducedRequirement) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ReducedInstanceType_ReducedRequirement) ProtoMessage() {}

func (x *ReducedInstanceType_ReducedRequirement) ProtoReflect() protoreflect.Message {
	mi := &file_SchedulingInput_proto_msgTypes[8]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ReducedInstanceType_ReducedRequirement.ProtoReflect.Descriptor instead.
func (*ReducedInstanceType_ReducedRequirement) Descriptor() ([]byte, []int) {
	return file_SchedulingInput_proto_rawDescGZIP(), []int{4, 0}
}

func (x *ReducedInstanceType_ReducedRequirement) GetKey() string {
	if x != nil {
		return x.Key
	}
	return ""
}

func (x *ReducedInstanceType_ReducedRequirement) GetNodeselectoroperator() string {
	if x != nil {
		return x.Nodeselectoroperator
	}
	return ""
}

func (x *ReducedInstanceType_ReducedRequirement) GetValues() []string {
	if x != nil {
		return x.Values
	}
	return nil
}

type ReducedInstanceType_ReducedOffering struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Requirements []*ReducedInstanceType_ReducedRequirement `protobuf:"bytes,1,rep,name=requirements,proto3" json:"requirements,omitempty"`
	Price        float64                                   `protobuf:"fixed64,2,opt,name=price,proto3" json:"price,omitempty"`
	Available    bool                                      `protobuf:"varint,3,opt,name=available,proto3" json:"available,omitempty"`
}

func (x *ReducedInstanceType_ReducedOffering) Reset() {
	*x = ReducedInstanceType_ReducedOffering{}
	if protoimpl.UnsafeEnabled {
		mi := &file_SchedulingInput_proto_msgTypes[9]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ReducedInstanceType_ReducedOffering) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ReducedInstanceType_ReducedOffering) ProtoMessage() {}

func (x *ReducedInstanceType_ReducedOffering) ProtoReflect() protoreflect.Message {
	mi := &file_SchedulingInput_proto_msgTypes[9]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ReducedInstanceType_ReducedOffering.ProtoReflect.Descriptor instead.
func (*ReducedInstanceType_ReducedOffering) Descriptor() ([]byte, []int) {
	return file_SchedulingInput_proto_rawDescGZIP(), []int{4, 1}
}

func (x *ReducedInstanceType_ReducedOffering) GetRequirements() []*ReducedInstanceType_ReducedRequirement {
	if x != nil {
		return x.Requirements
	}
	return nil
}

func (x *ReducedInstanceType_ReducedOffering) GetPrice() float64 {
	if x != nil {
		return x.Price
	}
	return 0
}

func (x *ReducedInstanceType_ReducedOffering) GetAvailable() bool {
	if x != nil {
		return x.Available
	}
	return false
}

type SchedulingMetadataMap_MappingEntry struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Action    string `protobuf:"bytes,1,opt,name=action,proto3" json:"action,omitempty"`
	Timestamp string `protobuf:"bytes,2,opt,name=timestamp,proto3" json:"timestamp,omitempty"`
}

func (x *SchedulingMetadataMap_MappingEntry) Reset() {
	*x = SchedulingMetadataMap_MappingEntry{}
	if protoimpl.UnsafeEnabled {
		mi := &file_SchedulingInput_proto_msgTypes[10]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SchedulingMetadataMap_MappingEntry) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SchedulingMetadataMap_MappingEntry) ProtoMessage() {}

func (x *SchedulingMetadataMap_MappingEntry) ProtoReflect() protoreflect.Message {
	mi := &file_SchedulingInput_proto_msgTypes[10]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SchedulingMetadataMap_MappingEntry.ProtoReflect.Descriptor instead.
func (*SchedulingMetadataMap_MappingEntry) Descriptor() ([]byte, []int) {
	return file_SchedulingInput_proto_rawDescGZIP(), []int{5, 0}
}

func (x *SchedulingMetadataMap_MappingEntry) GetAction() string {
	if x != nil {
		return x.Action
	}
	return ""
}

func (x *SchedulingMetadataMap_MappingEntry) GetTimestamp() string {
	if x != nil {
		return x.Timestamp
	}
	return ""
}

var File_SchedulingInput_proto protoreflect.FileDescriptor

var file_SchedulingInput_proto_rawDesc = []byte{
	0x0a, 0x15, 0x53, 0x63, 0x68, 0x65, 0x64, 0x75, 0x6c, 0x69, 0x6e, 0x67, 0x49, 0x6e, 0x70, 0x75,
	0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0xbe, 0x01, 0x0a, 0x0b, 0x44, 0x69, 0x66, 0x66,
	0x65, 0x72, 0x65, 0x6e, 0x63, 0x65, 0x73, 0x12, 0x2b, 0x0a, 0x05, 0x61, 0x64, 0x64, 0x65, 0x64,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x10, 0x2e, 0x53, 0x63, 0x68, 0x65, 0x64, 0x75, 0x6c,
	0x69, 0x6e, 0x67, 0x49, 0x6e, 0x70, 0x75, 0x74, 0x48, 0x00, 0x52, 0x05, 0x61, 0x64, 0x64, 0x65,
	0x64, 0x88, 0x01, 0x01, 0x12, 0x2f, 0x0a, 0x07, 0x72, 0x65, 0x6d, 0x6f, 0x76, 0x65, 0x64, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x10, 0x2e, 0x53, 0x63, 0x68, 0x65, 0x64, 0x75, 0x6c, 0x69,
	0x6e, 0x67, 0x49, 0x6e, 0x70, 0x75, 0x74, 0x48, 0x01, 0x52, 0x07, 0x72, 0x65, 0x6d, 0x6f, 0x76,
	0x65, 0x64, 0x88, 0x01, 0x01, 0x12, 0x2f, 0x0a, 0x07, 0x63, 0x68, 0x61, 0x6e, 0x67, 0x65, 0x64,
	0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x10, 0x2e, 0x53, 0x63, 0x68, 0x65, 0x64, 0x75, 0x6c,
	0x69, 0x6e, 0x67, 0x49, 0x6e, 0x70, 0x75, 0x74, 0x48, 0x02, 0x52, 0x07, 0x63, 0x68, 0x61, 0x6e,
	0x67, 0x65, 0x64, 0x88, 0x01, 0x01, 0x42, 0x08, 0x0a, 0x06, 0x5f, 0x61, 0x64, 0x64, 0x65, 0x64,
	0x42, 0x0a, 0x0a, 0x08, 0x5f, 0x72, 0x65, 0x6d, 0x6f, 0x76, 0x65, 0x64, 0x42, 0x0a, 0x0a, 0x08,
	0x5f, 0x63, 0x68, 0x61, 0x6e, 0x67, 0x65, 0x64, 0x22, 0xe7, 0x01, 0x0a, 0x0f, 0x53, 0x63, 0x68,
	0x65, 0x64, 0x75, 0x6c, 0x69, 0x6e, 0x67, 0x49, 0x6e, 0x70, 0x75, 0x74, 0x12, 0x1c, 0x0a, 0x09,
	0x74, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x09, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x12, 0x34, 0x0a, 0x0f, 0x70, 0x65,
	0x6e, 0x64, 0x69, 0x6e, 0x67, 0x70, 0x6f, 0x64, 0x5f, 0x64, 0x61, 0x74, 0x61, 0x18, 0x02, 0x20,
	0x03, 0x28, 0x0b, 0x32, 0x0b, 0x2e, 0x52, 0x65, 0x64, 0x75, 0x63, 0x65, 0x64, 0x50, 0x6f, 0x64,
	0x52, 0x0e, 0x70, 0x65, 0x6e, 0x64, 0x69, 0x6e, 0x67, 0x70, 0x6f, 0x64, 0x44, 0x61, 0x74, 0x61,
	0x12, 0x3b, 0x0a, 0x0f, 0x73, 0x74, 0x61, 0x74, 0x65, 0x6e, 0x6f, 0x64, 0x65, 0x73, 0x5f, 0x64,
	0x61, 0x74, 0x61, 0x18, 0x03, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x12, 0x2e, 0x53, 0x74, 0x61, 0x74,
	0x65, 0x4e, 0x6f, 0x64, 0x65, 0x57, 0x69, 0x74, 0x68, 0x50, 0x6f, 0x64, 0x73, 0x52, 0x0e, 0x73,
	0x74, 0x61, 0x74, 0x65, 0x6e, 0x6f, 0x64, 0x65, 0x73, 0x44, 0x61, 0x74, 0x61, 0x12, 0x43, 0x0a,
	0x12, 0x69, 0x6e, 0x73, 0x74, 0x61, 0x6e, 0x63, 0x65, 0x74, 0x79, 0x70, 0x65, 0x73, 0x5f, 0x64,
	0x61, 0x74, 0x61, 0x18, 0x04, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x14, 0x2e, 0x52, 0x65, 0x64, 0x75,
	0x63, 0x65, 0x64, 0x49, 0x6e, 0x73, 0x74, 0x61, 0x6e, 0x63, 0x65, 0x54, 0x79, 0x70, 0x65, 0x52,
	0x11, 0x69, 0x6e, 0x73, 0x74, 0x61, 0x6e, 0x63, 0x65, 0x74, 0x79, 0x70, 0x65, 0x73, 0x44, 0x61,
	0x74, 0x61, 0x22, 0x54, 0x0a, 0x0a, 0x52, 0x65, 0x64, 0x75, 0x63, 0x65, 0x64, 0x50, 0x6f, 0x64,
	0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04,
	0x6e, 0x61, 0x6d, 0x65, 0x12, 0x1c, 0x0a, 0x09, 0x6e, 0x61, 0x6d, 0x65, 0x73, 0x70, 0x61, 0x63,
	0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x6e, 0x61, 0x6d, 0x65, 0x73, 0x70, 0x61,
	0x63, 0x65, 0x12, 0x14, 0x0a, 0x05, 0x70, 0x68, 0x61, 0x73, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x05, 0x70, 0x68, 0x61, 0x73, 0x65, 0x22, 0xb7, 0x02, 0x0a, 0x11, 0x53, 0x74, 0x61,
	0x74, 0x65, 0x4e, 0x6f, 0x64, 0x65, 0x57, 0x69, 0x74, 0x68, 0x50, 0x6f, 0x64, 0x73, 0x12, 0x37,
	0x0a, 0x04, 0x6e, 0x6f, 0x64, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1e, 0x2e, 0x53,
	0x74, 0x61, 0x74, 0x65, 0x4e, 0x6f, 0x64, 0x65, 0x57, 0x69, 0x74, 0x68, 0x50, 0x6f, 0x64, 0x73,
	0x2e, 0x52, 0x65, 0x64, 0x75, 0x63, 0x65, 0x64, 0x4e, 0x6f, 0x64, 0x65, 0x48, 0x00, 0x52, 0x04,
	0x6e, 0x6f, 0x64, 0x65, 0x88, 0x01, 0x01, 0x12, 0x46, 0x0a, 0x09, 0x6e, 0x6f, 0x64, 0x65, 0x43,
	0x6c, 0x61, 0x69, 0x6d, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x23, 0x2e, 0x53, 0x74, 0x61,
	0x74, 0x65, 0x4e, 0x6f, 0x64, 0x65, 0x57, 0x69, 0x74, 0x68, 0x50, 0x6f, 0x64, 0x73, 0x2e, 0x52,
	0x65, 0x64, 0x75, 0x63, 0x65, 0x64, 0x4e, 0x6f, 0x64, 0x65, 0x43, 0x6c, 0x61, 0x69, 0x6d, 0x48,
	0x01, 0x52, 0x09, 0x6e, 0x6f, 0x64, 0x65, 0x43, 0x6c, 0x61, 0x69, 0x6d, 0x88, 0x01, 0x01, 0x12,
	0x1f, 0x0a, 0x04, 0x70, 0x6f, 0x64, 0x73, 0x18, 0x03, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x0b, 0x2e,
	0x52, 0x65, 0x64, 0x75, 0x63, 0x65, 0x64, 0x50, 0x6f, 0x64, 0x52, 0x04, 0x70, 0x6f, 0x64, 0x73,
	0x1a, 0x41, 0x0a, 0x0b, 0x52, 0x65, 0x64, 0x75, 0x63, 0x65, 0x64, 0x4e, 0x6f, 0x64, 0x65, 0x12,
	0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e,
	0x61, 0x6d, 0x65, 0x12, 0x1e, 0x0a, 0x0a, 0x6e, 0x6f, 0x64, 0x65, 0x73, 0x74, 0x61, 0x74, 0x75,
	0x73, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x0a, 0x6e, 0x6f, 0x64, 0x65, 0x73, 0x74, 0x61,
	0x74, 0x75, 0x73, 0x1a, 0x26, 0x0a, 0x10, 0x52, 0x65, 0x64, 0x75, 0x63, 0x65, 0x64, 0x4e, 0x6f,
	0x64, 0x65, 0x43, 0x6c, 0x61, 0x69, 0x6d, 0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x42, 0x07, 0x0a, 0x05, 0x5f,
	0x6e, 0x6f, 0x64, 0x65, 0x42, 0x0c, 0x0a, 0x0a, 0x5f, 0x6e, 0x6f, 0x64, 0x65, 0x43, 0x6c, 0x61,
	0x69, 0x6d, 0x22, 0xc3, 0x03, 0x0a, 0x13, 0x52, 0x65, 0x64, 0x75, 0x63, 0x65, 0x64, 0x49, 0x6e,
	0x73, 0x74, 0x61, 0x6e, 0x63, 0x65, 0x54, 0x79, 0x70, 0x65, 0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61,
	0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x4b,
	0x0a, 0x0c, 0x72, 0x65, 0x71, 0x75, 0x69, 0x72, 0x65, 0x6d, 0x65, 0x6e, 0x74, 0x73, 0x18, 0x02,
	0x20, 0x03, 0x28, 0x0b, 0x32, 0x27, 0x2e, 0x52, 0x65, 0x64, 0x75, 0x63, 0x65, 0x64, 0x49, 0x6e,
	0x73, 0x74, 0x61, 0x6e, 0x63, 0x65, 0x54, 0x79, 0x70, 0x65, 0x2e, 0x52, 0x65, 0x64, 0x75, 0x63,
	0x65, 0x64, 0x52, 0x65, 0x71, 0x75, 0x69, 0x72, 0x65, 0x6d, 0x65, 0x6e, 0x74, 0x52, 0x0c, 0x72,
	0x65, 0x71, 0x75, 0x69, 0x72, 0x65, 0x6d, 0x65, 0x6e, 0x74, 0x73, 0x12, 0x42, 0x0a, 0x09, 0x6f,
	0x66, 0x66, 0x65, 0x72, 0x69, 0x6e, 0x67, 0x73, 0x18, 0x03, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x24,
	0x2e, 0x52, 0x65, 0x64, 0x75, 0x63, 0x65, 0x64, 0x49, 0x6e, 0x73, 0x74, 0x61, 0x6e, 0x63, 0x65,
	0x54, 0x79, 0x70, 0x65, 0x2e, 0x52, 0x65, 0x64, 0x75, 0x63, 0x65, 0x64, 0x4f, 0x66, 0x66, 0x65,
	0x72, 0x69, 0x6e, 0x67, 0x52, 0x09, 0x6f, 0x66, 0x66, 0x65, 0x72, 0x69, 0x6e, 0x67, 0x73, 0x1a,
	0x72, 0x0a, 0x12, 0x52, 0x65, 0x64, 0x75, 0x63, 0x65, 0x64, 0x52, 0x65, 0x71, 0x75, 0x69, 0x72,
	0x65, 0x6d, 0x65, 0x6e, 0x74, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x32, 0x0a, 0x14, 0x6e, 0x6f, 0x64, 0x65, 0x73,
	0x65, 0x6c, 0x65, 0x63, 0x74, 0x6f, 0x72, 0x6f, 0x70, 0x65, 0x72, 0x61, 0x74, 0x6f, 0x72, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x14, 0x6e, 0x6f, 0x64, 0x65, 0x73, 0x65, 0x6c, 0x65, 0x63,
	0x74, 0x6f, 0x72, 0x6f, 0x70, 0x65, 0x72, 0x61, 0x74, 0x6f, 0x72, 0x12, 0x16, 0x0a, 0x06, 0x76,
	0x61, 0x6c, 0x75, 0x65, 0x73, 0x18, 0x03, 0x20, 0x03, 0x28, 0x09, 0x52, 0x06, 0x76, 0x61, 0x6c,
	0x75, 0x65, 0x73, 0x1a, 0x92, 0x01, 0x0a, 0x0f, 0x52, 0x65, 0x64, 0x75, 0x63, 0x65, 0x64, 0x4f,
	0x66, 0x66, 0x65, 0x72, 0x69, 0x6e, 0x67, 0x12, 0x4b, 0x0a, 0x0c, 0x72, 0x65, 0x71, 0x75, 0x69,
	0x72, 0x65, 0x6d, 0x65, 0x6e, 0x74, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x27, 0x2e,
	0x52, 0x65, 0x64, 0x75, 0x63, 0x65, 0x64, 0x49, 0x6e, 0x73, 0x74, 0x61, 0x6e, 0x63, 0x65, 0x54,
	0x79, 0x70, 0x65, 0x2e, 0x52, 0x65, 0x64, 0x75, 0x63, 0x65, 0x64, 0x52, 0x65, 0x71, 0x75, 0x69,
	0x72, 0x65, 0x6d, 0x65, 0x6e, 0x74, 0x52, 0x0c, 0x72, 0x65, 0x71, 0x75, 0x69, 0x72, 0x65, 0x6d,
	0x65, 0x6e, 0x74, 0x73, 0x12, 0x14, 0x0a, 0x05, 0x70, 0x72, 0x69, 0x63, 0x65, 0x18, 0x02, 0x20,
	0x01, 0x28, 0x01, 0x52, 0x05, 0x70, 0x72, 0x69, 0x63, 0x65, 0x12, 0x1c, 0x0a, 0x09, 0x61, 0x76,
	0x61, 0x69, 0x6c, 0x61, 0x62, 0x6c, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x08, 0x52, 0x09, 0x61,
	0x76, 0x61, 0x69, 0x6c, 0x61, 0x62, 0x6c, 0x65, 0x22, 0x9c, 0x01, 0x0a, 0x15, 0x53, 0x63, 0x68,
	0x65, 0x64, 0x75, 0x6c, 0x69, 0x6e, 0x67, 0x4d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x4d,
	0x61, 0x70, 0x12, 0x3d, 0x0a, 0x07, 0x65, 0x6e, 0x74, 0x72, 0x69, 0x65, 0x73, 0x18, 0x01, 0x20,
	0x03, 0x28, 0x0b, 0x32, 0x23, 0x2e, 0x53, 0x63, 0x68, 0x65, 0x64, 0x75, 0x6c, 0x69, 0x6e, 0x67,
	0x4d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x4d, 0x61, 0x70, 0x2e, 0x4d, 0x61, 0x70, 0x70,
	0x69, 0x6e, 0x67, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x07, 0x65, 0x6e, 0x74, 0x72, 0x69, 0x65,
	0x73, 0x1a, 0x44, 0x0a, 0x0c, 0x4d, 0x61, 0x70, 0x70, 0x69, 0x6e, 0x67, 0x45, 0x6e, 0x74, 0x72,
	0x79, 0x12, 0x16, 0x0a, 0x06, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x06, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x1c, 0x0a, 0x09, 0x74, 0x69, 0x6d,
	0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x74, 0x69,
	0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x42, 0x31, 0x5a, 0x2f, 0x73, 0x69, 0x67, 0x73, 0x2e,
	0x6b, 0x38, 0x73, 0x2e, 0x69, 0x6f, 0x2f, 0x6b, 0x61, 0x72, 0x70, 0x65, 0x6e, 0x74, 0x65, 0x72,
	0x2f, 0x70, 0x6b, 0x67, 0x2f, 0x63, 0x6f, 0x6e, 0x74, 0x72, 0x6f, 0x6c, 0x6c, 0x65, 0x72, 0x73,
	0x2f, 0x6f, 0x72, 0x62, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x33,
}

var (
	file_SchedulingInput_proto_rawDescOnce sync.Once
	file_SchedulingInput_proto_rawDescData = file_SchedulingInput_proto_rawDesc
)

func file_SchedulingInput_proto_rawDescGZIP() []byte {
	file_SchedulingInput_proto_rawDescOnce.Do(func() {
		file_SchedulingInput_proto_rawDescData = protoimpl.X.CompressGZIP(file_SchedulingInput_proto_rawDescData)
	})
	return file_SchedulingInput_proto_rawDescData
}

var file_SchedulingInput_proto_msgTypes = make([]protoimpl.MessageInfo, 11)
var file_SchedulingInput_proto_goTypes = []any{
	(*Differences)(nil),                            // 0: Differences
	(*SchedulingInput)(nil),                        // 1: SchedulingInput
	(*ReducedPod)(nil),                             // 2: ReducedPod
	(*StateNodeWithPods)(nil),                      // 3: StateNodeWithPods
	(*ReducedInstanceType)(nil),                    // 4: ReducedInstanceType
	(*SchedulingMetadataMap)(nil),                  // 5: SchedulingMetadataMap
	(*StateNodeWithPods_ReducedNode)(nil),          // 6: StateNodeWithPods.ReducedNode
	(*StateNodeWithPods_ReducedNodeClaim)(nil),     // 7: StateNodeWithPods.ReducedNodeClaim
	(*ReducedInstanceType_ReducedRequirement)(nil), // 8: ReducedInstanceType.ReducedRequirement
	(*ReducedInstanceType_ReducedOffering)(nil),    // 9: ReducedInstanceType.ReducedOffering
	(*SchedulingMetadataMap_MappingEntry)(nil),     // 10: SchedulingMetadataMap.MappingEntry
}
var file_SchedulingInput_proto_depIdxs = []int32{
	1,  // 0: Differences.added:type_name -> SchedulingInput
	1,  // 1: Differences.removed:type_name -> SchedulingInput
	1,  // 2: Differences.changed:type_name -> SchedulingInput
	2,  // 3: SchedulingInput.pendingpod_data:type_name -> ReducedPod
	3,  // 4: SchedulingInput.statenodes_data:type_name -> StateNodeWithPods
	4,  // 5: SchedulingInput.instancetypes_data:type_name -> ReducedInstanceType
	6,  // 6: StateNodeWithPods.node:type_name -> StateNodeWithPods.ReducedNode
	7,  // 7: StateNodeWithPods.nodeClaim:type_name -> StateNodeWithPods.ReducedNodeClaim
	2,  // 8: StateNodeWithPods.pods:type_name -> ReducedPod
	8,  // 9: ReducedInstanceType.requirements:type_name -> ReducedInstanceType.ReducedRequirement
	9,  // 10: ReducedInstanceType.offerings:type_name -> ReducedInstanceType.ReducedOffering
	10, // 11: SchedulingMetadataMap.entries:type_name -> SchedulingMetadataMap.MappingEntry
	8,  // 12: ReducedInstanceType.ReducedOffering.requirements:type_name -> ReducedInstanceType.ReducedRequirement
	13, // [13:13] is the sub-list for method output_type
	13, // [13:13] is the sub-list for method input_type
	13, // [13:13] is the sub-list for extension type_name
	13, // [13:13] is the sub-list for extension extendee
	0,  // [0:13] is the sub-list for field type_name
}

func init() { file_SchedulingInput_proto_init() }
func file_SchedulingInput_proto_init() {
	if File_SchedulingInput_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_SchedulingInput_proto_msgTypes[0].Exporter = func(v any, i int) any {
			switch v := v.(*Differences); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_SchedulingInput_proto_msgTypes[1].Exporter = func(v any, i int) any {
			switch v := v.(*SchedulingInput); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_SchedulingInput_proto_msgTypes[2].Exporter = func(v any, i int) any {
			switch v := v.(*ReducedPod); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_SchedulingInput_proto_msgTypes[3].Exporter = func(v any, i int) any {
			switch v := v.(*StateNodeWithPods); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_SchedulingInput_proto_msgTypes[4].Exporter = func(v any, i int) any {
			switch v := v.(*ReducedInstanceType); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_SchedulingInput_proto_msgTypes[5].Exporter = func(v any, i int) any {
			switch v := v.(*SchedulingMetadataMap); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_SchedulingInput_proto_msgTypes[6].Exporter = func(v any, i int) any {
			switch v := v.(*StateNodeWithPods_ReducedNode); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_SchedulingInput_proto_msgTypes[7].Exporter = func(v any, i int) any {
			switch v := v.(*StateNodeWithPods_ReducedNodeClaim); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_SchedulingInput_proto_msgTypes[8].Exporter = func(v any, i int) any {
			switch v := v.(*ReducedInstanceType_ReducedRequirement); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_SchedulingInput_proto_msgTypes[9].Exporter = func(v any, i int) any {
			switch v := v.(*ReducedInstanceType_ReducedOffering); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_SchedulingInput_proto_msgTypes[10].Exporter = func(v any, i int) any {
			switch v := v.(*SchedulingMetadataMap_MappingEntry); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	file_SchedulingInput_proto_msgTypes[0].OneofWrappers = []any{}
	file_SchedulingInput_proto_msgTypes[3].OneofWrappers = []any{}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_SchedulingInput_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   11,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_SchedulingInput_proto_goTypes,
		DependencyIndexes: file_SchedulingInput_proto_depIdxs,
		MessageInfos:      file_SchedulingInput_proto_msgTypes,
	}.Build()
	File_SchedulingInput_proto = out.File
	file_SchedulingInput_proto_rawDesc = nil
	file_SchedulingInput_proto_goTypes = nil
	file_SchedulingInput_proto_depIdxs = nil
}
