package main

import (
	"fmt"
	"sort"
)

// this file contains everything related to the metadata we need about
// the go types that are supported by the config definition.

type overrideExprType int

const (
	// compare against empty string
	osCompareString overrideExprType = iota
	// compare length of slice
	osLen
	// compare to literal 0
	osCompareZero
	// compare to a nil
	osCompareNil
	// generate a container based override instead
	osStruct
)

// typeInfo is metadata about a go type that is usable from the config definition
type typeInfo struct {
	// Name is the go name of the type, e.g. string, or []uint64
	Name string
	// OverrideFunc is the name of the generated function that can apply an override of this type.
	OverrideFunc string
	// overrideStyle indicates what type of expression is needed to decided to apply the override?
	overrideStyle overrideExprType
	// ExampleValues contains 3 different example values for this type, as go source code.
	ExampleValues []string
	// if the type is a struct, a pointer to the struct definition for this type
	structDef *structInfo
}

// OverrideExpr generates the source code of the expression that checks if
// the override should be applied, it is go type specific
// it assumes that the variable containing the override value is called 'o'
func (t *typeInfo) OverrideExpr() string {
	return t.overrideStyle.overrideExpr()
}

func (o overrideExprType) overrideExpr() string {
	switch o {
	case osCompareString:
		return `*o != ""`
	case osLen:
		return `len(*o) > 0`
	case osCompareZero:
		return `*o != 0`
	case osCompareNil:
		return `*o != nil`
	}
	return fmt.Sprintf(`*UNEXPECTED_OVERRIDE_STYLE_%d`, o)
}

// RequiresOverrideImpl returns true if this type requires an override impl
// that checks the override value first [i.e. the simple type overrides]
// if this returns false, then its typically a type that delegates overrides
// to other types [like the overrideFrom method that is generated for structs]
func (t *typeInfo) RequiresOverrideImpl() bool {
	return t.overrideStyle != osStruct
}

type typeInfos []*typeInfo

//
// Sort interface: Len, Less, Swap
//

func (t typeInfos) Len() int {
	return len(t)
}

func (t typeInfos) Less(a, b int) bool {
	return t[a].OverrideFunc < t[b].OverrideFunc
}

func (t typeInfos) Swap(a, b int) {
	t[a], t[b] = t[b], t[a]
}

var stdTypes typeInfos
var stdTypesByName map[string]*typeInfo

func init() {
	base := []typeInfo{
		{"string", "overrideString", osCompareString, []string{`"one"`, `"two"`, `"three"`}, nil},
		{"*bool", "overrideBool", osCompareNil, []string{`&trueVal`, `&falseVal`, `&trueVal`}, nil},
		{"int", "overrideInt", osCompareZero, []string{`-42`, `42`, `1234`}, nil},
		{"int64", "overrideInt64", osCompareZero, []string{`int64(-1234)`, `int64(123)`, `int64(19413241)`}, nil},
		{"uint64", "overrideUint64", osCompareZero, []string{`uint64(42)`, `uint64(1)`, `uint64(1234132)`}, nil},
		{"float64", "overrideFloat64", osCompareZero, []string{`float64(0.1)`, `float64(1.0)`, `float64(42.42)`}, nil},
		{"Duration", "overrideDuration", osCompareZero, []string{`Duration(time.Second)`, `Duration(time.Minute)`, `Duration(time.Hour)`}, nil},
		{"[]string", "overrideStrings", osLen, []string{`[]string{"a"}`, `[]string{"b", "b"}`, `[]string{"c", "c", "c"}`}, nil},
		{"[]bool", "overrideBools", osLen, []string{`[]bool{true}`, `[]bool{false}`, `[]bool{false,true}`}, nil},
		{"[]int", "overrideInts", osLen, []string{`[]int{1}`, `[]int{2, 2}`, `[]int{3, 3, 3}`}, nil},
		{"[]int64", "overrideInt64s", osLen, []string{`[]int64{1}`, `[]int64{-1, 1}`, `[]int64{-1, 0, 1}`}, nil},
		{"[]uint64", "overrideUint64s", osLen, []string{`[]uint64{1}`, `[]uint64{0, 1}`, `[]uint64{65}`}, nil},
		{"[]float64", "overrideFloat64s", osLen, []string{`[]float64{0}`, `[]float64{1, 1}`, `[]float64{-1, 0, 42.42}`}, nil},
		{"[]Duration", "overrideDurations", osLen, []string{`[]Duration{Duration(time.Second)}`, `[]Duration{Duration(time.Second), Duration(time.Minute)}`, `[]Duration{Duration(time.Hour), Duration(time.Minute *3), Duration(time.Second * 10)}`}, nil},
	}
	stdTypes = make([]*typeInfo, len(base))
	stdTypesByName = make(map[string]*typeInfo, len(base))
	for idx := range stdTypes {
		stdTypesByName[base[idx].Name] = &base[idx]
		stdTypes[idx] = &base[idx]
	}
	sort.Sort(stdTypes)
}

// stdTypeNames returns the list of all the go names of all the supported simple types
func stdTypeNames() []string {
	res := make([]string, len(stdTypes))
	for idx := range stdTypes {
		res[idx] = stdTypes[idx].Name
	}
	return res
}
