# Cache and cache

2 types of cache under 1 interface

### SimpleCache (V2)
Simple cache interface allows to store and retrieve arbitrary types from the cache.

There are 2 implementations of the SimpleCache interface:
- SimpleRedisCache
- TypedLibCache

#### Redis Simple Cache
With the Redis Simple Cache, you can choose the way to encode and decode the data to be stored in the cache.
If you don't provide a codec, the default codec will be used, which is the `json` codec.

#### Typed Lib Cache
With this cache you can store any type in memory, without having to serialize the data, with the caveat that you have to declare the type when you create the cache.

### Resource Coalescing Cache
Resource Coalescing Cache is a cache that allows you to coalesce the requests for the same resource,
this means that if there are multiple requests for the same resource, only one request will be made to the backend, 
and the rest of the requests will wait for the first request to finish.

It includes a cache, so once the first request finishes, it will set the value in the cache and, while is valid, new requests
will return the value from the cache.

### New Redis cache

```go
cache, err := NewRedisCache(context.Background(), "localhost:6379", 5*time.Second)
if err != nil {
    panic(err)
}

cache.Set("key1", []byte("hello1"))
cache.Set("key2", []byte("hello2"))

v, err := cache.Get("key2")
if err != nil {
    panic(err)
}

fmt.Println(v)
```

### New libcache (in-process cache)

```go
cache, err := NewLibcache(0, 5*time.Second)
if err != nil {
    panic(err)
}

cache.Set("key1", []byte("hello1"))
cache.Set("key2", []byte("hello2"))

v, err := cache.Get("key2")
if err != nil {
    panic(err)
}

fmt.Println(v)
```
