package apps

// hat tip https://gist.github.com/indraniel/1a91458984179ab4cf80

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"log"
	"os"
	"path"
	"strings"
)

func extractTarball(packageName string, r io.Reader) {
	s, err := gzip.NewReader(r)
	if err != nil {
		log.Fatal(err)
	}

	t := tar.NewReader(s)

	for true {
		h, err := t.Next()

		if err == io.EOF {
			break
		}

		if err != nil {
			log.Fatal(err)
		}

		nameComponents := strings.Split(h.Name, "/")
		nameComponents[0] = packageName
		name := path.Join(nameComponents...)

		switch h.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(path.Join(name), 0755); err != nil {
				log.Fatal(err)
			}
		case tar.TypeReg:
			if err := os.MkdirAll(path.Dir(path.Join(name)), 0755); err != nil {
				log.Fatal(err)
			}
			o, err := os.Create(path.Join(name))
			if err != nil {
				log.Fatal(err)
			}
			defer o.Close()
			if _, err := io.Copy(o, t); err != nil {
				log.Fatal(err)
			}
		default:
			log.Fatalf(
				"ExtractTarGz: uknown type: %s in %s",
				h.Typeflag,
				h.Name)
		}
	}
}
