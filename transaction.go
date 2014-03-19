/**
 * Date: 12/02/14
 * Time: 8:20 AM
 */
package opal

import (
	"database/sql"
)

// ******************************************** Transactions

// Txn is a sql.Tx wrapper which holds the action to run and its result
// It also carries a reference to the current Gem.
type Txn struct {

	//Embedded so Txn behaves like a standard Tx
	*sql.Tx

	// The base data access object is used as an access point to
	// model metadata and sql.Db connections
	gem *Gem

	// Action is a function which is called to perform the meat of
	// a Tx call
	funcAction Action

	// Result wraps a standard sql.Result and an error. This is wrapped
	// to provide user simplicity in writing their own exec stmt
	// and transactions
	result Result
}

// Transaction is a wrapper of a *Txn for building Transactions.
// It is used to simplify Transaction building from a user perspective.
// t Transaction is much easier to remember than t *Txn
type Transaction struct {
	*Txn
}

// User simplified function types to build transactions
type Action func(pTx Transaction) Result
type TxArgs func(pTx Transaction, pArgs interface{}) error
type Select func(pTx Transaction) ([]Model, error)

// Begin makes a new Transaction with the Action function
// E.g: To perform a transaction create a new instance with
// a new dynamic function func(t Transaction)Result {}
// You can access any package and local vars and the transaction
// instance to perform transaction a Result must be returned; it
// contains an errors that might have occurred.

// result, success := gemInstance.New((func(t Transaction)Result{
// 		person := domain.NewPerson()
//		person.SetKey(190)
//		person.Name.Set(name)
//		person.Save()
//		person = NewPerson()
//		person.SetKey(191)
//		person.Name.Set("Doong Gong Sum")
//		return t.Persist(person)
//	})).Go()
//
func (o *Gem) Begin(fAction Action) *Txn {
	tx := new(Txn)
	tx.funcAction = fAction
	tx.gem = o
	return tx
}

// Go runs the Transaction it has in from Gem.Begin Begin
// Running Go will defer control of Rollback and
// Commit functionality to the system.
// If required a user can manually run a transaction using the
// *sql.DB connection.
// Returns the result and whether the transaction was successful
func (o *Txn) Go() (result Result, success bool) {
	// Begin transaction
	o.Tx, o.result.Error = o.gem.DB.Begin()
	if o.result.Error != nil {
		return o.result, false
	}
	// Lock access to Gems protected variables
	o.gem.txMu.Lock()
	o.gem.tx = o

	// Do work
	o.result = o.funcAction(Transaction{o})
	for key, txStmt := range o.gem.txPreparedStatements {
		txStmt.Close()
		//delete(o.gem.txPreparedStatements, key)
		// TODO Delete here
		o.gem.txPreparedStatements[key] = nil
	}

	if o.result.Error != nil {
		goto rollback
	} else {
		goto commit
	}
rollback:
{
	o.result.Error = o.Rollback()
	result = o.result
	success = false
	goto done
}
commit:
{
	// Try commit - TODO [possible log reversal]
	o.result.Error = o.Commit()
	if o.result.Error != nil {
		goto rollback
	}
	result = o.result
	success = true
}
done:
	// Clean transactional and unlock
	o.gem.tx = nil
	o.gem.txMu.Unlock()
	return result, success
}

// Update will update the Model within a Transaction.
// Call this within a Txn func Action
func (o *Txn) Update(pModel Model) Result {
	return merge(o, pModel)
}

// Persist will insert the Model within a Transaction.
// Call this within a Txn func Action
func (o *Txn) Insert(pModel Model) Result {
	return persist(o, pModel)
}

// Persist will insert the Model within a Transaction.
// Call this within a Txn func Action
func (o *Txn) Create(pModel Model, pArgs Args) Result {
	if pArgs == nil {
		persist(o, pModel)
	}
	return persist(o, pModel, pArgs.Get()...)
}  // TODO consider moving the execor to replace the DAO so that a transaction takes over the role of the DAO when it is active

// Delete will delete the Model within a Transaction.
// Call this within a Txn func Action
func (o *Txn) Delete(pModel Model) Result {
	return remove(o, pModel)
}

func (o *Txn) Exec(pSql Sql, pArgs ...interface{}) Result {
	result, err := o.Tx.Exec(pSql.String(), pArgs...)
	return Result{result, err}
}

type StmtExec func(...interface{}) (*sql.Result, error)
type StmtQuery func(...interface{}) (*sql.Rows, error)
type StmtQueryRow func(...interface{}) (*sql.Row)

func (o *Txn) ExecorStmt(pModelName ModelName, pNamedStmt string) *sql.Stmt {
	stmt := o.gem.allModelsMetadata[pModelName].preparedStatements[pNamedStmt]
	return o.stmt(stmt)
}

func (o *Txn) stmt(pStmt *sql.Stmt) *sql.Stmt {
	if v, ok := o.gem.txPreparedStatements[pStmt]; ok && v != nil {   // TODO remove nil once delete works
		return v
	}
	txStmt := o.Stmt(pStmt)
	o.gem.txPreparedStatements[pStmt] = txStmt
	return txStmt
}



