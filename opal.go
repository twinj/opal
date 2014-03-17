/**
 * Date: 8/02/14
 * Time: 11:22 AM
 */
package opal

import (
	"database/sql"
	"fmt"
	"github.com/twinj/version"
	"log"
	"reflect"
)

const (
	// Group magic number
	group = 0xAA7E5441

	// Magic number specific to Opal
	opalMagic = 0x88280BA1

	// Version information
	release   = 1
	iteration = 0
	revision  = 0
	api       = 1
)

var (
	done bool

	// The current instance - instances can be swapped out
	// TODO create a package only handle set by the user
	currentGem *Gem
)

func init() {
	version.System(group, opalMagic, release, iteration, revision, api, "OPAL")
}

// Copy a Gem into the packages address to be used as the current service
func SwitchGem(pGem Gem) bool {
	*currentGem = pGem
	return true
}

// Retrieve a copy of Gem
func GetGem() Gem {
	return *currentGem
}

// ******************************************** Data Access

// ActiveRecordDAO acts as a data provider for a Model's Entity.
// It is a limited version of the full Domain access object.
type ActiveRecordDAO interface {

	// Takes a Model saves it as a new data-store entity and
	// returns a result
	Insert(pModel Model) Result

	// Takes a Model, updates an existing entity from the
	// data-store and returns a result
	Update(pModel Model) Result

	// Takes a Model and removes an existing entity from the
	// data-store
	Delete(pModel Model) Result
}

// ModelDAO acts a data provider for a Model's domain
// Methods are
type ModelDAO interface {

	OPAL
	ActiveRecordDAO

	// Find all models within the domain
	FindAllModels(pModelName ModelName) []Model

	// Find a specific Model using its keys
	FindModel(pModelName ModelName, pKeys ...interface{}) Model

	// Create a Sql Builder for the specified Model
	SqlBuilder(pModelName ModelName) *SqlBuilder
}

// ModelIDAO Implements ModelDAO
type ModelIDAO struct {
	gem *Gem
}

// Opal partially implements the opal OPAL interface
func (ModelIDAO) opal() OpalMagic {
	return opal
}

// Opal partially implements the opal OPAL interface
func (ModelIDAO) Kind() reflect.Kind {
	return DAO
}

// TODO betterway to handle DAO
func (o ModelIDAO) Gem() Gem {
	return *o.gem
}

func (o ModelIDAO) FindAllModels(pModelName ModelName) []Model {
	meta := o.gem.allModelsMetadata[pModelName]
	// TODO what if lose connection
	stmt := meta.preparedStatements[findAll]
	rows, err := stmt.Query()
	if err != nil {
		log.Print(err)
		return nil // TODO handle err
	}
	defer rows.Close()
	var models []Model
	for rows.Next() {
		model, args := meta.ScanInto()
		rows.Scan(args...)
		models = append(models, model)
	}
	return models
}

// TODO better key solution
func (o ModelIDAO) FindModel(pModelName ModelName, pKeys ...interface{}) Model {
	meta := o.gem.allModelsMetadata[pModelName]
	stmt := meta.preparedStatements[find]
	row := stmt.QueryRow(pKeys...)
	model, args := meta.ScanInto()
	err := row.Scan(args...)
	if err != nil {
		//TODO determine how errors should be handled
		fmt.Println(err)
		return nil
	}
	return model
}

func (o *ModelIDAO) SqlBuilder(pModelName ModelName) *SqlBuilder {
	meta := o.gem.allModelsMetadata[pModelName]
	builder := new(SqlBuilder)
	builder.ModelMetadata = &meta
	builder.Dialect = o.gem.Dialect
	return builder
}

// TODO find why insert prepared statement does not work
// TODO fix persis
// gets a primary key unique constraint error
func (o *ModelIDAO) Insert(pModel Model) Result {
	if o.gem.tx == nil {
		builder := o.SqlBuilder(pModel.ModelName()).Insert().Values()
		fPre, fPost := insertHooks(pModel)
		if fPre != nil {
			err := fPre()
			if err != nil {
				return Result{nil, err}
			}
		}
		result, err := o.gem.Exec(builder, insertArgs(pModel)...)
		if err != nil {
			return Result{result, err}
		}
		// TODO dialect for Id
		if id, err := result.LastInsertId(); err == nil {
			k := pModel.Keys()[0].(*Int64)
			k.Set(id)
		}
		if fPost != nil {
			err := fPost()
			if err != nil {
				return Result{nil, err}
			}
		}
		return Result{result, err}
	}
	return persist(o, pModel)
}

func (o *ModelIDAO) Update(pModel Model) Result {
	return merge(o, pModel)
}

func (o *ModelIDAO) Delete(pModel Model) Result {
	return remove(o, pModel)
}

func (o *ModelIDAO) ExecorStmt(pModelName ModelName, pNamedStmt string) *sql.Stmt {
	// TODO handle disconnections
	stmt := o.gem.allModelsMetadata[pModelName].preparedStatements[pNamedStmt]
	if o.gem.tx == nil {
		return stmt
	}
	return o.gem.tx.stmt(stmt)
}

// Future type for using when the opal sql has more of its own nuances
type OpalSql string

func (o OpalSql) String() string {
	return string(o)
}

type Rows struct {
	*sql.Rows // TODO explore embedded is wasteful?
}

type Sql interface {
	String() string
}

type StartArgs struct {
	BaseModel BaseModel
	DB *sql.DB
	Dialect   Dialect
	CreateEntity func(ModelName) Entity
}

func GEM(o StartArgs) *Gem {
	// TODO panic on nil options
	gem := new(Gem)
	gem.Dialect = o.Dialect
	gem.dao = &ModelIDAO{gem: gem}
	gem.DB = o.DB
	gem.funcCreateDomainEntity = o.CreateEntity
	if gem.funcCreateDomainEntity == nil {
		gem.funcCreateDomainEntity = NewEntity
	}
	models := o.BaseModel.Models()
	gem.allModelsMetadata = make(map[ModelName]ModelMetadata, len(models))
	gem.allModelsEntity = make(map[ModelName]*Entity, len(models))
	gem.txPreparedStatements = make(map[*sql.Stmt]*sql.Stmt)
	currentGem = gem
	for _, face := range models {
		model, ok := face.(Model)
		if !ok {
			panic("Opal.Start: You cannot pass a type which does not implement Model.")
		}
		// TODO option for fuller path name
		t := reflect.TypeOf(model).Elem()

		// Create the ModelMetadata and gather the
		// table and column information
		meta := NewMetadata(model, t)

		// Gather the metadata and save into the ModelMetadata holder
		modelName, entity, modelDAO := model.Gather(meta) // TODO somehow detach Gather from model and initialise another way

		// Add the ModelName to the map for retrieving metadata
		gem.modelNames = append(gem.modelNames, modelName)
		gem.allModelsMetadata[modelName] = *meta
		gem.allModelsEntity[modelName] = entity

		// Inject OpalDAOs into Model DAOs
		// TODO report
		modelDAO(gem.dao)

		// Save an entity instance into the provided address
		*gem.allModelsEntity[modelName] = gem.funcCreateDomainEntity(modelName)

		// Generate prepared statements
		builder := gem.dao.SqlBuilder(modelName)

		// Create tables if necessary
		table := builder.Create().Sql()
		log.Printf("Opal.Start: Create table statement: %s", table.String())
		gem.Exec(table)

		meta.addStmt(gem.DB, findAll, builder.Select().Sql())
		meta.addStmt(gem.DB, find, builder.Select().WherePk().Sql())
		meta.addStmt(gem.DB, insert, builder.Insert().Values().Sql())
		meta.addStmt(gem.DB, update, builder.Update().WherePk().Sql())
		meta.addStmt(gem.DB, delete, builder.Delete().WhereAll().Sql())
	}
	done = true
	return currentGem
}

// Base Model statement names
const (
	find    = "find"
	findAll = "findAll"
	insert  = "insert"
	update  = "update"
	delete  = "delete"
)

type SqlBuilderDialectEncoder func(*SqlBuilder, DialectEncoder) *SqlBuilder

type ModifyDB (func(ModelName, Model) (Result, error))

// *********************************************  SPECIAL TYPES

type PreparedQuery interface {
}

type ModelQueries interface {
	NamedQueries() []PreparedQuery
	DerivedQueries() []PreparedQuery
}

type Validation interface {
}


