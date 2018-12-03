package main

import (
	"fmt"
	"os"
)

/////////////////////////////////////////////////////////////////////// Errors Text Generation

func (self *MsgCompiler) GenerateErrorsText(errorsfile string) error {
	file, err := os.Create(errorsfile)
	if err != nil {
		return err
	}
	defer file.Close()

	fmt.Printf("Generating file: %s\n", errorsfile)
	for _, msg := range self.EnumList {
		if msg.Name == kErrCodeEnumName {
			for _, field := range msg.Fields {
				str := trimComment(field.SuffixComment)
				fmt.Fprintf(file, "%s = %s\n", field.FieldName, str)
			}
		}
	}
	return nil
}
