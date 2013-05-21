package main

import (
	"fmt"
	"gocode/gooday/domain"
	"gocode/gooday/router"
	"net/http"
	"os"
)

var (
	addr   = ":8080"
	r      = router.New()
	pwd, _ = os.Getwd()
)

func main() {
	fmt.Printf("GoodDay is now serving %s\n", addr)
	r.Get("/", day.GetHandler)
	http.Handle("/", r)
	http.ListenAndServe(addr, nil)
}
