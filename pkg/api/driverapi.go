package api

import (
	"goredis-driver/pkg/driver"
	"goredis-driver/pkg/config"
	"goredis-driver/pkg/log"
	"time"
)

type IDriverApi interface {
	Set(key string, val interface{}) (bool, error)

	Get(key string) (interface{}, error)

	Close() error
}

type DriverPool struct {

}

func (dp *DriverPool) loadConfig() Config.RedisConfig {
	return Config.RedisConfig{}
}

func (dp *DriverPool) GetClient() *DriverApi {
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

func (d *DriverApi) Set(key string, val interface{}) (bool, error) {
	if request, err := convertSetKeyValue(key, val); err != nil {
		return false, err
	} else {
		d.driver.Put(request)
		log.Factory.GetLogger().Info("put request to driver")
		request.Wait.Wait()

		log.Factory.GetLogger().Info("request process over")
		if request.Err != nil {
			println(request.Err.Error())
			return false, request.Err
		}

		if len(request.Clips) > 0 {
			if string(request.Clips[0].Value) != "OK" {
				return false, nil
			}
		}

		return true, nil
	}
}

func (d *DriverApi) SetExpIn(key string, val interface{}, dur time.Duration) (bool, error){
	if request, err := convertSetKeyValueExpire(key, val, dur); err != nil {
		return false, err
	} else {
		d.driver.Put(request)
		log.Factory.GetLogger().Info("put request to driver")
		request.Wait.Wait()

		log.Factory.GetLogger().Info("request process over")
		if request.Err != nil {
			println(request.Err.Error())
			return false, request.Err
		}

		if len(request.Clips) > 0 {
			if string(request.Clips[0].Value) != "OK" {
				return false, nil
			}
		}

		return true, nil
	}
}


func (d *DriverApi) Get(key string) (string, error) {
	if request, err := convertGetKeyValue(key); err != nil {
		return "", err
	} else {
		d.driver.Put(request)
		log.Factory.GetLogger().Info("put request to driver")
		request.Wait.Wait()

		log.Factory.GetLogger().Info("request process over")
		if request.Err != nil {
			println(request.Err.Error())
			return "", request.Err
		}

		if len(request.Clips) == 0 {
			return "", nil
		}

		log.Factory.GetLogger().Info(string(request.Clips[0].Value))

		return string(request.Clips[0].Value), nil
	}
}

func (d *DriverApi) MGet(keys ...string) ([]string, error) {
	if request, err := convertMGetKeyValue(keys); err != nil {
		return nil, err
	} else {
		d.driver.Put(request)
		log.Factory.GetLogger().Info("put request to driver")
		request.Wait.Wait()

		log.Factory.GetLogger().Info("request process over")
		if request.Err != nil {
			println(request.Err.Error())
			return nil, request.Err
		}

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
	} else {
		d.driver.Put(request)
		log.Factory.GetLogger().Info("put request to driver")
		request.Wait.Wait()

		log.Factory.GetLogger().Info("request process over")
		if request.Err != nil {
			println(request.Err.Error())
			return false, request.Err
		}

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