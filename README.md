# go-redis-driver

##### What's that?
a redis client written by golang

##### How to use?
````
apipool := &api.DriverPool{}

dapi := apipool.GetClientByConfig(yourconfig)

dapi.Set("who", "the king of fighter")

dapi.Get("who")
````
