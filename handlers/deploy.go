package handlers

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"sync"
)

type Deployment struct {
	Username string `json:"username"`
	Appname  string `json:"appname"`
	Port     int    `json:"port"`
}

var (
	portMu   sync.Mutex
	nextPort = 5001
	running  = map[string]*exec.Cmd{}
	stateFile = "/home/ubuntu/deployments/state.json"
)

func allocatePort() int {
	portMu.Lock()
	defer portMu.Unlock()
	p := nextPort
	nextPort++
	return p
}

func saveState(deps []Deployment) {
	data, _ := json.Marshal(deps)
	os.WriteFile(stateFile, data, 0644)
}

func loadState() []Deployment {
	data, err := os.ReadFile(stateFile)
	if err != nil {
		return nil
	}
	var deps []Deployment
	json.Unmarshal(data, &deps)
	return deps
}

func RestoreDeployments() {
	deps := loadState()
	for _, d := range deps {
		appDir := filepath.Join("/home/ubuntu/deployments", d.Username, d.Appname)
		cmd := exec.Command("python3", "app.py")
		cmd.Dir = appDir
		cmd.Env = append(os.Environ(), "PORT="+strconv.Itoa(d.Port))
		cmd.Start()
		key := d.Username + "/" + d.Appname
		running[key] = cmd
		if d.Port >= nextPort {
			nextPort = d.Port + 1
		}
		fmt.Printf("♻️  restored %s on port %d\n", key, d.Port)
	}
}

func Deploy(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respond(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}

	username := r.FormValue("username")
	appname := r.FormValue("appname")
	if username == "" || appname == "" {
		respond(w, http.StatusBadRequest, map[string]string{"error": "username and appname required"})
		return
	}

	file, _, err := r.FormFile("zipfile")
	if err != nil {
		respond(w, http.StatusBadRequest, map[string]string{"error": "zip file required"})
		return
	}
	defer file.Close()

	appDir := filepath.Join("/home/ubuntu/deployments", username, appname)
	os.MkdirAll(appDir, 0755)

	zipPath := filepath.Join(appDir, "app.zip")
	out, _ := os.Create(zipPath)
	io.Copy(out, file)
	out.Close()

	extractZip(zipPath, appDir)
	os.Remove(zipPath)

	key := username + "/" + appname
	if cmd, ok := running[key]; ok {
		cmd.Process.Kill()
		delete(running, key)
	}

	port := allocatePort()
	cmd := exec.Command("python3", "app.py")
	cmd.Dir = appDir
	cmd.Env = append(os.Environ(), "PORT="+strconv.Itoa(port))
	cmd.Start()
	running[key] = cmd

	deps := loadState()
	// Remove old entry for same app
	newDeps := []Deployment{}
	for _, d := range deps {
		if !(d.Username == username && d.Appname == appname) {
			newDeps = append(newDeps, d)
		}
	}
	newDeps = append(newDeps, Deployment{Username: username, Appname: appname, Port: port})
	saveState(newDeps)
	updateNginx()

	respond(w, http.StatusOK, map[string]interface{}{
		"message": "deployed successfully",
		"url":     fmt.Sprintf("http://13.200.117.42/%s/%s/", username, appname),
		"port":    port,
	})
}

func updateNginx() {
	deps := loadState()
	conf := `server {
    listen 80 default_server;
    listen [::]:80 default_server;
    root /var/www/html;
    index index.html;
    server_name _;
`
	for _, d := range deps {
		conf += fmt.Sprintf(`
    location /%s/%s/ {
        proxy_pass http://127.0.0.1:%d/;
        proxy_set_header Host $host;
    }
`, d.Username, d.Appname, d.Port)
	}
	conf += `
    location / {
        try_files $uri $uri/ =404;
    }
}`
	os.WriteFile("/tmp/nginx-mini-render.conf", []byte(conf), 0644)
	exec.Command("sudo", "cp", "/tmp/nginx-mini-render.conf", "/etc/nginx/sites-enabled/default").Run()
	exec.Command("sudo", "nginx", "-s", "reload").Run()
}

func extractZip(src, dest string) {
	r, err := zip.OpenReader(src)
	if err != nil {
		return
	}
	defer r.Close()
	for _, f := range r.File {
		path := filepath.Join(dest, f.Name)
		if f.FileInfo().IsDir() {
			os.MkdirAll(path, 0755)
			continue
		}
		os.MkdirAll(filepath.Dir(path), 0755)
		out, err := os.Create(path)
		if err != nil {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			out.Close()
			continue
		}
		io.Copy(out, rc)
		rc.Close()
		out.Close()
	}
}

func Status(w http.ResponseWriter, r *http.Request) {
	list := []string{}
	for k := range running {
		list = append(list, k)
	}
	respond(w, http.StatusOK, map[string]interface{}{
		"running": list,
		"count":   len(list),
	})
}
