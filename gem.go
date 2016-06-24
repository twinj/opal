package opal

import (
	"database/sql"
	"fmt"
	"log"
	"sync"
)

// Go entity manager
// This struct acts as a mediator between Models and a sql.DB
type Gem struct {
	*sql.DB
	Dialect

	// txMu locks access to daoMu during a transaction
	txMu sync.Mutex

	// Will usually be nil
	tx                   *Txn
	txPreparedStatements map[*sql.Stmt]*sql.Stmt

	dao               *ModelIDAO // TODO change to embedded dao?
	modelNames        []ModelName
	allModelsMetadata map[ModelName]ModelMetadata
	allModelsEntity   map[ModelName]*Entity

	funcCreateDomainEntity func(pModelName ModelName) Entity

	// TODO make it work without INIT add reflection based
	// so Entities can be made on the fly

}

// Gets a copy of the metadata associated with the Model
// which is identified by its ModelName
func (o Gem) Metadata(pModelName ModelName) ModelMetadata {
	return o.allModelsMetadata[pModelName]
}

// Runs a standard Db query which expects a slice of Models as a result,
// Will take any Sql interface and the ModelName to identify Model
func (o Gem) Query(pModelName ModelName, pSql Sql) ([]Model, error) {
	// Do query and convert results to Models
	// TODO assert right model
	rows, err := o.DB.Query(pSql.String())
	if err != nil {
		log.Print(err)
		return nil, err
	}
	defer rows.Close()
	var models []Model
	for rows.Next() {
		model, args := o.Metadata(pModelName).ScanInto()
		rows.Scan(args...)
		models = append(models, model)
	}
	return models, nil
}

// Runs a standard Db query which expects a Model as a result,
// Will take any Sql interface and the ModelName to identify Model
// TODO investigate do not support keyword as identifiers it's easier
func (o Gem) QueryRow(pModelName ModelName, pSql Sql, pArgs ...interface{}) Model {
	// Do query and convert results to Models
	row := o.DB.QueryRow(pSql.String(), pArgs...)
	model, args := o.Metadata(pModelName).ScanInto()
	err := row.Scan(args...)
	if err != nil {
		//TODO determine how errors should be handled
		return nil
	}
	return model
}

// TODO determine requirement error wrapping?
func (o Gem) Exec(pSql Sql, pArgs ...interface{}) (sql.Result, error) {
	// Do execution expect a result
	result, err := o.DB.Exec(pSql.String(), pArgs...)
	if err != nil {
		log.Print(err)
		return nil, err
	}
	return result, nil
}

// ******************************************** NOT DEPENDENT ON GEM

// calls the model exec method with delete args and hooks
func remove(pExecor Execor, pModel Model) Result {
	return exec(pExecor, pModel, delete, deleteArgs, deleteHooks)
}

// calls the model exec method with persist args and hooks
func persist(pExecor Execor, pModel Model, pArgs ...interface{}) Result {
	fArgs := insertArgs
	if len(pArgs) > 0 {
		fArgs = func(Model) []interface{} {
			return pArgs
		}
	}
	result := exec(pExecor, pModel, insert, fArgs, insertHooks)
	if result.Error != nil {
		return result
	}
	// TODO dialect for Id
	var id int64
	if id, result.Error = result.LastInsertId(); result.Error == nil {
		v, ok := pModel.Keys()[0].(*AutoIncrement)
		if ok {
			v.A(id)
		}
		v2, ok := pModel.Keys()[0].(*Int64)
		if ok {
			v2.A(id)
		}
	}
	return result
} // TODO metadata API and interface - check security

// calls the model exec method with update args and hooks
func merge(pExecor Execor, pModel Model) Result {
	return exec(pExecor, pModel, update, updateArgs, updateHooks)
}

// exec handles the execution of basic Models with no joins
func exec(pExecor Execor, pModel Model, pNamedStmt string, fArgs ModelArgs, fModelHooks func(Model) (ModelHook, ModelHook)) Result {
	fPre, fPost := fModelHooks(pModel)
	if fPre != nil {
		err := fPre()
		if err != nil {
			return Result{nil, err}
		}
	}
	// TODO delete bug
	if pNamedStmt == delete {
		fmt.Println(pModel.ModelName(), pNamedStmt, "Delete here")
		fmt.Printf("%#v", pExecor.ExecorStmt(pModel.ModelName(), pNamedStmt))
	}
	result, err := pExecor.ExecorStmt(pModel.ModelName(), pNamedStmt).Exec(fArgs(pModel)...)
	if err != nil {
		return Result{result, err}
	}
	if fPost != nil {
		err := fPost()
		if err != nil {
			return Result{result, err}
		}
	}
	return Result{result, nil}
}

type Execor interface {
	// Retrieve the statement required for the database work
	// TODO handle discons
	ExecorStmt(pModel ModelName, pNamedStmt string) *sql.Stmt
}

// Result is a wrapper around a sql.Result and any
// affiliated error. It is used to simplify the API
// TODO change package for Result as its associated to DAO
type Result struct {
	sql.Result
	Error error
}

func (o Result) String() string {
	var si, sj, sErr string
	i, errI := o.LastInsertId()
	j, errJ := o.RowsAffected()
	if o.Error == nil {
		sErr = "nil"
	} else {
		sErr = fmt.Sprint(o.Error)
	}
	if errI == nil {
		si = fmt.Sprintf("%d", i)
	} else {
		si = fmt.Sprint(errI)
	}
	if errJ == nil {
		sj = fmt.Sprintf("%d", j)
	} else {
		sj = fmt.Sprint(errJ)
	}
	return fmt.Sprintf("Last insert: %s; Rows affected: %s; Error: %s", si, sj, sErr)
}

// The Base model acts as the container to retrieve all Models
// and their corresponding metadata in the Domain.
//
// Each domain must define a BaseModel
//
// BaseModel / Entity overlap in function
// TODO re-consider use and requirement
type BaseModel interface {
	Models() []Domain
}
