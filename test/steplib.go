package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("no publish address provided! usage: steplib localhost:8088")
		os.Exit(1)
	}

	addr := os.Args[1]
	err := http.ListenAndServe(addr, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		f, err := os.Open("spec.json")
		if err != nil {
			fmt.Println(err)
		}

		data, err := ioutil.ReadAll(f)
		if err != nil {
			fmt.Println(err)
		}

		w.Header().Set("Content-type", "application/json")
		if _, err := w.Write(data); err != nil {
			fmt.Println(err)
		}
	}))

	if err != nil {
		fmt.Println(fmt.Sprintf("error: %s", err))
		os.Exit(1)
	}
}
