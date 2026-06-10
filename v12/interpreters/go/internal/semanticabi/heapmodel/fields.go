package heapmodel

import (
	"fmt"

	"able/interpreter-go/internal/semanticabi"
)

func NewFields(layoutID uint16) ([]FieldValue, error) {
	layout, ok := semanticabi.ObjectLayoutByID(layoutID)
	if !ok {
		return nil, fmt.Errorf("heapmodel: unknown object layout %d", layoutID)
	}
	nilCell, err := semanticabi.ImmediateCell(semanticabi.TagKindNil, 0, 0)
	if err != nil {
		return nil, err
	}
	result := make([]FieldValue, len(layout.Fields))
	for index, field := range layout.Fields {
		switch field.Storage {
		case semanticabi.FieldCell:
			result[index].Cells = []semanticabi.Cell{nilCell}
		case semanticabi.FieldObject:
			result[index].Objects = []semanticabi.ObjectID{0}
		}
	}
	return result, nil
}

func SetScalar(layoutID uint16, fields []FieldValue, name string, value uint64) error {
	index, err := fieldIndex(layoutID, fields, name, semanticabi.FieldScalar)
	if err != nil {
		return err
	}
	fields[index] = FieldValue{Scalar: value}
	return nil
}

func SetBytes(layoutID uint16, fields []FieldValue, name string, value []byte) error {
	index, err := fieldIndex(layoutID, fields, name, semanticabi.FieldBytes)
	if err != nil {
		return err
	}
	fields[index] = FieldValue{Bytes: append([]byte(nil), value...)}
	return nil
}

func SetCells(layoutID uint16, fields []FieldValue, name string, values ...semanticabi.Cell) error {
	layout, _ := semanticabi.ObjectLayoutByID(layoutID)
	index := -1
	for candidate, field := range layout.Fields {
		if field.Name == name && (field.Storage == semanticabi.FieldCell || field.Storage == semanticabi.FieldCells) {
			index = candidate
			break
		}
	}
	if index < 0 || len(fields) != len(layout.Fields) {
		return fmt.Errorf("heapmodel: layout %s has no compatible cell field %s", layout.Name, name)
	}
	if layout.Fields[index].Storage == semanticabi.FieldCell && len(values) != 1 {
		return fmt.Errorf("heapmodel: layout %s field %s requires one cell", layout.Name, name)
	}
	fields[index] = FieldValue{Cells: append([]semanticabi.Cell(nil), values...)}
	return nil
}

func SetObjects(layoutID uint16, fields []FieldValue, name string, values ...semanticabi.ObjectID) error {
	layout, _ := semanticabi.ObjectLayoutByID(layoutID)
	index := -1
	for candidate, field := range layout.Fields {
		if field.Name == name && (field.Storage == semanticabi.FieldObject || field.Storage == semanticabi.FieldObjects) {
			index = candidate
			break
		}
	}
	if index < 0 || len(fields) != len(layout.Fields) {
		return fmt.Errorf("heapmodel: layout %s has no compatible object field %s", layout.Name, name)
	}
	if layout.Fields[index].Storage == semanticabi.FieldObject && len(values) != 1 {
		return fmt.Errorf("heapmodel: layout %s field %s requires one object", layout.Name, name)
	}
	fields[index] = FieldValue{Objects: append([]semanticabi.ObjectID(nil), values...)}
	return nil
}

func fieldIndex(layoutID uint16, fields []FieldValue, name string, storage semanticabi.FieldStorage) (int, error) {
	layout, ok := semanticabi.ObjectLayoutByID(layoutID)
	if !ok {
		return -1, fmt.Errorf("heapmodel: unknown object layout %d", layoutID)
	}
	if len(fields) != len(layout.Fields) {
		return -1, fmt.Errorf("heapmodel: layout %s field count mismatch", layout.Name)
	}
	for index, field := range layout.Fields {
		if field.Name == name && field.Storage == storage {
			return index, nil
		}
	}
	return -1, fmt.Errorf("heapmodel: layout %s has no %s field %s", layout.Name, storageName(storage), name)
}

func storageName(storage semanticabi.FieldStorage) string {
	switch storage {
	case semanticabi.FieldScalar:
		return "scalar"
	case semanticabi.FieldBytes:
		return "bytes"
	default:
		return "requested"
	}
}
