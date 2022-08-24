package ae_ipfs_ds

import (
	"bytes"
	"context"
	"fmt"
	"github.com/cheggaaa/Anteater/aerpc"
	"github.com/cheggaaa/Anteater/aerpc/rpcclient"
	"github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/query"
	"github.com/valyala/fasthttp"
	"net/http"
	"net/rpc"
	"strconv"
	"strings"
)

type Config struct {
	BaseAddr string
	RpcAddr  string
}

func NewAEDS(c Config) (datastore.Batching, error) {
	aeds := &aeDs{
		baseAddr: c.BaseAddr,
	}
	addr := aerpc.NormalizeAddr(c.RpcAddr)
	var err error
	if aeds.rpc, err = rpcclient.NewClient(addr); err != nil {
		return nil, err
	}
	return aeds, nil
}

type aeDs struct {
	baseAddr string
	rpc      *rpc.Client
}

func (a *aeDs) Get(ctx context.Context, key datastore.Key) (value []byte, err error) {
	err = a.exec("GET", key, func(resp *fasthttp.Response) error {
		if resp.StatusCode() != http.StatusOK {
			return fmt.Errorf("not found")
		}
		buf := bytes.NewBuffer(make([]byte, 0, resp.Header.ContentLength()))
		resp.BodyWriteTo(buf)
		value = buf.Bytes()
		return nil
	})
	return
}

func (a *aeDs) Has(ctx context.Context, key datastore.Key) (exists bool, err error) {
	err = a.exec("HEAD", key, func(resp *fasthttp.Response) error {
		if resp.StatusCode() == http.StatusNoContent {
			exists = true
		}
		return nil
	})
	return
}

func (a *aeDs) GetSize(ctx context.Context, key datastore.Key) (size int, err error) {
	err = a.exec("HEAD", key, func(resp *fasthttp.Response) error {
		if resp.StatusCode() == http.StatusNoContent {
			var err error
			cl := resp.Header.Peek("X-Ae-Content-Length")
			size, err = strconv.Atoi(string(cl))
			if err != nil {
				return err
			}
		} else {
			return fmt.Errorf("not found")
		}
		return nil
	})
	return
}

func (a *aeDs) Query(ctx context.Context, q query.Query) (query.Results, error) {
	if q.Orders != nil || q.Filters != nil {
		return nil, fmt.Errorf("s3ds: filters or orders are not supported")
	}

	q.Prefix = strings.TrimPrefix(q.Prefix, "/")

	fl := &aerpc.RpcCommandFileList{}
	fl.SetArgs(strings.Split(q.Prefix, "/"))
	if q.Limit == 0 {
		q.Limit = 10000
	}

	index := q.Offset
	if err := fl.Execute(a.rpc); err != nil {
		return nil, err
	}
	list := fl.Data().([]string)
	nextValue := func() (query.Result, bool) {
		var err error
		if index >= len(list) {
			return query.Result{}, false
		}
		key := datastore.NewKey(list[index])

		entry := query.Entry{
			Key: key.String(),
		}
		if !q.KeysOnly {
			err = a.exec("GET", key, func(resp *fasthttp.Response) error {
				if resp.StatusCode() != http.StatusOK {
					return fmt.Errorf("not found")
				}
				entry.Size = resp.Header.ContentLength()
				buf := bytes.NewBuffer(make([]byte, 0, resp.Header.ContentLength()))
				resp.BodyWriteTo(buf)
				entry.Value = buf.Bytes()
				return nil
			})
			if err != nil {
				return query.Result{Error: err}, false
			}
		} else {
			if entry.Size, err = a.GetSize(ctx, key); err != nil {
				return query.Result{Error: err}, false
			}
		}
		index++
		return query.Result{Entry: entry}, index-q.Offset <= q.Limit
	}
	return query.ResultsFromIterator(q, query.Iterator{
		Close: func() error {
			return nil
		},
		Next: nextValue,
	}), nil
}

func (a *aeDs) Put(ctx context.Context, key datastore.Key, value []byte) error {
	var req = fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	var u = fasthttp.AcquireURI()
	defer fasthttp.ReleaseURI(u)
	u.SetHost(a.baseAddr)
	u.SetPath(key.String())
	req.SetURI(u)
	req.SetBody(value)
	req.Header.SetMethod("PUT")
	var resp = fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)
	if err := fasthttp.Do(req, resp); err != nil {
		return err
	}
	if resp.StatusCode() != http.StatusCreated {
		return fmt.Errorf("error: %v", resp.StatusCode())
	}
	return nil
}

func (a *aeDs) Delete(ctx context.Context, key datastore.Key) error {
	var req = fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)
	var u = fasthttp.AcquireURI()
	defer fasthttp.ReleaseURI(u)
	u.SetHost(a.baseAddr)
	u.SetPath(key.String())
	req.SetURI(u)
	req.Header.SetMethod("DELETE")
	var resp = fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)
	if err := fasthttp.Do(req, resp); err != nil {
		return err
	}
	if resp.StatusCode() != http.StatusNoContent {
		return fmt.Errorf("error: %v", resp.StatusCode())
	}
	return nil
}

func (a *aeDs) Sync(ctx context.Context, prefix datastore.Key) error {
	return nil
}

func (a *aeDs) Close() error {
	return nil
}

func (a *aeDs) Batch(ctx context.Context) (datastore.Batch, error) {
	return datastore.NewBasicBatch(a), nil
}

func (a *aeDs) exec(method string, key datastore.Key, cb func(resp *fasthttp.Response) error) error {
	var req = fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)
	var u = fasthttp.AcquireURI()
	defer fasthttp.ReleaseURI(u)
	u.SetHost(a.baseAddr)
	u.SetPath(key.String())
	req.SetURI(u)
	req.Header.SetMethod(method)
	var resp = fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)
	if err := fasthttp.Do(req, resp); err != nil {
		return err
	}
	return cb(resp)
}
