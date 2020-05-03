package main

import (
	"fmt"
	"net/http"
)

type Params struct {
	User string
	Pwd  string
	Addr string
	Idx  string
}

func GetParam(r *http.Request) (Params, error) {
	var err error

	value := func(name string) string {
		v := r.URL.Query().Get(name)
		if v == "" {
			if err == nil {
				err = fmt.Errorf("Missing '%s' QueryString\n", name)
			} else {
				err = fmt.Errorf("%wMissing '%s' QueryString\n", err, name)
			}
			return ""
		}
		return v
	}

	user := value("user")
	pwd := value("pwd")
	addr := value("addr")
	idx := r.Header.Get("X-Idx")

	return Params{user, pwd, addr, idx}, err
}
