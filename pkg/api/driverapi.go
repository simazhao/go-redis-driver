package api

import (
	"github.com/simazhao/go-redis-driver/pkg/driver"
	"github.com/simazhao/go-redis-driver/pkg/config"
	"github.com/simazhao/go-redis-driver/pkg/log"
	"time"
)

type IDriverApi interface {
	Set(key string, val interface{}) (bool, error)

	SetExpIn(key string, val interface{}, dur time.Duration) (bool, error)

	Get(key string) (string, error)

	MGet(keys ...string) ([]string, error)

	MSet(keyvals map[string]interface{}) (bool, error)

	Close() error
}

type DriverPool struct {

}

func (dp *DriverPool) loadConfig() Config.RedisConfig {
	return Config.RedisConfig{}
}

func (dp *DriverPool) GetClient() IDriverApi {
	api := &DriverApi{}
	api.driver = driver.NewDriver(Config.DefaultDriverConfig(), dp.loadConfig())
	return api
}

func (dp *DriverPool) GetClientByConfig(config Config.RedisConfig) *DriverApi {
	api := &DriverApi{}
	api.driver = driver.NewDriver(Config.DefaultDriverConfig(), config)
	return api
}

type DriverApi struct {
	IDriverApi

	driver *driver.RedisDriver
}

func (d *DriverApi) process(request *driver.Request) error {
	d.driver.Put(request)
	log.Factory.GetLogger().Info("put request to driver")
	request.Wait.Wait()

	log.Factory.GetLogger().Info("request process over")
	if request.Err != nil {
		return request.Err
	}

	return nil
}

func (d *DriverApi) Set(key string, val interface{}) (bool, error) {
	if request, err := convertSetKeyValue(key, val); err != nil {
		return false, err
	} else if err := d.process(request); err != nil {
		return false, err
	} else {
		return len(request.Clips) > 0 && string(request.Clips[0].Value) == "OK", nil
	}
}

func (d *DriverApi) SetExpIn(key string, val interface{}, dur time.Duration) (bool, error){
	if request, err := convertSetKeyValueExpire(key, val, dur); err != nil {
		return false, err
	} else if err := d.process(request); err != nil {
		return false, err
	} else {
		return len(request.Clips) > 0 && string(request.Clips[0].Value) == "OK", nil
	}
}

func (d *DriverApi) Get(key string) (string, error) {
	if request, err := convertGetKeyValue(key); err != nil {
		return "", err
	} else if err := d.process(request); err != nil {
		return "", err
	} else {
		if len(request.Clips) == 0 || len(request.Clips[0].Value) == 0{
			return "", nil
		}

		return string(request.Clips[0].Value), nil
	}
}

func (d *DriverApi) MGet(keys ...string) ([]string, error) {
	if request, err := convertMGetKeyValue(keys); err != nil {
		return nil, err
	} else if err := d.process(request); err != nil {
		return nil, err
	} else {
		if len(request.Clips) == 0 {
			return nil, nil
		}

		vals := make([]string, len(request.Clips))

		for i, clip := range request.Clips {
			vals[i] = string(clip.Value)
 		}

		return vals, nil
	}
}

func (d *DriverApi) MSet(keyvals map[string]interface{}) (bool, error) {
	if request, err := convertMSetKeyValue(keyvals); err != nil {
		return false, err
	} else if err := d.process(request); err != nil {
		return false, err
	}else {
		if len(request.Clips) == 0 {
			return false, nil
		}

		if len(request.Clips) > 0 {
			if string(request.Clips[0].Value) != "OK" {
				return false, nil
			}
		}

		return true, nil
	}
}

func (d *DriverApi) Close() error {
	return d.driver.Stop()
}