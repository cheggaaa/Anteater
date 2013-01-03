package storage

import (
	"config"
	"utils"
	"testing"
	mrand "math/rand"
	"fmt"
	"io"
)


func storageBConf() *config.Config {
	mb, _ := utils.BytesFromString("1m")
	return &config.Config{
		// Data path
		DataPath : "",
		ContainerSize : 10 * mb,
	}
}

func Benchmark_Add100B(b *testing.B) {
	b.StopTimer()
	conf := storageBConf()
	st := GetStorage(conf)
	benchmarkAdd(st, b, 100)
	st.Drop()
}

func Benchmark_Update100B(b *testing.B) {
	b.StopTimer()
	conf := storageBConf()
	st := GetStorage(conf)
	benchmarkUpdate(st, b, 100)
	st.Drop()
}

func Benchmark_Add100K(b *testing.B) {
	b.StopTimer()
	conf := storageBConf()
	conf.ContainerSize = 1024 * 1024 * 200
	st := GetStorage(conf)
	benchmarkAdd(st, b, 100 * 1024)
	st.Drop()
}

func Benchmark_Update100K(b *testing.B) {
	b.StopTimer()
	conf := storageBConf()
	conf.ContainerSize = 1024 * 1024 * 200
	st := GetStorage(conf)
	benchmarkUpdate(st, b, 100 * 1024)
	st.Drop()
}

func Benchmark_Add1Mb(b *testing.B) {
	b.StopTimer()
	conf := storageBConf()
	conf.ContainerSize = 1024 * 1024 * 200
	st := GetStorage(conf)
	benchmarkAdd(st, b, 1024 * 1024)
	st.Drop()
}

func Benchmark_AddRand(b *testing.B) {
	b.StopTimer()
	conf := storageBConf()
	conf.ContainerSize = 1024 * 1024 * 200
	st := GetStorage(conf)
	benchmarkAdd(st, b, 0)
	st.Drop()
}

func Benchmark_UpdateRand(b *testing.B) {
	b.StopTimer()
	conf := storageBConf()
	conf.ContainerSize = 1024 * 1024 * 200
	st := GetStorage(conf)
	benchmarkUpdate(st, b, 0)
	st.Drop()
}


func benchmarkUpdate(st *Storage, b *testing.B, size int) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
	 	var sz int64
		if size == 0 {
			sz = int64(mrand.Intn(1023 * 1024) + 1024)
		} else {
			sz = int64(size)
		}
		f := randReader(sz)
		n := fmt.Sprintf("b-%d", mrand.Intn(99))
		b.StartTimer()
		st.Delete(n)
		st.Add(n, f, sz)
		b.StopTimer()
		b.SetBytes(sz)
	}
}


func benchmarkAdd(st *Storage, b *testing.B, size int) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
	 	var sz int64
		if size == 0 {
			sz = int64(mrand.Intn(1023 * 1024) + 1024)
		} else {
			sz = int64(size)
		}
		f := randReader(sz)
		n := fmt.Sprintf("b-%d", i)
		b.StartTimer()
		st.Add(n, f, sz)
		b.StopTimer()
		b.SetBytes(sz)
	}
}

func Benchmark_Get100B(b *testing.B) {
	b.StopTimer()
	conf := storageBConf()
	st := GetStorage(conf)
	benchmarkGet(st, b, 100)
	st.Drop()
}

func Benchmark_Get100K(b *testing.B) {
	b.StopTimer()
	conf := storageBConf()
	st := GetStorage(conf)
	benchmarkGet(st, b, 100 * 1024)
	st.Drop()
}

func Benchmark_Get1Mb(b *testing.B) {
	b.StopTimer()
	conf := storageBConf()
	st := GetStorage(conf)
	benchmarkGet(st, b, 1024 * 1024)
	st.Drop()
}


func benchmarkGet(st *Storage, b *testing.B, size int64) {
	var n = "test"
	st.Add(n, randReader(size), size)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := &CWriter{}
		b.StartTimer()
		f,_ := st.Get(n)
		io.Copy(w, f.GetReader())
		b.StopTimer()
		b.SetBytes(int64(w.L))
	}
}

func BenchmarkDelete(b *testing.B) {
	b.StopTimer()
	conf := storageBConf()
	st := GetStorage(conf)
	benchmarkDelete(st, b)
	st.Drop()
}


func benchmarkDelete(st *Storage, b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		n := fmt.Sprintf("b-%d", i)
		st.Add(n, randReader(1), 1)
		b.StartTimer()
		st.Delete(n)
		b.StopTimer()
	}
}


type CWriter struct {
	L int
}
func (w *CWriter) Write(b []byte) (n int, err error) {
	n = len(b)
	w.L += n
	return
}