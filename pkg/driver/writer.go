package driver

import (
	"net"
	"strconv"
	"errors"
	"io"
	"fmt"
)


type RedisWriter struct {
	Writer *BufferWriter
}

func (e *RedisWriter) WriteRequest(r *Request) (err error) {
	if err = e.writeTotalLength(len(r.Clips)); err != nil {
		return
	}

	for _, clip := range r.Clips {
		if err = e.writeRequestClip(clip); err != nil {
			return
		}
	}

	_, err = e.Writer.Flush()

	return
}

func (e *RedisWriter) writeTotalLength(length int) (err error){
	if _, err = e.Writer.WriteByte(byte(SegArray)); err != nil {
		return
	}

	if err = e.writeInt(length); err != nil {
		return
	}

	if err = e.writeEndLine(); err != nil{
		return
	}

	return
}

func (e* RedisWriter) writeInt(n int) error {
	return e.WriteString(strconv.Itoa(n))
}

func (e *RedisWriter) WriteString(str string) error {
	_, err := e.Writer.Write([]byte(str))

	return err
}

const (
	CR = '\r'
	LF = '\n'
	CRLF = "\r\n"
)

var crlfBytes = []byte(CRLF)

func (e *RedisWriter) writeEndLine() (err error) {
	_, err = e.Writer.Write(crlfBytes)

	return
}

func (e *RedisWriter) writeRequestClip(r *RequestClip) (err error) {
	switch r.Type {
	case SegBulkBytes:
		return e.writeBulkRequest(r)
	case SegInt:
	case SegString:
		return e.WriteText(r)
	default:
		return errors.New(fmt.Sprint("do not supported request type :%s" , r.Type))
	}

	return
}

func (e *RedisWriter) writeBulkRequest(r *RequestClip) (err error){
	if err = e.writeRequstClipLength(len(r.Value)); err != nil {
		return
	}

	if _, err = e.Writer.Write(r.Value); err != nil {
		return
	}

	if err = e.writeEndLine(); err != nil{
		return
	}

	return
}

func (e *RedisWriter) writeRequstClipLength(length int) (err error) {
	if _, err = e.Writer.WriteByte(byte(SegBulkBytes)); err != nil {
		return
	}

	if err = e.writeInt(length); err != nil {
		return
	}

	if err = e.writeEndLine(); err != nil{
		return
	}

	return
}

func (e *RedisWriter) WriteText(r* RequestClip) (err error) {
	if _, err = e.Writer.Write(r.Value); err != nil {
		return
	}

	return e.writeEndLine()
}


type BufferWriter struct {
	Writer *connWriter
	buf []byte
	wpos int
}

func NewBufferWriter(conn net.Conn, bufferSize int) *BufferWriter {
	buffer := BufferWriter{Writer:&connWriter{Conn:conn}}
	buffer.buf = make([]byte, bufferSize)
	return &buffer
}

func (bw *BufferWriter) capacity() int {
	return len(bw.buf)
}

func (bw *BufferWriter) leftroom() int {
	return bw.capacity() - bw.wpos
}

func (bw *BufferWriter) WriteByte(s byte) (n int, err error) {
	if bw.leftroom() == 0 {
		if _, err = bw.Flush(); err != nil {
			return
		}
	}

	bw.buf[bw.wpos] = s
	bw.wpos += 1

	return 1, nil
}

func (bw *BufferWriter) Write(bytes []byte) (n int, err error) {
	for err == nil && len(bytes) > bw.leftroom(){
		nc := copy(bw.buf[bw.wpos:], bytes)
		bw.wpos += nc
		_, err = bw.Flush()

		n, bytes = n+nc, bytes[nc:]
	}

	if err != nil {
		return
	}

	if len(bytes) == 0 {
		err = errors.New("incorrect bytes length")
		return
	}

	nc := copy(bw.buf[bw.wpos:], bytes)
	n += nc
	bw.wpos += nc
	return
}

func (bw *BufferWriter) Flush() (n int, err error) {
	if bw.wpos == 0 {
		return
	}

	n, err = bw.Writer.Write(bw.buf[:bw.wpos])

	if err == nil {
		if n < bw.wpos {
			err = io.ErrShortWrite
		} else {
			bw.wpos = 0
		}
	}

	return
}


type connWriter struct {
	Conn net.Conn
}

func (w *connWriter) Write(bytes []byte) (int, error) {
	if w.Conn == nil {
		return 0, errors.New("no connection")
	}

	return w.Conn.Write(bytes)
}