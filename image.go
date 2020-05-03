package main

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"strings"
	"time"
)

func setupCam(ctx context.Context, prm Params, cam chan<- Cam, cleanIdx chan<- string) {
	var err error
	defer func() {
		cleanIdx <- prm.Idx
		if err != nil {
			cam <- Cam{
				err: err,
			}
		}
	}()

	request, err := http.NewRequestWithContext(ctx, "GET", prm.Addr, nil)
	if err != nil {
		fmt.Println("Error: REQUEST", err)
		return
	}

	request.SetBasicAuth(prm.User, prm.Pwd)

	c := &http.Client{}
	resp, err := c.Do(request)
	if err != nil {
		fmt.Println("Error: CLIENT", err)
		return
	}
	defer resp.Body.Close()
	// can't has body will always have cam
	// defer io.Copy(ioutil.Discard, resp.Body)
	if resp.StatusCode != 200 {
		fmt.Println("Bad request: ", resp.Status)
		err = fmt.Errorf("StatusCode != 200, go '%s'", resp.Status)
		return
	}

	ct := strings.Split(resp.Header.Get("Content-Type"), ";")
	if ct[0] != "multipart/x-mixed-replace" {
		fmt.Println("not multipart/x-mixed-replace")
		err = fmt.Errorf("Bad Content-Type: '%s'", resp.Header.Get("Content-Type"))
		return
	}
	boundary := strings.Split(ct[1], "=")[1]
	mr := multipart.NewReader(resp.Body, boundary)

	for {
		p, err := mr.NextPart()
		if err == io.EOF {
			fmt.Println("EOF")
			return
		}
		if err != nil {
			fmt.Println("NextPart: ", err)
			return
		}

		if p.Header.Get("Content-Type") != "image/jpeg" {
			fmt.Println("not image/jpeg")
			err = fmt.Errorf("Bad Content-Type: '%s'", resp.Header.Get("Content-Type"))
			return
		}

		b, err := ioutil.ReadAll(p)
		if err != nil {
			fmt.Println("Read: ", err)
			return
		}
		select {
		case cam <- Cam{buff: b}:
		case <-ctx.Done():
			err = fmt.Errorf("ctx Done")
			return
		case <-time.After(5 * time.Second):
			err = fmt.Errorf("Timeout")
			return
		}
		if prm.Idx == "" {
			return
		}
	}
}

type Cam struct {
	buff []byte
	err  error
}

func Image(ctx context.Context) http.Handler {
	clients := map[string]chan Cam{}

	cleanIdx := make(chan string)
	go func() {
		for {
			select {
			case idx := <-cleanIdx:
				delete(clients, idx)
			case <-ctx.Done():
				for len(clients) > 0 {
					delete(clients, <-cleanIdx)
				}
				return
			}
		}
	}()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")

		if r.Method != "GET" {
			http.NotFound(w, r)
		}

		prm, err := GetParam(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		cam, ok := clients[prm.Idx]
		if !ok {
			cam = make(chan Cam)
			clients[prm.Idx] = cam
			go setupCam(ctx, prm, cam, cleanIdx)
		}

		data := <-cam
		if data.err != nil {
			http.Error(w, data.err.Error(), http.StatusInternalServerError)
		}
		w.Write(data.buff)
	})
}
