// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: ccl/utilccl/licenseccl/license.proto

package licenseccl

import (
	fmt "fmt"
	github_com_cockroachdb_cockroach_pkg_util_uuid "github.com/labulakalia/sqlfmt/cockroach/pkg/util/uuid"
	_ "github.com/gogo/protobuf/gogoproto"
	proto "github.com/gogo/protobuf/proto"
	io "io"
	math "math"
	math_bits "math/bits"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.GoGoProtoPackageIsVersion3 // please upgrade the proto package

type License_Type int32

const (
	License_NonCommercial License_Type = 0
	License_Enterprise    License_Type = 1
	License_Evaluation    License_Type = 2
)

var License_Type_name = map[int32]string{
	0: "NonCommercial",
	1: "Enterprise",
	2: "Evaluation",
}

var License_Type_value = map[string]int32{
	"NonCommercial": 0,
	"Enterprise":    1,
	"Evaluation":    2,
}

func (x License_Type) String() string {
	return proto.EnumName(License_Type_name, int32(x))
}

func (License_Type) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_765066c8b2c94c63, []int{0, 0}
}

type License struct {
	ClusterID         []github_com_cockroachdb_cockroach_pkg_util_uuid.UUID `protobuf:"bytes,1,rep,name=cluster_id,json=clusterId,proto3,customtype=sqlfmt/cockroach/pkg/util/uuid.UUID" json:"cluster_id"`
	ValidUntilUnixSec int64                                                 `protobuf:"varint,2,opt,name=valid_until_unix_sec,json=validUntilUnixSec,proto3" json:"valid_until_unix_sec,omitempty"`
	Type              License_Type                                          `protobuf:"varint,3,opt,name=type,proto3,enum=cockroach.ccl.utilccl.licenseccl.License_Type" json:"type,omitempty"`
	OrganizationName  string                                                `protobuf:"bytes,4,opt,name=organization_name,json=organizationName,proto3" json:"organization_name,omitempty"`
}

func (m *License) Reset()         { *m = License{} }
func (m *License) String() string { return proto.CompactTextString(m) }
func (*License) ProtoMessage()    {}
func (*License) Descriptor() ([]byte, []int) {
	return fileDescriptor_765066c8b2c94c63, []int{0}
}
func (m *License) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *License) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	b = b[:cap(b)]
	n, err := m.MarshalToSizedBuffer(b)
	if err != nil {
		return nil, err
	}
	return b[:n], nil
}
func (m *License) XXX_Merge(src proto.Message) {
	xxx_messageInfo_License.Merge(m, src)
}
func (m *License) XXX_Size() int {
	return m.Size()
}
func (m *License) XXX_DiscardUnknown() {
	xxx_messageInfo_License.DiscardUnknown(m)
}

var xxx_messageInfo_License proto.InternalMessageInfo

func init() {
	proto.RegisterEnum("cockroach.ccl.utilccl.licenseccl.License_Type", License_Type_name, License_Type_value)
	proto.RegisterType((*License)(nil), "cockroach.ccl.utilccl.licenseccl.License")
}

func init() {
	proto.RegisterFile("ccl/utilccl/licenseccl/license.proto", fileDescriptor_765066c8b2c94c63)
}

var fileDescriptor_765066c8b2c94c63 = []byte{
	// 362 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x84, 0x91, 0x31, 0x8f, 0xda, 0x30,
	0x1c, 0xc5, 0x63, 0x40, 0xad, 0xb0, 0x5a, 0x14, 0x22, 0x86, 0xa8, 0x83, 0x89, 0x50, 0x87, 0x48,
	0xad, 0x1c, 0xa9, 0x4c, 0x5d, 0x81, 0x4a, 0x45, 0xaa, 0x18, 0xd2, 0x66, 0xe9, 0x12, 0x19, 0xc7,
	0x0a, 0x16, 0x8e, 0x1d, 0x25, 0x36, 0x82, 0x7e, 0x8a, 0x7e, 0x2c, 0xb6, 0x63, 0x44, 0x37, 0xa0,
	0xbb, 0xf0, 0x45, 0x4e, 0x09, 0xe8, 0x60, 0xbb, 0xc9, 0xcf, 0x7a, 0xef, 0xf7, 0xf7, 0xff, 0xc9,
	0xf0, 0x33, 0xa5, 0x22, 0x30, 0x9a, 0x8b, 0xfa, 0x14, 0x9c, 0x32, 0x59, 0xb2, 0x3b, 0x89, 0xf3,
	0x42, 0x69, 0xe5, 0x78, 0x54, 0xd1, 0x75, 0xa1, 0x08, 0x5d, 0x61, 0x4a, 0x05, 0xbe, 0xe6, 0xf1,
	0x2d, 0xff, 0x69, 0x90, 0xaa, 0x54, 0x35, 0xe1, 0xa0, 0x56, 0x17, 0x6e, 0xf4, 0xd0, 0x82, 0xef,
	0x7f, 0x5d, 0x42, 0x4e, 0x0a, 0x21, 0x15, 0xa6, 0xd4, 0xac, 0x88, 0x79, 0xe2, 0x02, 0xaf, 0xed,
	0x7f, 0x98, 0xfc, 0xdc, 0x9f, 0x86, 0xd6, 0xe3, 0x69, 0x38, 0x4e, 0xb9, 0x5e, 0x99, 0x25, 0xa6,
	0x2a, 0x0b, 0x5e, 0x9f, 0x4a, 0x96, 0x37, 0x1d, 0xe4, 0xeb, 0xb4, 0x59, 0x33, 0x30, 0x86, 0x27,
	0x38, 0x8a, 0xe6, 0xb3, 0xea, 0x34, 0xec, 0x4e, 0x2f, 0x03, 0xe7, 0xb3, 0xb0, 0x7b, 0x9d, 0x3d,
	0x4f, 0x9c, 0x00, 0x0e, 0x36, 0x44, 0xf0, 0x24, 0x36, 0x52, 0x73, 0x11, 0x1b, 0xc9, 0xb7, 0x71,
	0xc9, 0xa8, 0xdb, 0xf2, 0x80, 0xdf, 0x0e, 0xfb, 0x8d, 0x17, 0xd5, 0x56, 0x24, 0xf9, 0xf6, 0x37,
	0xa3, 0xce, 0x04, 0x76, 0xf4, 0x2e, 0x67, 0x6e, 0xdb, 0x03, 0x7e, 0xef, 0x1b, 0xc6, 0x6f, 0x95,
	0xc5, 0xd7, 0x4a, 0xf8, 0xcf, 0x2e, 0x67, 0x61, 0xc3, 0x3a, 0x5f, 0x60, 0x5f, 0x15, 0x29, 0x91,
	0xfc, 0x1f, 0xd1, 0x5c, 0xc9, 0x58, 0x92, 0x8c, 0xb9, 0x1d, 0x0f, 0xf8, 0xdd, 0xd0, 0xbe, 0x37,
	0x16, 0x24, 0x63, 0xa3, 0xef, 0xb0, 0x53, 0xa3, 0x4e, 0x1f, 0x7e, 0x5c, 0x28, 0x39, 0x55, 0x59,
	0xc6, 0x0a, 0xca, 0x89, 0xb0, 0x2d, 0xa7, 0x07, 0xe1, 0x0f, 0xa9, 0x59, 0x91, 0x17, 0xbc, 0x64,
	0x36, 0x68, 0xee, 0x1b, 0x22, 0x4c, 0x03, 0xdb, 0xad, 0xc9, 0xd7, 0xfd, 0x33, 0xb2, 0xf6, 0x15,
	0x02, 0x87, 0x0a, 0x81, 0x63, 0x85, 0xc0, 0x53, 0x85, 0xc0, 0xff, 0x33, 0xb2, 0x0e, 0x67, 0x64,
	0x1d, 0xcf, 0xc8, 0xfa, 0x0b, 0x6f, 0x8b, 0x2e, 0xdf, 0x35, 0xdf, 0x30, 0x7e, 0x09, 0x00, 0x00,
	0xff, 0xff, 0x36, 0x56, 0x7c, 0xf6, 0xe6, 0x01, 0x00, 0x00,
}

func (m *License) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *License) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *License) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.OrganizationName) > 0 {
		i -= len(m.OrganizationName)
		copy(dAtA[i:], m.OrganizationName)
		i = encodeVarintLicense(dAtA, i, uint64(len(m.OrganizationName)))
		i--
		dAtA[i] = 0x22
	}
	if m.Type != 0 {
		i = encodeVarintLicense(dAtA, i, uint64(m.Type))
		i--
		dAtA[i] = 0x18
	}
	if m.ValidUntilUnixSec != 0 {
		i = encodeVarintLicense(dAtA, i, uint64(m.ValidUntilUnixSec))
		i--
		dAtA[i] = 0x10
	}
	if len(m.ClusterID) > 0 {
		for iNdEx := len(m.ClusterID) - 1; iNdEx >= 0; iNdEx-- {
			{
				size := m.ClusterID[iNdEx].Size()
				i -= size
				if _, err := m.ClusterID[iNdEx].MarshalTo(dAtA[i:]); err != nil {
					return 0, err
				}
				i = encodeVarintLicense(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0xa
		}
	}
	return len(dAtA) - i, nil
}

func encodeVarintLicense(dAtA []byte, offset int, v uint64) int {
	offset -= sovLicense(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *License) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if len(m.ClusterID) > 0 {
		for _, e := range m.ClusterID {
			l = e.Size()
			n += 1 + l + sovLicense(uint64(l))
		}
	}
	if m.ValidUntilUnixSec != 0 {
		n += 1 + sovLicense(uint64(m.ValidUntilUnixSec))
	}
	if m.Type != 0 {
		n += 1 + sovLicense(uint64(m.Type))
	}
	l = len(m.OrganizationName)
	if l > 0 {
		n += 1 + l + sovLicense(uint64(l))
	}
	return n
}

func sovLicense(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozLicense(x uint64) (n int) {
	return sovLicense(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *License) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowLicense
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: License: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: License: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ClusterID", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowLicense
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				byteLen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if byteLen < 0 {
				return ErrInvalidLengthLicense
			}
			postIndex := iNdEx + byteLen
			if postIndex < 0 {
				return ErrInvalidLengthLicense
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			var v github_com_cockroachdb_cockroach_pkg_util_uuid.UUID
			m.ClusterID = append(m.ClusterID, v)
			if err := m.ClusterID[len(m.ClusterID)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 2:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field ValidUntilUnixSec", wireType)
			}
			m.ValidUntilUnixSec = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowLicense
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.ValidUntilUnixSec |= int64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 3:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Type", wireType)
			}
			m.Type = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowLicense
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Type |= License_Type(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field OrganizationName", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowLicense
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthLicense
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthLicense
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.OrganizationName = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipLicense(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthLicense
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func skipLicense(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowLicense
			}
			if iNdEx >= l {
				return 0, io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		wireType := int(wire & 0x7)
		switch wireType {
		case 0:
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowLicense
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				iNdEx++
				if dAtA[iNdEx-1] < 0x80 {
					break
				}
			}
		case 1:
			iNdEx += 8
		case 2:
			var length int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowLicense
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				length |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if length < 0 {
				return 0, ErrInvalidLengthLicense
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupLicense
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthLicense
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthLicense        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowLicense          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupLicense = fmt.Errorf("proto: unexpected end of group")
)
