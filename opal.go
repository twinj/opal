/**
 * Date: 8/02/14
 * Time: 11:22 AM
 */
package opal

import (
	"database/sql"
	"fmt"
	_ "github.com/twinj/version"
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
	// The current instance - instances can be swapped out
	// TODO create a package only handle set by the user
	currentGem *Gem
)

func init() {
	//version.Init(group, opalMagic, release, iteration, revision, api, "OPAL")
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
	Save(pModel Model) Result

	// Takes a Model and removes an existing entity from the
	// data-store
	Delete(pModel Model) Result
}

// ActiveRecordDAO acts as a data provider for a Model's Entity.
// It is a limited version of the full Domain access object.
type ActiveRecord interface {

	// Takes a Model saves it as a new data-store entity and
	// returns a result
	Insert(pModel Model) Result

	// Takes a Model, updates an existing entity from the
	// data-store and returns a result
	Save(pModel Model) Result

	// Takes a Model and removes an existing entity from the
	// data-store
	Delete(pModel Model) Result
}

// ModelDAO acts a data provider for a Model's domain
// Methods are
type ModelDAO interface {
	Opal
	ActiveRecordDAO

	// Find all models within the domain
	FindAllModels() []Model

	// Find a specific Model using its keys
	FindModel(pKeys ...interface{}) Model

	// Create a Sql Builder for the specified Model
	SqlBuilder() *SqlBuilder

	// Returns the ModelName associated with this instance
	Model() ModelName
}

// ModelIDAO Implements ModelDAO
type ModelIDAO struct {
	gem   *Gem
	model ModelName
}

func (o ModelIDAO) Model() ModelName {
	return o.model
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

func (o ModelIDAO) FindAllModels() []Model {
	meta := o.gem.allModelsMetadata[o.Model()]
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
func (o ModelIDAO) FindModel(pKeys ...interface{}) Model {
	meta := o.gem.allModelsMetadata[o.Model()]
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

func (o *ModelIDAO) SqlBuilder() *SqlBuilder {
	meta := o.gem.allModelsMetadata[o.Model()]
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
		// TODO remove?
		builder := o.SqlBuilder().Insert().Values()
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
			// TODO compound key
			v, ok := pModel.Keys()[0].(*AutoIncrement)
			if ok {
				v.A(id)
			}
			v2, ok := pModel.Keys()[0].(*Int64)
			if ok {
				v2.A(id)
			}
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

func (o *ModelIDAO) Save(pModel Model) Result {
	return merge(o, pModel)
}

func (o *ModelIDAO) Delete(pModel Model) Result {
	return remove(o, pModel)
}

func (o *ModelIDAO) ExecorStmt(pModel ModelName, pNamedStmt string) *sql.Stmt {
	// TODO handle disconnections
	stmt := o.gem.allModelsMetadata[pModel].preparedStatements[pNamedStmt]
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
	BaseModel    BaseModel
	DB           *sql.DB
	Dialect      Dialect
	CreateEntity func(ModelName) Entity
	Id           *OpalMagic
}

func GEM(o StartArgs) *Gem {
	// TODO panic on nil options
	gem := new(Gem)
	gem.Dialect = o.Dialect
	gem.dao = &ModelIDAO{gem: gem}
	gem.DB = o.DB
	gem.funcCreateDomainEntity = o.CreateEntity

	SetMagic(o.Id)
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
			panic("Opal.Start: You cannot pass a type which does not implement the Model interface.")
		}
		// TODO option for fuller path name
		t := reflect.TypeOf(model).Elem()

		// Create the ModelMetadata and gather the
		// table and column information
		meta := NewMetadata(model, t)

		// Gather the metadata and save into the ModelMetadata holder
		name, entity, modelDAOf := model.Gather(meta) // TODO somehow detach Gather from model and initialise another way

		// Inject OpalDAOs into Model DAOs
		// TODO report
		modelDAO := modelDAOf(&ModelIDAO{gem, name})

		// Add the ModelName to the map for retrieving metadata
		gem.modelNames = append(gem.modelNames, modelDAO.Model())
		gem.allModelsMetadata[modelDAO.Model()] = *meta
		gem.allModelsEntity[modelDAO.Model()] = entity

		// Save an entity instance into the provided address
		*gem.allModelsEntity[modelDAO.Model()] = gem.funcCreateDomainEntity(modelDAO.Model())

		// Generate prepared statements
		builder := modelDAO.SqlBuilder()

		// Create tables if necessary
		table := builder.Create().Sql()
		log.Printf("Opal.Start: Create table statement: %s", table.String())
		gem.Exec(table)

		// Add these first run
		meta.addStmt(gem.DB, findAll, builder.Select().Sql())
		meta.addStmt(gem.DB, find, builder.Select().WherePk().Sql())
		meta.addStmt(gem.DB, insert, builder.Insert().Values().Sql())
		meta.addStmt(gem.DB, update, builder.Update().WherePk().Sql())
		meta.addStmt(gem.DB, delete, builder.Delete().WherePk().Sql())
	}
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
