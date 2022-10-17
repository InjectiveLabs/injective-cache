# Cache and cache

2 types of cache under 1 interface

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
