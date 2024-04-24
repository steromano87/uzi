package workspace

type Archiver interface {
	Archive(folder string) ([]byte, error)
}

type Extractor interface {
	Extract(archive []byte, destinationFolder string) error
}
