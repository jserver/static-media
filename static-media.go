package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

var backing = flag.String("backing", "", "Backing store to retrieve media")
var secondary = flag.String("secondary", "", "Secondary backing store to retrieve media")
var port = flag.Int("port", 8001, "Port to listen on")
var servePath string

func GetMedia(store *string, path, fspath string) error {
	backing_url := strings.Replace(path, "/media/", "", 1)
	resp, err := http.Get(*store + backing_url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		err = fmt.Errorf("backing returned %d", resp.StatusCode)
		return err
	}
	err = os.MkdirAll(filepath.Dir(fspath), 0777)
	if err != nil {
		return err
	}
	file, err := os.Create(fspath)
	if err != nil {
		return err
	}
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	_, err = file.Write(b)
	if err != nil {
		return err
	}
	return nil
}

func AssetHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Print(r.URL.Path)
		fspath := servePath + r.URL.Path
		if _, err := os.Stat(fspath); err != nil {
			if os.IsNotExist(err) {
				if (strings.HasPrefix(r.URL.Path, "/media") || strings.HasPrefix(r.URL.Path, "/images")) && *backing != "" {
					fmt.Println("Trying Main Backup")
					err = GetMedia(backing, r.URL.Path, fspath)
					if err != nil && *secondary != "" {
						fmt.Println("Trying Secondary Backup")
						err = GetMedia(secondary, r.URL.Path, fspath)
					}
				}
			}
			if err != nil {
				log.Print(err)
				http.NotFound(w, r)
				return
			}
		}
		h.ServeHTTP(w, r)
	})
}

func main() {
	flag.Parse()
	if len(flag.Args()) != 1 {
		log.Fatal("Pass in path to serve")
	}
	servePath = flag.Arg(0)
	if servePath[:1] == "~" {
		servePath = "$HOME" + servePath[1:]
	}
	servePath = os.ExpandEnv(servePath)
	if servePath[len(servePath)-1:] == "/" {
		servePath = servePath[:len(servePath)-1]
	}
	log.Printf("Serving %s on port %d", servePath, *port)

	http.Handle("/", AssetHandler(http.FileServer(http.Dir(servePath))))
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *port), nil))
}
