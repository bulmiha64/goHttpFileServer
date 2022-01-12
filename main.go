package main

import (
	"bufio"
	"flag"
	"io"
	"log"
	"net/http"
	"os"
	"path"

	_ "embed"

	"github.com/gorilla/handlers"
	"golang.org/x/net/webdav"
)

//go:embed upload.html
var form []byte

// var formTemplate *template.Template
var dir *string

func uploadFile(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.Write(form)
		return
	}
	reader, err := r.MultipartReader()
	if err != nil {
		http.Error(w, "Wrong request", http.StatusBadRequest)
		return
	}
	for {
		part, err := reader.NextPart()
		if err != nil {
			return
		}

		if part.FormName() == "myFile" {
			outPath := part.FileName()
			outPath = path.Join(*dir, outPath)
			out, err := os.Create(outPath)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			defer out.Close()
			bufOut := bufio.NewWriter(out)
			defer bufOut.Flush()
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			_, err = io.Copy(bufOut, part)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				out.Close()
				os.Remove(part.FileName())
				continue
			}

			log.Printf("%s uploaded file: %s\n", r.RemoteAddr, part.FileName())
		}
	}
}

func main() {

	dir = flag.String("d", ".", "Directory to serve")
	b := flag.String("b", ":9999", "Address to bind to")

	flag.Parse()

	http.Handle("/webdav_handler/", http.StripPrefix("/webdav_handler/", &webdav.Handler{
		FileSystem: webdav.Dir(*dir),
		LockSystem: webdav.NewMemLS(),
		Logger:     nil}))
	http.HandleFunc("/upload", uploadFile)

	http.Handle("/", http.FileServer(http.Dir(*dir)))

	log.Println("Listening at ", *b)

	http.ListenAndServe(*b, handlers.CombinedLoggingHandler(os.Stdout, http.DefaultServeMux))
}
