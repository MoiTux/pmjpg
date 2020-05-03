package main

import (
	"fmt"
	"net/http"
	"os"
	"regexp"
	"text/template"
)

func Tpl(path string) http.Handler {
	fmt.Println("searching tpl file in:", path)

	dir, err := os.Open(path)
	if err != nil {
		fmt.Println("can't read path:", err)
		os.Exit(1)
	}

	files, err := dir.Readdirnames(0)
	if err != nil {
		fmt.Println("Can't find files:", err)
		os.Exit(1)
	}

	re := regexp.MustCompile(`^[a-zA-Z0-9]*\.html$`)
	var tpls []string
	for _, f := range files {
		if !re.MatchString(f) {
			continue
		}
		fmt.Println("registering:", f)
		tpls = append(tpls, path+"/"+f)
	}

	if tpls == nil {
		fmt.Println("No tpl file find")
		os.Exit(1)
	}

	tpl, err := template.ParseFiles(tpls...)
	if err != nil {
		fmt.Println("Can't read tpls files")
		os.Exit(1)
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")

		if r.Method != "GET" {
			http.NotFound(w, r)
			return
		}

		t := tpl.Lookup(r.URL.Path[1:])
		if t == nil {
			http.NotFound(w, r)
			return
		}

		params, err := GetParam(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		params.Idx = RandString(10)

		err = t.Execute(w, params)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})
}
