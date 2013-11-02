package brogger

import (
	"fmt"
	"net/http"
	"runtime"
)

type Brog struct {
	*logMux
	Config   *Config
	tmplMngr *TemplateManager
	postMngr *PostManager
}

func PrepareBrog() (*Brog, error) {
	config, err := loadConfig()
	if err != nil {
		return nil, fmt.Errorf("preparing brog's configuration, %v", err)
	}

	logMux, err := makeLogMux(config.LogFilename)
	if err != nil {
		return nil, fmt.Errorf("making log multiplex on path %s, %v", config.LogFilename, err)
	}

	brog := &Brog{
		logMux: logMux,
		Config: config,
	}

	runtime.GOMAXPROCS(config.MaxCPUs)

	return brog, nil
}

func (b *Brog) ListenAndServe() error {

	addr := fmt.Sprintf("%s:%d", b.Config.Hostname, b.Config.PortNumber)

	b.Ok("CAPTAIN: Open channel, %s", addr)
	b.Warn("ON SCREEN: We are the Brog. Resistance is futile.")

	if err := b.startWatchers(); err != nil {
		return err
	}

	http.HandleFunc("/heartbeat", b.heartBeat)
	http.HandleFunc("/", b.indexFunc)
	http.HandleFunc("/posts/", b.postFunc)
	http.Handle("/assets", http.StripPrefix("/assets", http.FileServer(http.Dir(b.Config.AssetPath))))

	b.Ok("Assimilation completed.")
	return http.ListenAndServe(addr, nil)
}

func (b *Brog) Close() {

	if b.postMngr != nil {
		defer b.postMngr.Close()
	}

	if b.tmplMngr != nil {
		defer b.tmplMngr.Close()
	}

	if b.logMux != nil {
		defer b.logMux.Close()
	}

}

func (b *Brog) startWatchers() error {
	CopyBrogBinaries(b.Config)

	tmplMngr, err := StartTemplateManager(b, b.Config.TemplatePath)
	if err != nil {
		return fmt.Errorf("starting template manager, %v", err)
	}
	b.tmplMngr = tmplMngr

	postMngr, err := StartPostManager(b, b.Config.PostPath)
	if err != nil {
		return fmt.Errorf("starting post manager, %v", err)
	}
	b.postMngr = postMngr
	return nil
}

// heartBeat answers 200 to any request.
func (b *Brog) heartBeat(rw http.ResponseWriter, req *http.Request) {
	rw.WriteHeader(http.StatusOK)
}

func (b *Brog) indexFunc(rw http.ResponseWriter, req *http.Request) {
	rw.WriteHeader(http.StatusOK)
	fmt.Fprint(rw, `<!doctype html>
<html>
<head><title>Hello</title></head>
<body>`)
	for _, post := range b.postMngr.GetAllPosts() {
		fmt.Fprintf(rw, "<h1>%s</h1>", post.Title)
	}
	fmt.Fprint(rw, "</body></html>")
}

func (b *Brog) postFunc(rw http.ResponseWriter, req *http.Request) {
	rw.WriteHeader(http.StatusNotImplemented)
}
