package wire

import (
	"bytes"

	"github.com/lucas-clemente/quic-go/internal/protocol"
	"github.com/lucas-clemente/quic-go/internal/utils"
	"github.com/lucas-clemente/quic-go/qerr"
)

// A StopSendingFrame is a STOP_SENDING frame
type StopSendingFrame struct {
	StreamID  protocol.StreamID
	ErrorCode qerr.ErrorCode
}

// ParseStopSendingFrame parses a STOP_SENDING frame
func ParseStopSendingFrame(r *bytes.Reader, version protocol.VersionNumber) (*StopSendingFrame, error) {
	if _, err := r.ReadByte(); err != nil { // read the TypeByte
		return nil, err
	}

	streamID, err := utils.ReadVarInt(r)
	if err != nil {
		return nil, err
	}
	errorCode, err := utils.BigEndian.ReadUint16(r)
	if err != nil {
		return nil, err
	}

	return &StopSendingFrame{
		StreamID:  protocol.StreamID(streamID),
		ErrorCode: qerr.ErrorCode(errorCode),
	}, nil
}

// MinLength of a written frame
func (f *StopSendingFrame) MinLength(version protocol.VersionNumber) (protocol.ByteCount, error) {
	return 1 + utils.VarIntLen(uint64(f.StreamID)) + 2, nil
}

func (f *StopSendingFrame) Write(b *bytes.Buffer, version protocol.VersionNumber) error {
	b.WriteByte(0x0c)
	utils.WriteVarInt(b, uint64(f.StreamID))
	utils.BigEndian.WriteUint16(b, uint16(f.ErrorCode))
	return nil
}