package bean

import (
	"encoding/json"
	"fmt"
	"strconv"
)

// ------------

// Format defines the type of the VariableObject.Value
type Format string

const (
	// FormatTypeString is the string type
	FormatTypeString Format = "STRING"
	// FormatTypeNumber is the number type
	FormatTypeNumber Format = "NUMBER"
	// FormatTypeBool is the boolean type
	FormatTypeBool Format = "BOOL"
	// FormatTypeDate is the date type
	FormatTypeDate Format = "DATE"
)

func (d Format) ValuesOf(format string) (Format, error) {
	if format == "NUMBER" || format == "number" {
		return FormatTypeNumber, nil
	} else if format == "BOOL" || format == "bool" || format == "boolean" {
		return FormatTypeBool, nil
	} else if format == "STRING" || format == "string" {
		return FormatTypeString, nil
	} else if format == "DATE" || format == "date" {
		return FormatTypeDate, nil
	}
	return FormatTypeString, fmt.Errorf("invalid Format: %s", format)
}

func (d Format) String() string {
	return string(d)
}

func (d Format) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.String())
}

func (d *Format) UnmarshalJSON(b []byte) error {
	var s string
	err := json.Unmarshal(b, &s)
	if err != nil {
		return err
	}
	format, err := d.ValuesOf(s)
	if err != nil {
		return err
	}
	*d = format
	return nil
}

func (d Format) Convert(value string) (interface{}, error) {
	switch d {
	case FormatTypeString:
		return value, nil
	case FormatTypeNumber:
		return strconv.ParseFloat(value, 8)
	case FormatTypeBool:
		return strconv.ParseBool(value)
	case FormatTypeDate:
		return value, nil
	default:
		return nil, fmt.Errorf("unsupported datatype")
	}
}

// VariableType defines the type of the VariableObject
type VariableType string

const (
	// VariableTypeValue is used to define new VariableObject value
	VariableTypeValue VariableType = "VALUE"
	// VariableTypeRefPreCi is used to refer to a VariableObject from the previous PRE-CI stage
	VariableTypeRefPreCi VariableType = "REF_PRE_CI"
	// VariableTypeRefPostCi is used to refer to a VariableObject from the previous POST-CI stage
	VariableTypeRefPostCi VariableType = "REF_POST_CI"
	// VariableTypeRefGlobal is used to refer to a VariableObject from the global scope
	VariableTypeRefGlobal VariableType = "REF_GLOBAL"
	// VariableTypeRefPlugin is used to refer to a VariableObject from the previous plugin scope
	VariableTypeRefPlugin VariableType = "REF_PLUGIN"
)

// String returns the string representation of the VariableType
func (t VariableType) String() string {
	return string(t)
}

// MarshalJSON marshals the VariableType into a JSON string
// Note: Json.Marshal will call this function internally for custom type marshalling
func (t VariableType) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.String())
}

// UnmarshalJSON unmarshal a JSON string into a VariableType
// Note: Json.Unmarshal will call this function internally for custom type unmarshalling
func (t *VariableType) UnmarshalJSON(data []byte) error {
	var variableType VariableType
	err := json.Unmarshal(data, &variableType)
	if err != nil {
		return err
	}
	switch variableType {
	case VariableTypeValue,
		VariableTypeRefPreCi,
		VariableTypeRefPostCi,
		VariableTypeRefGlobal,
		VariableTypeRefPlugin:
		*t = variableType
		return nil
	default:
		// If the variableType is not one of the above, return an error
		// This error will be returned by the Json.Unmarshal function
		return fmt.Errorf("invalid variableType %s", data)
	}
}

// ---------------

// VariableObject defines the structure of an environment variable
//   - Name: name of the variable
//   - Format: type of the variable value.
//     Possible values are STRING, NUMBER, BOOL, DATE
//   - Value: value of the variable
//   - VariableType: defines the scope-type of the variable.
//     Possible values are VALUE, REF_PRE_CI, REF_POST_CI, REF_GLOBAL, REF_PLUGIN
//   - ReferenceVariableName: name of the variable to refer to
//   - ReferenceVariableStepIndex: index of the script step to refer to
//   - VariableStepIndexInPlugin: index of the variable in the plugin
//   - TypedValue: formatted value of the variable after type conversion.
//     This field is for internal use only (not exposed in the JSON)
type VariableObject struct {
	Name   string `json:"name"`
	Format Format `json:"format"`
	// only for input type
	Value                      string       `json:"value"`
	VariableType               VariableType `json:"variableType"`
	ReferenceVariableName      string       `json:"referenceVariableName"`
	ReferenceVariableStepIndex int          `json:"referenceVariableStepIndex"`
	VariableStepIndexInPlugin  int          `json:"variableStepIndexInPlugin"`
	TypedValue                 interface{}  `json:"-"` // TypedValue is the formatted value of the variable after type conversion
}

// TypeCheck converts the VariableObject.Value to the VariableObject.Format type
// and stores the formatted value in the VariableObject.TypedValue field.
// If the conversion fails, it returns an error.
func (v *VariableObject) TypeCheck() error {
	typedValue, err := v.Format.Convert(v.Value)
	if err != nil {
		return err
	}
	v.TypedValue = typedValue
	return nil
}
