// config-ui is a local web utility for building the discord.activity section of
// barman's YAML config. Run it, open http://localhost:8080, fill the form, copy
// the generated snippet into configs/config.yaml.
package main

import (
	_ "embed"
	"flag"
	"fmt"
	"log"
	"net/http"
)

//go:embed index.html
var indexHTML []byte

func main() {
	addr := flag.String("addr", ":8080", "listen address")
	flag.Parse()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(indexHTML)
	})

	fmt.Printf("config-ui listening on http://localhost%s\n", *addr)
	log.Fatal(http.ListenAndServe(*addr, nil))
}
