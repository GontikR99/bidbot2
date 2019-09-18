package soundmanip

import (
	"bytes"
	"errors"
	"fmt"
	"io"
)

type oggPage struct {
	HeaderType            byte
	GranulePosition       uint64
	BitstreamSerialNumber uint32
	PageSequenceNumber    uint32
	Segments              [][]byte
}

func toInt(data []byte) uint64 {
	result := uint64(0)
	place := uint(0)
	for _, b := range data {
		result |= uint64(uint8(b)) << place
		place += 8
	}
	return result
}

func readPage(input io.Reader) (page *oggPage, err error) {
	result := &oggPage{}
	headBytes := make([]byte, 27)
	readLen, err := input.Read(headBytes)
	if err != nil {
		err = fmt.Errorf("Failed to read Ogg page: %v", err)
		return
	}
	if readLen != len(headBytes) {
		err = errors.New("Short read")
		return
	}
	if bytes.Compare([]byte("OggS"), headBytes[0:4]) != 0 {
		err = errors.New("Not an Ogg stream")
		return
	}
	if headBytes[4] != 0 {
		err = errors.New("Ogg version not supported")
		return
	}
	result.HeaderType = headBytes[5]
	result.GranulePosition = toInt(headBytes[6:14])
	result.BitstreamSerialNumber = uint32(toInt(headBytes[14:18]))
	result.BitstreamSerialNumber = uint32(toInt(headBytes[18:22]))
	segmentCount := uint8(headBytes[26])
	segmentTable := make([]byte, segmentCount)
	readLen, err = input.Read(segmentTable)
	if err != nil {
		err = fmt.Errorf("Failed to read Ogg page: %v", err)
		return
	}
	if readLen != len(segmentTable) {
		err = errors.New("Short read")
		return
	}
	result.Segments = make([][]byte, segmentCount)
	for i := uint8(0); i < segmentCount; i++ {
		result.Segments[i] = make([]byte, uint8(segmentTable[i]))
		readLen, err = input.Read(result.Segments[i])
		if err != nil {
			err = fmt.Errorf("Failed to read Ogg page: %v", err)
			return
		}
		if readLen != len(result.Segments[i]) {
			err = errors.New("Short read")
			return
		}
	}
	page = result
	return
}

type OggFile struct {
	Input       io.Reader
	currentPage *oggPage
	nextSegment int
}

func (ogf *OggFile) ReadPacket() (data []byte, err error) {
	result := make([]byte, 0)
	for {
		if ogf.currentPage == nil {
			ogf.currentPage, err = readPage(ogf.Input)
			if err != nil {
				return
			}
			ogf.nextSegment = 0
			continue
		}
		if ogf.nextSegment >= len(ogf.currentPage.Segments) {
			ogf.nextSegment = 0
			ogf.currentPage = nil
			continue
		}
		segment := ogf.currentPage.Segments[ogf.nextSegment]
		ogf.nextSegment += 1
		result = append(result, segment...)
		if len(segment) != 255 {
			data = result
			return
		}
	}
}

type OpusFile struct {
	OggFile
	ChannelCount    uint8
	PreSkip         uint16
	InputSampleRate uint32
	OutputGain      uint16
	MappingFamily   uint8
	EncoderName     string
	UserTags        []string
}

// Read an Ogg Opus file, and get packets out of it.
func NewOpusFile(input io.Reader) (file *OpusFile, err error) {
	result := &OpusFile{}
	result.Input = input
	idPacket, err := result.ReadPacket()
	if err != nil {
		return
	}
	if len(idPacket) < 19 {
		err = errors.New("Invalid Opus header")
		return
	}
	if bytes.Compare([]byte("OpusHead"), idPacket[0:8]) != 0 {
		err = errors.New("Not an Ogg Opus stream")
		return
	}
	if idPacket[8] != 1 {
		err = errors.New("Unsupported Ogg Opus version")
		return
	}
	result.ChannelCount = uint8(idPacket[9])
	result.PreSkip = uint16(toInt(idPacket[10:12]))
	result.InputSampleRate = uint32(toInt(idPacket[12:16]))
	result.OutputGain = uint16(toInt(idPacket[16:18]))
	result.MappingFamily = uint8(idPacket[18])

	commentPacket, err := result.ReadPacket()
	if err != nil {
		return
	}
	if len(commentPacket) < 12 {
		err = errors.New("Comment packet too short")
		return
	}
	if bytes.Compare([]byte("OpusTags"), commentPacket[0:8]) != 0 {
		err = errors.New("Invalid tags packet")
		return
	}
	vsl := int(toInt(commentPacket[8:12]))
	if len(commentPacket) < 16+vsl {
		err = errors.New("Short read decoding vendor string")
		return
	}
	result.EncoderName = string(commentPacket[12 : 12+vsl])
	numUserComments := int(toInt(commentPacket[12+vsl : 16+vsl]))
	result.UserTags = make([]string, numUserComments)
	offset := 16 + vsl
	for i := 0; i < numUserComments; i++ {
		if len(commentPacket) < offset+4 {
			err = errors.New("Short read decoding user string")
			return
		}
		userCommentLength := int(toInt(commentPacket[offset : offset+4]))
		offset += 4
		if len(commentPacket) < offset+userCommentLength {
			err = errors.New("Short read decoding user string")
			return
		}
		result.UserTags[i] = string(commentPacket[offset : offset+userCommentLength])
		offset += userCommentLength
	}
	file = result
	return
}
