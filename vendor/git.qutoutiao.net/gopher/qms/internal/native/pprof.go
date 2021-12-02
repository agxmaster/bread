package native

import (
	"net/http"
	"net/http/pprof"
	"runtime"
	"strconv"
	"sync/atomic"
	"time"
)

var inBlock, inMutex uint32

const (
	idling    uint32 = 0
	inService uint32 = 1
)

func block(w http.ResponseWriter, r *http.Request) {
	if !atomic.CompareAndSwapUint32(&inBlock, idling, inService) {
		w.WriteHeader(http.StatusConflict)
		w.Write([]byte("in service"))
		return
	}
	defer atomic.StoreUint32(&inBlock, idling)

	sec, _ := strconv.ParseInt(r.FormValue("seconds"), 10, 64)
	if sec == 0 {
		sec = 30
	}
	rate, _ := strconv.Atoi(r.FormValue("rate"))
	if rate == 0 {
		rate = 1000000
	}
	runtime.SetBlockProfileRate(rate)
	defer runtime.SetBlockProfileRate(0)

	time.Sleep(time.Duration(sec) * time.Second)
	pprof.Index(w, r)
}

func mutex(w http.ResponseWriter, r *http.Request) {
	if !atomic.CompareAndSwapUint32(&inMutex, idling, inService) {
		w.WriteHeader(http.StatusConflict)
		w.Write([]byte("in service"))
		return
	}
	defer atomic.StoreUint32(&inMutex, idling)

	sec, _ := strconv.ParseInt(r.FormValue("seconds"), 10, 64)
	if sec == 0 {
		sec = 30
	}
	rate, _ := strconv.Atoi(r.FormValue("rate"))
	if rate == 0 {
		rate = 1000
	}
	runtime.SetMutexProfileFraction(rate)
	defer runtime.SetMutexProfileFraction(0)
	time.Sleep(time.Duration(sec) * time.Second)
	pprof.Index(w, r)
}

// heap will pass the call from /debug/pprof/heap to pprof
func heap(w http.ResponseWriter, r *http.Request) {
	pprof.Handler("heap").ServeHTTP(w, r)
}

// goroutine will pass the call from /debug/pprof/goroutine to pprof
func goroutine(w http.ResponseWriter, r *http.Request) {
	pprof.Handler("goroutine").ServeHTTP(w, r)
}
