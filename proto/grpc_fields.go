package proto

import (
	"google.golang.org/protobuf/reflect/protoreflect"
)

type Field struct {
	Name     string
	FullName string
	JsonName string
	Type     string

	IsMap     bool
	IsEnum    bool
	IsMessage bool
	IsOneOf   bool

	EnumValues   []*Enum
	FieldMessage *Message
	FieldOneOf   *OneOf
}

type Enum struct {
	Name     string
	FullName string
}

type Message struct {
	Name   string
	Fields []*Field
}

type OneOf struct {
	Name        string
	OneOfKeys   []string
	OneOfValues map[string][]*Field
}

type FieldsType int

const (
	Input FieldsType = iota
	Output
)

// GetFields will populate a structured representative of the fields associated with a grpc method
func (gcd *GrpcConnection) GetFields(methodName string, fieldsType FieldsType) ([]*Field, error) {
	var fields []*Field

	gcd.FileRegistry.RangeFiles(func(fd protoreflect.FileDescriptor) bool {
		var oneOfs []*Field
		svs := fd.Services()
		for i := 0; i < svs.Len(); i++ {
			methods := svs.Get(i).Methods()
			for j := 0; j < methods.Len(); j++ {
				if string(methods.Get(j).FullName()) == methodName {
					var fds protoreflect.FieldDescriptors
					if fieldsType == Input {
						fds = methods.Get(j).Input().Fields()
					} else {
						fds = methods.Get(j).Output().Fields()
					}
					for k := 0; k < fds.Len(); k++ {
						field := &Field{
							Name:      string(fds.Get(k).Name()),
							FullName:  string(fds.Get(k).FullName()),
							JsonName:  fds.Get(k).JSONName(),
							Type:      fds.Get(k).Kind().String(),
							IsMap:     fds.Get(k).IsMap(),
							IsMessage: fds.Get(k).Kind().String() == "message",
							IsEnum:    fds.Get(k).Kind().String() == "enum",
						}
						if field.IsEnum {
							field.EnumValues = gcd.getFieldsEnum(fds.Get(k))
						}
						if field.IsMessage {
							msg, err := gcd.getFieldsMessage(fds.Get(k).Message())
							if err != nil {
								return false
							}
							field.FieldMessage = msg
						}
						if fds.Get(k).ContainingOneof() != nil {
							oneOf, found := gcd.getFieldsOneOf(oneOfs, field, fds.Get(k).ContainingOneof())
							if !found {
								oneOfs = append(oneOfs, oneOf)
								fields = append(fields, oneOf)
							}
						} else {
							fields = append(fields, field)
						}
					}
				}
			}
		}
		return true
	})

	return fields, nil
}

func (gcd *GrpcConnection) getFieldsEnum(desc protoreflect.FieldDescriptor) []*Enum {
	var enumValues []*Enum
	enum := desc.Enum()
	evs := enum.Values()
	for i := 0; i < evs.Len(); i++ {
		val := evs.Get(i)
		enumValues = append(enumValues, &Enum{
			Name:     string(val.Name()),
			FullName: string(val.FullName()),
		})
	}
	return enumValues
}

func (gcd *GrpcConnection) getFieldsMessage(desc protoreflect.MessageDescriptor) (*Message, error) {
	fds := desc.Fields()

	var fields []*Field
	var oneOfs []*Field
	for k := 0; k < fds.Len(); k++ {
		field := &Field{
			Name:      string(fds.Get(k).Name()),
			FullName:  string(fds.Get(k).FullName()),
			JsonName:  fds.Get(k).JSONName(),
			Type:      fds.Get(k).Kind().String(),
			IsMessage: fds.Get(k).Kind().String() == "message",
			IsEnum:    fds.Get(k).Kind().String() == "enum",
		}
		if field.IsEnum {
			field.EnumValues = gcd.getFieldsEnum(fds.Get(k))
		}
		if field.IsMessage {
			msg, err := gcd.getFieldsMessage(fds.Get(k).Message())
			if err != nil {
				return nil, err
			}
			field.FieldMessage = msg
		}
		if fds.Get(k).ContainingOneof() != nil {
			oneOf, found := gcd.getFieldsOneOf(oneOfs, field, fds.Get(k).ContainingOneof())
			if !found {
				oneOfs = append(oneOfs, oneOf)
				fields = append(fields, oneOf)
			}
		} else {
			fields = append(fields, field)
		}
	}

	return &Message{
		Name:   string(desc.FullName()),
		Fields: fields,
	}, nil
}

func (gcd *GrpcConnection) getFieldsOneOf(oneOfs []*Field, field *Field, desc protoreflect.OneofDescriptor) (*Field, bool) {
	var oneOfFound *Field
	var found bool
	for _, oneOfField := range oneOfs {
		if oneOfField.Name == string(desc.Name()) {
			oneOfFound = oneOfField
			break
		}
	}
	if oneOfFound == nil {
		oneOfFound = &Field{
			Name:     string(desc.Name()),
			FullName: string(desc.FullName()),
			Type:     "oneof",
			IsOneOf:  true,
			FieldOneOf: &OneOf{
				Name:        string(desc.Name()),
				OneOfValues: make(map[string][]*Field),
			},
		}
	} else {
		found = true
	}
	oneOfFound.FieldOneOf.OneOfKeys = append(oneOfFound.FieldOneOf.OneOfKeys, field.Name)
	oneOfFound.FieldOneOf.OneOfValues[field.Name] = append(oneOfFound.FieldOneOf.OneOfValues[field.Name], field)
	return oneOfFound, found
}
