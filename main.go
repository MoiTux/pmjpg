package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"path"
	"syscall"
	"time"
)

type Conf struct {
	Listen string
	Tpl    string
}

func main() {
	ctx, done := context.WithCancel(context.Background())
	defer done()

	pwd, err := os.Executable()
	if err != nil {
		fmt.Println("Can't find PWD:", err)
		os.Exit(1)
	}
	pwd = path.Dir(pwd)

	confFile, err := ioutil.ReadFile(pwd + "/conf.json")
	if err != nil {
		fmt.Println("Can't read conf:", err)
		os.Exit(1)
	}

	conf := Conf{}
	err = json.Unmarshal(confFile, &conf)
	if err != nil {
		fmt.Println("Error in conf file", err)
		os.Exit(1)
	}

	router := http.NewServeMux()
	router.Handle("/", Tpl(conf.Tpl))
	router.Handle("/image", Image(ctx))

	server := &http.Server{
		Addr:    conf.Listen,
		Handler: router,
	}

	go func() {
		err := server.ListenAndServe()
		if err != nil {
			fmt.Println(err)
		}
	}()

	fmt.Println("Ready, waiting for client")
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	done()

	server.SetKeepAlivesEnabled(false)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	server.Shutdown(ctx)
}
