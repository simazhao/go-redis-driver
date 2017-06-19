package driver

import (
	"net"
	"goredis-driver/pkg/log"
	"strconv"
	"errors"
)

type RedisReader struct {
	Reader *BufferReader
}

func (r *RedisReader) ReadResponse() ([]*RequestClip, error) {
	response := make([]*RequestClip, 0)

	if err := r.readResponse(&response); err != nil {
		return nil, err
	}

	return response, nil
}

func (r *RedisReader) readResponse(clips *[]*RequestClip) error {
	if head, err := r.readPrefix(clips); err != nil {
		return err
	} else {
		return r.readMore(SegmentType(head), clips)
	}

	return nil
}

func (r *RedisReader) readPrefix(clips *[]*RequestClip) (byte, error) {
	if head, err := r.Reader.ReadByte(); err != nil {
		return 0, err
	} else {
		log.Factory.GetLogger().InfoFormat("get prefix %s", string(head))
		if head != byte(SegArray) {
			clip := &RequestClip{Type:SegmentType(head)}
			*clips = append(*clips, clip)
		}

		return head, nil
	}
}

func (r *RedisReader) readMore(segtype SegmentType, clips *[]*RequestClip) error {
	switch segtype {
	case SegString:
		return r.readString(clips)
	case SegError:
		return r.readError(clips)
	case SegInt:
		return r.readString(clips)
	case SegBulkBytes:
		return r.readBulk(clips)
	case SegArray:
		return r.readArray(clips)
	default:
		return errors.New("not support")
	}
}

func (r *RedisReader) readBulk(clips *[]*RequestClip) error {
	if lengthstr, err := r.Reader.ReadLine(); err != nil {
		return err
	} else if length, err := strconv.Atoi(string(lengthstr[:len(lengthstr)-2])); err != nil {
		return errors.New("data length is illegal")
	} else if length < 0 {
		clip := (*clips)[len(*clips) - 1]
		clip.Value = make([]byte, 0)
		return nil
	}else if bytes, err := r.Reader.Read(length + 2); err != nil {
		return err
	} else{
		log.Factory.GetLogger().InfoFormat("get bulk: %s", string(bytes))
		clip := (*clips)[len(*clips) - 1]
		clip.Value = make([]byte, len(bytes) - 2)
		copy(clip.Value, bytes)
		return nil
	}
}

func (r *RedisReader) readArray(clips *[]*RequestClip) error {
	if lengthstr, err := r.Reader.ReadLine(); err != nil {
		return err
	} else if length, err := strconv.Atoi(string(lengthstr[:len(lengthstr)-2])); err != nil || length <= 0{
		return errors.New("data length is illegal")
	} else {
		for i:=0;i<length;i++ {
			if err := r.readResponse(clips); err != nil{
				return err
			}
		}
	}

	return nil
}

func (r* RedisReader) readError(clips *[]*RequestClip) error {
	if bytes, err := r.Reader.ReadLine(); err != nil {
		return err
	} else {
		return errors.New(string(bytes))
	}
}

func (r* RedisReader) readString(clips *[]*RequestClip) error {
	if bytes, err := r.Reader.ReadLine(); err != nil {
		return err
	} else {
		clip := (*clips)[len(*clips) - 1]
		clip.Value = make([]byte, len(bytes) - 2)
		copy(clip.Value, bytes)
		return nil
	}
}

type BufferReader struct{
	Reader *connReader
	buf []byte
	rpos int
	wpos int
}

func (br *BufferReader) leftroom() int {
	return len(br.buf) - br.wpos
}

func NewBufferReader(conn net.Conn, bufferSize int) *BufferReader {
	buffer := &BufferReader{Reader:&connReader{Conn:conn}}
	buffer.buf = make([]byte, bufferSize)
	return buffer
}

const(
	MaxBuffSize = 2*1024*1024
)

func (br *BufferReader) ReadLine() ([]byte, error) {
	rpos := br.rpos
	if n := containLine(br.buf, br.rpos, br.wpos); n >= 0 {
		br.rpos = rpos + n + 1
		return br.buf[rpos:br.rpos], nil
	}

	left := make([]byte, len(br.buf))
	copiedlen := br.wpos - rpos
	copy(left, br.buf[rpos:br.wpos])

	for true {
		if br.leftroom() == 0 {
			br.rpos, br.wpos = 0, 0
		}

		wpos := br.wpos
		if n, err := br.Reader.Read(br.buf[br.wpos:]); err != nil {
			return nil, err
		} else {
			br.wpos += n
		}

		if n := containLine(br.buf, wpos, br.wpos); n >= 0 {
			if len(left) - copiedlen < n {
				var err2 error
				if left, err2 = newSizeBuffer(left, copiedlen + n); err2 != nil {
					return nil, err2
				}
			}

			l := copy(left[copiedlen:], br.buf[wpos:wpos+n+1])
			br.rpos = wpos + n + 1
			copiedlen += l

			return left[:copiedlen], nil
		} else {
			if len(left) - copiedlen < br.wpos - wpos {
				var err2 error
				if left, err2 = newSizeBuffer(left, br.wpos - wpos + copiedlen); err2 != nil {
					return nil, err2
				}
			}

			l := copy(left[copiedlen:], br.buf[wpos:br.wpos])

			br.rpos = br.wpos
			copiedlen += l
		}
	}

	return nil, errors.New("failed to read data")
}

func newSizeBuffer(buffer []byte, size int) ([]byte, error) {
	if size > MaxBuffSize {
		return nil, errors.New("data is too big")
	}

	newBuffer := make([]byte, size)
	copy(newBuffer, buffer)

	return newBuffer, nil
}

func doubleSizeBuffer(buffer []byte) ([]byte, error) {
	if len(buffer) * 2 > MaxBuffSize {
		return nil, errors.New("data is too big")
	}

	newBuffer := make([]byte, len(buffer) * 2)
	copy(newBuffer, buffer)

	return newBuffer, nil
}

func (br *BufferReader) Read(further int) ([]byte, error) {
	rpos := br.rpos
	buf := br.buf
	exists := br.wpos - br.rpos
	copiedlen := 0

	if further > exists {
		buf = make([]byte, further)
		copy(buf, br.buf[br.rpos:br.wpos])
		rpos, copiedlen = 0, exists
	} else {
		br.rpos += further
	}

	for further > exists {
		if br.leftroom() == 0 {
			rpos, br.rpos, br.wpos = 0, 0, 0
		}

		if n, err := br.Reader.Read(br.buf[br.wpos:]); err != nil {
			return nil, err
		} else {
			br.wpos += n
			exists += n

			l := copy(buf[copiedlen:], br.buf[br.rpos:br.wpos])
			copiedlen += l
			br.rpos += l
		}
	}

	return buf[rpos:rpos+further], nil
}

func containLine(data []byte, beginpos int, endpos int) int {
	if beginpos > endpos {
		return -1
	}

	return containLine2(data[beginpos:endpos])
}

func containLine2(data []byte) int {
	length := len(data)
	if length < 2 {
		return -1
	}

	for i:=0; i<length-1; i++ {
		if data[i] == CR && data[i+1] == LF {
			return i+1
		}
	}

	return -1
}

func (br *BufferReader) ReadByte() (byte, error) {
	rpos := br.rpos
	if rpos < br.wpos {
		br.rpos += 1
		return br.buf[rpos], nil
	}

	if br.leftroom() == 0 {
		rpos, br.rpos, br.wpos = 0, 0, 0
	}

	n, err := br.Reader.Read(br.buf[br.wpos:])

	if err != nil {
		return 0, nil
	}

	br.wpos += n
	br.rpos += 1

	return br.buf[rpos], nil
}

type connReader struct{
	Conn net.Conn
}

func (c *connReader) Read(bytes []byte) (int, error) {
	return c.Conn.Read(bytes)
}
