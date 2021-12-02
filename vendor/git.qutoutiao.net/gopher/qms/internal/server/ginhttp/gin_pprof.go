package ginhttp

import (
	"net/http"
	"net/http/pprof"
	"runtime"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
)

func ginIndex(c *gin.Context) {
	pprof.Index(c.Writer, c.Request)
}

func ginCmdline(c *gin.Context) {
	pprof.Cmdline(c.Writer, c.Request)
}

func ginProfile(c *gin.Context) {
	pprof.Profile(c.Writer, c.Request)
}

func ginSymbol(c *gin.Context) {
	pprof.Profile(c.Writer, c.Request)
}

func ginTrace(c *gin.Context) {
	pprof.Profile(c.Writer, c.Request)
}

var inBlock, inMutex uint32

const (
	idling    uint32 = 0
	inService uint32 = 1
)

// ginBlock will pass the call from /debug/pprof/block to pprof
func ginBlock(c *gin.Context) {
	if !atomic.CompareAndSwapUint32(&inBlock, idling, inService) {
		c.String(http.StatusConflict, "in service")
		return
	}
	defer atomic.StoreUint32(&inBlock, idling)

	sec, _ := strconv.ParseInt(c.Query("seconds"), 10, 64)
	if sec == 0 {
		sec = 30
	}
	rate, _ := strconv.Atoi(c.Query("rate"))
	if rate == 0 {
		rate = 1000000
	}
	runtime.SetBlockProfileRate(rate)
	defer runtime.SetBlockProfileRate(0)

	time.Sleep(time.Duration(sec) * time.Second)
	pprof.Index(c.Writer, c.Request)
}

func ginMutex(c *gin.Context) {
	if !atomic.CompareAndSwapUint32(&inMutex, idling, inService) {
		c.String(http.StatusConflict, "in service")
		return
	}
	defer atomic.StoreUint32(&inMutex, idling)

	sec, _ := strconv.ParseInt(c.Query("seconds"), 10, 64)
	if sec == 0 {
		sec = 30
	}
	rate, _ := strconv.Atoi(c.Query("rate"))
	if rate == 0 {
		rate = 1000
	}
	runtime.SetMutexProfileFraction(rate)
	defer runtime.SetMutexProfileFraction(0)
	time.Sleep(time.Duration(sec) * time.Second)
	pprof.Index(c.Writer, c.Request)
}

// ginHeap will pass the call from /debug/pprof/heap to pprof
func ginHeap(c *gin.Context) {
	pprof.Handler("heap").ServeHTTP(c.Writer, c.Request)
}

// ginGoroutine will pass the call from /debug/pprof/goroutine to pprof
func ginGoroutine(c *gin.Context) {
	pprof.Handler("goroutine").ServeHTTP(c.Writer, c.Request)
}
