package driver

import (
	"go-redis-driver/pkg/config"
	"net"
	"errors"
	"go-redis-driver/pkg/log"
	"time"
)

type RedisDriver struct {
	Conn net.Conn
	DriverConfig Config.DriverConfig
	RedisConfig Config.RedisConfig
	State int

	requests chan *Request
}

const (
	cold = 0
	running = 1
	stopped = 2
)

func NewDriver(dconfig Config.DriverConfig, rconfig Config.RedisConfig) *RedisDriver{
	rdriver := &RedisDriver{DriverConfig:dconfig, RedisConfig:rconfig}
	rdriver.requests = make(chan *Request, dconfig.RequsetChanLength)

	rdriver.Run()

	return rdriver
}

func (dr *RedisDriver) Run() (err error) {
	if dr.State == running {
		return nil
	}

	if dr.State == stopped {
		return errors.New("driver has stopped")
	}

	dr.State = running

	go dr.tryWork()

	return nil
}

func (dr *RedisDriver) IsRunning() bool {
	return dr.State == running
}

func (dr *RedisDriver) tryWork()  {
	for true {
		shouldDelay := false
		if c, err := dr.connect(); err != nil {
			shouldDelay = true
		} else {
			dr.Conn = c
			if err := dr.work(c); err != nil{
				shouldDelay = true
			}
		}

		if !dr.IsRunning() {
			dr.disconnect()
			return
		}

		if shouldDelay {
			delay := time.After(time.Second * 30)
			select {
			case <-delay:
				continue
			case r, ok := <- dr.requests:
				if ok {
					r.Clips = nil
					r.Err = errors.New("redis do not connected")
					if r.Wait != nil  {
						r.Wait.Done()
					}
					log.Factory.GetLogger().WarnFormat( "request %s can not be handled", r)
				}
			}
		}
	}
}

func (dr *RedisDriver) work(conn net.Conn) error {
	defer func() {
		if len(dr.requests) > 0{
			for request := range dr.requests {
				log.Factory.GetLogger().WarnFormat( "request %s is abandoned", request)
			}
		}
	}()

	connbuffer := NewConnBuffer(conn, dr.DriverConfig.BufferSize)
	readChan := make(chan *Request, 1024)
	go dr.tryRead(readChan, connbuffer.BuffReader)


	for request := range dr.requests {
		if !dr.IsRunning() {
			return nil
		}

		// use same buffer to handle write
		if err := connbuffer.BuffWriter.WriteRequest(request); err != nil{
			log.Factory.GetLogger().ErrorFormat("handleRequest error: %s", err.Error())
			dr.setResponse(request, nil, err)
			return err
		} else {
		}

		log.Factory.GetLogger().Info("put request to read chan")
		readChan <- request
	}

	return errors.New("work abort")
}

func (dr *RedisDriver) tryRead(readChan chan *Request, reader *RedisReader) error {
	defer func() {
		if len(readChan) > 0 {
			for request := range readChan {
				log.Factory.GetLogger().ErrorFormat( "request %s is abandoned", request)
			}
		}
	}()

	for request := range readChan {
		if !dr.IsRunning() {
			return dr.setResponse(request, nil, errors.New("driver is not running"))
		}

		log.Factory.GetLogger().Info("read request from readChan")
		// use same buffer to handle read
		if response, err := reader.ReadResponse(); err != nil {
			dr.setResponse(request, nil, err)
		} else {
			dr.setResponse(request, response, nil)
		}
	}

	return errors.New("read abort")
}

func (dr *RedisDriver) setResponse(request *Request, res []*RequestClip, err error) error {
	request.Clips, request.Err = res, err

	if request.Wait != nil {
		request.Wait.Done()
	}

	return err
}

func (dr *RedisDriver) Stop() error {
	if dr.State == stopped {
		return nil
	}

	defer func() {
		dr.requests <- quit
	}()

	dr.State = stopped
	return nil
}

func (dr *RedisDriver) connect() (net.Conn, error){
	if c, err := net.DialTimeout("tcp", dr.RedisConfig.RedisAddr, dr.RedisConfig.Timeout); err != nil {
		return nil, err
	} else {
		return c, nil
	}
}

var quit = &Request{}

func (dr *RedisDriver) disconnect() error {
	return dr.Conn.Close()
}

func (dr *RedisDriver) Put(request *Request) error {
	if !dr.IsRunning() {
		return errors.New("driver is not running")
	}

	if request.Wait != nil {
		request.Wait.Add(1)
	}

	dr.requests <- request

	return nil
}



