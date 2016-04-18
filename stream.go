package quic

import (
	"github.com/lucas-clemente/quic-go/frames"
	"github.com/lucas-clemente/quic-go/protocol"
	"github.com/lucas-clemente/quic-go/utils"
)

// A Stream assembles the data from StreamFrames and provides a super-convenient Read-Interface
type Stream struct {
	Session        *Session
	StreamID       protocol.StreamID
	StreamFrames   chan *frames.StreamFrame
	CurrentFrame   *frames.StreamFrame
	ReadPosInFrame int
	WriteOffset    uint64
}

// NewStream creates a new Stream
func NewStream(session *Session, StreamID protocol.StreamID) *Stream {
	return &Stream{
		Session:      session,
		StreamID:     StreamID,
		StreamFrames: make(chan *frames.StreamFrame, 8), // ToDo: add config option for this number
	}
}

// Read reads data
func (s *Stream) Read(p []byte) (int, error) {
	bytesRead := 0
	for bytesRead < len(p) {
		if s.CurrentFrame == nil {
			select {
			case s.CurrentFrame = <-s.StreamFrames:
			default:
				if bytesRead == 0 {
					s.CurrentFrame = <-s.StreamFrames
				} else {
					return bytesRead, nil
				}
			}
			s.ReadPosInFrame = 0
		}
		m := utils.Min(len(p)-bytesRead, len(s.CurrentFrame.Data)-s.ReadPosInFrame)
		copy(p[bytesRead:], s.CurrentFrame.Data[s.ReadPosInFrame:])
		s.ReadPosInFrame += m
		bytesRead += m
		if s.ReadPosInFrame >= len(s.CurrentFrame.Data) {
			s.CurrentFrame = nil
		}
	}

	return bytesRead, nil
}

// ReadByte implements io.ByteReader
func (s *Stream) ReadByte() (byte, error) {
	// TODO: Optimize
	p := make([]byte, 1)
	n, err := s.Read(p)
	if err != nil {
		return 0, err
	}
	if n != 1 {
		panic("Stream: should have returned error")
	}
	return p[0], nil
}

func (s *Stream) Write(p []byte) (int, error) {
	frame := &frames.StreamFrame{
		StreamID: s.StreamID,
		Offset:   s.WriteOffset,
		Data:     p,
	}
	err := s.Session.SendFrames([]frames.Frame{frame})
	if err != nil {
		return 0, err
	}
	s.WriteOffset += uint64(len(p))
	return len(p), nil
}

func (s *Stream) Close() error {
	return s.Session.SendFrame(&frames.StreamFrame{
		StreamID: s.StreamID,
		Offset:   s.WriteOffset,
		FinBit:   true,
	})
}

// AddStreamFrame adds a new stream frame
func (s *Stream) AddStreamFrame(frame *frames.StreamFrame) error {
	s.StreamFrames <- frame
	return nil
}