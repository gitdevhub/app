package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"runtime/pprof"
	"time"

	"github.com/gin-gonic/gin"
)

func init() {
	config.Setup()
	logger.Setup()
	db.Setup()
	model.Setup()
	gredis.Setup()
	racelimit.Setup()
	mailchecker.Setup()
	geoip.Setup()
	profanityfilter.Setup()
	racelimit.Setup()
	jwt.InitJWTAuthentication()
}

func main() {
	defer geoip.Close()
	defer logger.Close()
	defer db.CloseDB()

	var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to `file`")
	var memprofile = flag.String("memprofile", "", "write memory profile to `file`")

	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			logger.Fatal("could not create CPU profile: ", err)
		}
		if err := pprof.StartCPUProfile(f); err != nil {
			logger.Fatal("could not start CPU profile: ", err)
		}
		defer pprof.StopCPUProfile()
	}

	gin.SetMode(config.Server.Mode)

	if config.Server.Mode == gin.ReleaseMode {
		gin.DisableConsoleColor()
	}

	// Logging to a file.
	f, _ := os.Create("gin.log")
	gin.DefaultWriter = io.MultiWriter(f, os.Stdout)

	routersInit := router.InitRouter()
	port := fmt.Sprintf(":%d", config.Server.Port)
	readTimeout := config.Server.ReadTimeout * time.Second
	writeTimeout := config.Server.WriteTimeout * time.Second
	maxHeaderBytes := config.App.MaxHeaderBytes << 20 // 2 MiB

	// Set a lower memory limit for multipart forms (default is 32 MiB)
	// app.MaxMultipartMemory = config.Server.MaxMultipartMemory << 20 // 10 MiB

	logger.Info(fmt.Sprintf("[info] start http server listening %s", port))
	logger.Info(fmt.Sprintf("[info] readTimeout %s", readTimeout))
	logger.Info(fmt.Sprintf("[info] writeTimeout %s", writeTimeout))
	logger.Info(fmt.Sprintf("[info] maxHeaderBytes %d", maxHeaderBytes))

	server := &http.Server{
		Addr:    port,
		Handler: routersInit,
		// ReadTimeout:    readTimeout,
		// WriteTimeout:   writeTimeout,
		MaxHeaderBytes: maxHeaderBytes,
	}

	_ = server.ListenAndServe()

	if *memprofile != "" {
		f, err := os.Create(*memprofile)
		if err != nil {
			logger.Fatal("could not create memory profile: ", err)
		}
		runtime.GC() // get up-to-date statistics
		if err := pprof.WriteHeapProfile(f); err != nil {
			logger.Fatal("could not write memory profile: ", err)
		}
		_ = f.Close()
	}
}
