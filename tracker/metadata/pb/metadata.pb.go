// Code generated by protoc-gen-go. DO NOT EDIT.
// source: metadata.proto

/*
Package metadata_pb is a generated protocol buffer package.

It is generated from these files:
	metadata.proto

It has these top-level messages:
	CheckFileExistReq
	CheckFileExistResp
	ReplicaProvider
	UploadFilePrepareReq
	PieceHashAndSize
	UploadFilePrepareResp
	ErasureCodeProvider
	PieceHashAuth
	UploadFileDoneReq
	Partition
	Block
	UploadFileDoneResp
	ListFilesReq
	ListFilesResp
	FileOrFolder
	RetrieveFileReq
	RetrieveFileResp
*/
package metadata_pb

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

import (
	context "golang.org/x/net/context"
	grpc "google.golang.org/grpc"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

type FileStoreType int32

const (
	FileStoreType_ErasureCode  FileStoreType = 0
	FileStoreType_MultiReplica FileStoreType = 1
)

var FileStoreType_name = map[int32]string{
	0: "ErasureCode",
	1: "MultiReplica",
}
var FileStoreType_value = map[string]int32{
	"ErasureCode":  0,
	"MultiReplica": 1,
}

func (x FileStoreType) String() string {
	return proto.EnumName(FileStoreType_name, int32(x))
}
func (FileStoreType) EnumDescriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

type CheckFileExistReq struct {
	Version     uint32 `protobuf:"varint,1,opt,name=version" json:"version,omitempty"`
	NodeId      []byte `protobuf:"bytes,2,opt,name=nodeId,proto3" json:"nodeId,omitempty"`
	Timestamp   uint64 `protobuf:"varint,3,opt,name=timestamp" json:"timestamp,omitempty"`
	FilePath    string `protobuf:"bytes,4,opt,name=filePath" json:"filePath,omitempty"`
	FileHash    []byte `protobuf:"bytes,5,opt,name=fileHash,proto3" json:"fileHash,omitempty"`
	FileSize    uint64 `protobuf:"varint,6,opt,name=fileSize" json:"fileSize,omitempty"`
	FileName    string `protobuf:"bytes,7,opt,name=fileName" json:"fileName,omitempty"`
	FileModTime uint64 `protobuf:"varint,8,opt,name=fileModTime" json:"fileModTime,omitempty"`
	FileData    []byte `protobuf:"bytes,9,opt,name=fileData,proto3" json:"fileData,omitempty"`
	Interactive bool   `protobuf:"varint,10,opt,name=interactive" json:"interactive,omitempty"`
	Sign        []byte `protobuf:"bytes,11,opt,name=sign,proto3" json:"sign,omitempty"`
}

func (m *CheckFileExistReq) Reset()                    { *m = CheckFileExistReq{} }
func (m *CheckFileExistReq) String() string            { return proto.CompactTextString(m) }
func (*CheckFileExistReq) ProtoMessage()               {}
func (*CheckFileExistReq) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

func (m *CheckFileExistReq) GetVersion() uint32 {
	if m != nil {
		return m.Version
	}
	return 0
}

func (m *CheckFileExistReq) GetNodeId() []byte {
	if m != nil {
		return m.NodeId
	}
	return nil
}

func (m *CheckFileExistReq) GetTimestamp() uint64 {
	if m != nil {
		return m.Timestamp
	}
	return 0
}

func (m *CheckFileExistReq) GetFilePath() string {
	if m != nil {
		return m.FilePath
	}
	return ""
}

func (m *CheckFileExistReq) GetFileHash() []byte {
	if m != nil {
		return m.FileHash
	}
	return nil
}

func (m *CheckFileExistReq) GetFileSize() uint64 {
	if m != nil {
		return m.FileSize
	}
	return 0
}

func (m *CheckFileExistReq) GetFileName() string {
	if m != nil {
		return m.FileName
	}
	return ""
}

func (m *CheckFileExistReq) GetFileModTime() uint64 {
	if m != nil {
		return m.FileModTime
	}
	return 0
}

func (m *CheckFileExistReq) GetFileData() []byte {
	if m != nil {
		return m.FileData
	}
	return nil
}

func (m *CheckFileExistReq) GetInteractive() bool {
	if m != nil {
		return m.Interactive
	}
	return false
}

func (m *CheckFileExistReq) GetSign() []byte {
	if m != nil {
		return m.Sign
	}
	return nil
}

type CheckFileExistResp struct {
	Code             uint32             `protobuf:"varint,1,opt,name=code" json:"code,omitempty"`
	ErrMsg           string             `protobuf:"bytes,2,opt,name=errMsg" json:"errMsg,omitempty"`
	StoreType        FileStoreType      `protobuf:"varint,3,opt,name=storeType,enum=metadata.pb.FileStoreType" json:"storeType,omitempty"`
	DataPieceCount   int32              `protobuf:"varint,4,opt,name=dataPieceCount" json:"dataPieceCount,omitempty"`
	VerifyPieceCount int32              `protobuf:"varint,5,opt,name=verifyPieceCount" json:"verifyPieceCount,omitempty"`
	ReplicaCount     int32              `protobuf:"varint,6,opt,name=replicaCount" json:"replicaCount,omitempty"`
	Provider         []*ReplicaProvider `protobuf:"bytes,7,rep,name=provider" json:"provider,omitempty"`
}

func (m *CheckFileExistResp) Reset()                    { *m = CheckFileExistResp{} }
func (m *CheckFileExistResp) String() string            { return proto.CompactTextString(m) }
func (*CheckFileExistResp) ProtoMessage()               {}
func (*CheckFileExistResp) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

func (m *CheckFileExistResp) GetCode() uint32 {
	if m != nil {
		return m.Code
	}
	return 0
}

func (m *CheckFileExistResp) GetErrMsg() string {
	if m != nil {
		return m.ErrMsg
	}
	return ""
}

func (m *CheckFileExistResp) GetStoreType() FileStoreType {
	if m != nil {
		return m.StoreType
	}
	return FileStoreType_ErasureCode
}

func (m *CheckFileExistResp) GetDataPieceCount() int32 {
	if m != nil {
		return m.DataPieceCount
	}
	return 0
}

func (m *CheckFileExistResp) GetVerifyPieceCount() int32 {
	if m != nil {
		return m.VerifyPieceCount
	}
	return 0
}

func (m *CheckFileExistResp) GetReplicaCount() int32 {
	if m != nil {
		return m.ReplicaCount
	}
	return 0
}

func (m *CheckFileExistResp) GetProvider() []*ReplicaProvider {
	if m != nil {
		return m.Provider
	}
	return nil
}

type ReplicaProvider struct {
	NodeId    []byte `protobuf:"bytes,1,opt,name=nodeId,proto3" json:"nodeId,omitempty"`
	Server    string `protobuf:"bytes,2,opt,name=server" json:"server,omitempty"`
	Port      uint32 `protobuf:"varint,3,opt,name=port" json:"port,omitempty"`
	Timestamp uint64 `protobuf:"varint,4,opt,name=timestamp" json:"timestamp,omitempty"`
	Ticket    string `protobuf:"bytes,5,opt,name=ticket" json:"ticket,omitempty"`
	Auth      []byte `protobuf:"bytes,6,opt,name=auth,proto3" json:"auth,omitempty"`
}

func (m *ReplicaProvider) Reset()                    { *m = ReplicaProvider{} }
func (m *ReplicaProvider) String() string            { return proto.CompactTextString(m) }
func (*ReplicaProvider) ProtoMessage()               {}
func (*ReplicaProvider) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{2} }

func (m *ReplicaProvider) GetNodeId() []byte {
	if m != nil {
		return m.NodeId
	}
	return nil
}

func (m *ReplicaProvider) GetServer() string {
	if m != nil {
		return m.Server
	}
	return ""
}

func (m *ReplicaProvider) GetPort() uint32 {
	if m != nil {
		return m.Port
	}
	return 0
}

func (m *ReplicaProvider) GetTimestamp() uint64 {
	if m != nil {
		return m.Timestamp
	}
	return 0
}

func (m *ReplicaProvider) GetTicket() string {
	if m != nil {
		return m.Ticket
	}
	return ""
}

func (m *ReplicaProvider) GetAuth() []byte {
	if m != nil {
		return m.Auth
	}
	return nil
}

type UploadFilePrepareReq struct {
	Version   uint32              `protobuf:"varint,1,opt,name=version" json:"version,omitempty"`
	NodeId    []byte              `protobuf:"bytes,2,opt,name=nodeId,proto3" json:"nodeId,omitempty"`
	Timestamp uint64              `protobuf:"varint,3,opt,name=timestamp" json:"timestamp,omitempty"`
	FileHash  []byte              `protobuf:"bytes,4,opt,name=fileHash,proto3" json:"fileHash,omitempty"`
	FileSize  uint64              `protobuf:"varint,5,opt,name=fileSize" json:"fileSize,omitempty"`
	Piece     []*PieceHashAndSize `protobuf:"bytes,6,rep,name=piece" json:"piece,omitempty"`
	Sign      []byte              `protobuf:"bytes,7,opt,name=sign,proto3" json:"sign,omitempty"`
}

func (m *UploadFilePrepareReq) Reset()                    { *m = UploadFilePrepareReq{} }
func (m *UploadFilePrepareReq) String() string            { return proto.CompactTextString(m) }
func (*UploadFilePrepareReq) ProtoMessage()               {}
func (*UploadFilePrepareReq) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{3} }

func (m *UploadFilePrepareReq) GetVersion() uint32 {
	if m != nil {
		return m.Version
	}
	return 0
}

func (m *UploadFilePrepareReq) GetNodeId() []byte {
	if m != nil {
		return m.NodeId
	}
	return nil
}

func (m *UploadFilePrepareReq) GetTimestamp() uint64 {
	if m != nil {
		return m.Timestamp
	}
	return 0
}

func (m *UploadFilePrepareReq) GetFileHash() []byte {
	if m != nil {
		return m.FileHash
	}
	return nil
}

func (m *UploadFilePrepareReq) GetFileSize() uint64 {
	if m != nil {
		return m.FileSize
	}
	return 0
}

func (m *UploadFilePrepareReq) GetPiece() []*PieceHashAndSize {
	if m != nil {
		return m.Piece
	}
	return nil
}

func (m *UploadFilePrepareReq) GetSign() []byte {
	if m != nil {
		return m.Sign
	}
	return nil
}

type PieceHashAndSize struct {
	Hash []byte `protobuf:"bytes,1,opt,name=hash,proto3" json:"hash,omitempty"`
	Size uint32 `protobuf:"varint,2,opt,name=size" json:"size,omitempty"`
}

func (m *PieceHashAndSize) Reset()                    { *m = PieceHashAndSize{} }
func (m *PieceHashAndSize) String() string            { return proto.CompactTextString(m) }
func (*PieceHashAndSize) ProtoMessage()               {}
func (*PieceHashAndSize) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{4} }

func (m *PieceHashAndSize) GetHash() []byte {
	if m != nil {
		return m.Hash
	}
	return nil
}

func (m *PieceHashAndSize) GetSize() uint32 {
	if m != nil {
		return m.Size
	}
	return 0
}

type UploadFilePrepareResp struct {
	Provider []*ErasureCodeProvider `protobuf:"bytes,1,rep,name=provider" json:"provider,omitempty"`
}

func (m *UploadFilePrepareResp) Reset()                    { *m = UploadFilePrepareResp{} }
func (m *UploadFilePrepareResp) String() string            { return proto.CompactTextString(m) }
func (*UploadFilePrepareResp) ProtoMessage()               {}
func (*UploadFilePrepareResp) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{5} }

func (m *UploadFilePrepareResp) GetProvider() []*ErasureCodeProvider {
	if m != nil {
		return m.Provider
	}
	return nil
}

type ErasureCodeProvider struct {
	NodeId    []byte           `protobuf:"bytes,1,opt,name=nodeId,proto3" json:"nodeId,omitempty"`
	Server    string           `protobuf:"bytes,2,opt,name=server" json:"server,omitempty"`
	Port      uint32           `protobuf:"varint,3,opt,name=port" json:"port,omitempty"`
	Timestamp uint64           `protobuf:"varint,4,opt,name=timestamp" json:"timestamp,omitempty"`
	HashAuth  []*PieceHashAuth `protobuf:"bytes,5,rep,name=hashAuth" json:"hashAuth,omitempty"`
}

func (m *ErasureCodeProvider) Reset()                    { *m = ErasureCodeProvider{} }
func (m *ErasureCodeProvider) String() string            { return proto.CompactTextString(m) }
func (*ErasureCodeProvider) ProtoMessage()               {}
func (*ErasureCodeProvider) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{6} }

func (m *ErasureCodeProvider) GetNodeId() []byte {
	if m != nil {
		return m.NodeId
	}
	return nil
}

func (m *ErasureCodeProvider) GetServer() string {
	if m != nil {
		return m.Server
	}
	return ""
}

func (m *ErasureCodeProvider) GetPort() uint32 {
	if m != nil {
		return m.Port
	}
	return 0
}

func (m *ErasureCodeProvider) GetTimestamp() uint64 {
	if m != nil {
		return m.Timestamp
	}
	return 0
}

func (m *ErasureCodeProvider) GetHashAuth() []*PieceHashAuth {
	if m != nil {
		return m.HashAuth
	}
	return nil
}

type PieceHashAuth struct {
	Hash   []byte `protobuf:"bytes,1,opt,name=hash,proto3" json:"hash,omitempty"`
	Ticket string `protobuf:"bytes,2,opt,name=ticket" json:"ticket,omitempty"`
	Auth   []byte `protobuf:"bytes,3,opt,name=auth,proto3" json:"auth,omitempty"`
}

func (m *PieceHashAuth) Reset()                    { *m = PieceHashAuth{} }
func (m *PieceHashAuth) String() string            { return proto.CompactTextString(m) }
func (*PieceHashAuth) ProtoMessage()               {}
func (*PieceHashAuth) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{7} }

func (m *PieceHashAuth) GetHash() []byte {
	if m != nil {
		return m.Hash
	}
	return nil
}

func (m *PieceHashAuth) GetTicket() string {
	if m != nil {
		return m.Ticket
	}
	return ""
}

func (m *PieceHashAuth) GetAuth() []byte {
	if m != nil {
		return m.Auth
	}
	return nil
}

type UploadFileDoneReq struct {
	Version   uint32       `protobuf:"varint,1,opt,name=version" json:"version,omitempty"`
	NodeId    []byte       `protobuf:"bytes,2,opt,name=nodeId,proto3" json:"nodeId,omitempty"`
	Timestamp uint64       `protobuf:"varint,3,opt,name=timestamp" json:"timestamp,omitempty"`
	FileHash  []byte       `protobuf:"bytes,4,opt,name=fileHash,proto3" json:"fileHash,omitempty"`
	FileSize  uint64       `protobuf:"varint,5,opt,name=fileSize" json:"fileSize,omitempty"`
	Partition []*Partition `protobuf:"bytes,6,rep,name=partition" json:"partition,omitempty"`
	Sign      []byte       `protobuf:"bytes,7,opt,name=sign,proto3" json:"sign,omitempty"`
}

func (m *UploadFileDoneReq) Reset()                    { *m = UploadFileDoneReq{} }
func (m *UploadFileDoneReq) String() string            { return proto.CompactTextString(m) }
func (*UploadFileDoneReq) ProtoMessage()               {}
func (*UploadFileDoneReq) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{8} }

func (m *UploadFileDoneReq) GetVersion() uint32 {
	if m != nil {
		return m.Version
	}
	return 0
}

func (m *UploadFileDoneReq) GetNodeId() []byte {
	if m != nil {
		return m.NodeId
	}
	return nil
}

func (m *UploadFileDoneReq) GetTimestamp() uint64 {
	if m != nil {
		return m.Timestamp
	}
	return 0
}

func (m *UploadFileDoneReq) GetFileHash() []byte {
	if m != nil {
		return m.FileHash
	}
	return nil
}

func (m *UploadFileDoneReq) GetFileSize() uint64 {
	if m != nil {
		return m.FileSize
	}
	return 0
}

func (m *UploadFileDoneReq) GetPartition() []*Partition {
	if m != nil {
		return m.Partition
	}
	return nil
}

func (m *UploadFileDoneReq) GetSign() []byte {
	if m != nil {
		return m.Sign
	}
	return nil
}

type Partition struct {
	Block []*Block `protobuf:"bytes,1,rep,name=block" json:"block,omitempty"`
}

func (m *Partition) Reset()                    { *m = Partition{} }
func (m *Partition) String() string            { return proto.CompactTextString(m) }
func (*Partition) ProtoMessage()               {}
func (*Partition) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{9} }

func (m *Partition) GetBlock() []*Block {
	if m != nil {
		return m.Block
	}
	return nil
}

type Block struct {
	Hash        []byte   `protobuf:"bytes,1,opt,name=hash,proto3" json:"hash,omitempty"`
	Size        uint32   `protobuf:"varint,2,opt,name=size" json:"size,omitempty"`
	BlockSeq    uint32   `protobuf:"varint,3,opt,name=blockSeq" json:"blockSeq,omitempty"`
	Checksum    bool     `protobuf:"varint,4,opt,name=checksum" json:"checksum,omitempty"`
	StoreNodeId [][]byte `protobuf:"bytes,5,rep,name=storeNodeId,proto3" json:"storeNodeId,omitempty"`
}

func (m *Block) Reset()                    { *m = Block{} }
func (m *Block) String() string            { return proto.CompactTextString(m) }
func (*Block) ProtoMessage()               {}
func (*Block) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{10} }

func (m *Block) GetHash() []byte {
	if m != nil {
		return m.Hash
	}
	return nil
}

func (m *Block) GetSize() uint32 {
	if m != nil {
		return m.Size
	}
	return 0
}

func (m *Block) GetBlockSeq() uint32 {
	if m != nil {
		return m.BlockSeq
	}
	return 0
}

func (m *Block) GetChecksum() bool {
	if m != nil {
		return m.Checksum
	}
	return false
}

func (m *Block) GetStoreNodeId() [][]byte {
	if m != nil {
		return m.StoreNodeId
	}
	return nil
}

type UploadFileDoneResp struct {
	Done bool `protobuf:"varint,1,opt,name=done" json:"done,omitempty"`
}

func (m *UploadFileDoneResp) Reset()                    { *m = UploadFileDoneResp{} }
func (m *UploadFileDoneResp) String() string            { return proto.CompactTextString(m) }
func (*UploadFileDoneResp) ProtoMessage()               {}
func (*UploadFileDoneResp) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{11} }

func (m *UploadFileDoneResp) GetDone() bool {
	if m != nil {
		return m.Done
	}
	return false
}

type ListFilesReq struct {
	Version   uint32 `protobuf:"varint,1,opt,name=version" json:"version,omitempty"`
	NodeId    []byte `protobuf:"bytes,2,opt,name=nodeId,proto3" json:"nodeId,omitempty"`
	Timestamp uint64 `protobuf:"varint,3,opt,name=timestamp" json:"timestamp,omitempty"`
	Path      string `protobuf:"bytes,4,opt,name=path" json:"path,omitempty"`
	Sign      []byte `protobuf:"bytes,5,opt,name=sign,proto3" json:"sign,omitempty"`
}

func (m *ListFilesReq) Reset()                    { *m = ListFilesReq{} }
func (m *ListFilesReq) String() string            { return proto.CompactTextString(m) }
func (*ListFilesReq) ProtoMessage()               {}
func (*ListFilesReq) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{12} }

func (m *ListFilesReq) GetVersion() uint32 {
	if m != nil {
		return m.Version
	}
	return 0
}

func (m *ListFilesReq) GetNodeId() []byte {
	if m != nil {
		return m.NodeId
	}
	return nil
}

func (m *ListFilesReq) GetTimestamp() uint64 {
	if m != nil {
		return m.Timestamp
	}
	return 0
}

func (m *ListFilesReq) GetPath() string {
	if m != nil {
		return m.Path
	}
	return ""
}

func (m *ListFilesReq) GetSign() []byte {
	if m != nil {
		return m.Sign
	}
	return nil
}

type ListFilesResp struct {
	Fof []*FileOrFolder `protobuf:"bytes,1,rep,name=fof" json:"fof,omitempty"`
}

func (m *ListFilesResp) Reset()                    { *m = ListFilesResp{} }
func (m *ListFilesResp) String() string            { return proto.CompactTextString(m) }
func (*ListFilesResp) ProtoMessage()               {}
func (*ListFilesResp) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{13} }

func (m *ListFilesResp) GetFof() []*FileOrFolder {
	if m != nil {
		return m.Fof
	}
	return nil
}

type FileOrFolder struct {
	Folder   bool   `protobuf:"varint,1,opt,name=folder" json:"folder,omitempty"`
	Name     string `protobuf:"bytes,2,opt,name=name" json:"name,omitempty"`
	ModTime  uint64 `protobuf:"varint,3,opt,name=modTime" json:"modTime,omitempty"`
	FileHash []byte `protobuf:"bytes,4,opt,name=fileHash,proto3" json:"fileHash,omitempty"`
	FileSize uint64 `protobuf:"varint,5,opt,name=fileSize" json:"fileSize,omitempty"`
}

func (m *FileOrFolder) Reset()                    { *m = FileOrFolder{} }
func (m *FileOrFolder) String() string            { return proto.CompactTextString(m) }
func (*FileOrFolder) ProtoMessage()               {}
func (*FileOrFolder) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{14} }

func (m *FileOrFolder) GetFolder() bool {
	if m != nil {
		return m.Folder
	}
	return false
}

func (m *FileOrFolder) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *FileOrFolder) GetModTime() uint64 {
	if m != nil {
		return m.ModTime
	}
	return 0
}

func (m *FileOrFolder) GetFileHash() []byte {
	if m != nil {
		return m.FileHash
	}
	return nil
}

func (m *FileOrFolder) GetFileSize() uint64 {
	if m != nil {
		return m.FileSize
	}
	return 0
}

type RetrieveFileReq struct {
	Version   uint32 `protobuf:"varint,1,opt,name=version" json:"version,omitempty"`
	NodeId    []byte `protobuf:"bytes,2,opt,name=nodeId,proto3" json:"nodeId,omitempty"`
	Timestamp uint64 `protobuf:"varint,3,opt,name=timestamp" json:"timestamp,omitempty"`
	FileHash  []byte `protobuf:"bytes,4,opt,name=fileHash,proto3" json:"fileHash,omitempty"`
	FileSize  uint64 `protobuf:"varint,5,opt,name=fileSize" json:"fileSize,omitempty"`
	Sign      []byte `protobuf:"bytes,6,opt,name=sign,proto3" json:"sign,omitempty"`
}

func (m *RetrieveFileReq) Reset()                    { *m = RetrieveFileReq{} }
func (m *RetrieveFileReq) String() string            { return proto.CompactTextString(m) }
func (*RetrieveFileReq) ProtoMessage()               {}
func (*RetrieveFileReq) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{15} }

func (m *RetrieveFileReq) GetVersion() uint32 {
	if m != nil {
		return m.Version
	}
	return 0
}

func (m *RetrieveFileReq) GetNodeId() []byte {
	if m != nil {
		return m.NodeId
	}
	return nil
}

func (m *RetrieveFileReq) GetTimestamp() uint64 {
	if m != nil {
		return m.Timestamp
	}
	return 0
}

func (m *RetrieveFileReq) GetFileHash() []byte {
	if m != nil {
		return m.FileHash
	}
	return nil
}

func (m *RetrieveFileReq) GetFileSize() uint64 {
	if m != nil {
		return m.FileSize
	}
	return 0
}

func (m *RetrieveFileReq) GetSign() []byte {
	if m != nil {
		return m.Sign
	}
	return nil
}

type RetrieveFileResp struct {
	FileData  []byte       `protobuf:"bytes,1,opt,name=fileData,proto3" json:"fileData,omitempty"`
	Partition []*Partition `protobuf:"bytes,6,rep,name=partition" json:"partition,omitempty"`
}

func (m *RetrieveFileResp) Reset()                    { *m = RetrieveFileResp{} }
func (m *RetrieveFileResp) String() string            { return proto.CompactTextString(m) }
func (*RetrieveFileResp) ProtoMessage()               {}
func (*RetrieveFileResp) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{16} }

func (m *RetrieveFileResp) GetFileData() []byte {
	if m != nil {
		return m.FileData
	}
	return nil
}

func (m *RetrieveFileResp) GetPartition() []*Partition {
	if m != nil {
		return m.Partition
	}
	return nil
}

func init() {
	proto.RegisterType((*CheckFileExistReq)(nil), "metadata.pb.CheckFileExistReq")
	proto.RegisterType((*CheckFileExistResp)(nil), "metadata.pb.CheckFileExistResp")
	proto.RegisterType((*ReplicaProvider)(nil), "metadata.pb.ReplicaProvider")
	proto.RegisterType((*UploadFilePrepareReq)(nil), "metadata.pb.UploadFilePrepareReq")
	proto.RegisterType((*PieceHashAndSize)(nil), "metadata.pb.PieceHashAndSize")
	proto.RegisterType((*UploadFilePrepareResp)(nil), "metadata.pb.UploadFilePrepareResp")
	proto.RegisterType((*ErasureCodeProvider)(nil), "metadata.pb.ErasureCodeProvider")
	proto.RegisterType((*PieceHashAuth)(nil), "metadata.pb.PieceHashAuth")
	proto.RegisterType((*UploadFileDoneReq)(nil), "metadata.pb.UploadFileDoneReq")
	proto.RegisterType((*Partition)(nil), "metadata.pb.Partition")
	proto.RegisterType((*Block)(nil), "metadata.pb.Block")
	proto.RegisterType((*UploadFileDoneResp)(nil), "metadata.pb.UploadFileDoneResp")
	proto.RegisterType((*ListFilesReq)(nil), "metadata.pb.ListFilesReq")
	proto.RegisterType((*ListFilesResp)(nil), "metadata.pb.ListFilesResp")
	proto.RegisterType((*FileOrFolder)(nil), "metadata.pb.FileOrFolder")
	proto.RegisterType((*RetrieveFileReq)(nil), "metadata.pb.RetrieveFileReq")
	proto.RegisterType((*RetrieveFileResp)(nil), "metadata.pb.RetrieveFileResp")
	proto.RegisterEnum("metadata.pb.FileStoreType", FileStoreType_name, FileStoreType_value)
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// Client API for MatadataService service

type MatadataServiceClient interface {
	CheckFileExist(ctx context.Context, in *CheckFileExistReq, opts ...grpc.CallOption) (*CheckFileExistResp, error)
	UploadFilePrepare(ctx context.Context, in *UploadFilePrepareReq, opts ...grpc.CallOption) (*UploadFilePrepareResp, error)
	UploadFileDone(ctx context.Context, in *UploadFileDoneReq, opts ...grpc.CallOption) (*UploadFileDoneResp, error)
	ListFiles(ctx context.Context, in *ListFilesReq, opts ...grpc.CallOption) (*ListFilesResp, error)
	RetrieveFile(ctx context.Context, in *RetrieveFileReq, opts ...grpc.CallOption) (*RetrieveFileResp, error)
}

type matadataServiceClient struct {
	cc *grpc.ClientConn
}

func NewMatadataServiceClient(cc *grpc.ClientConn) MatadataServiceClient {
	return &matadataServiceClient{cc}
}

func (c *matadataServiceClient) CheckFileExist(ctx context.Context, in *CheckFileExistReq, opts ...grpc.CallOption) (*CheckFileExistResp, error) {
	out := new(CheckFileExistResp)
	err := grpc.Invoke(ctx, "/metadata.pb.MatadataService/CheckFileExist", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *matadataServiceClient) UploadFilePrepare(ctx context.Context, in *UploadFilePrepareReq, opts ...grpc.CallOption) (*UploadFilePrepareResp, error) {
	out := new(UploadFilePrepareResp)
	err := grpc.Invoke(ctx, "/metadata.pb.MatadataService/UploadFilePrepare", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *matadataServiceClient) UploadFileDone(ctx context.Context, in *UploadFileDoneReq, opts ...grpc.CallOption) (*UploadFileDoneResp, error) {
	out := new(UploadFileDoneResp)
	err := grpc.Invoke(ctx, "/metadata.pb.MatadataService/UploadFileDone", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *matadataServiceClient) ListFiles(ctx context.Context, in *ListFilesReq, opts ...grpc.CallOption) (*ListFilesResp, error) {
	out := new(ListFilesResp)
	err := grpc.Invoke(ctx, "/metadata.pb.MatadataService/ListFiles", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *matadataServiceClient) RetrieveFile(ctx context.Context, in *RetrieveFileReq, opts ...grpc.CallOption) (*RetrieveFileResp, error) {
	out := new(RetrieveFileResp)
	err := grpc.Invoke(ctx, "/metadata.pb.MatadataService/RetrieveFile", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for MatadataService service

type MatadataServiceServer interface {
	CheckFileExist(context.Context, *CheckFileExistReq) (*CheckFileExistResp, error)
	UploadFilePrepare(context.Context, *UploadFilePrepareReq) (*UploadFilePrepareResp, error)
	UploadFileDone(context.Context, *UploadFileDoneReq) (*UploadFileDoneResp, error)
	ListFiles(context.Context, *ListFilesReq) (*ListFilesResp, error)
	RetrieveFile(context.Context, *RetrieveFileReq) (*RetrieveFileResp, error)
}

func RegisterMatadataServiceServer(s *grpc.Server, srv MatadataServiceServer) {
	s.RegisterService(&_MatadataService_serviceDesc, srv)
}

func _MatadataService_CheckFileExist_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CheckFileExistReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MatadataServiceServer).CheckFileExist(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/metadata.pb.MatadataService/CheckFileExist",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MatadataServiceServer).CheckFileExist(ctx, req.(*CheckFileExistReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _MatadataService_UploadFilePrepare_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UploadFilePrepareReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MatadataServiceServer).UploadFilePrepare(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/metadata.pb.MatadataService/UploadFilePrepare",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MatadataServiceServer).UploadFilePrepare(ctx, req.(*UploadFilePrepareReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _MatadataService_UploadFileDone_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UploadFileDoneReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MatadataServiceServer).UploadFileDone(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/metadata.pb.MatadataService/UploadFileDone",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MatadataServiceServer).UploadFileDone(ctx, req.(*UploadFileDoneReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _MatadataService_ListFiles_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListFilesReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MatadataServiceServer).ListFiles(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/metadata.pb.MatadataService/ListFiles",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MatadataServiceServer).ListFiles(ctx, req.(*ListFilesReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _MatadataService_RetrieveFile_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RetrieveFileReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MatadataServiceServer).RetrieveFile(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/metadata.pb.MatadataService/RetrieveFile",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MatadataServiceServer).RetrieveFile(ctx, req.(*RetrieveFileReq))
	}
	return interceptor(ctx, in, info, handler)
}

var _MatadataService_serviceDesc = grpc.ServiceDesc{
	ServiceName: "metadata.pb.MatadataService",
	HandlerType: (*MatadataServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "CheckFileExist",
			Handler:    _MatadataService_CheckFileExist_Handler,
		},
		{
			MethodName: "UploadFilePrepare",
			Handler:    _MatadataService_UploadFilePrepare_Handler,
		},
		{
			MethodName: "UploadFileDone",
			Handler:    _MatadataService_UploadFileDone_Handler,
		},
		{
			MethodName: "ListFiles",
			Handler:    _MatadataService_ListFiles_Handler,
		},
		{
			MethodName: "RetrieveFile",
			Handler:    _MatadataService_RetrieveFile_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "metadata.proto",
}

func init() { proto.RegisterFile("metadata.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 965 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xcc, 0x56, 0xdd, 0x8e, 0x1b, 0x35,
	0x14, 0xee, 0x6c, 0x32, 0xd9, 0xcc, 0xc9, 0xcf, 0xa6, 0x06, 0x56, 0x43, 0xd4, 0x42, 0x98, 0x0b,
	0x14, 0x15, 0x69, 0x2f, 0xb6, 0x80, 0x2a, 0xd4, 0x9b, 0xb2, 0xdb, 0x0a, 0x24, 0xd2, 0xae, 0x9c,
	0xf6, 0x8e, 0x9b, 0xd9, 0x19, 0xa7, 0xb1, 0x36, 0x99, 0x99, 0x7a, 0x9c, 0x88, 0xf6, 0x01, 0x90,
	0x90, 0x78, 0x07, 0x5e, 0x80, 0x9f, 0xc7, 0xe0, 0x49, 0x10, 0xaf, 0x81, 0xce, 0x99, 0x5f, 0x4f,
	0xd2, 0x05, 0x55, 0x2a, 0xea, 0xdd, 0x39, 0x3e, 0x9f, 0x8f, 0x8f, 0x3f, 0x7f, 0xc7, 0x36, 0x0c,
	0xd7, 0x42, 0xfb, 0xa1, 0xaf, 0xfd, 0x93, 0x44, 0xc5, 0x3a, 0x66, 0xbd, 0xca, 0xbf, 0xf4, 0xfe,
	0x3c, 0x80, 0x9b, 0x67, 0x4b, 0x11, 0x5c, 0x3d, 0x92, 0x2b, 0xf1, 0xf0, 0x07, 0x99, 0x6a, 0x2e,
	0x5e, 0x30, 0x17, 0x0e, 0xb7, 0x42, 0xa5, 0x32, 0x8e, 0x5c, 0x6b, 0x62, 0x4d, 0x07, 0xbc, 0x70,
	0xd9, 0x31, 0x74, 0xa2, 0x38, 0x14, 0xdf, 0x86, 0xee, 0xc1, 0xc4, 0x9a, 0xf6, 0x79, 0xee, 0xb1,
	0x5b, 0xe0, 0x68, 0xb9, 0x16, 0xa9, 0xf6, 0xd7, 0x89, 0xdb, 0x9a, 0x58, 0xd3, 0x36, 0xaf, 0x06,
	0xd8, 0x18, 0xba, 0x0b, 0xb9, 0x12, 0x17, 0xbe, 0x5e, 0xba, 0xed, 0x89, 0x35, 0x75, 0x78, 0xe9,
	0x17, 0xb1, 0x6f, 0xfc, 0x74, 0xe9, 0xda, 0x94, 0xb3, 0xf4, 0x8b, 0xd8, 0x5c, 0xbe, 0x12, 0x6e,
	0x87, 0x92, 0x96, 0x7e, 0x11, 0x7b, 0xec, 0xaf, 0x85, 0x7b, 0x58, 0xe5, 0x44, 0x9f, 0x4d, 0xa0,
	0x87, 0xf6, 0x2c, 0x0e, 0x9f, 0xca, 0xb5, 0x70, 0xbb, 0x34, 0xb5, 0x3e, 0x54, 0xcc, 0x3e, 0xf7,
	0xb5, 0xef, 0x3a, 0xd5, 0xaa, 0xe8, 0xe3, 0x6c, 0x19, 0x69, 0xa1, 0xfc, 0x40, 0xcb, 0xad, 0x70,
	0x61, 0x62, 0x4d, 0xbb, 0xbc, 0x3e, 0xc4, 0x18, 0xb4, 0x53, 0xf9, 0x3c, 0x72, 0x7b, 0x34, 0x93,
	0x6c, 0xef, 0xd7, 0x03, 0x60, 0x4d, 0x26, 0xd3, 0x04, 0xa1, 0x41, 0x1c, 0x8a, 0x9c, 0x47, 0xb2,
	0x91, 0x44, 0xa1, 0xd4, 0x2c, 0x7d, 0x4e, 0x24, 0x3a, 0x3c, 0xf7, 0xd8, 0x3d, 0x70, 0x52, 0x1d,
	0x2b, 0xf1, 0xf4, 0x65, 0x22, 0x88, 0xc4, 0xe1, 0xe9, 0xf8, 0xa4, 0x76, 0x5a, 0x27, 0x98, 0x7a,
	0x5e, 0x20, 0x78, 0x05, 0x66, 0x9f, 0xc2, 0x10, 0x31, 0x17, 0x52, 0x04, 0xe2, 0x2c, 0xde, 0x44,
	0x9a, 0x68, 0xb6, 0x79, 0x63, 0x94, 0xdd, 0x81, 0xd1, 0x56, 0x28, 0xb9, 0x78, 0x59, 0x43, 0xda,
	0x84, 0xdc, 0x19, 0x67, 0x1e, 0xf4, 0x95, 0x48, 0x56, 0x32, 0xf0, 0x33, 0x5c, 0x87, 0x70, 0xc6,
	0x18, 0xbb, 0x07, 0xdd, 0x44, 0xc5, 0x5b, 0x19, 0x0a, 0xe5, 0x1e, 0x4e, 0x5a, 0xd3, 0xde, 0xe9,
	0x2d, 0xa3, 0x60, 0x9e, 0x81, 0x2f, 0x72, 0x0c, 0x2f, 0xd1, 0xde, 0x2f, 0x16, 0x1c, 0x35, 0xa2,
	0x35, 0x71, 0x59, 0x86, 0xb8, 0x8e, 0xa1, 0x93, 0x0a, 0xb5, 0x15, 0xaa, 0xe0, 0x2b, 0xf3, 0x90,
	0xdb, 0x24, 0x56, 0x9a, 0xa8, 0x1a, 0x70, 0xb2, 0x4d, 0x21, 0xb6, 0x9b, 0x42, 0x3c, 0x86, 0x8e,
	0x96, 0xc1, 0x95, 0xc8, 0x76, 0xed, 0xf0, 0xdc, 0xc3, 0x4c, 0xfe, 0x46, 0x2f, 0x69, 0x8f, 0x7d,
	0x4e, 0xb6, 0xf7, 0xb7, 0x05, 0xef, 0x3f, 0x4b, 0x56, 0xb1, 0x1f, 0x22, 0xed, 0x17, 0x4a, 0x24,
	0xbe, 0x12, 0x6f, 0xb1, 0x3b, 0xa8, 0x03, 0xda, 0xd7, 0x74, 0x80, 0xdd, 0xe8, 0x80, 0xbb, 0x60,
	0x27, 0x78, 0x5c, 0x6e, 0x87, 0x98, 0xbf, 0x6d, 0x30, 0x4f, 0x07, 0x89, 0x29, 0x1e, 0x44, 0x21,
	0xa2, 0x79, 0x86, 0x2d, 0xa5, 0x7b, 0x58, 0x93, 0xee, 0x57, 0x30, 0x6a, 0xc2, 0x11, 0xb7, 0xc4,
	0x82, 0xb2, 0x93, 0x20, 0x3b, 0x9b, 0xfb, 0x4a, 0xd0, 0xe6, 0x06, 0x9c, 0x6c, 0xef, 0x19, 0x7c,
	0xb0, 0x87, 0xa4, 0x34, 0x61, 0xf7, 0x6b, 0xd2, 0xb0, 0xa8, 0xc0, 0x89, 0x51, 0xe0, 0x43, 0xe5,
	0xa7, 0x1b, 0x25, 0xce, 0xe2, 0x50, 0xec, 0x91, 0xc7, 0x1f, 0x16, 0xbc, 0xb7, 0x07, 0xf1, 0x3f,
	0x48, 0xe4, 0x4b, 0xe8, 0xe2, 0x66, 0x1f, 0xa0, 0x1c, 0x6c, 0xaa, 0x7b, 0xfc, 0x1a, 0x62, 0x37,
	0x7a, 0xc9, 0x4b, 0xac, 0xf7, 0x04, 0x06, 0x46, 0x68, 0x2f, 0x83, 0x95, 0xfe, 0x0e, 0xf6, 0xea,
	0xaf, 0x55, 0xd3, 0xdf, 0x5f, 0x16, 0xdc, 0xac, 0xa8, 0x3d, 0x8f, 0xa3, 0x77, 0x4a, 0x7c, 0x9f,
	0x83, 0x93, 0xf8, 0x4a, 0x4b, 0x8d, 0x95, 0x64, 0x02, 0x3c, 0x36, 0x79, 0x2a, 0xa2, 0xbc, 0x02,
	0xee, 0x55, 0xdf, 0x17, 0xe0, 0x94, 0x58, 0x36, 0x05, 0xfb, 0x72, 0x15, 0x07, 0x57, 0xb9, 0x64,
	0x98, 0x91, 0xf2, 0x6b, 0x8c, 0xf0, 0x0c, 0xe0, 0xfd, 0x64, 0x81, 0x4d, 0x03, 0xff, 0x55, 0xaa,
	0xb8, 0x1d, 0x9a, 0x3a, 0x17, 0x2f, 0x72, 0x3d, 0x94, 0x3e, 0xc6, 0x02, 0xbc, 0xbc, 0xd3, 0xcd,
	0x9a, 0x68, 0xe8, 0xf2, 0xd2, 0xc7, 0xf7, 0x80, 0x6e, 0xda, 0xc7, 0x19, 0xbb, 0x28, 0x8a, 0x3e,
	0xaf, 0x0f, 0x79, 0x53, 0x60, 0xcd, 0x93, 0xca, 0xae, 0xfe, 0x30, 0x8e, 0xb2, 0xab, 0xbf, 0xcb,
	0xc9, 0xf6, 0x7e, 0xb4, 0xa0, 0xff, 0x9d, 0x4c, 0x35, 0x02, 0xd3, 0xb7, 0x71, 0x9e, 0x28, 0xf8,
	0xea, 0x99, 0x25, 0xbb, 0x64, 0xdd, 0xae, 0xb1, 0x7e, 0x1f, 0x06, 0xb5, 0x3a, 0xd2, 0x84, 0x7d,
	0x06, 0xad, 0x45, 0xbc, 0xc8, 0x79, 0xff, 0x70, 0xe7, 0xd9, 0x79, 0xa2, 0x1e, 0xc5, 0x2b, 0xec,
	0x51, 0x44, 0x79, 0x3f, 0x5b, 0xd0, 0xaf, 0x8f, 0x62, 0xb1, 0x0b, 0xb2, 0xf2, 0xdd, 0xe6, 0x1e,
	0x2e, 0x1d, 0xe1, 0x0b, 0x9d, 0xc9, 0x9d, 0x6c, 0xdc, 0xf2, 0x3a, 0x7f, 0x99, 0xb3, 0xf2, 0x0b,
	0xf7, 0x4d, 0xc5, 0xe8, 0xfd, 0x46, 0x8f, 0x89, 0x56, 0x52, 0x6c, 0x05, 0x96, 0xf5, 0x2e, 0x35,
	0x4a, 0x41, 0x7e, 0xa7, 0x46, 0x7e, 0x08, 0x23, 0xb3, 0xdc, 0x34, 0x31, 0x7e, 0x24, 0x56, 0xe3,
	0x47, 0xf2, 0x46, 0xcd, 0x76, 0xe7, 0x14, 0x06, 0xc6, 0x87, 0x81, 0x1d, 0x41, 0xaf, 0x76, 0xa7,
	0x8e, 0x6e, 0xb0, 0x11, 0xf4, 0x67, 0x9b, 0x95, 0x96, 0xf9, 0x43, 0x3c, 0xb2, 0x4e, 0x7f, 0x6f,
	0xc1, 0xd1, 0xcc, 0xcf, 0x12, 0xcf, 0x85, 0xda, 0xca, 0x40, 0xb0, 0x39, 0x0c, 0xcd, 0x8f, 0x0d,
	0xfb, 0xc8, 0x58, 0x7c, 0xe7, 0xff, 0x38, 0xfe, 0xf8, 0xda, 0x78, 0x9a, 0x78, 0x37, 0xd8, 0xf7,
	0xf5, 0xcb, 0x2d, 0x7f, 0x37, 0xd8, 0x27, 0xc6, 0xbc, 0x7d, 0x8f, 0xef, 0xd8, 0xfb, 0x37, 0x08,
	0x65, 0x9f, 0xc3, 0xd0, 0x6c, 0xc8, 0x46, 0xc9, 0x3b, 0xf7, 0x6a, 0xa3, 0xe4, 0xdd, 0x6e, 0xf6,
	0x6e, 0xb0, 0x73, 0x70, 0xca, 0x96, 0x61, 0x66, 0x87, 0xd4, 0x5b, 0x7a, 0x3c, 0x7e, 0x5d, 0x88,
	0xb2, 0xcc, 0xa0, 0x5f, 0x3f, 0x7b, 0xd6, 0xfc, 0x30, 0x19, 0x2a, 0x1e, 0xdf, 0xbe, 0x26, 0x8a,
	0xe9, 0x2e, 0x3b, 0xf4, 0xa9, 0xbf, 0xfb, 0x4f, 0x00, 0x00, 0x00, 0xff, 0xff, 0x3f, 0x2a, 0x4a,
	0xfd, 0xe6, 0x0b, 0x00, 0x00,
}