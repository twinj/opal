/**
 * Date: 9/02/14
 * Time: 7:03 PM
 */
package opal

import (
	"database/sql"
	"fmt"
	"log"
	"reflect"
	"strings"
)

type ModelMetadata struct {
	Model
	this  reflect.Type
	table Table

	// The metadata columns
	columns            []Column
	columnsByIndex     map[int]*Column
	columnsByFieldName map[string]*Column
	keysByIndex        map[int]*Column
	keysByFieldName    map[string]*Column
	// TODO maps are not guaranteed to keep insert order
	// Sort an int array
	// Also combine two maps where necessary here for
	// easier insertion of parameters

	// Prepared query store
	preparedStatements map[string]*sql.Stmt

	// A map of a columns standard Validators
	// mapped by field name
	validators map[string][]Validator
}

type Validator int

const (
	Presence Validator = iota
	Length
)

type Validate interface {
	Validate(pOpal OPAL) bool
}

/**
 * Hold special data about a domain object model
 */
func NewMetadata(pModel Model, pType reflect.Type) *ModelMetadata {
	o := new(ModelMetadata)
	o.this = pType
	o.Model = pModel
	o.columnsByIndex = make(map[int]*Column)
	o.columnsByFieldName = make(map[string]*Column)
	o.keysByIndex = make(map[int]*Column)
	o.keysByFieldName = make(map[string]*Column)
	o.preparedStatements = make(map[string]*sql.Stmt)
	return o
}

func (o *ModelMetadata) addStmt(pDB *sql.DB, pKey string, pValue Sql) {
	sql := pValue.String()
	log.Printf("Opal.ModelMetadata.addStmt: %s", sql)
	stmt, err := pDB.Prepare(sql)
	if err != nil {
		panic(err) // TODO wtf?
	}
	o.preparedStatements[pKey] = stmt
	log.Printf("Opal.ModelMetadata.addStmt: %#v", o.preparedStatements[pKey])

}

/**
 * Get the columns metadata
 */
func (o ModelMetadata) Table() Table {
	return o.table
}

/**
 * Get the columns metadata
 */
func (o ModelMetadata) Columns() []Column {
	return o.columns
}

/**
 * Get the column metadata by its natural index
 */
func (o ModelMetadata) ColumnByIndex(index int) Column {
	return o.columns[index]
}

/**
 * Get the column metadata by the domains field name
 */
func (o ModelMetadata) Column(pField string) Column {
	return *o.columnsByFieldName[pField]
}

/**
 * Get the column metadata by the domains field index
 */
func (o ModelMetadata) ColumnByFieldIndex(pIndex int) Column {
	return *o.columnsByIndex[pIndex]
}

/**
 * Get the type of the entity domain parent
 */
func (o ModelMetadata) Type() reflect.Type {
	return o.this
}

func (o *ModelMetadata) AddTable(pTable Table, pKeyFieldNames ...string) {
	o.table = pTable
	o.table.Key = pKeyFieldNames
}

type GenerationType int

const (
	AUTO GenerationType = iota
	INCREMENT
	TABLE
	IDENTITY
	SEQUENCE
)

type KeyColumn struct {
	Auto bool
	Type GenerationType
}

func (o *ModelMetadata) AddKey(pField string, pIndex int, pColumn Column, pKind reflect.Kind) {
	//	k := Key{Auto: true, Type: INCREMENT}
	//	keyTag := ExtractOpalTags(o.this.FieldByName("Key").Tag)
	//	if keyTag.Get("Auto") != "" {
	//		k.Nilable = pColumn.Nilable
	//	}
	tag := ExtractOpalTags(o.this.Field(pIndex).Tag)
	// Set the default values if not specified
	c := Column{Insertable: true, Length: 255, Nilable: false, Updatable: true}
	if tag.Get("Nilable") != "" {
		c.Nilable = pColumn.Nilable
	}
	if tag.Get("Insertable") != "" {
		c.Insertable = pColumn.Insertable
	}
	if tag.Get("Updatable") != "" {
		c.Updatable = pColumn.Updatable
	}
	if tag.Get("Length") != "" {
		c.Length = pColumn.Length
	}
	c.Identifier = pField
	c.Name = pColumn.Name
	if tag.Get("Name") == "" {
		c.Name = c.Identifier
	}
	c.Unique = pColumn.Unique
	c.Precision = pColumn.Precision
	c.Scale = pColumn.Scale
	c.Kind = pKind

	o.columns = append(o.columns, c)
	o.keysByFieldName[pField] = &o.columns[len(o.columns)-1]
	o.keysByIndex[pIndex] = o.keysByFieldName[pField]
}

func (o *ModelMetadata) AddColumn(pField string, pIndex int, pColumn Column, pKind reflect.Kind) {
	tag := ExtractOpalTags(o.this.Field(pIndex).Tag)

	// Set the default values if not specified
	c := Column{Insertable: true, Length: 255, Nilable: true, Updatable: true}
	if tag.Get("Nilable") != "" {
		c.Nilable = pColumn.Nilable
	}
	if tag.Get("Insertable") != "" {
		c.Insertable = pColumn.Insertable
	}
	if tag.Get("Updatable") != "" {
		c.Updatable = pColumn.Updatable
	}
	if tag.Get("Length") != "" {
		c.Length = pColumn.Length
	}
	c.Identifier = pField
	c.Name = pColumn.Name
	if tag.Get("Name") == "" {
		c.Name = c.Identifier
	}
	c.Unique = pColumn.Unique
	c.Precision = pColumn.Precision
	c.Scale = pColumn.Scale
	c.Kind = pKind

	o.columns = append(o.columns, c)
	o.columnsByFieldName[pField] = &o.columns[len(o.columns)-1]
	o.columnsByIndex[pIndex] = o.columnsByFieldName[pField]
}

func (o *ModelMetadata) AddValidators(pField string, pIndex int, pValidators ...Validator) {

}

func (o *ModelMetadata) ReplaceTableIdentifiers(sql string, fDialect DialectEncoder) string {
	return strings.Replace(sql, o.this.Name(), fDialect(o.table.Name), -1)
}

func (o *ModelMetadata) ReplaceColumnIdentifiers(sql string, fDialect DialectEncoder) string {
	for _, column := range o.columns {
		sql = strings.Replace(sql, column.Name, fDialect(column.Name), -1)
	}
	return sql
}

func (o *ModelMetadata) ColumnsList(pBuilder *SqlBuilder, fDialect DialectEncoder) *SqlBuilder {
	for _, column := range o.columns {
		pBuilder.Add(column.Name).Add(", ")
	}
	return pBuilder.Truncate(2)
}

// Adds columns onto a sql builder in the form
// :Name, :Name,...  or ?, ?...
func (o *ModelMetadata) ColumnsBindList(pBuilder *SqlBuilder, fDialect DialectEncoder) *SqlBuilder {
	for i := 0; i < len(o.columns); i++ {
		pBuilder.Add("?, ")
	}
	return pBuilder.Truncate(2)
}

// sqlite seems to ignore incorrect set pk bindings when updating
// TODO need to ignore primary keys columns
func (o *ModelMetadata) NonKeyListEqualsNonKeyBindList(pBuilder *SqlBuilder, fDialect DialectEncoder) *SqlBuilder {
	for _, column := range o.columnsByFieldName {
		pBuilder.Add(column.Name).Add(" = ?, ")
	} // TODO the assumption is there is always a column
	return pBuilder.Truncate(2)
}

func (o *ModelMetadata) ColumnsListEqualsColumnsBindList(pBuilder *SqlBuilder, fDialect DialectEncoder) *SqlBuilder {
	for _, column := range o.columns {
		pBuilder.Add(column.Name).Add(" = ? AND ")
	}
	return pBuilder.Truncate(5)
}

func (o *ModelMetadata) KeyListEqualsKeyBindList(pBuilder *SqlBuilder, fDialect DialectEncoder) *SqlBuilder {
	for _, key := range o.keysByFieldName {
		pBuilder.Add(key.Name).Add(" = ? AND ")
	}
	return pBuilder.Truncate(5)
}

func (o *ModelMetadata) ColumnListWithConstraints(pBuilder *SqlBuilder, fDialect DialectEncoder) *SqlBuilder {
	if len(o.keysByFieldName) == 1 {
		for _, key := range o.keysByIndex {
			pBuilder.Add(key.Name)
			// TODO REMOVE fmt.Println(key.Kind.String())
			if key.Kind == reflect.Int64 {
				pBuilder.Add(" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT").Add(", ")
			} // TODO a better way
			// TODO handle other types
		}
	} else {
		for _, key := range o.keysByFieldName {
			key.BuildColumnSchema(pBuilder).Add(", ")
		}
	}
	for _, column := range o.columnsByFieldName {
		column.BuildColumnSchema(pBuilder).Add(", ")
	}
	return pBuilder.Truncate(2)
}

type JoinColumn struct {
	Identifier string
	Name       string
	Unique     bool
	Nilable    bool
	Insertable bool
	Updatable  bool
	Length     uint
	Precision  uint
	Scale      uint
	Kind       reflect.Kind
}

type Column struct {
	Identifier string
	Name       string
	Unique     bool
	Nilable    bool
	Insertable bool
	Updatable  bool
	Length     uint
	Precision  uint
	Scale      uint
	Kind       reflect.Kind
}

func (o Column) BuildColumnSchema(pBuilder *SqlBuilder) *SqlBuilder {
	pBuilder.Add(o.Name)
	pBuilder.Add(o.ToSqlType())
	o.unique(pBuilder)
	o.nilable(pBuilder)
	return pBuilder
}

func (o Column) unique(pBuilder *SqlBuilder) {
	if o.Unique {
		pBuilder.Add(" UNIQUE")
	}
}

func (o Column) nilable(pBuilder *SqlBuilder) {
	if !o.Nilable {
		pBuilder.Add(" NOT NULL")
	}
}

func (o Column) ToSqlType() string {
	switch o.Kind {
	case reflect.Ptr:
		return " crash" // TODO sort out types
	case reflect.Bool:
		return " BOOLEAN"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return " INTEGER"
	case reflect.Float64, reflect.Float32:
		return " FLOAT"
	case reflect.Slice:
		return " BLOB"
	case OpalTime:
		return " DATETIME"
	}

	//	switch val.Name() {
	//	case "Int64":
	//		return "INTEGER"
	//	case "Float64":
	//		return "REAL"
	//	case "Bool":
	//		return "INTEGER"
	//	case "RawBytes":  // TODO support properly
	//		return "BLOB"
	//	case "Time":
	//		return "DATETIME"
	//	}

	return fmt.Sprintf(" VARCHAR(%d)", o.Length)
}

type Table struct {
	Name string
	Key  []string
}
