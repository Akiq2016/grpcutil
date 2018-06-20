package gentstypes

// TODO: eliminate stderr debug printing
// TODO: add services support
// TODO: add nested messages support
// TODO: add nested enum support
// TODO: add better output filename

import (
	"bytes"
	"fmt"
	"log"
	"sort"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	plugin "github.com/golang/protobuf/protoc-gen-go/plugin"
	"github.com/jhump/protoreflect/desc"
)

const indent = "  "

type Generator struct {
	*bytes.Buffer
	indent   string
	Request  *plugin.CodeGeneratorRequest
	Response *plugin.CodeGeneratorResponse
}

func New() *Generator {
	return &Generator{
		Buffer:   new(bytes.Buffer),
		Request:  new(plugin.CodeGeneratorRequest),
		Response: new(plugin.CodeGeneratorResponse),
	}
}
func (g *Generator) incIndent() {
	g.indent += indent
}
func (g *Generator) decIndent() {
	g.indent = string(g.indent[:len(g.indent)-len(indent)])
}

func (g *Generator) W(s string) {
	g.Buffer.WriteString(g.indent)
	g.Buffer.WriteString(s)
	g.Buffer.WriteString("\n")
}

/*
var s = &spew.ConfigState{
	Indent:         " ",
	DisableMethods: true,
}
*/

func (g *Generator) GenerateAllFiles() {
	g.W("// generated by protoc-gen-tstypes\n")
	files, err := desc.CreateFileDescriptors(g.Request.ProtoFile)
	if err != nil {
		log.Fatal(err)
	}
	names := []string{}
	for fname := range files {
		names = append(names, fname)
	}
	sort.Strings(names)
	for _, n := range names {
		f := files[n]
		g.W(fmt.Sprintf("declare namespace %s {\n", f.GetPackage()))
		g.incIndent()
		g.generate(f)
		g.decIndent()
		g.W("}\n")
	}

	//s.Fdump(os.Stderr, g.Request)

	g.Response.File = append(g.Response.File, &plugin.CodeGeneratorResponse_File{
		Name:    proto.String("output.d.ts"),
		Content: proto.String(g.String()),
	})
}

func (g *Generator) generate(f *desc.FileDescriptor) {
	// TODO reorder
	g.generateEnums(f.GetEnumTypes())
	g.generateMessages(f.GetMessageTypes())
	g.generateServices(f.GetServices())
	//s.Fdump(os.Stderr, f.GetName())
}

func (g *Generator) generateMessages(messages []*desc.MessageDescriptor) {
	for _, m := range messages {
		g.generateMessage(m)
	}
}
func (g *Generator) generateEnums(enums []*desc.EnumDescriptor) {
	for _, e := range enums {
		g.generateEnum(e)
	}
}
func (g *Generator) generateServices(f []*desc.ServiceDescriptor) {
}

func (g *Generator) generateMessage(m *desc.MessageDescriptor) {
	// TODO: namespace messages?
	for _, e := range m.GetNestedEnumTypes() {
		g.generateEnum(e)
	}
	g.W(fmt.Sprintf("interface %s {", m.GetName()))
	for _, f := range m.GetFields() {
		g.W(fmt.Sprintf(indent+"%s?: %s;", f.GetName(), fieldType(f)))
	}
	g.W("}\n")
}

func fieldType(f *desc.FieldDescriptor) string {
	t := rawFieldType(f)
	if f.IsMap() {
		return fmt.Sprintf("{ [key: %s]: %s }", rawFieldType(f.GetMapKeyType()), rawFieldType(f.GetMapValueType()))
	}
	if f.IsRepeated() {
		return fmt.Sprintf("Array<%s>", t)
	}
	return t
}

func rawFieldType(f *desc.FieldDescriptor) string {
	switch f.GetType() {
	case descriptor.FieldDescriptorProto_TYPE_DOUBLE:
		fallthrough
	case descriptor.FieldDescriptorProto_TYPE_FLOAT:
		fallthrough
	case descriptor.FieldDescriptorProto_TYPE_INT64:
		fallthrough
	case descriptor.FieldDescriptorProto_TYPE_UINT64:
		fallthrough
	case descriptor.FieldDescriptorProto_TYPE_INT32:
		fallthrough
	case descriptor.FieldDescriptorProto_TYPE_FIXED64:
		fallthrough
	case descriptor.FieldDescriptorProto_TYPE_FIXED32:
		fallthrough
	case descriptor.FieldDescriptorProto_TYPE_UINT32:
		fallthrough
	case descriptor.FieldDescriptorProto_TYPE_SFIXED32:
		fallthrough
	case descriptor.FieldDescriptorProto_TYPE_SFIXED64:
		fallthrough
	case descriptor.FieldDescriptorProto_TYPE_SINT32:
		fallthrough
	case descriptor.FieldDescriptorProto_TYPE_SINT64:
		return "number"
	case descriptor.FieldDescriptorProto_TYPE_BOOL:
		return "boolean"
	case descriptor.FieldDescriptorProto_TYPE_STRING:
		return "string"
	case descriptor.FieldDescriptorProto_TYPE_BYTES:
		return "Uint8Array"
	case descriptor.FieldDescriptorProto_TYPE_ENUM:
		return f.GetEnumType().GetName()
	case descriptor.FieldDescriptorProto_TYPE_MESSAGE:
		t := f.GetMessageType()
		if t.GetFile().GetPackage() != f.GetFile().GetPackage() {
			return t.GetFullyQualifiedName()
		}
		return t.GetName()
	}
	return "any /*unknown*/"
}

func (g *Generator) generateEnum(e *desc.EnumDescriptor) {
	g.W(fmt.Sprintf("enum %s {", e.GetName()))
	for _, v := range e.GetValues() {
		g.W(fmt.Sprintf("    %s = %v,", v.GetName(), v.GetNumber()))
	}
	g.W("}")
}
