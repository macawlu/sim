package server

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jpillora/archive"
)

const fileNumberLimit = 65535

type fsNode struct {
	Name     string
	Size     int64
	Modified time.Time
	Children []*fsNode
}

func (s *Server) listFiles() *fsNode {
	rootDir := s.state.Config.DownloadDirectory
	root := &fsNode{}
	if info, err := os.Stat(rootDir); err == nil {
		if err := list(rootDir, info, root, new(uint)); err != nil {
			log.Printf("File listing failed: %s", err)
		}
	}
	return root
}

func (s *Server) serveFiles(w http.ResponseWriter, r *http.Request) {
	if strings.HasPrefix(r.URL.Path, "/download/") {
		url := strings.TrimPrefix(r.URL.Path, "/download/")
		//dldir is absolute
		dldir := s.state.Config.DownloadDirectory
		file := filepath.Join(dldir, url)
		//only allow fetches/deletes inside the dl dir
		if !strings.HasPrefix(file, dldir) || dldir == file {
			http.Error(w, "Nice try\n"+dldir+"\n"+file, http.StatusBadRequest)
			return
		}
		info, err := os.Stat(file)
		if err != nil {
			http.Error(w, "File stat error: "+err.Error(), http.StatusBadRequest)
			return
		}
		switch r.Method {
		case "GET":
			if info.IsDir() {
				w.Header().Set("Content-Type", "application/zip")
				w.WriteHeader(200)
				//write .zip archive directly into response
				a := archive.NewZipWriter(w)
				a.AddDir(file)
				a.Close()
			} else {
				http.ServeFile(w, r, file)
			}
		case "DELETE":
			if err := os.RemoveAll(file); err != nil {
				http.Error(w, "Delete failed: "+err.Error(), http.StatusInternalServerError)
			}
		default:
			http.Error(w, "Not allowed", http.StatusMethodNotAllowed)
		}
		return
	}
	s.static.ServeHTTP(w, r)
}

//custom directory walk

func list(path string, info os.FileInfo, node *fsNode, n *uint) error {
	if (!info.IsDir() && !info.Mode().IsRegular()) || strings.HasPrefix(info.Name(), ".") {
		return errors.New("Non-regular file")
	}
	(*n)++
	if (*n) > fileNumberLimit {
		return errors.New("Over file limit") //limit number of files walked
	}
	node.Name = info.Name()
	node.Size = info.Size()
	node.Modified = info.ModTime()
	if !info.IsDir() {
		return nil
	}
	children, err := ioutil.ReadDir(path)
	if err != nil {
		return fmt.Errorf("Failed to list files: %w", err)
	}
	node.Size = 0
	for _, i := range children {
		c := &fsNode{}
		p := filepath.Join(path, i.Name())
		if err := list(p, i, c, n); err != nil {
			continue
		}
		node.Size += c.Size
		node.Children = append(node.Children, c)
	}

	return nil
}

func (s *Server) DoneCmd(path, hash, ttype string, size, ts int64) (string, []string, error) {

	if s.state.Config.DoneCmd == "" {
		return "", nil, fmt.Errorf("unconfigred Donecmd")
	}

	env := []string{
		fmt.Sprintf("CLD_DIR=%s", s.state.Config.DownloadDirectory),
		fmt.Sprintf("CLD_RESTAPI=%s", s.RestAPI),
		fmt.Sprintf("CLD_PATH=%s", path),
		fmt.Sprintf("CLD_HASH=%s", hash),
		fmt.Sprintf("CLD_TYPE=%s", ttype),
		fmt.Sprintf("CLD_SIZE=%d", size),
		fmt.Sprintf("CLD_STARTTS=%d", ts),
	}
	env = append(env, os.Environ()...)
	return s.state.Config.DoneCmd, env, nil
}
