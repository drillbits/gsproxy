package main

import (
	"context"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"path/filepath"

	"fmt"

	"cloud.google.com/go/storage"
	"google.golang.org/api/option"
)

// Workspace represents a workspace.
type Workspace struct {
	client      *storage.Client
	dir         string
	downloadDir string
	destDir     string
	src         *url.URL
	dest        *url.URL
}

// NewWorkspace creates a new workspace.
func NewWorkspace(ctx context.Context, src *url.URL, dest *url.URL) (*Workspace, error) {
	w := &Workspace{
		src:  src,
		dest: dest,
	}

	cfg := ConfigFromContext(ctx)
	client, err := storage.NewClient(ctx, option.WithServiceAccountFile(cfg.KeyFile))
	if err != nil {
		return nil, err
	}
	w.client = client

	err = w.setup()
	if err != nil {
		return nil, err
	}

	return w, nil
}

func (w *Workspace) setup() error {
	workspace, err := ioutil.TempDir("", name)
	if err != nil {
		return err
	}
	w.dir = workspace

	w.downloadDir = filepath.Join(w.dir, "download")
	w.destDir = filepath.Join(w.dir, "dest")

	err = os.MkdirAll(filepath.Dir(w.Src()), os.ModePerm)
	if err != nil {
		return err
	}
	log.Printf("download directory created: %s", filepath.Dir(w.Src()))

	err = os.MkdirAll(w.Dest(), os.ModePerm)
	if err != nil {
		return err
	}
	log.Printf("dest directory created: %s", w.Dest())

	return nil
}

// Close cleans directories of the workspace.
func (w *Workspace) Close() {
	os.RemoveAll(w.dir)
}

// Src returns source file path.
func (w *Workspace) Src() string {
	return filepath.Join(w.downloadDir, w.src.Path[1:])
}

// Dest returns destination directory path.
func (w *Workspace) Dest() string {
	return filepath.Dir(filepath.Join(w.destDir, w.src.Path[1:]))
}

// Download downloads all files to the workspace.
func (w *Workspace) Download(ctx context.Context) error {
	u := w.src
	bucket := w.client.Bucket(u.Host)
	r, err := bucket.Object(u.Path[1:]).NewReader(ctx)
	if err != nil {
		return err
	}
	defer r.Close()

	b, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}
	log.Printf("download from gs://%s/%s", u.Host, u.Path[1:])

	dest := filepath.Join(w.downloadDir, u.Path[1:])
	err = ioutil.WriteFile(dest, b, 0666)
	if err != nil {
		return err
	}

	return nil
}

// Upload uploads all files from destination directory to Google Cloud Storage.
func (w *Workspace) Upload(ctx context.Context) error {
	u := w.dest
	err := filepath.Walk(w.Dest(), func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}

		b, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}

		obj := fmt.Sprintf("%s/%s", u.Path[1:], filepath.Base(path))
		bucket := w.client.Bucket(u.Host)
		w := bucket.Object(obj).NewWriter(ctx)
		defer w.Close()

		_, err = w.Write(b)
		if err != nil {
			return err
		}
		log.Printf("upload to gs://%s/%s", u.Host, obj)

		return nil
	})
	if err != nil {
		return err
	}
	return nil
}
