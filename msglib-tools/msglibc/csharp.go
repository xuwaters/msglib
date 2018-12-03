package main

import (
	"fmt"
	"os"
	"path"
	"text/template"
)

/////////////////////////////////////////////////////////////////////// CSharp Code Generation

var tmplCSharpCode string = `
{{- define "Enum"}}
    /// <summary>{{range $i,$line := .Comments}}
    /// {{TrimComment $line }}{{end}}
    /// </summary>
    public enum {{.Name}}
    {{print "{"}}
    {{- range $i, $field := .Fields}}{{with $field}}
        {{- if len .Comments | lt 0}}
        /// <summary>{{range $i,$line := .Comments}}
        /// {{TrimComment $line }}{{end}}
        /// </summary>
        {{- end}}
        {{.FieldName}} = {{.FieldValue}}, {{.SuffixComment}}
    {{- end}}{{end}}
    }
{{- end}}
{{- define "Message"}}
    /// <summary>{{range $i,$line := .Comments}}
    /// {{TrimComment $line }}{{end}}
    /// </summary>
    public class {{.Name}}
    {{print "{"}}
        internal byte[] EncodeAsBytes() { return MIO_{{.Name}}.I.Encode(this); }
        internal static {{.Name}} Decode(MemoryStream ms) { return MIO_{{.Name}}.I.Decode(ms); }

    {{- range $i, $field := .Fields}}{{with $field}}
        {{- if HasComments .Comments .SuffixComment}}
        /// <summary>{{range $i,$line := .Comments}}
        /// {{TrimComment $line }}{{end}}
        /// {{TrimComment .SuffixComment}}
        /// </summary>
        {{- end}}
        public {{GetCSharpType .TypeName .TypeParams}} {{.FieldName}};

    {{- end}}{{end}}
    }

    internal class MIO_{{.Name}} : MIO_Base<{{.Name}}>
    {
        internal static readonly MIO_{{.Name}} I = new MIO_{{.Name}}();

        internal enum E : int
        {{print "{"}}
        {{- range $i, $field := .Fields}}{{with $field}}
            {{.FieldName}} = {{.FieldID}},
        {{- end}}{{end}}
        }

        private MIO_{{.Name}}()
            : base(new MioStruct(new MioField[] {{print "{"}}
        {{- range $i, $field := .Fields}}{{with $field}}
            new MioField((int)E.{{.FieldName}}, {{GetCSharpMioStr .TypeName .TypeParams}}),
        {{- end}}{{end}}
        })) { }

        internal override {{.Name}} FromMioObject(object o)
        {
            MioObject obj = o as MioObject;
            if (obj == null) return null;
            {{.Name}} data = new {{.Name}}();
        {{- range $i, $field := .Fields}}{{with $field}}
            data.{{.FieldName}} = {{MakeMioGet .}};
        {{- end}}{{end}}
            return data;
        }

        internal override object ToMioObject({{.Name}} data)
        {
            MioObject obj = new MioObject();
        {{- range $i, $field := .Fields}}{{with $field}}
            obj.SetObject((int)E.{{.FieldName}}, {{MakeMioSet .}});
        {{- end}}{{end}}
            return obj;
        }
    }

{{- end}}
{{- define "Main" -}}
using System;
using MsgLib;
using System.Collections.Generic;
using System.Collections;
using System.IO;

namespace {{.ProtoName}}
{
{{range $idx, $elem := .EnumList}}
    {{- template "Enum" $elem}}
{{end}}
{{range $idx, $elem := .Messages}}
    {{- template "Message" $elem}}
{{end}}
}
{{- end}}
{{- template "Main" .}}
`

func (self *MsgCompiler) GenerateCSharpCode(outdir string) error {
	var err error
	if err = os.MkdirAll(outdir, os.ModePerm); err != nil {
		return err
	}

	filename := kCSharpFilename
	fullname := path.Join(outdir, filename)
	w, err := os.Create(fullname)
	if err != nil {
		return err
	}
	defer w.Close()

	allTemplates := []string{
		tmplCSharpCode,
	}

	funcMap := template.FuncMap{
		"HasEnum": func(e []*EnumSchema) bool {
			return len(e) > 0
		},
		"TrimComment": func(comment string) string {
			return trimComment(comment)
		},
		"HasComments": func(comments []string, suffix string) bool {
			return len(comments) > 0 || len(suffix) > 0
		},
		"GetCSharpType": func(typename string, typeparams []string) string {
			res := self.getCSharpTypeStr(typename, typeparams)
			return res
		},
		"GetCSharpMioStr": func(typename string, typeparams []string) string {
			return self.makeCSharpMioStr(typename, typeparams)
		},
		"MakeMioGet": func(field *FieldSchema) string {
			return self.makeCSharpMioGet(field)
		},
		"MakeMioSet": func(field *FieldSchema) string {
			return self.makeCSharpMioSet(field)
		},
	}

	tmpl, err := CreateTemplate(filename, allTemplates, funcMap)

	err = tmpl.Execute(w, self)
	if err != nil {
		return err
	}

	fmt.Printf("Generating csharp code done: %s\n", fullname)
	return nil
}

func (self *MsgCompiler) getCSharpTypeStr(typename string, typeparams []string) string {
	switch typename {
	case "list":
		return fmt.Sprintf("List<%s>", self.getCSharpTypeStr(typeparams[0], nil))
	case "set":
		return fmt.Sprintf("HashSet<%s>", self.getCSharpTypeStr(typeparams[0], nil))
	case "map":
		return fmt.Sprintf("Dictionary<%s,%s>",
			self.getCSharpTypeStr(typeparams[0], nil), self.getCSharpTypeStr(typeparams[1], nil))
	case "bytes":
		return "byte[]"
	case "bool":
		return "bool"
	case "byte":
		return "byte"
	case "string":
		return "string"
	case "int32":
		return "int"
	case "int64":
		return "long"
	case "float":
		return "float"
	case "double":
		return "double"
	default:
		return typename
	}
}

func (self *MsgCompiler) makeCSharpMioStr(typename string, typeparams []string) string {
	if self.IsEnumType(typename) {
		return "Mio.I32"
	}
	switch typename {
	case "list", "set":
		return fmt.Sprintf("new MioList(%s)", self.makeCSharpMioStr(typeparams[0], nil))
	case "map":
		return fmt.Sprintf("new MioMap(%s,%s)",
			self.makeCSharpMioStr(typeparams[0], nil), self.makeCSharpMioStr(typeparams[1], nil))
	case "bytes":
		return "Mio.Binary"
	case "bool":
		return "Mio.Bool"
	case "byte":
		return "Mio.Byte"
	case "string":
		return "Mio.String"
	case "int32":
		return "Mio.I32"
	case "int64":
		return "Mio.I64"
	case "float":
		return "Mio.Float"
	case "double":
		return "Mio.Double"
	}
	return fmt.Sprintf("MIO_%s.I.M", typename)
}

func (self *MsgCompiler) makeCSharpMioInstV2(typename string) string {
	if self.IsEnumType(typename) {
		return fmt.Sprintf("MIO_Enum<%s>.I", typename)
	}
	switch typename {
	case "list", "set", "map":
		return fmt.Sprintf("<UnSupport '%s'>", typename)

	case "bytes":
		return "MIO_Binary.I"
	case "bool":
		return "MIO_Bool.I"
	case "byte":
		return "MIO_Byte.I"
	case "string":
		return "MIO_String.I"
	case "int32":
		return "MIO_I32.I"
	case "int64":
		return "MIO_I64.I"
	case "float":
		return "MIO_Float.I"
	case "double":
		return "MIO_Double.I"
	}
	return fmt.Sprintf("MIO_%s.I", typename)
}

func (self *MsgCompiler) makeCSharpMioGet(field *FieldSchema) string {
	if self.IsEnumType(field.TypeName) {
		return fmt.Sprintf("(%s)obj.GetInt((int)E.%s)", field.TypeName, field.FieldName)
	}
	switch field.TypeName {
	case "list", "set":
		return fmt.Sprintf("FromMioList(obj.GetObject((int)E.%s), %s)",
			field.FieldName, self.makeCSharpMioInstV2(field.TypeParams[0]))

	case "map":
		return fmt.Sprintf("FromMioDictionary(obj.GetObject((int)E.%s), %s, %s)",
			field.FieldName, self.makeCSharpMioInstV2(field.TypeParams[0]), self.makeCSharpMioInstV2(field.TypeParams[1]))

	case "bytes":
		return fmt.Sprintf("obj.GetBytes((int)E.%s)", field.FieldName)
	case "bool":
		return fmt.Sprintf("obj.GetBool((int)E.%s)", field.FieldName)
	case "byte":
		return fmt.Sprintf("obj.GetByte((int)E.%s)", field.FieldName)
	case "string":
		return fmt.Sprintf("obj.GetString((int)E.%s)", field.FieldName)
	case "int32":
		return fmt.Sprintf("obj.GetInt((int)E.%s)", field.FieldName)
	case "int64":
		return fmt.Sprintf("obj.GetLong((int)E.%s)", field.FieldName)
	case "float":
		return fmt.Sprintf("obj.GetFloat((int)E.%s)", field.FieldName)
	case "double":
		return fmt.Sprintf("obj.GetDouble((int)E.%s)", field.FieldName)
	}
	return fmt.Sprintf("MIO_%s.I.FromMioObject(obj.GetObject((int)E.%s))", field.TypeName, field.FieldName)
}

func isCollectionListOrSet(typename string) bool {
	return typename == "list" || typename == "set"
}

func isCollectionMap(typename string) bool {
	return typename == "map"
}
func isSimplePrimitive(typename string) bool {
	switch typename {
	case "bytes", "bool", "byte", "string", "int32", "int64", "float", "double":
		return true
	}
	return false
}

func (self *MsgCompiler) makeCSharpMioSet(field *FieldSchema) string {
	typename := field.TypeName
	typeparams := field.TypeParams
	fieldname := field.FieldName
	if isCollectionListOrSet(typename) {

		return fmt.Sprintf("ToMioList(data.%s, %s)", fieldname, self.makeCSharpMioInstV2(typeparams[0]))

	} else if isCollectionMap(typename) {

		return fmt.Sprintf("ToMioDictionary(data.%s, %s, %s)",
			fieldname, self.makeCSharpMioInstV2(typeparams[0]), self.makeCSharpMioInstV2(typeparams[1]))

	} else if isSimplePrimitive(typename) {

		return fmt.Sprintf("data.%s", fieldname)

	} else if self.IsEnumType(typename) {

		return fmt.Sprintf("(int)data.%s", fieldname)

	} else {

		return fmt.Sprintf("MIO_%s.I.ToMioObject(data.%s)", typename, fieldname)
	}
}
