package driver

import "sync"

type SegmentType byte

const (
	SegString    SegmentType = '+'
	SegError     SegmentType = '-'
	SegInt       SegmentType = ':'
	SegBulkBytes SegmentType = '$'
	SegArray     SegmentType = '*'
)


type Request struct{
	Clips []*RequestClip
	Wait *sync.WaitGroup

	Err error
}

type RequestClip struct {
	Type SegmentType
	Value []byte
}




