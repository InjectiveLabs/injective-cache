# Cache and cache

2 types of cache

### New Redis cache

```

```

### New libcache (in-process cache)

```
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
