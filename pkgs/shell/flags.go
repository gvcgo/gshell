package shell

import "fmt"

func GetFlagKey(parent, child string) string {
	if parent == "" && child != "" {
		return child
	} else if parent != "" && child != "" {
		return fmt.Sprintf("%s.%s", parent, child)
	}
	return ""
}

type FlagType string

const (
	OptionTypeString FlagType = "string"
	OptionTypeBool   FlagType = "bool"
	OptionTypeInt    FlagType = "int"
	OptionTypeFloat  FlagType = "float"
)

/*
Flags
*/
type IShellFlag interface {
	GetName() string
	GetShort() string
	GetType() FlagType
	GetDefault() string
	GetUsage() string
}

type Flag struct {
	Name    string   // flag name
	Short   string   // flag shorthand
	Type    FlagType // flag type
	Default string   // default value
	Usage   string   // flag help info
}

func (f *Flag) GetName() string {
	return f.Name
}

func (f *Flag) GetShort() string {
	return f.Short
}

func (f *Flag) GetType() FlagType {
	return f.Type
}

func (f *Flag) GetDefault() string {
	return f.Default
}

func (f *Flag) GetUsage() string {
	return f.Usage
}
