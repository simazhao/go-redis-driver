package api

import (
	"github.com/simazhao/go-redis-driver/pkg/driver"
	"encoding/json"
	"sync"
	"time"
	"fmt"
)

type converter struct {

}

func newRequest() (*driver.Request) {
	return &driver.Request{Wait:&sync.WaitGroup{}}
}

func convertSetKeyValue(key string, val interface{}) (*driver.Request, error){
	if valbytes, err := json.Marshal(val); err != nil {
		return nil, err
	} else {
		request := newRequest()
		request.Clips = makeRequestClips(convertToBytes(driver.SET), convertToBytes(key), valbytes)
		return request, nil
	}
}



func newDuration(hours int64, minutes int64, seconds int64) time.Duration {
	return time.Duration(hours * int64(time.Hour) + minutes * int64(time.Minute) + seconds * int64(time.Second))
}

func convertSetKeyValueExpire(key string, val interface{}, dur time.Duration) (*driver.Request, error){
	if valbytes, err := json.Marshal(val); err != nil {
		return nil, err
	} else {
		request := newRequest()
		seconds := int64(dur/time.Second)
		request.Clips = makeRequestClips(convertToBytes(driver.SETEX), convertToBytes(key), convertIntToBytes(seconds), valbytes)
		return request, nil
	}
}

func convertGetKeyValue(key string) (*driver.Request, error) {
	request := newRequest()
	request.Clips = makeRequestClips(convertToBytes(driver.GET), convertToBytes(key))
	return request, nil
}

func convertMGetKeyValue(keys []string) (*driver.Request, error) {
	request := newRequest()
	request.Clips = make([]*driver.RequestClip, len(keys) + 1)
	request.Clips[0] = newRequestClip(convertToBytes(driver.MGET))

	for i, key := range keys {
		request.Clips[i+1] = newRequestClip(convertToBytes(key))
	}

	return request, nil
}

func convertMSetKeyValue(keyvals map[string]interface{}) (*driver.Request, error) {
	request := newRequest()
	request.Clips = make([]*driver.RequestClip, len(keyvals)*2 + 1)
	request.Clips[0] = newRequestClip(convertToBytes(driver.MSET))

	i := 1
	for key, val := range keyvals {
		request.Clips[i] = newRequestClip(convertToBytes(key))

		if valbytes, err := json.Marshal(val); err != nil {
			return nil, err
		} else {
			request.Clips[i+1] = newRequestClip(valbytes)
		}

		i += 2
	}

	return request, nil
}

func convertExpireAt(key string, dt time.Time) (*driver.Request, error) {
	request := newRequest()

	request.Clips = makeRequestClips(convertToBytes(driver.EXPIREAT), convertToBytes(key), convertIntToBytes(dt.Unix()))

	return request, nil
}

func convertIntToBytes(p int64) []byte {
	return []byte(fmt.Sprintf("%d", p))
}

func convertToBytes(p string) []byte {
	return []byte(p)
}

func makeRequestClips(params ...[]byte) ([]*driver.RequestClip){
	clips := make([]*driver.RequestClip, len(params))
	for i,param := range params {
		clips[i] = newRequestClip(param)
	}

	return clips
}

func newRequestClip(param []byte) *driver.RequestClip {
	return &driver.RequestClip{Type:driver.SegBulkBytes, Value:param}
}