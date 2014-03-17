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

	dao               *ModelIDAO
	modelNames        []ModelName
	allModelsMetadata map[ModelName]ModelMetadata
	allModelsEntity map[ModelName]*Entity

	funcCreateDomainEntity func(pModelName ModelName) Entity
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
		fArgs = func(Model)[]interface{}{
			return pArgs
		}
	}
	return exec(pExecor, pModel, insert, fArgs, insertHooks)
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
	result, err := pExecor.ExecorStmt(pModel.ModelName(), pNamedStmt).Exec(fArgs(pModel)...)
	if err != nil {
		return Result{result, err}
	}
	// TODO dialect for Id
	if id, err := result.LastInsertId(); err == nil {
		k := pModel.Keys()[0].(*Int64)
		k.Set(id)
	}
	// TODO delete bug
	if pNamedStmt == delete {
		fmt.Println(fArgs(pModel)...)
	}
	if fPost != nil {
		err := fPost()
		if err != nil {
			return Result{nil, err}
		}
	}
	return Result{result, err}
}


type Execor interface {
	// Retrieve the statement required for the database work
	// TODO handle discons
	ExecorStmt(pModelName ModelName, pNamedStmt string) *sql.Stmt
}


