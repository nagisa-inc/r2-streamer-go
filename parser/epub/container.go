package epub

//Container META-INF/container.xml file
type Container struct {
	Rootfile  Rootfile
	Rootfiles []Rootfile `xml:"rootfiles>rootfile"`
}

//Rootfile root file
type Rootfile struct {
	Path    string `xml:"full-path,attr"`
	Type    string `xml:"media-type,attr"`
	Version string `xml:"version,attr"`
}
