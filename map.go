package jsonmanu

const (
	Object int = iota
	Array
	String
	Number
	Boolean
	Null
)

type JsonNode struct {
	Path string
	Type int
}

type Transformer interface {
	Transform(value any) any
}

type SplitTransformer struct {
	Delim string
	Index int
}

func (t SplitTransformer) Transform(value any) any {
	return value
}

type ReplaceTransformer struct {
	OldVal string
	NewVal string
}

func (t ReplaceTransformer) Transform(value any) any {
	return value
}

type Mapper struct {
	SrcNode     JsonNode
	DstNode     JsonNode
	Transformer Transformer
}

func Map(src any, dst any, mappers []Mapper) error {

	for _, mapper := range mappers {
		value, err := Get(src, mapper.SrcNode.Path)
		if err != nil {
			return err
		}

		if mapper.Transformer != nil {
			value = mapper.Transformer.Transform(value)
		}

		err = Put(dst, mapper.DstNode.Path, value)
		if err != nil {
			return err
		}

	}
	return nil
}
