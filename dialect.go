/**
 * Date: 22/02/14
 * Time: 11:17 AM
 */
package opal

import ()

// The Dialect interface performs sql syntax modification to conform
// to the differences in implementations of the SQL standard.
// Initially the Dialect will be made with the differences of Sqlite3
// in mind. Will be expanded in the future.
type Dialect interface {

	// All identifiers will be transformed into a string
	// compatible with the dialect to ensure that any keywords
	// that might be used as identifiers do not cause Sql errors.
	// E.g. In Sqlite if you want to use a keyword as a name,
	// you need to quote it:
	//
	// 'keyword'	A keyword in single quotes is a string
	// 		literal.
	//
	// "keyword"	A keyword in double-quotes is an identifier.
	//
	// [keyword]	A keyword enclosed in square brackets is an
	//		identifier. This is not standard SQL. This quoting
	// 		mechanism is used by MS Access and  SQL Server and
	// 		is included in SQLite for  compatibility.
	//
	// `keyword`	A keyword enclosed in grave accents (ASCII
	// 		code 96) is an identifier. This is not standard SQL.
	// 		This quoting mechanism is used  by MySQL and is
	//		included in SQLite for compatibility.
	//
	// Sqlite enforces quoted string literals as identifiers based on
	// context
	EncodeIdentifier(pIdentifier string) string

	TransformTypeDeclaration(pColumn Column) string
}

type DialectEncoder (func(string) string)

// Sqlite3 implements the Dialect interface
type DefaultDialect struct {
}
