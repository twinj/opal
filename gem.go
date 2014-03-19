package opal

import (
	"database/sql"
	"fmt"
	"log"
	"sync"
)

type Gem struct {
	*sql.DB
	Dialect

	// txMu locks access to daoMu during a transaction
	txMu sync.Mutex

	// Will usually be nil
	tx *Txn
	txPreparedStatements map[*sql.Stmt]*sql.Stmt

	dao               *ModelIDAO // TODO change to embedded dao?
	modelNames        []ModelName
	allModelsMetadata map[ModelName]ModelMetadata
	allModelsEntity map[ModelName]*Entity

	funcCreateDomainEntity func(pModelName ModelName) Entity

	// TODO make it work without INIT add reflection based like Gorp
	// so Entities can be made on the fly
}

func (o Gem) Metadata(pModelName ModelName) ModelMetadata {
	return o.allModelsMetadata[pModelName]
}

func (o Gem) Query(pModelName ModelName, pSql Sql) ([]Model, error) {
	// Do query and convert results to Models
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

func remove(pExecor Execor, pModel Model) Result {
	return exec(pExecor, pModel, delete, deleteArgs, deleteHooks)
}

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
		k := pModel.Keys()[0].(*Int64)
		k.Set(id)
	}
	return result
}

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
		fmt.Printf("%#v",  pExecor.ExecorStmt(pModel.ModelName(), pNamedStmt))
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
	ExecorStmt(pModelName ModelName, pNamedStmt string) *sql.Stmt
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
	if (errI == nil) {
		si = fmt.Sprintf("%d", i)
	} else {
		si = fmt.Sprint(errI)
	}
	if (errJ == nil) {
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


