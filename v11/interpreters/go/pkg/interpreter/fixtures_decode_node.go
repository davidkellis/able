package interpreter

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"math/big"

	"able/interpreter-go/pkg/ast"
)

type nodeCategoryDecoder func(map[string]any, string) (ast.Node, bool, error)

var nodeDecoders []nodeCategoryDecoder

func init() {
	nodeDecoders = []nodeCategoryDecoder{
		decodeModuleNodes,
		decodeLiteralNodes,
		decodeExpressionNodes,
		decodeControlFlowNodes,
		decodeDefinitionNodes,
	}
}

func decodeNode(node map[string]any) (ast.Node, error) {
	typ, _ := node["type"].(string)
	for _, decoder := range nodeDecoders {
		decoded, handled, err := decoder(node, typ)
		if err != nil {
			return nil, fmt.Errorf("decode %s: %w", typ, err)
		}
		if handled {
			return decoded, nil
		}
	}
	return nil, fs.ErrInvalid
}

func parseBigInt(value interface{}) *big.Int {
	if value == nil {
		return big.NewInt(0)
	}
	switch v := value.(type) {
	case float64:
		return big.NewInt(int64(v))
	case int64:
		return big.NewInt(v)
	case string:
		if bi, ok := new(big.Int).SetString(v, 10); ok {
			return bi
		}
	case json.Number:
		if i, err := v.Int64(); err == nil {
			return big.NewInt(i)
		}
	}
	return big.NewInt(0)
}
