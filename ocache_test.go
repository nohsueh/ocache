package ocache

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"reflect"
	"testing"
)

func Test_Getter(t *testing.T) {
	f := GetterFunc(func(key string) ([]byte, error) {
		return []byte(key), nil
	})

	expect := []byte("key")
	if bytes, _ := f.Get("key"); !reflect.DeepEqual(bytes, expect) {
		t.Errorf("callback failed")
	}
}

var db = map[string]string{
	"Tom":  "630",
	"Jack": "589",
	"Sam":  "567",
}

func Test_Get(t *testing.T) {
	loadCounts := make(map[string]int, len(db))
	r := NewRelation("Person", 1<<10, GetterFunc(
		func(key string) ([]byte, error) {
			log.Println("[SlowDB] Search key", key)
			if v, ok := db[key]; ok {
				if _, ok := loadCounts[key]; !ok {
					loadCounts[key] = 0
				}
				loadCounts[key] += 1
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s not exist", key)
		},
	))

	for k, v := range db {
		// load from callback function
		if view, err := r.Get(k); err != nil || view.String() != v {
			t.Fatal("Failed to get bytes of Tom")
		}
		// Cache hit
		if _, err := r.Get(k); err != nil || loadCounts[k] > 1 {
			t.Fatalf("Cache %s miss", k)
		}
	}

	if value, err := r.Get("unknown"); err == nil {
		t.Fatalf("The bytes of unknow should be empty, but %s got", value)
	}
}

func createRelation() *Relation {
	return NewRelation("Person", 2<<10, GetterFunc(
		func(key string) ([]byte, error) {
			log.Println("[SlowDB] Search key", key)
			if v, ok := db[key]; ok {
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s not exist", key)
		}))
}

func startCacheServer(addr string, addrs []string, r *Relation) {
	peers := NewHTTPPool(addr)
	peers.Set(addrs...)
	r.RegisterPeers(peers)
	log.Println("Ocache is running at", addr)
	log.Fatal(http.ListenAndServe(addr[7:], peers))
}

func startAPIServer(apiAddr string, c *Relation) {
	http.Handle("/api", http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			key := r.URL.Query().Get("key")
			view, err := c.Get(key)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/octet-stream")
			_, err = w.Write(view.ByteSlice())
			if err != nil {
				return
			}
		}))
	log.Println("Frontend server is running at", apiAddr)
	log.Fatal(http.ListenAndServe(apiAddr[7:], nil))

}

func Test_Server(t *testing.T) {
	var port int
	var api bool
	flag.IntVar(&port, "port", 8001, "Ocache server port")
	flag.BoolVar(&api, "api", false, "Start a api server?")
	flag.Parse()

	apiAddr := "http://localhost:9999"
	addrMap := map[int]string{
		8001: "http://localhost:8001",
		8002: "http://localhost:8002",
		8003: "http://localhost:8003",
	}

	var addrs []string
	for _, v := range addrMap {
		addrs = append(addrs, v)
	}

	r := createRelation()
	if api {
		go startAPIServer(apiAddr, r)
	}
	startCacheServer(addrMap[port], addrs, r)
}
