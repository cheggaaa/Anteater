package ae_ipfs_ds

import (
	"context"
	"fmt"
	"github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/query"
	"testing"
)

var testConf = Config{
	BaseAddr: "127.0.0.1:8081",
}

var ctx = context.Background()

func TestAeDs(t *testing.T) {
	s, err := NewAEDS(testConf)
	if err != nil {
		t.Fatal(err)
	}
	err = s.Put(ctx, datastore.NewKey("my/key"), []byte("my value"))
	if err != nil {
		t.Fatal(err)
	}
	val, err := s.Get(ctx, datastore.NewKey("my/key"))
	if err != nil {
		t.Fatal(err)
	}
	if string(val) != "my value" {
		t.Errorf("unexpected get val: %s", string(val))
	}
	size, err := s.GetSize(ctx, datastore.NewKey("my/key"))
	if err != nil {
		t.Fatal(err)
	}
	if size != len("my value") {
		t.Errorf("unexpected size: %d", size)
	}
	for i := 0; i < 100; i++ {
		err = s.Put(ctx, datastore.NewKey(fmt.Sprintf("query/%d.txt", i)), []byte("my value"))
		if err != nil {
			t.Fatal(err)
		}
	}

	it, err := s.Query(ctx, query.Query{
		Prefix:   "query",
		Limit:    10,
		Offset:   10,
		KeysOnly: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	entries, err := it.Rest()
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 10 {
		t.Errorf("unexpected entries len: %d", len(entries))
	}
}
