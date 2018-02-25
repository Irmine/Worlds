package io

import (
	"bytes"
	"encoding/binary"
	"io"
	"math"
	"os"
	"time"
)

// CompressionType is the compression used to compress chunk data.
// All chunk data is compressed by a byte indicating the compression type.
type CompressionType byte

const (
	// HeaderSize is the total header size of the chunk.
	// The chunk data proceeds after this.
	HeaderSize = 8192
	// SectorSize is the total size of a block.
	// Every sector must be padded to have a length of a multiple of this.
	SectorSize = 4096
	// LengthOffset is the amount of bytes that form the length of chunk data.
	// It prefixes all chunk data and does not include padding.
	LengthOffset = 4
	// Gzip compression is not actually used in Minecraft.
	// It is only there for legacy regions and should no longer be used.
	CompressionGzip CompressionType = 1
	// Zlib compression is the main compression for region files in Minecraft.
	CompressionZlib CompressionType = 2
)

// RegionHeader contains locations of chunks, as well as timestamps for these.
type RegionHeader struct {
	Locations  [1024]*Location
	Timestamps [1024]int32
}

// A location holds information about the offset of chunks, the count of sectors and the timestamp of those.
type Location struct {
	Offset       int32
	SectorLength int32
}

// A region holds a reference to the attached file, and has a header containing information of the region.
type Region struct {
	Header RegionHeader
	File   *os.File
}

// NewRegion returns a new region struct with data at the given path.
// It does not load the header, and therefore OpenRegion is recommended for usage.
func NewRegion(path string) (*Region, error) {
	var file, err = os.OpenFile(path, os.O_RDWR, 0644)
	return &Region{RegionHeader{}, file}, err
}

// OpenRegion opens a region at the given path.
// It loads the header and returns any error that might occur.
func OpenRegion(path string) (*Region, error) {
	var region, err = NewRegion(path)
	region.LoadHeader()
	return region, err
}

// GetChunkLocationIndex returns the chunk location index, where information of a chunk can be found.
func GetChunkLocationIndex(x, z int32) int {
	return int((x & 31) + (z&31)*32)
}

// Close closes the region file, cleans garbage and writes the region header.
func (r *Region) Close(save bool) {
	r.Save()
	r.File.Close()
}

// Save saves the region file, cleaning the garbage and writing the header.
func (r *Region) Save() {
	r.CleanGarbage()
	r.WriteHeader()
}

// LoadHeader loads the header of the region.
// This includes the loading of timestamps and chunk locations.
func (r *Region) LoadHeader() {
	var buff = make([]byte, 8192)
	r.File.Read(buff)

	var o int32 = 0

	for i := 0; i < 1024; i++ {
		index := i * LengthOffset
		b := bytes.NewBuffer(buff[index : index+4])

		binary.Read(b, binary.BigEndian, &o)
		offset := o >> 8

		r.Header.Locations[i] = &Location{offset << 12, o & 0xff}
	}

	var in int32 = 0
	var buffer *bytes.Buffer

	for i := 0; i < 1024; i++ {
		buffer = bytes.NewBuffer(buff[HeaderSize/2+i*LengthOffset : HeaderSize/2+i*LengthOffset+LengthOffset])
		binary.Read(buffer, binary.BigEndian, &in)
		r.Header.Timestamps[i] = in
	}
}

// CleanGarbage cleans all garbage of the region file.
// This function
func (r *Region) CleanGarbage() {
	var l int32
	var lastOffset int64 = HeaderSize

	var fileBuffer = bytes.NewBuffer([]byte{})

	for i := 0; i < 1024; i++ {
		loc := r.Header.Locations[i]
		if loc.Offset == 0 {
			continue
		}

		b := make([]byte, 4)
		r.File.ReadAt(b, int64(loc.Offset))
		buffer := bytes.NewBuffer(b)
		binary.Read(buffer, binary.BigEndian, &l)

		if l <= 0 {
			loc.Offset = 0
			continue
		}

		data := make([]byte, l)
		r.File.ReadAt(data, int64(loc.Offset)+4)
		buffer.Write(data)
		paddingLength := int32(math.Ceil(float64(buffer.Len())/SectorSize) * SectorSize)
		buffer.Write(make([]byte, paddingLength-l))

		fileBuffer.Write(buffer.Bytes())
		lastOffset += int64(len(buffer.Bytes()))
	}

	io.Copy(r.File, fileBuffer)
	r.File.Truncate(lastOffset)
}

// WriteHeader writes the header to the file.
// This includes the writing of locations and timestamps.
func (r *Region) WriteHeader() {
	var header = bytes.NewBuffer([]byte{})
	var offsets []int32
	for i := 0; i < 1024; i++ {
		offsetL := r.Header.Locations[i].SectorLength
		offsetI := r.Header.Locations[i].Offset >> 12 << 8
		offset := offsetI | offsetL
		offsets = append(offsets, offset)
	}
	binary.Write(header, binary.BigEndian, offsets)

	var timestamps = bytes.NewBuffer([]byte{})
	binary.Write(timestamps, binary.BigEndian, r.Header.Timestamps[:])

	var headerBytes = append(header.Bytes(), timestamps.Bytes()...)
	r.File.WriteAt(headerBytes, 0)
}

// GetLocation returns the location of a chunk with the given X and Z in the region.
func (r *Region) GetLocation(x, z int32) *Location {
	return r.Header.Locations[GetChunkLocationIndex(x, z)]
}

// GetChunkData returns the chunk data of a chunk with the given X and Z in the region.
// It also provides the compression type, in order to know how to decompress it.
func (r *Region) GetChunkData(x, z int32) (compressionType CompressionType, chunkData []byte) {
	var loc = r.GetLocation(x, z)
	if loc.Offset == 0 {
		return 0, []byte{}
	}
	var buff = make([]byte, 5)

	r.File.ReadAt(buff, int64(loc.Offset))

	var buffer = bytes.NewBuffer(buff[:4])

	var length int32
	binary.Read(buffer, binary.BigEndian, &length)
	compressionType = CompressionType(buff[4])

	chunkData = make([]byte, length)
	r.File.ReadAt(chunkData, int64(loc.Offset+5))
	return
}

// WriteChunkData writes the given chunk data at the given X and Z.
// Compression type should be CompressionZlib. (or CompressionGzip)
func (r *Region) WriteChunkData(x, z int32, data []byte, compressionType byte) {
	var loc = r.GetLocation(x, z)
	var buff = make([]byte, LengthOffset)

	r.File.ReadAt(buff, int64(loc.Offset))

	var buffer = bytes.NewBuffer(buff[:4])

	var oldLength int32
	var newLength = int32(len(data)) + 5
	binary.Read(buffer, binary.BigEndian, &oldLength)
	oldLength += 4
	var oldLengthPadded = int32(math.Ceil(float64(oldLength)/SectorSize) * SectorSize)

	var sectorLength = int32(math.Ceil(float64(newLength) / SectorSize))
	var newLengthPadded = sectorLength

	var padding []byte
	if newLengthPadded-newLength > 0 {
		padding = make([]byte, newLengthPadded-newLength)
	}

	var offset = loc.Offset
	if newLengthPadded > oldLengthPadded {
		d, _ := r.File.Stat()
		offset = int32(d.Size())
		loc.Offset = offset
	}
	loc.SectorLength = sectorLength

	r.Header.Timestamps[GetChunkLocationIndex(x, z)] = int32(time.Now().Unix())

	buffer = bytes.NewBuffer([]byte{})
	binary.Write(buffer, binary.BigEndian, newLength)

	buffer.WriteByte(compressionType)
	buffer.Write(data)
	buffer.Write(padding)

	r.File.WriteAt(buffer.Bytes(), int64(offset))
}

// HasChunkGenerated checks if the region has a chunk with the given X and Z generated.
func (r *Region) HasChunkGenerated(x, z int32) bool {
	return r.GetLocation(x, z).IsExistent()
}

// Checks if the location is existent.
func (location *Location) IsExistent() bool {
	return location.Offset >= HeaderSize && location.SectorLength != 0
}
