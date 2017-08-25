// Code generated by protoc-gen-go.
// source: mapnificent.proto
// DO NOT EDIT!

/*
Package mapnificent is a generated protocol buffer package.

It is generated from these files:
	mapnificent.proto

It has these top-level messages:
	MapnificentNetwork
*/
package mapnificent

import proto "code.google.com/p/goprotobuf/proto"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = math.Inf

type MapnificentNetwork struct {
	Cityid           *string                    `protobuf:"bytes,1,opt" json:"Cityid,omitempty"`
	Stops            []*MapnificentNetwork_Stop `protobuf:"bytes,2,rep" json:"Stops,omitempty"`
	Lines            []*MapnificentNetwork_Line `protobuf:"bytes,3,rep" json:"Lines,omitempty"`
	XXX_unrecognized []byte                     `json:"-"`
}

func (m *MapnificentNetwork) Reset()         { *m = MapnificentNetwork{} }
func (m *MapnificentNetwork) String() string { return proto.CompactTextString(m) }
func (*MapnificentNetwork) ProtoMessage()    {}

func (m *MapnificentNetwork) GetCityid() string {
	if m != nil && m.Cityid != nil {
		return *m.Cityid
	}
	return ""
}

func (m *MapnificentNetwork) GetStops() []*MapnificentNetwork_Stop {
	if m != nil {
		return m.Stops
	}
	return nil
}

func (m *MapnificentNetwork) GetLines() []*MapnificentNetwork_Line {
	if m != nil {
		return m.Lines
	}
	return nil
}

type MapnificentNetwork_Stop struct {
	Latitude         *float64                                `protobuf:"fixed64,1,opt" json:"Latitude,omitempty"`
	Longitude        *float64                                `protobuf:"fixed64,2,opt" json:"Longitude,omitempty"`
	TravelOptions    []*MapnificentNetwork_Stop_TravelOption `protobuf:"bytes,3,rep" json:"TravelOptions,omitempty"`
	Name             *string                                 `protobuf:"bytes,4,opt" json:"Name,omitempty"`
	XXX_unrecognized []byte                                  `json:"-"`
}

func (m *MapnificentNetwork_Stop) Reset()         { *m = MapnificentNetwork_Stop{} }
func (m *MapnificentNetwork_Stop) String() string { return proto.CompactTextString(m) }
func (*MapnificentNetwork_Stop) ProtoMessage()    {}

func (m *MapnificentNetwork_Stop) GetLatitude() float64 {
	if m != nil && m.Latitude != nil {
		return *m.Latitude
	}
	return 0
}

func (m *MapnificentNetwork_Stop) GetLongitude() float64 {
	if m != nil && m.Longitude != nil {
		return *m.Longitude
	}
	return 0
}

func (m *MapnificentNetwork_Stop) GetTravelOptions() []*MapnificentNetwork_Stop_TravelOption {
	if m != nil {
		return m.TravelOptions
	}
	return nil
}

func (m *MapnificentNetwork_Stop) GetName() string {
	if m != nil && m.Name != nil {
		return *m.Name
	}
	return ""
}

type MapnificentNetwork_Stop_TravelOption struct {
	Stop             *int32  `protobuf:"varint,1,opt" json:"Stop,omitempty"`
	TravelTime       *int32  `protobuf:"varint,2,opt" json:"TravelTime,omitempty"`
	StayTime         *int32  `protobuf:"varint,3,opt" json:"StayTime,omitempty"`
	Line             *string `protobuf:"bytes,4,opt" json:"Line,omitempty"`
	WalkDistance     *int32  `protobuf:"varint,5,opt" json:"WalkDistance,omitempty"`
	XXX_unrecognized []byte  `json:"-"`
}

func (m *MapnificentNetwork_Stop_TravelOption) Reset()         { *m = MapnificentNetwork_Stop_TravelOption{} }
func (m *MapnificentNetwork_Stop_TravelOption) String() string { return proto.CompactTextString(m) }
func (*MapnificentNetwork_Stop_TravelOption) ProtoMessage()    {}

func (m *MapnificentNetwork_Stop_TravelOption) GetStop() int32 {
	if m != nil && m.Stop != nil {
		return *m.Stop
	}
	return 0
}

func (m *MapnificentNetwork_Stop_TravelOption) GetTravelTime() int32 {
	if m != nil && m.TravelTime != nil {
		return *m.TravelTime
	}
	return 0
}

func (m *MapnificentNetwork_Stop_TravelOption) GetStayTime() int32 {
	if m != nil && m.StayTime != nil {
		return *m.StayTime
	}
	return 0
}

func (m *MapnificentNetwork_Stop_TravelOption) GetLine() string {
	if m != nil && m.Line != nil {
		return *m.Line
	}
	return ""
}

func (m *MapnificentNetwork_Stop_TravelOption) GetWalkDistance() int32 {
	if m != nil && m.WalkDistance != nil {
		return *m.WalkDistance
	}
	return 0
}

type MapnificentNetwork_Line struct {
	LineId           *string                             `protobuf:"bytes,1,opt" json:"LineId,omitempty"`
	LineTimes        []*MapnificentNetwork_Line_LineTime `protobuf:"bytes,2,rep" json:"LineTimes,omitempty"`
	Name             *string                             `protobuf:"bytes,3,opt" json:"Name,omitempty"`
	XXX_unrecognized []byte                              `json:"-"`
}

func (m *MapnificentNetwork_Line) Reset()         { *m = MapnificentNetwork_Line{} }
func (m *MapnificentNetwork_Line) String() string { return proto.CompactTextString(m) }
func (*MapnificentNetwork_Line) ProtoMessage()    {}

func (m *MapnificentNetwork_Line) GetLineId() string {
	if m != nil && m.LineId != nil {
		return *m.LineId
	}
	return ""
}

func (m *MapnificentNetwork_Line) GetLineTimes() []*MapnificentNetwork_Line_LineTime {
	if m != nil {
		return m.LineTimes
	}
	return nil
}

func (m *MapnificentNetwork_Line) GetName() string {
	if m != nil && m.Name != nil {
		return *m.Name
	}
	return ""
}

type MapnificentNetwork_Line_LineTime struct {
	Interval         *int32 `protobuf:"varint,1,opt" json:"Interval,omitempty"`
	Start            *int32 `protobuf:"varint,2,opt" json:"Start,omitempty"`
	Stop             *int32 `protobuf:"varint,3,opt" json:"Stop,omitempty"`
	Weekday          *int32 `protobuf:"varint,4,opt" json:"Weekday,omitempty"`
	XXX_unrecognized []byte `json:"-"`
}

func (m *MapnificentNetwork_Line_LineTime) Reset()         { *m = MapnificentNetwork_Line_LineTime{} }
func (m *MapnificentNetwork_Line_LineTime) String() string { return proto.CompactTextString(m) }
func (*MapnificentNetwork_Line_LineTime) ProtoMessage()    {}

func (m *MapnificentNetwork_Line_LineTime) GetInterval() int32 {
	if m != nil && m.Interval != nil {
		return *m.Interval
	}
	return 0
}

func (m *MapnificentNetwork_Line_LineTime) GetStart() int32 {
	if m != nil && m.Start != nil {
		return *m.Start
	}
	return 0
}

func (m *MapnificentNetwork_Line_LineTime) GetStop() int32 {
	if m != nil && m.Stop != nil {
		return *m.Stop
	}
	return 0
}

func (m *MapnificentNetwork_Line_LineTime) GetWeekday() int32 {
	if m != nil && m.Weekday != nil {
		return *m.Weekday
	}
	return 0
}

func init() {
}
