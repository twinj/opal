/**
 * Date: 12/03/14
 * Time: 06:34 PM
 */
package opal

import (
	"bytes"
)

type SqlBuilder struct {
	*ModelMetadata
	Dialect
	bytes.Buffer
}

func (o *SqlBuilder) Add(p string) *SqlBuilder {
	o.WriteString(p)
	return o
}

func (o *SqlBuilder) Do(...interface{}) *SqlBuilder {
	return o
}

func (o *SqlBuilder) Create() *SqlBuilder {
	o.Add("CREATE TABLE IF NOT EXISTS ").Add(o.table.Name)
	return o.Add("(").With(o.ColumnListWithConstraints, o.EncodeIdentifier).Add(")")
}

func (o *SqlBuilder) Select(pColumns ...string) *SqlBuilder {
	return o.Add("SELECT * FROM ").Add(o.table.Name)
}

func (o *SqlBuilder) Insert() *SqlBuilder {
	o.Add("INSERT INTO ").Add(o.table.Name)
	return o.Add("(").With(o.ColumnsList, o.EncodeIdentifier).Add(")")
}

func (o *SqlBuilder) With(fBuild SqlBuilderDialectEncoder, fDialect DialectEncoder) *SqlBuilder {
	return fBuild(o, fDialect)
}

func (o *SqlBuilder) Values() *SqlBuilder {
	return o.Add(" VALUES (").With(o.ColumnsBindList, o.EncodeIdentifier).Add(")")
}

func (o *SqlBuilder) Update() *SqlBuilder {
	o.Add("UPDATE ").Add(o.table.Name)
	return o.Add(" SET ").With(o.NonKeyListEqualsNonKeyBindList, o.EncodeIdentifier)
}

func (o *SqlBuilder) Delete() *SqlBuilder {
	return o.Add("DELETE FROM ").Add(o.table.Name)
}

func (o *SqlBuilder) WherePk() *SqlBuilder {
	return o.Add(" WHERE ").With(o.KeyListEqualsKeyBindList, o.EncodeIdentifier)
}

func (o *SqlBuilder) WhereAll() *SqlBuilder {
	return o.Add(" WHERE ").With(o.ColumnsListEqualsColumnsBindList, o.EncodeIdentifier)
}

func (o *SqlBuilder) Truncate(pInt int) *SqlBuilder {
	o.Buffer.Truncate(o.Len() - pInt)
	return o
}

func (o *SqlBuilder) Sql() Sql {
	sql := OpalSql(o.Buffer.String())
	o.Reset()
	return &sql
}
