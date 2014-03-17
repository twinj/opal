/**
 * Date: 11/02/14
 * Time: 7:17 PM
 */
package opal

import (
	"database/sql/driver"
	"errors"
	"fmt"
	_ "log"
	"reflect"
	"time"
)

const (
	PrimaryKey reflect.Kind = iota + reflect.UnsafePointer + 1
	OpalTime
	Embedded
	DAO
)

var (
	opal        OpalMagic = opalMagic
	nilPtrError           = errors.New("Destination pointer is nil") // embedded in descriptive error
)

// *****************************************  OPAL TYPE SYSTEM

// TODO allow programmers to specify their own magic
// A simple type to return the identifying global opal
type OpalMagic int

// Interface OPAL acts as a shoe horn for valid types in a model
type OPAL interface {
	opal() OpalMagic
	Kind() reflect.Kind
}

// An embeddable struct with which to create your own types
type Opal struct{}

// Opal partially implements the opal.OPAL interface
func (Opal) opal() OpalMagic {
	return opal
}

// Opal partially implements the opal.OPAL interface
func (Opal) Kind() reflect.Kind {
	return Embedded
}

func Kind() reflect.Kind {
	var opal OPAL
	switch opal.(type) {
	case String:
		return reflect.String
	case Int64:
		return reflect.Int64
	case Float64:
		return reflect.Float64
	case Bool:
		return reflect.Bool
		// TODO FINISH TYPES
	default:
		panic("Opal.Kind: non supported type")
	}
}

// **********************************************  SPECIAL TYPES

// Key is a special type of opal.OPAL
type Key interface {
OPAL
}

// ************************************************  STRING TYPE

// String represents an int64 that may be null.
// String implements the Scanner interface so
// It can be used as a scan destination, similar to sql.NullString.
type String struct {
	Str *string
}

// String implements the opal.OPAL interface
func (String) opal() OpalMagic {
	return opal
}

// Convenience constructor for a String
func NewString(p string) String {
	return String{&p}
}

// Convenience setting method
func (o *String) Set(p string) {
	o.Str = &p
}

// String implements the sql.Scanner interface.
func (o *String) Scan(pValue interface{}) error {
	o.Str = new(string)
	switch s := pValue.(type) {
	case string:
		*o.Str = s
		return nil
	case []byte:
		*o.Str = string(s)
		return nil
	}
	if pValue == nil {
		return nil
	}
	return errors.New("Opal.String: invalid value to scan into")
}

// String implements the driver Valuer interface.
func (o String) Value() (driver.Value, error) {
	if o.Str == nil {
		return nil, nil
	}
	return *o.Str, nil
}

// Returns the primitive type
func (o String) Kind() reflect.Kind {
	return reflect.String
}

// Prints the value
func (o String) String() string {
	if o.Str != nil {
		return *o.Str
	}
	return fmt.Sprint(nil)
}

// ************************************************  Slice TYPE

// Slice implements the Scanner interface so
// It can be used as a scan destination, similar to sql.NullString.
type Slice []byte

// Slice implements the opal.OPAL interface
func (Slice) opal() OpalMagic {
	return opal
}

// Convenience constructor for a Slice
func NewSlice(p []byte) Slice {
	return Slice(p)
}

// Convenience setting method
func (o *Slice) Set(p []byte) {
	*o = Slice(p)
}

// Slice implements the sql.Scanner interface.
func (o *Slice) Scan(pValue interface{}) error {
	if value, ok := pValue.([]byte); ok {
		*o = make([]byte, len(value))
		copy(*o, value)
		return nil
	}
	if pValue == nil {
		return nil
	}
	return errors.New("Opal.Slice: invalid value to scan into Int64")
}

// Slice implements the driver Valuer interface.
func (o *Slice) Value() (driver.Value, error) {
	if *o == nil {
		return nil, nil
	}
	return *o, nil
}

// Returns the primitive type
func (o Slice) Kind() reflect.Kind {
	return reflect.Slice
}

// Prints the value
func (o Slice) String() string {
	return string(o)
}

// **************************************************  INT64 TYPE

// Int64 represents an int64 that may be null.
// Int64 implements the Scanner interface so
// It can be used as a scan destination, similar to sql.NullString.
type Int64 struct {
	Int64 *int64
}

// Int64 implements the opal.OPAL interface
func (Int64) opal() OpalMagic {
	return opal
}

// Convenience constructor for an Int64
func NewInt64(p int64) Int64 {
	return Int64{&p}
}

// Convenience setting method
func (o *Int64) Set(p int64) {
	o.Int64 = &p
}

// Int64 implements the Scanner interface.
func (o *Int64) Scan(pValue interface{}) error {
	if value, ok := pValue.(int64); ok {
		o.Int64 = new(int64)
		*o.Int64 = value
		return nil
	}
	if pValue == nil {
		return nil
	}
	return errors.New("Opal.Int64: invalid value to scan into Int64")
}

// Int64 implements the driver Valuer interface.
func (o Int64) Value() (driver.Value, error) {
	if o.Int64 == nil {
		return nil, nil
	}
	return *o.Int64, nil
}

// Returns the primitive type
func (o Int64) Kind() reflect.Kind {
	return reflect.Int64
}

// Prints the value
func (o Int64) String() string {
	if o.Int64 != nil {
		return fmt.Sprintf("%d", *o.Int64)
	}
	return fmt.Sprint(nil)
}

// ************************************************  FLOAT64 TYPE

// Float64 represents an int64 that may be null.
// Float64 implements the Scanner interface so
// It can be used as a scan destination, similar to sql.NullString.
type Float64 struct {
	Float64 *float64
}

// Float64 implements the opal.OPAL interface
func (Float64) opal() OpalMagic {
	return opal
}

// Convenience constructor for an Float64
func NewFloat64(p float64) Float64 {
	return Float64{&p}
}

// Convenience setting method
func (o *Float64) Set(p float64) {
	o.Float64 = &p
}

// Float64 the Scanner interface.
func (o *Float64) Scan(pValue interface{}) error {
	if value, ok := pValue.(float64); ok {
		o.Float64 = new(float64)
		*o.Float64 = value
		return nil
	}
	if pValue == nil {
		return nil
	}
	return errors.New("Opal.Float64: invalid value to scan into")
}

// Float64 implements the driver Valuer interface.
func (o Float64) Value() (driver.Value, error) {
	if o.Float64 == nil {
		return nil, nil
	}
	return *o.Float64, nil
}

// Returns the primitive type
func (o Float64) Kind() reflect.Kind {
	return reflect.Float64
}

// Prints the value
func (o Float64) String() string {
	if o.Float64 != nil {
		return fmt.Sprintf("%f", *o.Float64)
	}
	return fmt.Sprint(nil)
}

// ************************************************  Bool TYPE

// Bool represents an int64 that may be null.
// Bool implements the Scanner interface so
// It can be used as a scan destination, similar to sql.NullString.
type Bool struct {
	Bool *bool
}

// Bool implements the opal.OPAL interface
func (Bool) opal() OpalMagic {
	return opal
}

// Convenience constructor for an Bool
func NewBool(p bool) Bool {
	return Bool{&p}
}

// Convenience setting method
func (o *Bool) Set(p bool) {
	o.Bool = &p
}

// Bool implements the sql.Scanner interface.
func (o *Bool) Scan(pValue interface{}) error {
	if value, ok := pValue.(bool); ok {
		o.Bool = new(bool)
		*o.Bool = value
		return nil
	}
	if pValue == nil {
		return nil
	}
	return errors.New("Opal.Bool: invalid value to scan into")
}

// Bool implements the driver Valuer interface.
func (o Bool) Value() (driver.Value, error) {
	if o.Bool == nil {
		return nil, nil
	}
	return *o.Bool, nil
}

// Returns the primitive type
func (o Bool) Kind() reflect.Kind {
	return reflect.Bool
}

// Prints the value
func (o Bool) String() string {
	if o.Bool != nil {
		return fmt.Sprintf("%t", *o.Bool)
	}
	return fmt.Sprint(nil)
}

// ************************************************  Date TYPE

// Time represents an time.Time that may be null.
// Time implements the Scanner interface so
// It can be used as a scan destination, similar to sql.NullString.
type Time struct {
	*time.Time
}

// Time implements the opal.OPAL interface
func (Time) opal() OpalMagic {
	return opal
}

// Convenience constructor for an Bool
func NewTime(p time.Time) Time {
	return Time{&p}
}

// Convenience setting method
func (o *Time) Set(p time.Time) {
	o.Time = &p
}

// Time implements the sql.Scanner interface.
func (o *Time) Scan(pValue interface{}) error {
	if value, ok := pValue.(time.Time); ok {
		o.Time = new(time.Time)
		*o.Time = value
		return nil
	}
	if pValue == nil {
		return nil
	}
	return errors.New("Opal.Time: invalid value to scan into")
}

// Time implements the driver Valuer interface.
func (o Time) Value() (driver.Value, error) {
	if o.Time == nil {
		return nil, nil
	}
	return *o.Time, nil
}

// Returns the primitive type
func (o Time) Kind() reflect.Kind {
	return reflect.String
}

// Prints the value
func (o Time) String() string {
	if o.Time != nil {
		return o.Time.String()
	}
	return fmt.Sprint(nil)
}

// ************************************************
