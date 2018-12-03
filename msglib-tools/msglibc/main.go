package main

import (
	"flag"
	"fmt"
	"os"
)

const (
	kVersion         = "1.0.0 (by xuwaters@gmail.com)"
	kErrCodeEnumName = "ErrCode"
	kCSharpFilename  = "messages.cs"
)

func main() {
	var (
		PrintVersion bool
		Language     string
		ProtoFile    string
		Outdir       string // output directory for
		ErrCodeFile  string // output file for generated error messages
	)
	flag.BoolVar(&PrintVersion, "version", false, "Print version")
	flag.StringVar(&ProtoFile, "proto", "", "'.proto' file path")
	flag.StringVar(&Language, "language", "", "generated language, supported options: java, csharp")
	flag.StringVar(&Outdir, "outdir", "out", "output directory for generated source code")
	flag.StringVar(&ErrCodeFile, "errors", "", "output file for error texts, empty value indicates do not generate error texts")
	flag.Parse()
	if !flag.Parsed() {
		flag.Usage()
		return
	}
	if PrintVersion {
		fmt.Printf("Version %s\n", kVersion)
		return
	}
	if Language == "" || ProtoFile == "" {
		flag.Usage()
		return
	}

	var err error
	compiler := NewMsgCompiler()
	if err = compiler.ParseProtoFile(ProtoFile); err != nil {
		fmt.Fprintf(os.Stderr, "ProtoFile format error: %v\n", err)
		return
	}

	//
	switch Language {
	case "java":
		err = compiler.GenerateJavaCode(Outdir)
	case "csharp":
		err = compiler.GenerateCSharpCode(Outdir)
	default:
		fmt.Fprintf(os.Stderr, "Unsupported language %s\n", Language)
		return
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating code: %v\n", err)
		return
	}

	if ErrCodeFile != "" {
		err = compiler.GenerateErrorsText(ErrCodeFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error generating errorcode file: %v\n", err)
			return
		}
	}
}
