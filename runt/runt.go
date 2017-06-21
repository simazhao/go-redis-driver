package main

import (
	"github.com/simazhao/go-redis-driver/pkg/api"
	"time"
	"bufio"
	"os"
	"github.com/simazhao/go-redis-driver/pkg/config"
	"encoding/json"
	"fmt"
)

func main() {
	print(string([]byte("begin\r\n")))

	dapi := whynotrun2()

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
	ts := api.NewTimeSpan3(0, 2, 30)

	if b,err := dapi.Set("who", "the king of fighter"); err != nil {
		println(err.Error())
	} else {
		println(b)
	}


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

	n := 500
	for i:=1;i<n;i++ {
		key := fmt.Sprintf("forcons-%d", i)
		val := makestringn("abcd-", i)
		//dapi.Set(key, val)
		if b, err := dapi.SetExpIn(key, val, ts.GetDuration()); !b || err != nil {
			println("set fail", key, b)
		}

		//time.Sleep(api.NewTimeSpan1(0,0,0, 0,10).GetDuration())

		if real, err := dapi.Get(key); err != nil {
			println("get key error:", key, err.Error())
			break
		} else if fmt.Sprintf("\"%s\"", val) == real{
			println("get key ok:", key)
		} else {
			println("get key incorrect:", key)
			break
		}
	}

	println("all over")

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

func makestringn(content string, n int) string {
	ret := content
	for i:=1;i<n;i++ {
		ret += content
	}

	return ret
}

type Cell struct {
	Width int
	Height int
	Length int
}

func whynotrun4() (*api.DriverApi){
	apipool := &api.DriverPool{}
	dapi := apipool.GetClientByConfig(config4test())

	if b,err := dapi.Set("cell", Cell{Width:1,Height:2,Length:3}); err != nil {
		println(err.Error())
	} else {
		println(b)
	}

	if b, err := dapi.Get("cell"); err != nil {
		println(err.Error())
	} else {
		var cell Cell
		if err := json.Unmarshal([]byte(b), &cell); err != nil {
			print("wrong value type of key")
		} else {
			print(cell.Width, cell.Height, cell.Length)
		}
	}

	return dapi
}

func config4test() Config.RedisConfig {
	return  Config.RedisConfig{
		RedisAddr:"192.168.144.128:6379",
		Timeout:time.Second*6,
	}
}