package filetype

import (
	"github.com/h2non/filetype"
)

type MIME struct {
	Type      string `json:"type"`
	Subtype   string `json:"sub_type"`
	Value     string `json:"value"`
	Extension string `json:"extension"`
}

type SupportType map[string]MIME

func FileType(filename string) MIME {
	kind, unknown := filetype.MatchFile(filename)
	if unknown != nil {
		return MIME{Type: "unknown", Subtype: "unknown", Value: "unknown", Extension: "unknown"}
	}
	return MIME{Type: kind.MIME.Type, Subtype: kind.MIME.Subtype, Value: kind.MIME.Value, Extension: kind.Extension}

}

func SupportTypes() SupportType {
	supportTypeMap := make(map[string]MIME)
	for kind := range filetype.Matchers {
		supportTypeMap[kind.MIME.Value] = MIME{Type: kind.MIME.Type, Subtype: kind.MIME.Subtype, Value: kind.MIME.Value, Extension: kind.Extension}

	}
	return supportTypeMap
}

func (s SupportType) GetTypeAndExtension(fileType string) (string, string) {
	filetypeObj, ok := s[fileType]
	if !ok {
		return "unknown", "unknown"
	}
	return filetypeObj.Type, filetypeObj.Extension
}
