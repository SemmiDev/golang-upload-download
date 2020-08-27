package main

import (
	"encoding/json"
	"fmt"
	"gorilla/mux"
	"html/template"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

func main() {
	handleRequest()
}

func handleRequest()  {
	myRouter := mux.NewRouter()
	// upload route
	myRouter.HandleFunc("/indexUpload", IndexUpload)
	myRouter.HandleFunc("/upload", Upload)
	// download route
	myRouter.HandleFunc("/list-files", ListFiles)
	myRouter.HandleFunc("/", IndexDownload)
	myRouter.HandleFunc("/download", Download)
	myRouter.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("assets"))))
	http.ListenAndServe(":9000", myRouter)
}


func IndexUpload(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("indexUpload.html"))
	if err := tmpl.Execute(w, nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func Upload(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "NOT ALLOWED EXCEPT POST REQUEST", http.StatusBadRequest)
		return
	}
	// return absolute information where the application executed
	basePath, _ := os.Getwd()

	reader, err := r.MultipartReader()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	for {
		part, err := reader.NextPart()
		if err == io.EOF {
			break
		}
		// filepath.join -> fushion the items with separator which is appropiate with OS
		// where the program executed (\ -> windows, / -> linux/unix)
		fileLocation := filepath.Join(basePath, "files", part.FileName())
		dst, err := os.Create(fileLocation)
		if dst != nil {
			defer dst.Close()
		}
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if _, err := io.Copy(dst, part); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	w.Write([]byte(`UPLOADED`))
}

func IndexDownload(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("indexDownload.html"))
	if err := tmpl.Execute(w, nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

type M map[string]interface{}

func ListFiles(w http.ResponseWriter, r *http.Request) {
	files := []M{}
	basePath, _ := os.Getwd()
	filesLocation := filepath.Join(basePath, "files")
	// walk = read value the directory
	err := filepath.Walk(filesLocation, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		files = append(files, M{"filename": info.Name(), "path": path})

		return nil
	})

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	res, err := json.Marshal(files)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(res)
}

func Download(w http.ResponseWriter, r *http.Request)  {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	path := r.FormValue("path")
	f, err := os.Open(path)
	if f != nil {
		defer f.Close()
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// content-disposition is one of the one MIME protocol extension
	// to give information to browser bagaimana dia harus berinteraksi dengan output
	// salah satu value nya adalah attachmant/file yg kemudina akan di download oleh browser
	contentDisposition := fmt.Sprintf("attachment; filename=%s", f.Name())
	w.Header().Set("Content-Disposition", contentDisposition)

	if _, err := io.Copy(w, f); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}