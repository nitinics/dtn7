package bundle

import (
	"bytes"
	"encoding/binary"
	"hash/crc32"

	"github.com/howeyc/crc16"
	"github.com/ugorji/go/codec"
)

// CRCType indicates which CRC type is used. Only the three defined consts
// CRCNo, CRC16 and CRC32 are valid, as specified in section 4.1.1.
type CRCType uint

const (
	CRCNo CRCType = 0
	CRC16 CRCType = 1
	CRC32 CRCType = 2
)

func (c CRCType) String() string {
	switch c {
	case CRCNo:
		return "no"
	case CRC16:
		return "16"
	case CRC32:
		return "32"
	default:
		return "unknown"
	}
}

var (
	crc16table = crc16.MakeTable(crc16.CCITT)
	crc32table = crc32.MakeTable(crc32.Castagnoli)
)

// blockToBytes encodes a Block to a byte array based on the CBOR encoding. It
// temporary sets the present CRC value to zero. Therefore this function is not
// thread safe.
func blockToBytes(blck block) []byte {
	var b []byte = make([]byte, 0, 64)
	var enc *codec.Encoder = codec.NewEncoderBytes(&b, new(codec.CborHandle))

	var blockCRC = blck.getCRC()

	blck.resetCRC()
	enc.MustEncode(blck)
	blck.setCRC(blockCRC)

	return b
}

// calculateCRC calculates a Block's CRC value based on its CRCType. The CRC
// value will be set to zero temporary during calcuation. Thereforce this
// function is not thread safe.
// The returned value is a byte array containing the CRC in network byte order
// (big endian) and its length is 4 for CRC32 or 2 for CRC16.
func calculateCRC(blck block) []byte {
	var data = blockToBytes(blck)
	var arr = emptyCRC(blck.GetCRCType())

	switch blck.GetCRCType() {
	case CRCNo:

	case CRC16:
		binary.BigEndian.PutUint16(arr, crc16.Checksum(data, crc16table))

	case CRC32:
		binary.BigEndian.PutUint32(arr, crc32.Checksum(data, crc32table))

	default:
		panic("Unknown CRCType")
	}

	return arr
}

// emptyCRC returns the "default" CRC value for the given CRC Type.
func emptyCRC(crcType CRCType) (arr []byte) {
	switch crcType {
	case CRCNo:
		arr = nil

	case CRC16:
		arr = make([]byte, 2)

	case CRC32:
		arr = make([]byte, 4)

	default:
		panic("Unknown CRCType")
	}

	return
}

// checkCRC returns true if the stored CRC value matches the calculated one or
// the CRC Type is none.
// This method changes the block's CRC value temporary and is not thread safe.
func checkCRC(blck block) bool {
	if !blck.HasCRC() {
		return true
	}

	return bytes.Equal(blck.getCRC(), calculateCRC(blck))
}
