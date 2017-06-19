package driver

import "net"

type ConnBuffer struct {
	BuffReader *RedisReader
	BuffWriter *RedisWriter
}

func NewConnBuffer(conn net.Conn, bufferSize int) *ConnBuffer{
	buffer := &ConnBuffer{}
	buffer.BuffReader = &RedisReader{Reader:NewBufferReader(conn, bufferSize)}
	buffer.BuffWriter = &RedisWriter{Writer:NewBufferWriter(conn, bufferSize)}
	return buffer
}
