package parsers

type Parser interface {
	TryParse(content []byte) Metadata

}

type Metadata interface {
}

type BasicMetadata struct {
	Type string `json:"type"`
}

var parsers = []Parser{	// order matter
	HttpRequestParser{},
	HttpResponseParser{},
}

func Parse(content []byte) Metadata {
	for _, parser := range parsers {
		if metadata := parser.TryParse(content); metadata != nil {
			return metadata
		}
	}

	return nil
}
