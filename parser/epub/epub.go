package epub

import (
	"archive/zip"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
)

// Epub represent epub data
type Epub struct {
	Ncx        Ncx
	NcxPath    string
	Opf        Opf
	Container  Container
	Encryption Encryption
	LCP        LCP

	zipFd     *zip.ReadCloser
	directory string
}

//Open open resource file
func (epub *Epub) Open(filepath string) (io.ReadCloser, error) {
	return epub.open(epub.filename(filepath))
}

//RawOpen open resource file without filepath transform
func (epub *Epub) RawOpen(filepath string) (io.ReadCloser, error) {
	return epub.open(filepath)
}

//Close close file reader
func (epub *Epub) Close() {
	epub.zipFd.Close()
}

//-----------------------------------------------------------------------------
func (epub *Epub) filename(name string) string {
	return path.Join(path.Dir(epub.Container.Rootfile.Path), name)
}

func (epub *Epub) parseXML(filename string, v interface{}) error {
	fd, err := epub.open(filename)
	if err != nil {
		return nil
	}
	defer fd.Close()
	dec := xml.NewDecoder(fd)
	return dec.Decode(v)
}

func (epub *Epub) parseJSON(filename string, v interface{}) error {
	fd, err := epub.open(filename)
	if err != nil {
		return nil
	}
	defer fd.Close()
	dec := json.NewDecoder(fd)
	return dec.Decode(v)
}

// GetData return raw data from file
func (epub *Epub) GetData(filename string) ([]byte, error) {
	fd, err := epub.open(filename)
	if err != nil {
		return nil, nil
	}
	defer fd.Close()

	return ioutil.ReadAll(fd)

}

func (epub *Epub) open(filename string) (io.ReadCloser, error) {
	if epub.directory != "" {
		filenameEpub := path.Join(path.Dir(epub.directory+string(filepath.Separator)), filename)
		return os.Open(filenameEpub)
	}

	for _, f := range epub.zipFd.File {
		if f.Name == filename {
			return f.Open()
		}
	}
	return nil, fmt.Errorf("can't find file or directory %s", filename)
}

// ZipReader return the internal file descriptor
func (epub *Epub) ZipReader() *zip.ReadCloser {
	return epub.zipFd
}

// GetSMIL parse and return SMIL structure
func (epub *Epub) GetSMIL(ressource string) SMIL {
	var smil SMIL

	epub.parseXML(ressource, &smil)

	return smil
}

//OpenEpub open and parse epub
func OpenEpub(fn string) (*Epub, error) {
	zipFile, err := zip.OpenReader(fn)
	if err != nil {
		return nil, err
	}
	defer zipFile.Close()

	return openEpub(zipFile)
}

//OpenEpubReader open and parse epub
func OpenEpubReader(zipFile *zip.ReadCloser) (*Epub, error) {
	return openEpub(zipFile)
}

func openEpub(zipFile *zip.ReadCloser) (*Epub, error) {
	epb := Epub{zipFd: zipFile}
	errCont := epb.parseXML("META-INF/container.xml", &epb.Container)
	if errCont != nil {
		return nil, errCont
	}
	if len(epb.Container.Rootfiles) > 0 {
		epb.Container.Rootfile = epb.Container.Rootfiles[0]
	}

	errOpf := epb.parseXML(epb.Container.Rootfile.Path, &epb.Opf)
	if errOpf != nil {
		return nil, errOpf
	}

	epb.parseXML("META-INF/encryption.xml", &epb.Encryption)
	epb.parseJSON("META-INF/license.lcpl", &epb.LCP)

	for _, manf := range epb.Opf.Manifest {
		if manf.ID == epb.Opf.Spine.Toc {
			epb.NcxPath = epb.filename(manf.Href)
			errToc := epb.parseXML(epb.filename(manf.Href), &epb.Ncx)
			if errToc != nil {
				// return nil, errToc
			}
			break
		}
	}

	return &epb, nil
}

//OpenDir open a opf file
func OpenDir(filename string) (*Epub, error) {

	epb := Epub{directory: filename}

	errCont := epb.parseXML("META-INF/container.xml", &epb.Container)
	if errCont != nil {
		return nil, errCont
	}
	if len(epb.Container.Rootfiles) > 0 {
		epb.Container.Rootfile = epb.Container.Rootfiles[0]
	}

	errOpf := epb.parseXML(epb.Container.Rootfile.Path, &epb.Opf)
	if errOpf != nil {
		return nil, errOpf
	}

	epb.parseXML("META-INF/encryption.xml", &epb.Encryption)
	epb.parseJSON("META-INF/license.lcpl", &epb.LCP)

	for _, manf := range epb.Opf.Manifest {
		if manf.ID == epb.Opf.Spine.Toc {
			errToc := epb.parseXML(epb.filename(manf.Href), &epb.Ncx)
			if errToc != nil {
				return nil, errToc
			}
			break
		}
	}

	return &epb, nil
}
