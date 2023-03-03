package server

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
	"os"
)

//go:embed static
var static embed.FS

// staticFiles returns a http.FileSystem for the static directory. If useOS is
// true the returned filesystem will represent the actual OS filesystem. In
// general, this should be used when developing so that changes are reflected
// immediately.
//
// If useOS is false, the returned filesystem will be a virtual filesystem built
// from the static directory. In general, this should be used in production so
// that we can ship a single binary that includes the static files in it.
func staticFiles(useOS bool) http.FileSystem {
	if useOS {
		log.Print("using live mode")
		return http.FS(os.DirFS("static"))
	}

	log.Print("using embed mode")
	fsys, err := fs.Sub(static, "static")
	if err != nil {
		panic(err)
	}

	return http.FS(fsys)
}
