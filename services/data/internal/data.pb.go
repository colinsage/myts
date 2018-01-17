// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: data.proto

/*
Package internal is a generated protocol buffer package.

It is generated from these files:
	data.proto

It has these top-level messages:
	WriteShardRequest
	WriteShardResponse
	ExecuteStatementRequest
	ExecuteStatementResponse
	CreateIteratorRequest
	CreateIteratorResponse
	Field
	FieldDimensionsRequest
	FieldDimensionsResponse
	MapTypeRequest
	MapTypeResponse
	IteratorCostRequest
	IteratorCostResponse
	IteratorCost
	Point
	Aux
	IteratorOptions
	Measurements
	Measurement
	Interval
	IteratorStats
	VarRef
*/
package internal

import proto "github.com/gogo/protobuf/proto"
import fmt "fmt"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.GoGoProtoPackageIsVersion2 // please upgrade the proto package

type WriteShardRequest struct {
	ShardID          *uint64  `protobuf:"varint,1,req,name=ShardID" json:"ShardID,omitempty"`
	Points           [][]byte `protobuf:"bytes,2,rep,name=Points" json:"Points,omitempty"`
	Database         *string  `protobuf:"bytes,3,opt,name=Database" json:"Database,omitempty"`
	RetentionPolicy  *string  `protobuf:"bytes,4,opt,name=RetentionPolicy" json:"RetentionPolicy,omitempty"`
	XXX_unrecognized []byte   `json:"-"`
}

func (m *WriteShardRequest) Reset()                    { *m = WriteShardRequest{} }
func (m *WriteShardRequest) String() string            { return proto.CompactTextString(m) }
func (*WriteShardRequest) ProtoMessage()               {}
func (*WriteShardRequest) Descriptor() ([]byte, []int) { return fileDescriptorData, []int{0} }

func (m *WriteShardRequest) GetShardID() uint64 {
	if m != nil && m.ShardID != nil {
		return *m.ShardID
	}
	return 0
}

func (m *WriteShardRequest) GetPoints() [][]byte {
	if m != nil {
		return m.Points
	}
	return nil
}

func (m *WriteShardRequest) GetDatabase() string {
	if m != nil && m.Database != nil {
		return *m.Database
	}
	return ""
}

func (m *WriteShardRequest) GetRetentionPolicy() string {
	if m != nil && m.RetentionPolicy != nil {
		return *m.RetentionPolicy
	}
	return ""
}

type WriteShardResponse struct {
	Code             *int32  `protobuf:"varint,1,req,name=Code" json:"Code,omitempty"`
	Message          *string `protobuf:"bytes,2,opt,name=Message" json:"Message,omitempty"`
	XXX_unrecognized []byte  `json:"-"`
}

func (m *WriteShardResponse) Reset()                    { *m = WriteShardResponse{} }
func (m *WriteShardResponse) String() string            { return proto.CompactTextString(m) }
func (*WriteShardResponse) ProtoMessage()               {}
func (*WriteShardResponse) Descriptor() ([]byte, []int) { return fileDescriptorData, []int{1} }

func (m *WriteShardResponse) GetCode() int32 {
	if m != nil && m.Code != nil {
		return *m.Code
	}
	return 0
}

func (m *WriteShardResponse) GetMessage() string {
	if m != nil && m.Message != nil {
		return *m.Message
	}
	return ""
}

type ExecuteStatementRequest struct {
	Statement        *string `protobuf:"bytes,1,req,name=Statement" json:"Statement,omitempty"`
	Database         *string `protobuf:"bytes,2,req,name=Database" json:"Database,omitempty"`
	XXX_unrecognized []byte  `json:"-"`
}

func (m *ExecuteStatementRequest) Reset()                    { *m = ExecuteStatementRequest{} }
func (m *ExecuteStatementRequest) String() string            { return proto.CompactTextString(m) }
func (*ExecuteStatementRequest) ProtoMessage()               {}
func (*ExecuteStatementRequest) Descriptor() ([]byte, []int) { return fileDescriptorData, []int{2} }

func (m *ExecuteStatementRequest) GetStatement() string {
	if m != nil && m.Statement != nil {
		return *m.Statement
	}
	return ""
}

func (m *ExecuteStatementRequest) GetDatabase() string {
	if m != nil && m.Database != nil {
		return *m.Database
	}
	return ""
}

type ExecuteStatementResponse struct {
	Code             *int32  `protobuf:"varint,1,req,name=Code" json:"Code,omitempty"`
	Message          *string `protobuf:"bytes,2,opt,name=Message" json:"Message,omitempty"`
	XXX_unrecognized []byte  `json:"-"`
}

func (m *ExecuteStatementResponse) Reset()                    { *m = ExecuteStatementResponse{} }
func (m *ExecuteStatementResponse) String() string            { return proto.CompactTextString(m) }
func (*ExecuteStatementResponse) ProtoMessage()               {}
func (*ExecuteStatementResponse) Descriptor() ([]byte, []int) { return fileDescriptorData, []int{3} }

func (m *ExecuteStatementResponse) GetCode() int32 {
	if m != nil && m.Code != nil {
		return *m.Code
	}
	return 0
}

func (m *ExecuteStatementResponse) GetMessage() string {
	if m != nil && m.Message != nil {
		return *m.Message
	}
	return ""
}

type CreateIteratorRequest struct {
	Measurement      []byte   `protobuf:"bytes,1,req,name=Measurement" json:"Measurement,omitempty"`
	Opt              []byte   `protobuf:"bytes,2,req,name=Opt" json:"Opt,omitempty"`
	ShardIDs         []uint64 `protobuf:"varint,3,rep,name=ShardIDs" json:"ShardIDs,omitempty"`
	XXX_unrecognized []byte   `json:"-"`
}

func (m *CreateIteratorRequest) Reset()                    { *m = CreateIteratorRequest{} }
func (m *CreateIteratorRequest) String() string            { return proto.CompactTextString(m) }
func (*CreateIteratorRequest) ProtoMessage()               {}
func (*CreateIteratorRequest) Descriptor() ([]byte, []int) { return fileDescriptorData, []int{4} }

func (m *CreateIteratorRequest) GetMeasurement() []byte {
	if m != nil {
		return m.Measurement
	}
	return nil
}

func (m *CreateIteratorRequest) GetOpt() []byte {
	if m != nil {
		return m.Opt
	}
	return nil
}

func (m *CreateIteratorRequest) GetShardIDs() []uint64 {
	if m != nil {
		return m.ShardIDs
	}
	return nil
}

type CreateIteratorResponse struct {
	Err              *string `protobuf:"bytes,1,opt,name=Err" json:"Err,omitempty"`
	XXX_unrecognized []byte  `json:"-"`
}

func (m *CreateIteratorResponse) Reset()                    { *m = CreateIteratorResponse{} }
func (m *CreateIteratorResponse) String() string            { return proto.CompactTextString(m) }
func (*CreateIteratorResponse) ProtoMessage()               {}
func (*CreateIteratorResponse) Descriptor() ([]byte, []int) { return fileDescriptorData, []int{5} }

func (m *CreateIteratorResponse) GetErr() string {
	if m != nil && m.Err != nil {
		return *m.Err
	}
	return ""
}

type Field struct {
	Name             *string `protobuf:"bytes,1,req,name=name" json:"name,omitempty"`
	Type             *uint64 `protobuf:"varint,2,req,name=type" json:"type,omitempty"`
	XXX_unrecognized []byte  `json:"-"`
}

func (m *Field) Reset()                    { *m = Field{} }
func (m *Field) String() string            { return proto.CompactTextString(m) }
func (*Field) ProtoMessage()               {}
func (*Field) Descriptor() ([]byte, []int) { return fileDescriptorData, []int{6} }

func (m *Field) GetName() string {
	if m != nil && m.Name != nil {
		return *m.Name
	}
	return ""
}

func (m *Field) GetType() uint64 {
	if m != nil && m.Type != nil {
		return *m.Type
	}
	return 0
}

type FieldDimensionsRequest struct {
	Measurement      *string  `protobuf:"bytes,1,req,name=Measurement" json:"Measurement,omitempty"`
	ShardIDs         []uint64 `protobuf:"varint,2,rep,name=ShardIDs" json:"ShardIDs,omitempty"`
	XXX_unrecognized []byte   `json:"-"`
}

func (m *FieldDimensionsRequest) Reset()                    { *m = FieldDimensionsRequest{} }
func (m *FieldDimensionsRequest) String() string            { return proto.CompactTextString(m) }
func (*FieldDimensionsRequest) ProtoMessage()               {}
func (*FieldDimensionsRequest) Descriptor() ([]byte, []int) { return fileDescriptorData, []int{7} }

func (m *FieldDimensionsRequest) GetMeasurement() string {
	if m != nil && m.Measurement != nil {
		return *m.Measurement
	}
	return ""
}

func (m *FieldDimensionsRequest) GetShardIDs() []uint64 {
	if m != nil {
		return m.ShardIDs
	}
	return nil
}

type FieldDimensionsResponse struct {
	Fields           []*Field `protobuf:"bytes,1,rep,name=Fields" json:"Fields,omitempty"`
	Dimensions       []string `protobuf:"bytes,2,rep,name=Dimensions" json:"Dimensions,omitempty"`
	Err              *string  `protobuf:"bytes,3,opt,name=Err" json:"Err,omitempty"`
	XXX_unrecognized []byte   `json:"-"`
}

func (m *FieldDimensionsResponse) Reset()                    { *m = FieldDimensionsResponse{} }
func (m *FieldDimensionsResponse) String() string            { return proto.CompactTextString(m) }
func (*FieldDimensionsResponse) ProtoMessage()               {}
func (*FieldDimensionsResponse) Descriptor() ([]byte, []int) { return fileDescriptorData, []int{8} }

func (m *FieldDimensionsResponse) GetFields() []*Field {
	if m != nil {
		return m.Fields
	}
	return nil
}

func (m *FieldDimensionsResponse) GetDimensions() []string {
	if m != nil {
		return m.Dimensions
	}
	return nil
}

func (m *FieldDimensionsResponse) GetErr() string {
	if m != nil && m.Err != nil {
		return *m.Err
	}
	return ""
}

type MapTypeRequest struct {
	ShardIDs         []uint64 `protobuf:"varint,1,rep,name=ShardIDs" json:"ShardIDs,omitempty"`
	Measurement      *string  `protobuf:"bytes,2,req,name=Measurement" json:"Measurement,omitempty"`
	Field            *string  `protobuf:"bytes,3,req,name=Field" json:"Field,omitempty"`
	XXX_unrecognized []byte   `json:"-"`
}

func (m *MapTypeRequest) Reset()                    { *m = MapTypeRequest{} }
func (m *MapTypeRequest) String() string            { return proto.CompactTextString(m) }
func (*MapTypeRequest) ProtoMessage()               {}
func (*MapTypeRequest) Descriptor() ([]byte, []int) { return fileDescriptorData, []int{9} }

func (m *MapTypeRequest) GetShardIDs() []uint64 {
	if m != nil {
		return m.ShardIDs
	}
	return nil
}

func (m *MapTypeRequest) GetMeasurement() string {
	if m != nil && m.Measurement != nil {
		return *m.Measurement
	}
	return ""
}

func (m *MapTypeRequest) GetField() string {
	if m != nil && m.Field != nil {
		return *m.Field
	}
	return ""
}

type MapTypeResponse struct {
	Type             *int64  `protobuf:"varint,1,req,name=Type" json:"Type,omitempty"`
	Err              *string `protobuf:"bytes,2,opt,name=Err" json:"Err,omitempty"`
	XXX_unrecognized []byte  `json:"-"`
}

func (m *MapTypeResponse) Reset()                    { *m = MapTypeResponse{} }
func (m *MapTypeResponse) String() string            { return proto.CompactTextString(m) }
func (*MapTypeResponse) ProtoMessage()               {}
func (*MapTypeResponse) Descriptor() ([]byte, []int) { return fileDescriptorData, []int{10} }

func (m *MapTypeResponse) GetType() int64 {
	if m != nil && m.Type != nil {
		return *m.Type
	}
	return 0
}

func (m *MapTypeResponse) GetErr() string {
	if m != nil && m.Err != nil {
		return *m.Err
	}
	return ""
}

type IteratorCostRequest struct {
	ShardIDs         []uint64         `protobuf:"varint,1,rep,name=ShardIDs" json:"ShardIDs,omitempty"`
	Measurement      *Measurement     `protobuf:"bytes,2,req,name=Measurement" json:"Measurement,omitempty"`
	IteratorOptions  *IteratorOptions `protobuf:"bytes,3,req,name=IteratorOptions" json:"IteratorOptions,omitempty"`
	XXX_unrecognized []byte           `json:"-"`
}

func (m *IteratorCostRequest) Reset()                    { *m = IteratorCostRequest{} }
func (m *IteratorCostRequest) String() string            { return proto.CompactTextString(m) }
func (*IteratorCostRequest) ProtoMessage()               {}
func (*IteratorCostRequest) Descriptor() ([]byte, []int) { return fileDescriptorData, []int{11} }

func (m *IteratorCostRequest) GetShardIDs() []uint64 {
	if m != nil {
		return m.ShardIDs
	}
	return nil
}

func (m *IteratorCostRequest) GetMeasurement() *Measurement {
	if m != nil {
		return m.Measurement
	}
	return nil
}

func (m *IteratorCostRequest) GetIteratorOptions() *IteratorOptions {
	if m != nil {
		return m.IteratorOptions
	}
	return nil
}

type IteratorCostResponse struct {
	Cost             *IteratorCost `protobuf:"bytes,1,req,name=Cost" json:"Cost,omitempty"`
	Err              *string       `protobuf:"bytes,2,opt,name=Err" json:"Err,omitempty"`
	XXX_unrecognized []byte        `json:"-"`
}

func (m *IteratorCostResponse) Reset()                    { *m = IteratorCostResponse{} }
func (m *IteratorCostResponse) String() string            { return proto.CompactTextString(m) }
func (*IteratorCostResponse) ProtoMessage()               {}
func (*IteratorCostResponse) Descriptor() ([]byte, []int) { return fileDescriptorData, []int{12} }

func (m *IteratorCostResponse) GetCost() *IteratorCost {
	if m != nil {
		return m.Cost
	}
	return nil
}

func (m *IteratorCostResponse) GetErr() string {
	if m != nil && m.Err != nil {
		return *m.Err
	}
	return ""
}

type IteratorCost struct {
	NumShards        *int64 `protobuf:"varint,1,req,name=NumShards" json:"NumShards,omitempty"`
	NumSeries        *int64 `protobuf:"varint,2,req,name=NumSeries" json:"NumSeries,omitempty"`
	CachedValues     *int64 `protobuf:"varint,3,req,name=CachedValues" json:"CachedValues,omitempty"`
	NumFiles         *int64 `protobuf:"varint,4,req,name=NumFiles" json:"NumFiles,omitempty"`
	BlocksRead       *int64 `protobuf:"varint,5,req,name=BlocksRead" json:"BlocksRead,omitempty"`
	BlockSize        *int64 `protobuf:"varint,6,req,name=BlockSize" json:"BlockSize,omitempty"`
	XXX_unrecognized []byte `json:"-"`
}

func (m *IteratorCost) Reset()                    { *m = IteratorCost{} }
func (m *IteratorCost) String() string            { return proto.CompactTextString(m) }
func (*IteratorCost) ProtoMessage()               {}
func (*IteratorCost) Descriptor() ([]byte, []int) { return fileDescriptorData, []int{13} }

func (m *IteratorCost) GetNumShards() int64 {
	if m != nil && m.NumShards != nil {
		return *m.NumShards
	}
	return 0
}

func (m *IteratorCost) GetNumSeries() int64 {
	if m != nil && m.NumSeries != nil {
		return *m.NumSeries
	}
	return 0
}

func (m *IteratorCost) GetCachedValues() int64 {
	if m != nil && m.CachedValues != nil {
		return *m.CachedValues
	}
	return 0
}

func (m *IteratorCost) GetNumFiles() int64 {
	if m != nil && m.NumFiles != nil {
		return *m.NumFiles
	}
	return 0
}

func (m *IteratorCost) GetBlocksRead() int64 {
	if m != nil && m.BlocksRead != nil {
		return *m.BlocksRead
	}
	return 0
}

func (m *IteratorCost) GetBlockSize() int64 {
	if m != nil && m.BlockSize != nil {
		return *m.BlockSize
	}
	return 0
}

type Point struct {
	Name             *string        `protobuf:"bytes,1,req,name=Name" json:"Name,omitempty"`
	Tags             *string        `protobuf:"bytes,2,req,name=Tags" json:"Tags,omitempty"`
	Time             *int64         `protobuf:"varint,3,req,name=Time" json:"Time,omitempty"`
	Nil              *bool          `protobuf:"varint,4,req,name=Nil" json:"Nil,omitempty"`
	Aux              []*Aux         `protobuf:"bytes,5,rep,name=Aux" json:"Aux,omitempty"`
	Aggregated       *uint32        `protobuf:"varint,6,opt,name=Aggregated" json:"Aggregated,omitempty"`
	FloatValue       *float64       `protobuf:"fixed64,7,opt,name=FloatValue" json:"FloatValue,omitempty"`
	IntegerValue     *int64         `protobuf:"varint,8,opt,name=IntegerValue" json:"IntegerValue,omitempty"`
	StringValue      *string        `protobuf:"bytes,9,opt,name=StringValue" json:"StringValue,omitempty"`
	BooleanValue     *bool          `protobuf:"varint,10,opt,name=BooleanValue" json:"BooleanValue,omitempty"`
	UnsignedValue    *uint64        `protobuf:"varint,12,opt,name=UnsignedValue" json:"UnsignedValue,omitempty"`
	Stats            *IteratorStats `protobuf:"bytes,11,opt,name=Stats" json:"Stats,omitempty"`
	Trace            []byte         `protobuf:"bytes,13,opt,name=Trace" json:"Trace,omitempty"`
	XXX_unrecognized []byte         `json:"-"`
}

func (m *Point) Reset()                    { *m = Point{} }
func (m *Point) String() string            { return proto.CompactTextString(m) }
func (*Point) ProtoMessage()               {}
func (*Point) Descriptor() ([]byte, []int) { return fileDescriptorData, []int{14} }

func (m *Point) GetName() string {
	if m != nil && m.Name != nil {
		return *m.Name
	}
	return ""
}

func (m *Point) GetTags() string {
	if m != nil && m.Tags != nil {
		return *m.Tags
	}
	return ""
}

func (m *Point) GetTime() int64 {
	if m != nil && m.Time != nil {
		return *m.Time
	}
	return 0
}

func (m *Point) GetNil() bool {
	if m != nil && m.Nil != nil {
		return *m.Nil
	}
	return false
}

func (m *Point) GetAux() []*Aux {
	if m != nil {
		return m.Aux
	}
	return nil
}

func (m *Point) GetAggregated() uint32 {
	if m != nil && m.Aggregated != nil {
		return *m.Aggregated
	}
	return 0
}

func (m *Point) GetFloatValue() float64 {
	if m != nil && m.FloatValue != nil {
		return *m.FloatValue
	}
	return 0
}

func (m *Point) GetIntegerValue() int64 {
	if m != nil && m.IntegerValue != nil {
		return *m.IntegerValue
	}
	return 0
}

func (m *Point) GetStringValue() string {
	if m != nil && m.StringValue != nil {
		return *m.StringValue
	}
	return ""
}

func (m *Point) GetBooleanValue() bool {
	if m != nil && m.BooleanValue != nil {
		return *m.BooleanValue
	}
	return false
}

func (m *Point) GetUnsignedValue() uint64 {
	if m != nil && m.UnsignedValue != nil {
		return *m.UnsignedValue
	}
	return 0
}

func (m *Point) GetStats() *IteratorStats {
	if m != nil {
		return m.Stats
	}
	return nil
}

func (m *Point) GetTrace() []byte {
	if m != nil {
		return m.Trace
	}
	return nil
}

type Aux struct {
	DataType         *int32   `protobuf:"varint,1,req,name=DataType" json:"DataType,omitempty"`
	FloatValue       *float64 `protobuf:"fixed64,2,opt,name=FloatValue" json:"FloatValue,omitempty"`
	IntegerValue     *int64   `protobuf:"varint,3,opt,name=IntegerValue" json:"IntegerValue,omitempty"`
	StringValue      *string  `protobuf:"bytes,4,opt,name=StringValue" json:"StringValue,omitempty"`
	BooleanValue     *bool    `protobuf:"varint,5,opt,name=BooleanValue" json:"BooleanValue,omitempty"`
	UnsignedValue    *uint64  `protobuf:"varint,6,opt,name=UnsignedValue" json:"UnsignedValue,omitempty"`
	XXX_unrecognized []byte   `json:"-"`
}

func (m *Aux) Reset()                    { *m = Aux{} }
func (m *Aux) String() string            { return proto.CompactTextString(m) }
func (*Aux) ProtoMessage()               {}
func (*Aux) Descriptor() ([]byte, []int) { return fileDescriptorData, []int{15} }

func (m *Aux) GetDataType() int32 {
	if m != nil && m.DataType != nil {
		return *m.DataType
	}
	return 0
}

func (m *Aux) GetFloatValue() float64 {
	if m != nil && m.FloatValue != nil {
		return *m.FloatValue
	}
	return 0
}

func (m *Aux) GetIntegerValue() int64 {
	if m != nil && m.IntegerValue != nil {
		return *m.IntegerValue
	}
	return 0
}

func (m *Aux) GetStringValue() string {
	if m != nil && m.StringValue != nil {
		return *m.StringValue
	}
	return ""
}

func (m *Aux) GetBooleanValue() bool {
	if m != nil && m.BooleanValue != nil {
		return *m.BooleanValue
	}
	return false
}

func (m *Aux) GetUnsignedValue() uint64 {
	if m != nil && m.UnsignedValue != nil {
		return *m.UnsignedValue
	}
	return 0
}

type IteratorOptions struct {
	Expr             *string        `protobuf:"bytes,1,opt,name=Expr" json:"Expr,omitempty"`
	Aux              []string       `protobuf:"bytes,2,rep,name=Aux" json:"Aux,omitempty"`
	Fields           []*VarRef      `protobuf:"bytes,17,rep,name=Fields" json:"Fields,omitempty"`
	Sources          []*Measurement `protobuf:"bytes,3,rep,name=Sources" json:"Sources,omitempty"`
	Interval         *Interval      `protobuf:"bytes,4,opt,name=Interval" json:"Interval,omitempty"`
	Dimensions       []string       `protobuf:"bytes,5,rep,name=Dimensions" json:"Dimensions,omitempty"`
	GroupBy          []string       `protobuf:"bytes,19,rep,name=GroupBy" json:"GroupBy,omitempty"`
	Fill             *int32         `protobuf:"varint,6,opt,name=Fill" json:"Fill,omitempty"`
	FillValue        *float64       `protobuf:"fixed64,7,opt,name=FillValue" json:"FillValue,omitempty"`
	Condition        *string        `protobuf:"bytes,8,opt,name=Condition" json:"Condition,omitempty"`
	StartTime        *int64         `protobuf:"varint,9,opt,name=StartTime" json:"StartTime,omitempty"`
	EndTime          *int64         `protobuf:"varint,10,opt,name=EndTime" json:"EndTime,omitempty"`
	Location         *string        `protobuf:"bytes,21,opt,name=Location" json:"Location,omitempty"`
	Ascending        *bool          `protobuf:"varint,11,opt,name=Ascending" json:"Ascending,omitempty"`
	Limit            *int64         `protobuf:"varint,12,opt,name=Limit" json:"Limit,omitempty"`
	Offset           *int64         `protobuf:"varint,13,opt,name=Offset" json:"Offset,omitempty"`
	SLimit           *int64         `protobuf:"varint,14,opt,name=SLimit" json:"SLimit,omitempty"`
	SOffset          *int64         `protobuf:"varint,15,opt,name=SOffset" json:"SOffset,omitempty"`
	StripName        *bool          `protobuf:"varint,22,opt,name=StripName" json:"StripName,omitempty"`
	Dedupe           *bool          `protobuf:"varint,16,opt,name=Dedupe" json:"Dedupe,omitempty"`
	MaxSeriesN       *int64         `protobuf:"varint,18,opt,name=MaxSeriesN" json:"MaxSeriesN,omitempty"`
	Ordered          *bool          `protobuf:"varint,20,opt,name=Ordered" json:"Ordered,omitempty"`
	XXX_unrecognized []byte         `json:"-"`
}

func (m *IteratorOptions) Reset()                    { *m = IteratorOptions{} }
func (m *IteratorOptions) String() string            { return proto.CompactTextString(m) }
func (*IteratorOptions) ProtoMessage()               {}
func (*IteratorOptions) Descriptor() ([]byte, []int) { return fileDescriptorData, []int{16} }

func (m *IteratorOptions) GetExpr() string {
	if m != nil && m.Expr != nil {
		return *m.Expr
	}
	return ""
}

func (m *IteratorOptions) GetAux() []string {
	if m != nil {
		return m.Aux
	}
	return nil
}

func (m *IteratorOptions) GetFields() []*VarRef {
	if m != nil {
		return m.Fields
	}
	return nil
}

func (m *IteratorOptions) GetSources() []*Measurement {
	if m != nil {
		return m.Sources
	}
	return nil
}

func (m *IteratorOptions) GetInterval() *Interval {
	if m != nil {
		return m.Interval
	}
	return nil
}

func (m *IteratorOptions) GetDimensions() []string {
	if m != nil {
		return m.Dimensions
	}
	return nil
}

func (m *IteratorOptions) GetGroupBy() []string {
	if m != nil {
		return m.GroupBy
	}
	return nil
}

func (m *IteratorOptions) GetFill() int32 {
	if m != nil && m.Fill != nil {
		return *m.Fill
	}
	return 0
}

func (m *IteratorOptions) GetFillValue() float64 {
	if m != nil && m.FillValue != nil {
		return *m.FillValue
	}
	return 0
}

func (m *IteratorOptions) GetCondition() string {
	if m != nil && m.Condition != nil {
		return *m.Condition
	}
	return ""
}

func (m *IteratorOptions) GetStartTime() int64 {
	if m != nil && m.StartTime != nil {
		return *m.StartTime
	}
	return 0
}

func (m *IteratorOptions) GetEndTime() int64 {
	if m != nil && m.EndTime != nil {
		return *m.EndTime
	}
	return 0
}

func (m *IteratorOptions) GetLocation() string {
	if m != nil && m.Location != nil {
		return *m.Location
	}
	return ""
}

func (m *IteratorOptions) GetAscending() bool {
	if m != nil && m.Ascending != nil {
		return *m.Ascending
	}
	return false
}

func (m *IteratorOptions) GetLimit() int64 {
	if m != nil && m.Limit != nil {
		return *m.Limit
	}
	return 0
}

func (m *IteratorOptions) GetOffset() int64 {
	if m != nil && m.Offset != nil {
		return *m.Offset
	}
	return 0
}

func (m *IteratorOptions) GetSLimit() int64 {
	if m != nil && m.SLimit != nil {
		return *m.SLimit
	}
	return 0
}

func (m *IteratorOptions) GetSOffset() int64 {
	if m != nil && m.SOffset != nil {
		return *m.SOffset
	}
	return 0
}

func (m *IteratorOptions) GetStripName() bool {
	if m != nil && m.StripName != nil {
		return *m.StripName
	}
	return false
}

func (m *IteratorOptions) GetDedupe() bool {
	if m != nil && m.Dedupe != nil {
		return *m.Dedupe
	}
	return false
}

func (m *IteratorOptions) GetMaxSeriesN() int64 {
	if m != nil && m.MaxSeriesN != nil {
		return *m.MaxSeriesN
	}
	return 0
}

func (m *IteratorOptions) GetOrdered() bool {
	if m != nil && m.Ordered != nil {
		return *m.Ordered
	}
	return false
}

type Measurements struct {
	Items            []*Measurement `protobuf:"bytes,1,rep,name=Items" json:"Items,omitempty"`
	XXX_unrecognized []byte         `json:"-"`
}

func (m *Measurements) Reset()                    { *m = Measurements{} }
func (m *Measurements) String() string            { return proto.CompactTextString(m) }
func (*Measurements) ProtoMessage()               {}
func (*Measurements) Descriptor() ([]byte, []int) { return fileDescriptorData, []int{17} }

func (m *Measurements) GetItems() []*Measurement {
	if m != nil {
		return m.Items
	}
	return nil
}

type Measurement struct {
	Database         *string `protobuf:"bytes,1,opt,name=Database" json:"Database,omitempty"`
	RetentionPolicy  *string `protobuf:"bytes,2,opt,name=RetentionPolicy" json:"RetentionPolicy,omitempty"`
	Name             *string `protobuf:"bytes,3,opt,name=Name" json:"Name,omitempty"`
	Regex            *string `protobuf:"bytes,4,opt,name=Regex" json:"Regex,omitempty"`
	IsTarget         *bool   `protobuf:"varint,5,opt,name=IsTarget" json:"IsTarget,omitempty"`
	SystemIterator   *string `protobuf:"bytes,6,opt,name=SystemIterator" json:"SystemIterator,omitempty"`
	XXX_unrecognized []byte  `json:"-"`
}

func (m *Measurement) Reset()                    { *m = Measurement{} }
func (m *Measurement) String() string            { return proto.CompactTextString(m) }
func (*Measurement) ProtoMessage()               {}
func (*Measurement) Descriptor() ([]byte, []int) { return fileDescriptorData, []int{18} }

func (m *Measurement) GetDatabase() string {
	if m != nil && m.Database != nil {
		return *m.Database
	}
	return ""
}

func (m *Measurement) GetRetentionPolicy() string {
	if m != nil && m.RetentionPolicy != nil {
		return *m.RetentionPolicy
	}
	return ""
}

func (m *Measurement) GetName() string {
	if m != nil && m.Name != nil {
		return *m.Name
	}
	return ""
}

func (m *Measurement) GetRegex() string {
	if m != nil && m.Regex != nil {
		return *m.Regex
	}
	return ""
}

func (m *Measurement) GetIsTarget() bool {
	if m != nil && m.IsTarget != nil {
		return *m.IsTarget
	}
	return false
}

func (m *Measurement) GetSystemIterator() string {
	if m != nil && m.SystemIterator != nil {
		return *m.SystemIterator
	}
	return ""
}

type Interval struct {
	Duration         *int64 `protobuf:"varint,1,opt,name=Duration" json:"Duration,omitempty"`
	Offset           *int64 `protobuf:"varint,2,opt,name=Offset" json:"Offset,omitempty"`
	XXX_unrecognized []byte `json:"-"`
}

func (m *Interval) Reset()                    { *m = Interval{} }
func (m *Interval) String() string            { return proto.CompactTextString(m) }
func (*Interval) ProtoMessage()               {}
func (*Interval) Descriptor() ([]byte, []int) { return fileDescriptorData, []int{19} }

func (m *Interval) GetDuration() int64 {
	if m != nil && m.Duration != nil {
		return *m.Duration
	}
	return 0
}

func (m *Interval) GetOffset() int64 {
	if m != nil && m.Offset != nil {
		return *m.Offset
	}
	return 0
}

type IteratorStats struct {
	SeriesN          *int64 `protobuf:"varint,1,opt,name=SeriesN" json:"SeriesN,omitempty"`
	PointN           *int64 `protobuf:"varint,2,opt,name=PointN" json:"PointN,omitempty"`
	XXX_unrecognized []byte `json:"-"`
}

func (m *IteratorStats) Reset()                    { *m = IteratorStats{} }
func (m *IteratorStats) String() string            { return proto.CompactTextString(m) }
func (*IteratorStats) ProtoMessage()               {}
func (*IteratorStats) Descriptor() ([]byte, []int) { return fileDescriptorData, []int{20} }

func (m *IteratorStats) GetSeriesN() int64 {
	if m != nil && m.SeriesN != nil {
		return *m.SeriesN
	}
	return 0
}

func (m *IteratorStats) GetPointN() int64 {
	if m != nil && m.PointN != nil {
		return *m.PointN
	}
	return 0
}

type VarRef struct {
	Val              *string `protobuf:"bytes,1,req,name=Val" json:"Val,omitempty"`
	Type             *int32  `protobuf:"varint,2,opt,name=Type" json:"Type,omitempty"`
	XXX_unrecognized []byte  `json:"-"`
}

func (m *VarRef) Reset()                    { *m = VarRef{} }
func (m *VarRef) String() string            { return proto.CompactTextString(m) }
func (*VarRef) ProtoMessage()               {}
func (*VarRef) Descriptor() ([]byte, []int) { return fileDescriptorData, []int{21} }

func (m *VarRef) GetVal() string {
	if m != nil && m.Val != nil {
		return *m.Val
	}
	return ""
}

func (m *VarRef) GetType() int32 {
	if m != nil && m.Type != nil {
		return *m.Type
	}
	return 0
}

func init() {
	proto.RegisterType((*WriteShardRequest)(nil), "internal.WriteShardRequest")
	proto.RegisterType((*WriteShardResponse)(nil), "internal.WriteShardResponse")
	proto.RegisterType((*ExecuteStatementRequest)(nil), "internal.ExecuteStatementRequest")
	proto.RegisterType((*ExecuteStatementResponse)(nil), "internal.ExecuteStatementResponse")
	proto.RegisterType((*CreateIteratorRequest)(nil), "internal.CreateIteratorRequest")
	proto.RegisterType((*CreateIteratorResponse)(nil), "internal.CreateIteratorResponse")
	proto.RegisterType((*Field)(nil), "internal.Field")
	proto.RegisterType((*FieldDimensionsRequest)(nil), "internal.FieldDimensionsRequest")
	proto.RegisterType((*FieldDimensionsResponse)(nil), "internal.FieldDimensionsResponse")
	proto.RegisterType((*MapTypeRequest)(nil), "internal.MapTypeRequest")
	proto.RegisterType((*MapTypeResponse)(nil), "internal.MapTypeResponse")
	proto.RegisterType((*IteratorCostRequest)(nil), "internal.IteratorCostRequest")
	proto.RegisterType((*IteratorCostResponse)(nil), "internal.IteratorCostResponse")
	proto.RegisterType((*IteratorCost)(nil), "internal.IteratorCost")
	proto.RegisterType((*Point)(nil), "internal.Point")
	proto.RegisterType((*Aux)(nil), "internal.Aux")
	proto.RegisterType((*IteratorOptions)(nil), "internal.IteratorOptions")
	proto.RegisterType((*Measurements)(nil), "internal.Measurements")
	proto.RegisterType((*Measurement)(nil), "internal.Measurement")
	proto.RegisterType((*Interval)(nil), "internal.Interval")
	proto.RegisterType((*IteratorStats)(nil), "internal.IteratorStats")
	proto.RegisterType((*VarRef)(nil), "internal.VarRef")
}

func init() { proto.RegisterFile("data.proto", fileDescriptorData) }

var fileDescriptorData = []byte{
	// 1026 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x8c, 0x55, 0xcd, 0x6f, 0xeb, 0xc4,
	0x17, 0x95, 0x63, 0x3b, 0x4d, 0x6e, 0x9c, 0xa6, 0x75, 0xbf, 0xe6, 0xf7, 0xdb, 0x60, 0xf9, 0x95,
	0x27, 0x0b, 0xa1, 0x0a, 0x05, 0x36, 0x6c, 0x40, 0xfd, 0x44, 0x11, 0xaf, 0xe9, 0x53, 0x53, 0xca,
	0x82, 0x05, 0x1a, 0xe2, 0x5b, 0xbf, 0x11, 0x8e, 0x6d, 0x66, 0xc6, 0x28, 0x65, 0xc3, 0x82, 0x05,
	0x6b, 0xfe, 0x29, 0xfe, 0x2e, 0x34, 0xd7, 0x76, 0x3e, 0x5b, 0xc4, 0x72, 0xee, 0xcc, 0xdc, 0x7b,
	0xce, 0x3d, 0x67, 0xee, 0x00, 0xc4, 0x5c, 0xf3, 0xb3, 0x42, 0xe6, 0x3a, 0xf7, 0x3b, 0x22, 0xd3,
	0x28, 0x33, 0x9e, 0x86, 0x3f, 0xc2, 0xfe, 0xf7, 0x52, 0x68, 0x9c, 0x7c, 0xe0, 0x32, 0xbe, 0xc7,
	0x5f, 0x4a, 0x54, 0xda, 0x1f, 0xc0, 0x0e, 0xad, 0x47, 0x57, 0xcc, 0x0a, 0x5a, 0x91, 0xe3, 0xef,
	0x42, 0xfb, 0x7d, 0x2e, 0x32, 0xad, 0x58, 0x2b, 0xb0, 0x23, 0xcf, 0xdf, 0x83, 0xce, 0x15, 0xd7,
	0xfc, 0x27, 0xae, 0x90, 0xd9, 0x81, 0x15, 0x75, 0xfd, 0x13, 0x18, 0xdc, 0xa3, 0xc6, 0x4c, 0x8b,
	0x3c, 0x7b, 0x9f, 0xa7, 0x62, 0xfa, 0xcc, 0x1c, 0xb3, 0x11, 0x7e, 0x0e, 0xfe, 0x6a, 0x01, 0x55,
	0xe4, 0x99, 0x42, 0xdf, 0x03, 0xe7, 0x32, 0x8f, 0x91, 0xd2, 0xbb, 0xa6, 0xde, 0x2d, 0x2a, 0xc5,
	0x13, 0x64, 0x2d, 0xba, 0xf4, 0x15, 0x9c, 0x5c, 0xcf, 0x71, 0x5a, 0x6a, 0x9c, 0x68, 0xae, 0x71,
	0x86, 0x99, 0x6e, 0xb0, 0xed, 0x43, 0x77, 0x11, 0xa3, 0xeb, 0xdd, 0x35, 0x34, 0x2d, 0x13, 0x09,
	0xbf, 0x04, 0xb6, 0x7d, 0xff, 0xbf, 0x95, 0xfe, 0x16, 0x8e, 0x2e, 0x25, 0x72, 0x8d, 0x23, 0x8d,
	0x92, 0xeb, 0x5c, 0x36, 0x85, 0x0f, 0xa0, 0x77, 0x8b, 0x5c, 0x95, 0x72, 0x59, 0xda, 0xf3, 0x7b,
	0x60, 0xdf, 0x15, 0x9a, 0xaa, 0x52, 0x57, 0xea, 0xb6, 0x29, 0x66, 0x07, 0x76, 0xe4, 0x84, 0x1f,
	0xc3, 0xf1, 0x66, 0xb2, 0x1a, 0x45, 0x0f, 0xec, 0x6b, 0x29, 0x99, 0x45, 0x35, 0xdf, 0x80, 0x7b,
	0x23, 0x30, 0x8d, 0x0d, 0xb6, 0x8c, 0xcf, 0xb0, 0xe6, 0xe5, 0x81, 0xa3, 0x9f, 0x8b, 0x8a, 0x93,
	0x13, 0x7e, 0x0d, 0xc7, 0x74, 0xe8, 0x4a, 0xcc, 0x30, 0x53, 0x22, 0xcf, 0xd4, 0xbf, 0x20, 0xeb,
	0xae, 0x81, 0x69, 0x11, 0x98, 0x1f, 0xe0, 0x64, 0x2b, 0x41, 0x8d, 0xe6, 0x23, 0x68, 0xd3, 0x96,
	0x62, 0x56, 0x60, 0x47, 0xbd, 0xe1, 0xe0, 0xac, 0x31, 0xc8, 0x59, 0x05, 0xcc, 0x07, 0x58, 0x5e,
	0xa3, 0x7c, 0xdd, 0x86, 0x02, 0xe9, 0x1f, 0xde, 0xc0, 0xee, 0x2d, 0x2f, 0x1e, 0x9e, 0x0b, 0x6c,
	0x50, 0xad, 0x02, 0x30, 0x59, 0x9d, 0x4d, 0x9c, 0x24, 0x95, 0xdf, 0xaf, 0xb9, 0x33, 0x9b, 0x94,
	0xfb, 0x14, 0x06, 0x8b, 0x3c, 0x4b, 0xc1, 0xcc, 0x9a, 0x78, 0xd9, 0x4d, 0xd5, 0x4a, 0xac, 0x3f,
	0x2d, 0x38, 0x68, 0x5a, 0x7b, 0x99, 0x2b, 0xfd, 0x7a, 0xed, 0x4f, 0xb6, 0x6b, 0xf7, 0x86, 0x47,
	0x4b, 0x9a, 0x2b, 0x9b, 0xfe, 0x10, 0x06, 0x4d, 0xd2, 0xbb, 0x42, 0x13, 0x63, 0x9b, 0xce, 0xff,
	0x6f, 0x79, 0x7e, 0xe3, 0x40, 0x38, 0x82, 0xc3, 0x75, 0x20, 0x35, 0xf8, 0x53, 0xe3, 0x36, 0x55,
	0x89, 0xd2, 0x1b, 0x1e, 0x6f, 0x27, 0x30, 0xbb, 0xeb, 0xa4, 0xfe, 0xb0, 0xc0, 0x5b, 0xdb, 0xdd,
	0x87, 0xee, 0xb8, 0x9c, 0x11, 0x21, 0x55, 0x77, 0xa1, 0x0e, 0xa1, 0x14, 0xa8, 0x88, 0x8c, 0xed,
	0x1f, 0x82, 0x77, 0xc9, 0xa7, 0x1f, 0x30, 0x7e, 0xe4, 0x69, 0x89, 0x15, 0x64, 0xdb, 0x74, 0x62,
	0x5c, 0xce, 0x6e, 0x44, 0x8a, 0x8a, 0x39, 0x14, 0xf1, 0x01, 0x2e, 0xd2, 0x7c, 0xfa, 0xb3, 0xba,
	0x47, 0x1e, 0x33, 0xb7, 0x49, 0x47, 0xb1, 0x89, 0xf8, 0x0d, 0x59, 0xdb, 0x84, 0xc2, 0xbf, 0x5a,
	0xe0, 0xd2, 0x9b, 0x37, 0xfd, 0x1f, 0xaf, 0x99, 0xf2, 0x81, 0x27, 0xaa, 0x56, 0xcf, 0xac, 0xc4,
	0x0c, 0xeb, 0x62, 0x3d, 0xb0, 0xc7, 0x22, 0xa5, 0x3a, 0x1d, 0xff, 0xff, 0x60, 0x9f, 0x97, 0x73,
	0xe6, 0x92, 0xa1, 0xfa, 0x4b, 0xe2, 0xe7, 0xe5, 0xdc, 0x60, 0x38, 0x4f, 0x12, 0x89, 0x09, 0xd7,
	0x18, 0xb3, 0x76, 0x60, 0x45, 0x7d, 0x13, 0xbb, 0x49, 0x73, 0xae, 0x09, 0x3e, 0xdb, 0x09, 0xac,
	0xc8, 0x32, 0x9c, 0x46, 0x99, 0xc6, 0x04, 0x65, 0x15, 0xed, 0x04, 0x56, 0x64, 0x1b, 0x1f, 0x4d,
	0xb4, 0x14, 0x59, 0x52, 0x05, 0xbb, 0x34, 0x80, 0x0e, 0xc1, 0xbb, 0xc8, 0xf3, 0x14, 0x79, 0x56,
	0x45, 0x21, 0xb0, 0xa2, 0x8e, 0x7f, 0x04, 0xfd, 0xef, 0x32, 0x25, 0x92, 0xac, 0x6e, 0x0b, 0xf3,
	0x02, 0x2b, 0x72, 0xfc, 0xb7, 0xe0, 0x9a, 0xc1, 0xa0, 0x58, 0x2f, 0xb0, 0xa2, 0xde, 0xf0, 0x64,
	0x5b, 0x16, 0xda, 0x36, 0xe6, 0x7c, 0x90, 0x7c, 0x8a, 0xac, 0x1f, 0x58, 0x91, 0x67, 0x94, 0x31,
	0x9c, 0x9a, 0x81, 0xb3, 0x70, 0xa5, 0xbb, 0x01, 0xbe, 0xf5, 0x22, 0x78, 0xfb, 0x25, 0xf0, 0xce,
	0x8b, 0xe0, 0xdd, 0x97, 0xc1, 0x9b, 0x46, 0x39, 0xe1, 0xdf, 0xf6, 0x96, 0x3f, 0x8d, 0x0e, 0xd7,
	0xf3, 0xa2, 0x9e, 0x27, 0x46, 0x07, 0xd3, 0xfa, 0xea, 0x99, 0x06, 0x8b, 0xb7, 0xbd, 0x4f, 0x52,
	0xec, 0x2d, 0xc9, 0x3e, 0x72, 0x79, 0x8f, 0x4f, 0xfe, 0x5b, 0xd8, 0x99, 0xe4, 0xa5, 0x9c, 0x62,
	0x35, 0xb6, 0x5e, 0x7d, 0x17, 0xa7, 0xd0, 0x31, 0x84, 0xe4, 0xaf, 0x3c, 0x25, 0xdc, 0xbd, 0xa1,
	0xbf, 0xd2, 0xb8, 0x7a, 0x67, 0x63, 0x54, 0xb8, 0x84, 0x61, 0x00, 0x3b, 0xdf, 0xc8, 0xbc, 0x2c,
	0x2e, 0x9e, 0xd9, 0x01, 0x05, 0x3c, 0x70, 0x6e, 0x44, 0x9a, 0x12, 0x23, 0xd7, 0xd8, 0xcf, 0xac,
	0x56, 0x95, 0xdf, 0x87, 0xee, 0x65, 0x9e, 0xc5, 0xc2, 0xd0, 0x23, 0xd9, 0xbb, 0xf5, 0xe4, 0x97,
	0x9a, 0x0c, 0xd7, 0xa5, 0x66, 0x0e, 0x60, 0xe7, 0x3a, 0x8b, 0x29, 0x00, 0x14, 0xd8, 0x83, 0xce,
	0xbb, 0x7c, 0xca, 0xe9, 0xd6, 0x51, 0x73, 0xeb, 0x5c, 0x4d, 0x31, 0x8b, 0x45, 0x96, 0x90, 0xdc,
	0x1d, 0xa3, 0xea, 0x3b, 0x31, 0x13, 0x9a, 0xcc, 0x60, 0x9b, 0xcf, 0xed, 0xee, 0xe9, 0x49, 0xa1,
	0x26, 0x95, 0x69, 0x3d, 0xa9, 0xf6, 0x77, 0x9b, 0x22, 0x93, 0xfa, 0xc0, 0x80, 0x02, 0x04, 0x44,
	0x8a, 0x82, 0x5e, 0xc5, 0x31, 0xa5, 0xdc, 0x85, 0xf6, 0x15, 0xc6, 0x65, 0x81, 0x6c, 0x8f, 0xd6,
	0x3e, 0xc0, 0x2d, 0x9f, 0x57, 0xef, 0x73, 0xcc, 0xfc, 0x26, 0xcf, 0x9d, 0x8c, 0x51, 0x62, 0xcc,
	0x0e, 0xcd, 0xa1, 0xf0, 0x0b, 0xf0, 0x56, 0xda, 0xab, 0xfc, 0x53, 0x70, 0x47, 0x1a, 0x67, 0xcd,
	0x10, 0x7e, 0x59, 0x85, 0xf0, 0xf7, 0xb5, 0x49, 0xb6, 0xf6, 0xf9, 0x59, 0xaf, 0x7d, 0xc5, 0x34,
	0x58, 0x16, 0x0f, 0xb9, 0xfa, 0xb1, 0xfb, 0xe0, 0xde, 0x63, 0x82, 0xf3, 0xda, 0x82, 0x7b, 0xd0,
	0x19, 0xa9, 0x07, 0x2e, 0x13, 0xd4, 0xb5, 0xfd, 0x8e, 0x61, 0x77, 0xf2, 0xac, 0x34, 0xce, 0x1a,
	0xb3, 0x91, 0x5a, 0x66, 0x44, 0x2f, 0x6c, 0x40, 0xd5, 0x4b, 0x59, 0xf5, 0xdb, 0xda, 0xe8, 0xa6,
	0x29, 0x6a, 0x87, 0x9f, 0x41, 0x7f, 0xfd, 0x4d, 0x99, 0x76, 0xd6, 0x7d, 0x59, 0xdc, 0xa0, 0x41,
	0x33, 0xae, 0x6f, 0xbc, 0x81, 0x76, 0x6d, 0xcc, 0x1e, 0xd8, 0x8f, 0x3c, 0x5d, 0x19, 0x3c, 0xd5,
	0x6f, 0x68, 0x45, 0xee, 0x3f, 0x01, 0x00, 0x00, 0xff, 0xff, 0x69, 0xef, 0x3a, 0xab, 0xce, 0x08,
	0x00, 0x00,
}
