package main

import (
	"bufio"
	"flag"
	"io"
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
			break
		}

		if part.FormName() == "myFile" {
			outPath := part.FileName()
			outPath = path.Join(*dir, outPath)
			out, err := os.Create(outPath)
			if err != nil {
				continue
			}
			bufOut := bufio.NewWriter(out)
			// if err != nil {
			// 	// http.Error(w, err.Error(), http.StatusInternalServerError)
			// 	break
			// }
			_, err = io.Copy(bufOut, part)
			bufOut.Flush()
			out.Close()
			if err != nil {
				os.Remove(outPath)
				// http.Error(w, err.Error(), http.StatusInternalServerError)
				// break
			}
		}
	}
	// clientFile, header, err := r.FormFile("myFile")
	// if err != nil {
	// 	http.Error(w, "Wrong request", http.StatusBadRequest)
	// 	return
	// }
	// out, err := os.Create(header.Filename)
	// if err != nil {
	// 	http.Error(w, "Can't create the file", http.StatusInternalServerError)
	// 	return
	// }
	// defer out.Close()
	// _, err = io.Copy(out, clientFile)
	// if err != nil {
	// 	http.Error(w, err.Error(), http.StatusInternalServerError)
	// }
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

	http.ListenAndServe(*b, handlers.CombinedLoggingHandler(os.Stdout, http.DefaultServeMux))
}
