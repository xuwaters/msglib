package main

import (
	"bytes"
	"fmt"
	"os"
	"path"
	"strings"
	"text/template"
)

/////////////////////////////////////////////////////////////////////// Java Code Generation

var tmplJavaCode string = `
{{- define "Enum"}}
    /**{{range $i,$line := .Comments}}
     * {{TrimComment $line }}{{end}}
     */
    public static enum {{.Name}} {{print "{"}}
    {{- range $i, $field := .Fields}}{{with $field}}
        {{- if len .Comments | lt 0}}
        /**{{range $i,$line := .Comments}}
         * {{TrimComment $line }}{{end}}
         */
        {{- end}}
        {{.FieldName}}({{.FieldValue}}), {{.SuffixComment}}
    {{- end}}{{end}}
        ;
        private final int code;
        private final String[] params;

        private {{.Name}}(int code, String ...params) {
            this.code = code;
            this.params = params;
        }

        public int getCode() {
            return code;
        }

        public static {{.Name}} getByCode(int value) {
            for ({{.Name}} e : values()) {
                if (e.getCode() == value) {
                    return e;
                }
            }
            return null;
        }
{{- if eq .Name "ErrCode"}}

        // ErrData methods
    {{- range $i, $field := .Fields}}{{with $field}}{{$args := GetErrCodeArgs .SuffixComment}}
        /**
         * {{TrimComment .SuffixComment}}
         */
        public static ErrData make{{.FieldName}}({{MakeMethodParams $args}}) {
            ErrData data = new ErrData();
            data.code = {{.FieldName}}.code;
            data.keyhash = CryptoUtil.makeDBHash("{{.FieldName}}");
            data.args = Arrays.asList(new String[] { {{MakeDataList $args}} });
            return data;
        }

    {{- end}}{{end}}
{{- end}}
    }
{{- end}}
{{- define "Message"}}
    /**{{range $i,$line := .Comments}}
     * {{TrimComment $line }}{{end}}
     */
    @MsgStruct
    public static final class {{.Name}} {{print "{"}}
    {{- range $i, $field := .Fields}}{{with $field}}
        {{- if HasComments .Comments .SuffixComment}}
        /**{{range $i,$line := .Comments}}
         * {{TrimComment $line }}{{end}}
         * {{TrimComment .SuffixComment}}
         */
        {{- end}}
        @MsgField(Id = {{.FieldID}})
        {{- $extra := GetJavaExtraAnnotation .TypeName .TypeParams}}{{if eq $extra "" | not}}
        {{$extra}}
        {{- end}}
        public {{GetJavaType .TypeName .TypeParams}} {{.FieldName}};

    {{- end}}{{end}}
    }
{{- end}}
{{- define "Main" -}}
package {{ .JavaPackage }};

import java.util.*;
import msglib.annotation.*;
{{if HasEnum .EnumList -}}
import com.haoyou.common.crypto.*;
{{- end}}

public final class {{.ProtoName}} {
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

func (self *MsgCompiler) GenerateJavaCode(outdir string) error {
	var err error
	if err = os.MkdirAll(outdir, os.ModePerm); err != nil {
		return err
	}

	filename := self.ProtoName + ".java"
	fullname := path.Join(outdir, filename)
	w, err := os.Create(fullname)
	if err != nil {
		return err
	}
	defer w.Close()

	allTemplates := []string{
		tmplJavaCode,
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
		"GetErrCodeArgs": func(comment string) []string {
			args := ParseArgsFromErrMessage(comment)
			return args
		},
		"MakeMethodParams": func(args []string) string {
			if len(args) == 0 {
				return ""
			}
			return "String " + strings.Join(args, ", String ")
		},
		"MakeDataList": func(args []string) string {
			buff := bytes.NewBuffer(nil)
			for i, arg := range args {
				if i > 0 {
					buff.WriteString(", ")
				}
				buff.WriteRune('"')
				buff.WriteString(arg)
				buff.WriteString(`", `)
				buff.WriteString(arg)
			}
			return buff.String()
		},
		"GetJavaType": func(typename string, typeparams []string) string {
			res := self.getJavaTypeStr(typename, typeparams)
			return res
		},
		"GetJavaExtraAnnotation": func(typename string, typeparams []string) string {
			res := self.getJavaExtraAnnotation(typename, typeparams)
			return res
		},
	}

	tmpl, err := CreateTemplate(filename, allTemplates, funcMap)

	err = tmpl.Execute(w, self)
	if err != nil {
		return err
	}

	fmt.Printf("Generating java code done: %s\n", fullname)
	return nil
}

func (self *MsgCompiler) getJavaTypeStr(typename string, typeparams []string) string {

	if self.IsEnumType(typename) {
		return "int"
	}
	switch typename {
	case "list":
		return fmt.Sprintf("List<%s>", self.getJavaObjectTypeStr(typeparams[0]))
	case "set":
		return fmt.Sprintf("Set<%s>", self.getJavaObjectTypeStr(typeparams[0]))
	case "map":
		return fmt.Sprintf("Map<%s, %s>", self.getJavaObjectTypeStr(typeparams[0]), self.getJavaObjectTypeStr(typeparams[1]))
	case "bytes":
		return "byte[]"
	case "bool":
		return "boolean"
	case "byte":
		return "byte"
	case "string":
		return "String"
	case "int16":
		return "short"
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

func (self *MsgCompiler) getJavaObjectTypeStr(typename string) string {
	if self.IsEnumType(typename) {
		return "Integer"
	}
	switch typename {
	case "bool":
		return "Boolean"
	case "byte":
		return "Byte"
	case "string":
		return "String"
	case "int16":
		return "Short"
	case "int32":
		return "Integer"
	case "int64":
		return "Long"
	case "float":
		return "Float"
	case "double":
		return "Double"
	default:
		return typename
	}
}

func (self *MsgCompiler) getJavaExtraAnnotation(typename string, typeparams []string) string {
	switch typename {
	case "list", "set":
		return fmt.Sprintf("@MsgList(ElemClass = %s.class)", self.getJavaObjectTypeStr(typeparams[0]))
	case "map":
		return fmt.Sprintf("@MsgMap(KeyClass = %s.class, ValueClass = %s.class)",
			self.getJavaObjectTypeStr(typeparams[0]), self.getJavaObjectTypeStr(typeparams[1]))
	}
	return ""
}
