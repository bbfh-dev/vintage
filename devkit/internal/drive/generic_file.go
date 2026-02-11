package drive

type GenericFile struct {
	Ext  string
	Body []byte
}

func NewGenericFile(ext string, body []byte) *GenericFile {
	return &GenericFile{
		Ext:  ext,
		Body: body,
	}
}

func (file *GenericFile) Extension() string {
	return file.Ext
}

func (file *GenericFile) Contents() []byte {
	return file.Body
}

func (file *GenericFile) Clone() *GenericFile {
	return NewGenericFile(file.Ext, file.Body)
}
