package main

import (
	"goredis-driver/pkg/api"
	"time"
	"bufio"
	"os"
	"goredis-driver/pkg/config"
)

func main() {
	print(string([]byte("begin\r\n")))

	dapi := whynotrun3()

	reader := bufio.NewReader(os.Stdin)
	println("press YNA key #1")
	reader.ReadLine()

	dapi.Close()

	println("press YNA key #2")
	reader.ReadLine()
}

func whynotrun2() (*api.DriverApi){
	apipool := &api.DriverPool{}
	dapi := apipool.GetClientByConfig(config4test())

	if b,err := dapi.Set("who", "the king of fighter"); err != nil {
		println(err.Error())
	} else {
		println(b)
	}

	ts := api.NewTimeSpan3(0, 2, 30)

	if b,err := dapi.SetExpIn("best", "cao ji jin", ts.GetDuration()); err != nil {
		println(err.Error())
	} else {
		println(b)
	}

	if b, err := dapi.Get("who"); err != nil {
		println(err.Error())
	} else {
		println(b)
	}

	if b, err := dapi.Get("best"); err != nil {
		println(err.Error())
	} else {
		println(b)
	}

	if b, err := dapi.MGet("who", "best"); err != nil {
		println(err.Error())
	} else {
		println(b[0], b[1])
	}

	return dapi
}

func whynotrun3() (*api.DriverApi){
	apipool := &api.DriverPool{}
	dapi := apipool.GetClientByConfig(config4test())

	keyvalues := make(map[string]interface{})
	keyvalues["name"] = "a hua"
	keyvalues["age"] = 10



	if b,err := dapi.MSet(keyvalues); err != nil {
		println(err.Error())
	} else {
		println(b)
	}


	if b, err := dapi.MGet("name", "age"); err != nil {
		println(err.Error())
	} else {
		println(b[0], b[1])
	}

	return dapi
}

func config4test() Config.RedisConfig {
	return  Config.RedisConfig{
		RedisAddr:"192.168.144.128:6379",
		Timeout:time.Second*6,
	}
}