package drive

type File interface {
	Extension() string
	Contents() []byte
}
