/**
 * Date: 11/02/14
 * Time: 7:17 PM
 */
package opal

import (
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"time"
)

const (
	PrimaryKey reflect.Kind = iota + reflect.UnsafePointer + 1
	OpalTime
	Embedded
	DAO
)

var (
	lockOpal bool
	opal     OpalMagic = opalMagic
)

// Borrowed from sql library
var errNilPtr = errors.New("destination pointer is nil") // embedded in descriptive error

// Allows a user to set their own id
// Can only be run once
// Initial Gem startup will run this to check
func SetMagic(pId *OpalMagic) {
	if lockOpal {
		return
	}
	lockOpal = true
	if pId != nil {
		opal = *pId
	}
}

// *****************************************  Opal TYPE SYSTEM

// A simple type to return the global identifying opal
type OpalMagic int

// Interface Opal acts as a shoe horn for valid types in a model
type Opal interface {
	opal() OpalMagic
	Kind() reflect.Kind
}

// An embeddable struct with which to create your own types
type T struct{}

// T partially implements the opal.Opal interface
func (T) opal() OpalMagic {
	return opal
}

// T partially implements the opal.Opal interface
func (T) Kind() reflect.Kind {
	return Embedded
}

// **********************************************  SPECIAL TYPES

// Key is a special type of opal.Opal
type Key interface {
	Opal
}

// ************************************************  STRING TYPE

// String represents an int64 that may be null.
// String implements the Scanner interface so
// It can be used as a scan destination, similar to sql.NullString.
type String struct {
	Str *string
}

// String implements the opal.Opal interface
func (String) opal() OpalMagic {
	return opal
}

// Convenience constructor for a String
func NewString(p string) String {
	return String{&p}
}

// Convenience setting method
func (o *String) A(p string) {
	o.Str = &p
}

// String implements the sql.Scanner interface.
func (o *String) Scan(pValue interface{}) error {
	if pValue == nil {
		return nil
	}
	o.Str = new(string)
	switch s := pValue.(type) {
	case string:
		*o.Str = s
		return nil
	case *string:
		o.Str = s
		return nil
	case []byte:
		*o.Str = string(s)
		return nil
	}
	return errors.New("Opal.String: invalid value to scan into String")
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
type Slice struct {
	Slice []byte
}

// Slice implements the opal.Opal interface
func (Slice) opal() OpalMagic {
	return opal
}

// Convenience constructor for a Slice
func NewSlice(p []byte) Slice {
	return Slice{p}
}

// Convenience setting method
func (o *Slice) A(p []byte) {
	o.Slice = p
}

// Slice implements the sql.Scanner interface.
func (o *Slice) Scan(pValue interface{}) error {
	if pValue == nil {
		return nil
	}
	switch value := pValue.(type) {
	case []byte:
		o.Slice = make([]byte, len(value))
		copy(o.Slice, value)
		return nil
	case *[]byte:
		o.Slice = make([]byte, len(*value))
		copy(o.Slice, *value)
		return nil
	}
	//	if value, ok := pValue.([]byte); ok {
	//		*o = make([]byte, len(value))
	//		copy(*o, value)
	//		return nil
	//	}
	panic("Opal.String: invalid value to scan into String")
	return nil
}

// Slice implements the driver Valuer interface.
func (o *Slice) Value() (driver.Value, error) {
	return o.Slice, nil
}

// Returns the primitive type
func (o Slice) Kind() reflect.Kind {
	return reflect.Slice
} // TODO reduce slice to itself to get rid of scan

// Prints the value
func (o Slice) String() string {
	return string(o.Slice)
}

// **************************************************  INT64 TYPE

// Int64 represents an int64 that may be null.
// Int64 implements the Scanner interface so
// It can be used as a scan destination, similar to sql.NullString.
type Int64 struct {
	Int64 *int64
}

type Type interface {
	Scan(interface{}) error
	Interface() interface{}
	Value() (driver.Value, error)
	String() string
}

// Int64 implements the opal.Opal interface
func (Int64) opal() OpalMagic {
	return opal
}

// Convenience constructor for an Int64
func NewInt64(p int64) Int64 {
	return Int64{&p}
}

// Convenience setting method
func (o *Int64) A(p int64) {
	o.Int64 = &p
}

// Convenience setting method
func (o *Int64) Assign(p int64) {
	o.A(p)
}

// Int64 implements the Scanner interface.
func (o *Int64) Scan(pValue interface{}) error {
	if pValue == nil {
		return nil
	}
	o.Int64 = new(int64)
	switch value := pValue.(type) {
	case int64:
		*o.Int64 = value
		return nil
	case int:
		*o.Int64 = int64(value)
		return nil
	case *int64:
		v := int64(*value)
		o.Int64 = &v
		return nil
	case *int:
		v := int64(*value)
		o.Int64 = &v
		return nil
	}
	panic("Opal.Int64: invalid value to scan into Int64")
	return nil
}

// Int64 implements the driver Valuer interface.
func (o Int64) Value() (driver.Value, error) {
	return o.Interface(), nil
}

func (o Int64) Interface() interface{} {
	if o.Int64 == nil {
		return nil
	}
	return int64(*o.Int64)
}

// Returns the primitive value of Int64 if nil returns 0
// Same as Primitive()
func (o Int64) Primitive() int64 {
	if v := o.Interface(); v != nil {
		return v.(int64)
	}
	return 0 // TODO maybe return max instead
}

// Prints the value
func (o Int64) String() string {
	return fmt.Sprint(o.Interface())
}

// Returns the primitive type
func (o Int64) Kind() reflect.Kind {
	return reflect.Int64
} // TODO reduce slice to itself to get rid of scan


// Is an Int64 under the covers where its name flags its use
type AutoIncrement struct {
	Int64
}

// Makes a new AutoIncrement
func NewAutoIncrement(p int64) *AutoIncrement {
	return &AutoIncrement{NewInt64(p)}
}

// ************************************************  FLOAT64 TYPE

// Float64 represents an int64 that may be null.
// Float64 implements the Scanner interface so
// It can be used as a scan destination, similar to sql.NullString.
type Float64 struct {
	Float64 *float64
}

// Float64 implements the opal.Opal interface
func (Float64) opal() OpalMagic {
	return opal
}

// Convenience constructor for an Float64
func NewFloat64(p float64) Float64 {
	return Float64{&p}
}

// Convenience setting method
func (o *Float64) A(p float64) {
	o.Float64 = &p
}

// Float64 the Scanner interface.
func (o *Float64) Scan(pValue interface{}) error {
	if pValue == nil {
		return nil
	}
	o.Float64 = new(float64)
	switch value := pValue.(type) {
	case float64:
		*o.Float64 = value
		return nil
	case *float64:
		o.Float64 = value
		return nil
	}
	panic("Opal.Float64: invalid value to scan into Float64")
	return nil
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

// Bool implements the opal.Opal interface
func (Bool) opal() OpalMagic {
	return opal
}

// Convenience constructor for an Bool
func NewBool(p bool) Bool {
	return Bool{&p}
}

// Convenience setting method
func (o *Bool) A(p bool) {
	o.Bool = &p
}

// Bool implements the sql.Scanner interface.
func (o *Bool) Scan(pValue interface{}) error {
	if pValue == nil {
		return nil
	}
	o.Bool = new(bool)
	switch value := pValue.(type) {
	case bool:
		*o.Bool = value
		return nil
	case *bool:
		o.Bool = value
		return nil
	}
	panic("Opal.Bool: invalid value to scan into")
	return nil
}

// Bool implements the driver Valuer interface.
func (o Bool) Value() (driver.Value, error) {
	if o.Bool == nil {
		return nil, nil
	}
	return *o.Bool, nil
}

// Prints the value
func (o Bool) String() string {
	if o.Bool != nil {
		return fmt.Sprintf("%t", *o.Bool)
	}
	return fmt.Sprint(nil)
}

// Returns the primitive type
func (o Bool) Kind() reflect.Kind {
	return reflect.Bool
} // TODO reduce slice to itself to get rid of scan


// ************************************************  Date TYPE

// Time represents an time.Time that may be null.
// Time implements the Scanner interface so
// It can be used as a scan destination, similar to sql.NullString.
type Time struct {
	*time.Time
}

// Time implements the opal.Opal interface
func (Time) opal() OpalMagic {
	return opal
}

// Convenience constructor for an Bool
func NewTime(p time.Time) Time {
	return Time{&p}
}

// Convenience setting method
func (o *Time) A(p time.Time) {
	o.Time = &p
}

// Time implements the sql.Scanner interface.
func (o *Time) Scan(pValue interface{}) error {
	if pValue == nil {
		return nil
	}
	o.Time = new(time.Time)
	switch value := pValue.(type) {
	case time.Time:
		*o.Time = value
		return nil
	case *time.Time:
		o.Time = value
		return nil
	}
	// TODO better support for time as it can be string, bytes pointers...
	panic("Opal.Time: invalid value to scan into Time")
	return nil
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

func ABool(p bool) *bool {
	return &p
}

// convertAssign copies to dest the value in src, converting it if possible.
// An error is returned if the copy would result in loss of information.
// dest should be a pointer type.
func convertAssign(dest, src interface{}) error {
	// Common cases, without reflect.
	switch s := src.(type) {
	case string:
		switch d := dest.(type) {
		case *string:
			if d == nil {
				return errNilPtr
			}
			*d = s
			return nil
		case *[]byte:
			if d == nil {
				return errNilPtr
			}
			*d = []byte(s)
			return nil
		}
	case []byte:
		switch d := dest.(type) {
		case *string:
			if d == nil {
				return errNilPtr
			}
			*d = string(s)
			return nil
		case *interface{}:
			if d == nil {
				return errNilPtr
			}
			*d = cloneBytes(s)
			return nil
		case *[]byte:
			if d == nil {
				return errNilPtr
			}
			*d = cloneBytes(s)
			return nil
		case *sql.RawBytes:
			if d == nil {
				return errNilPtr
			}
			*d = s
			return nil
		}
	case nil:
		switch d := dest.(type) {
		case *interface{}:
			if d == nil {
				return errNilPtr
			}
			*d = nil
			return nil
		case *[]byte:
			if d == nil {
				return errNilPtr
			}
			*d = nil
			return nil
		case *sql.RawBytes:
			if d == nil {
				return errNilPtr
			}
			*d = nil
			return nil
		}
	}

	var sv reflect.Value

	switch d := dest.(type) {
	case *string:
		sv = reflect.ValueOf(src)
		switch sv.Kind() {
		case reflect.Bool,
			reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
			reflect.Float32, reflect.Float64:
			*d = fmt.Sprintf("%v", src)
			return nil
		}
	case *[]byte:
		sv = reflect.ValueOf(src)
		switch sv.Kind() {
		case reflect.Bool,
			reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
			reflect.Float32, reflect.Float64:
			*d = []byte(fmt.Sprintf("%v", src))
			return nil
		}
	case *sql.RawBytes:
		sv = reflect.ValueOf(src)
		switch sv.Kind() {
		case reflect.Bool,
			reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
			reflect.Float32, reflect.Float64:
			*d = sql.RawBytes(fmt.Sprintf("%v", src))
			return nil
		}
	case *bool:
		bv, err := driver.Bool.ConvertValue(src)
		if err == nil {
			*d = bv.(bool)
		}
		return err
	case *interface{}:
		*d = src
		return nil
	}

	if scanner, ok := dest.(sql.Scanner); ok {
		return scanner.Scan(src)
	}

	dpv := reflect.ValueOf(dest)
	if dpv.Kind() != reflect.Ptr {
		return errors.New("destination not a pointer")
	}
	if dpv.IsNil() {
		return errNilPtr
	}

	if !sv.IsValid() {
		sv = reflect.ValueOf(src)
	}

	dv := reflect.Indirect(dpv)
	if dv.Kind() == sv.Kind() {
		dv.Set(sv)
		return nil
	}

	switch dv.Kind() {
	case reflect.Ptr:
		if src == nil {
			dv.Set(reflect.Zero(dv.Type()))
			return nil
		} else {
			dv.Set(reflect.New(dv.Type().Elem()))
			return convertAssign(dv.Interface(), src)
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		s := asString(src)
		i64, err := strconv.ParseInt(s, 10, dv.Type().Bits())
		if err != nil {
			return fmt.Errorf("converting string %q to a %s: %v", s, dv.Kind(), err)
		}
		dv.SetInt(i64)
		return nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		s := asString(src)
		u64, err := strconv.ParseUint(s, 10, dv.Type().Bits())
		if err != nil {
			return fmt.Errorf("converting string %q to a %s: %v", s, dv.Kind(), err)
		}
		dv.SetUint(u64)
		return nil
	case reflect.Float32, reflect.Float64:
		s := asString(src)
		f64, err := strconv.ParseFloat(s, dv.Type().Bits())
		if err != nil {
			return fmt.Errorf("converting string %q to a %s: %v", s, dv.Kind(), err)
		}
		dv.SetFloat(f64)
		return nil
	}

	return fmt.Errorf("unsupported driver -> Scan pair: %T -> %T", src, dest)
}

func asString(src interface{}) string {
	switch v := src.(type) {
	case string:
		return v
	case []byte:
		return string(v)
	}
	return fmt.Sprintf("%v", src)
}

func cloneBytes(b []byte) []byte {
	if b == nil {
		return nil
	} else {
		c := make([]byte, len(b))
		copy(c, b)
		return c
	}
}
