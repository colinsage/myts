package data

import (
	"errors"
	"fmt"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/influxdata/influxdb/query"
	"github.com/influxdata/influxdb/models"
	"gitlab.alibaba-inc.com/yuansheng.xt/myts/services/data/internal"
	"github.com/influxdata/influxql"
	"regexp"
)

//go:generate protoc --gogo_out=. internal/data.proto

// WritePointsRequest represents a request to write point data to the cluster
//type WritePointsRequest struct {
//	Database         string
//	RetentionPolicy  string
//	ConsistencyLevel ConsistencyLevel
//	Points           []models.Point
//}
//
//// AddPoint adds a point to the WritePointRequest with field key 'value'
//func (w *WritePointsRequest) AddPoint(name string, value interface{}, timestamp time.Time, tags models.Tags) {
//	pt, err := models.NewPoint(
//		name, tags, map[string]interface{}{"value": value}, timestamp,
//	)
//	if err != nil {
//		return
//	}
//	w.Points = append(w.Points, pt)
//}

const (
	writeShardRequestMessage byte = iota + 1
	writeShardResponseMessage

	executeStatementRequestMessage
	executeStatementResponseMessage

	createIteratorRequestMessage
	createIteratorResponseMessage

	iteratorCostRequestMessage
	iteratorCostResponseMessage

	fieldDimensionsRequestMessage
	fieldDimensionsResponseMessage

	mapTypeRequestMessage
	mapTypeResponseMessage

)

// WriteShardRequest represents the a request to write a slice of points to a shard
type WriteShardRequest struct {
	pb internal.WriteShardRequest
}

// WriteShardResponse represents the response returned from a remote WriteShardRequest call
type WriteShardResponse struct {
	pb internal.WriteShardResponse
}

// SetShardID sets the ShardID
func (w *WriteShardRequest) SetShardID(id uint64) { w.pb.ShardID = &id }

// ShardID gets the ShardID
func (w *WriteShardRequest) ShardID() uint64 { return w.pb.GetShardID() }

func (w *WriteShardRequest) SetDatabase(db string) { w.pb.Database = &db }

func (w *WriteShardRequest) SetRetentionPolicy(rp string) { w.pb.RetentionPolicy = &rp }

func (w *WriteShardRequest) Database() string { return w.pb.GetDatabase() }

func (w *WriteShardRequest) RetentionPolicy() string { return w.pb.GetRetentionPolicy() }

// Points returns the time series Points
func (w *WriteShardRequest) Points() []models.Point { return w.unmarshalPoints() }

// AddPoint adds a new time series point
func (w *WriteShardRequest) AddPoint(name string, value interface{}, timestamp time.Time, tags models.Tags) {
	pt, err := models.NewPoint(
		name, tags, map[string]interface{}{"value": value}, timestamp,
	)
	if err != nil {
		return
	}
	w.AddPoints([]models.Point{pt})
}

// AddPoints adds a new time series point
func (w *WriteShardRequest) AddPoints(points []models.Point) {
	for _, p := range points {
		b, err := p.MarshalBinary()
		if err != nil {
			// A error here means that we create a point higher in the stack that we could
			// not marshal to a byte slice.  If that happens, the endpoint that created that
			// point needs to be fixed.
			panic(fmt.Sprintf("failed to marshal point: `%v`: %v", p, err))
		}
		w.pb.Points = append(w.pb.Points, b)
	}
}

// MarshalBinary encodes the object to a binary format.
func (w *WriteShardRequest) MarshalBinary() ([]byte, error) {
	return proto.Marshal(&w.pb)
}

// UnmarshalBinary populates WritePointRequest from a binary format.
func (w *WriteShardRequest) UnmarshalBinary(buf []byte) error {
	if err := proto.Unmarshal(buf, &w.pb); err != nil {
		return err
	}
	return nil
}

func (w *WriteShardRequest) unmarshalPoints() []models.Point {
	points := make([]models.Point, len(w.pb.GetPoints()))
	for i, p := range w.pb.GetPoints() {
		pt, err := models.NewPointFromBytes(p)
		if err != nil {
			// A error here means that one node created a valid point and sent us an
			// unparseable version.  We could log and drop the point and allow
			// anti-entropy to resolve the discrepancy, but this shouldn't ever happen.
			panic(fmt.Sprintf("failed to parse point: `%v`: %v", string(p), err))
		}

		points[i] = pt
	}
	return points
}

// SetCode sets the Code
func (w *WriteShardResponse) SetCode(code int) { w.pb.Code = proto.Int32(int32(code)) }

// SetMessage sets the Message
func (w *WriteShardResponse) SetMessage(message string) { w.pb.Message = &message }

// Code returns the Code
func (w *WriteShardResponse) Code() int { return int(w.pb.GetCode()) }

// Message returns the Message
func (w *WriteShardResponse) Message() string { return w.pb.GetMessage() }

// MarshalBinary encodes the object to a binary format.
func (w *WriteShardResponse) MarshalBinary() ([]byte, error) {
	return proto.Marshal(&w.pb)
}

// UnmarshalBinary populates WritePointRequest from a binary format.
func (w *WriteShardResponse) UnmarshalBinary(buf []byte) error {
	if err := proto.Unmarshal(buf, &w.pb); err != nil {
		return err
	}
	return nil
}

// ExecuteStatementRequest represents the a request to execute a statement on a node.
type ExecuteStatementRequest struct {
	pb internal.ExecuteStatementRequest
}

// Statement returns the InfluxQL statement.
func (r *ExecuteStatementRequest) Statement() string { return r.pb.GetStatement() }

// SetStatement sets the InfluxQL statement.
func (r *ExecuteStatementRequest) SetStatement(statement string) {
	r.pb.Statement = proto.String(statement)
}

// Database returns the database name.
func (r *ExecuteStatementRequest) Database() string { return r.pb.GetDatabase() }

// SetDatabase sets the database name.
func (r *ExecuteStatementRequest) SetDatabase(database string) { r.pb.Database = proto.String(database) }

// MarshalBinary encodes the object to a binary format.
func (r *ExecuteStatementRequest) MarshalBinary() ([]byte, error) {
	return proto.Marshal(&r.pb)
}

// UnmarshalBinary populates ExecuteStatementRequest from a binary format.
func (r *ExecuteStatementRequest) UnmarshalBinary(buf []byte) error {
	if err := proto.Unmarshal(buf, &r.pb); err != nil {
		return err
	}
	return nil
}

// ExecuteStatementResponse represents the response returned from a remote ExecuteStatementRequest call.
type ExecuteStatementResponse struct {
	pb internal.WriteShardResponse
}

// Code returns the response code.
func (w *ExecuteStatementResponse) Code() int { return int(w.pb.GetCode()) }

// SetCode sets the Code
func (w *ExecuteStatementResponse) SetCode(code int) { w.pb.Code = proto.Int32(int32(code)) }

// Message returns the repsonse message.
func (w *ExecuteStatementResponse) Message() string { return w.pb.GetMessage() }

// SetMessage sets the Message
func (w *ExecuteStatementResponse) SetMessage(message string) { w.pb.Message = &message }

// MarshalBinary encodes the object to a binary format.
func (w *ExecuteStatementResponse) MarshalBinary() ([]byte, error) {
	return proto.Marshal(&w.pb)
}

// UnmarshalBinary populates ExecuteStatementResponse from a binary format.
func (w *ExecuteStatementResponse) UnmarshalBinary(buf []byte) error {
	if err := proto.Unmarshal(buf, &w.pb); err != nil {
		return err
	}
	return nil
}

// CreateIteratorRequest represents a request to create a remote iterator.
type CreateIteratorRequest struct {
	Measurement *influxql.Measurement
	Opt      query.IteratorOptions
	ShardIDs []uint64
}

func MarchalBinayMeasurement(mm influxql.Measurement) ([]byte,error){
	pb := &internal.Measurement{
		Database:        proto.String(mm.Database),
		RetentionPolicy: proto.String(mm.RetentionPolicy),
		Name:            proto.String(mm.Name),
		SystemIterator:  proto.String(mm.SystemIterator),
		IsTarget:        proto.Bool(mm.IsTarget),
	}
	if mm.Regex != nil {
		pb.Regex = proto.String(mm.Regex.Val.String())
	}

	return proto.Marshal(pb)
}

func UnMarchalBinayMeasurement(pb *internal.Measurement) (*influxql.Measurement, error) {
	mm := &influxql.Measurement{
		Database:        pb.GetDatabase(),
		RetentionPolicy: pb.GetRetentionPolicy(),
		Name:            pb.GetName(),
		SystemIterator:  pb.GetSystemIterator(),
		IsTarget:        pb.GetIsTarget(),
	}

	if pb.Regex != nil {
		regex, err := regexp.Compile(pb.GetRegex())
		if err != nil {
			return nil, fmt.Errorf("invalid binary measurement regex: value=%q, err=%s", pb.GetRegex(), err)
		}
		mm.Regex = &influxql.RegexLiteral{Val: regex}
	}

	return mm, nil
}

// MarshalBinary encodes r to a binary format.
func (r *CreateIteratorRequest) MarshalBinary() ([]byte, error) {
	buf, err := r.Opt.MarshalBinary()
	if err != nil {
		return nil, err
	}

	mmBuf,err := MarchalBinayMeasurement(*r.Measurement)
	if err != nil {
		return nil, err
	}
	return proto.Marshal(&internal.CreateIteratorRequest{
		Measurement: mmBuf,
		ShardIDs: r.ShardIDs,
		Opt:      buf,
	})
}

// UnmarshalBinary decodes data into r.
func (r *CreateIteratorRequest) UnmarshalBinary(data []byte) error {
	var pb internal.CreateIteratorRequest
	if err := proto.Unmarshal(data, &pb); err != nil {
		return err
	}

	r.ShardIDs = pb.GetShardIDs()
	if err := r.Opt.UnmarshalBinary(pb.GetOpt()); err != nil {
		return err
	}
	var mm internal.Measurement
	if err := proto.Unmarshal(pb.Measurement, &mm); err != nil {
		return err
	}
	r.Measurement,_ = UnMarchalBinayMeasurement(&mm)

	return nil
}

// CreateIteratorResponse represents a response from remote iterator creation.
type CreateIteratorResponse struct {
	Err error
}

// MarshalBinary encodes r to a binary format.
func (r *CreateIteratorResponse) MarshalBinary() ([]byte, error) {
	var pb internal.CreateIteratorResponse
	if r.Err != nil {
		pb.Err = proto.String(r.Err.Error())
	}
	return proto.Marshal(&pb)
}

// UnmarshalBinary decodes data into r.
func (r *CreateIteratorResponse) UnmarshalBinary(data []byte) error {
	var pb internal.CreateIteratorResponse
	if err := proto.Unmarshal(data, &pb); err != nil {
		return err
	}
	if pb.Err != nil {
		r.Err = errors.New(pb.GetErr())
	}
	return nil
}

// FieldDimensionsRequest represents a request to retrieve unique fields & dimensions.
type FieldDimensionsRequest struct {
	Measurement string
	ShardIDs []uint64
	//Sources  influxql.Sources
}

// MarshalBinary encodes r to a binary format.
func (r *FieldDimensionsRequest) MarshalBinary() ([]byte, error) {
	//buf, err := r.Sources.MarshalBinary()
	//if err != nil {
	//	return nil, err
	//}
	return proto.Marshal(&internal.FieldDimensionsRequest{
		Measurement: &r.Measurement,
		ShardIDs: r.ShardIDs,
		///Sources:  buf,
	})
}

// UnmarshalBinary decodes data into r.
func (r *FieldDimensionsRequest) UnmarshalBinary(data []byte) error {
	var pb internal.FieldDimensionsRequest
	if err := proto.Unmarshal(data, &pb); err != nil {
		return err
	}

	r.ShardIDs = pb.GetShardIDs()
	//if err := r.Sources.UnmarshalBinary(pb.GetSources()); err != nil {
	//	return err
	//}
	r.Measurement = pb.GetMeasurement()
	return nil
}

// FieldDimensionsResponse represents a response from remote iterator creation.
type FieldDimensionsResponse struct {
	Fields     map[string]influxql.DataType
	Dimensions map[string]struct{}
	Err        error
}

// MarshalBinary encodes r to a binary format.
func (r *FieldDimensionsResponse) MarshalBinary() ([]byte, error) {
	var pb internal.FieldDimensionsResponse

	pb.Fields = make([]*internal.Field, 0, len(r.Fields))
	for k,v := range r.Fields {
		iv := uint64(v)
		pb.Fields = append(pb.Fields, &internal.Field{Name:&k, Type:&iv})
	}

	pb.Dimensions = make([]string, 0, len(r.Dimensions))
	for k := range r.Dimensions {
		pb.Dimensions = append(pb.Dimensions, k)
	}

	if r.Err != nil {
		pb.Err = proto.String(r.Err.Error())
	}
	return proto.Marshal(&pb)
}

// UnmarshalBinary decodes data into r.
func (r *FieldDimensionsResponse) UnmarshalBinary(data []byte) error {
	var pb internal.FieldDimensionsResponse
	if err := proto.Unmarshal(data, &pb); err != nil {
		return err
	}

	r.Fields = make(map[string]influxql.DataType, len(pb.GetFields()))
	for _, s := range pb.GetFields() {
		r.Fields[*s.Name] = influxql.DataType(*s.Type)
	}

	r.Dimensions = make(map[string]struct{}, len(pb.GetDimensions()))
	for _, s := range pb.GetDimensions() {
		r.Dimensions[s] = struct{}{}
	}

	if pb.Err != nil {
		r.Err = errors.New(pb.GetErr())
	}
	return nil
}

type MapTypeRequest struct {
	ShardIDs []uint64
	Measurement string
	Field string
}

func (m *MapTypeRequest) MarshalBinary() ([]byte, error) {
	return proto.Marshal(&internal.MapTypeRequest{
		ShardIDs: m.ShardIDs,
		Measurement: &m.Measurement,
		Field: &m.Field,
	})
}

func (m *MapTypeRequest) UnmarshalBinary(data []byte) ( error) {

	var pb internal.MapTypeRequest
	if err := proto.Unmarshal(data, &pb); err != nil {
		return err
	}

	m.ShardIDs = pb.GetShardIDs()
	m.Measurement = pb.GetMeasurement()
	m.Field = pb.GetField()
	return nil
}

type MapTypeResponse struct {
	Type int
	Err  error
}

func (m *MapTypeResponse) MarshalBinary() ([]byte, error) {
	var pb internal.MapTypeResponse
	t := int64(m.Type)
	pb.Type = &t

	if m.Err != nil {
		pb.Err = proto.String(m.Err.Error())
	}
	return proto.Marshal(&pb)
}

func (m *MapTypeResponse) UnmarshalBinary(data []byte) (error) {
	var pb internal.MapTypeResponse

	if err := proto.Unmarshal(data, &pb); err != nil {
		return err
	}
	m.Type = int(*pb.Type)
	if pb.Err != nil {
		m.Err = errors.New(pb.GetErr())
	}
	return nil
}


