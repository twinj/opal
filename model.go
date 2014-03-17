/**
 * Date: 9/02/14
 * Time: 6:40 PM
 */
package opal

import (
	"strings"
)

// Compile time check of Domain implementation
var _ Domain = &OpalEntity{}

// Compile time check of Entity implementations
var _ Entity = &OpalEntity{}


// ***********************************************  Model...

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

// A ModelName is an identifying string specific to a Model.
// It must be unique e.g: domain.Person
type ModelName string

// Name splits a model name from its package and returns the
// Model only name as a string
// This should return the implemented struct name
func (m ModelName) Name() string {
	return strings.Split(string(m), ".")[1]
}

// Casts a ModelName into a string
func (m ModelName) String() string {
	return string(m)
}

// A Model is the interface type which can represent any object
// that can be stored as a Database store item
type Model interface {

	// A Model must implement the Entity interface.
	// This is achieved by Embedding the Entity implementation
	// into the Model implementation.
	Entity

	// Gather creates the column and table data metadata
	// required to perform all database tasks associated

	// with the Model
	// It retrieves an address in which to store the models
	// Entity implementation. This implementation is required
	// to perform basic database functions such as saving and
	// deleting or other functions associated with a single instance.

	// It also retrieves a function to create the global model
	// specific DAO. This object can be used to perform tasks
	// associated with a Model rather than a single instance.
	Gather(pMetadata *ModelMetadata) (ModelName, *Entity, func(*ModelIDAO)ModelDAO)

	// ScanInto should return a new Model which can interact with
	// the base DAO system and addresses to its columns so data
	// can be scanned into it.
	ScanInto() (Model, []interface{})

	// Returns all columns' primary keys which can be scanned
	// into.
	Keys() []interface{}

	// Returns all columns' non primary keys which can be scanned
	// into.
	Parameters() []interface{}
}

// ******************************************** Model hooks

// ModelHook can be used by a Model implementation to do
// a specific task on one of the pre defined events.
// A Model must implement one of the Hook interfaces below for the
// ModelHook function to run during Model related execution.
type ModelHook func() error

// Insert hooks ******************************************

// Pre insert hook interface
type PreInsert interface {
	PreInsertHook() error
}

// Post insert hook interface
type PostInsert interface {
	PostInsertHook() error
}

// Gets the hook function from the model if they exist
func insertHooks(pModel Model) (preHook ModelHook, postHook ModelHook) {
	if hook, ok := pModel.(PreInsert); ok {
		preHook = hook.PreInsertHook
	}
	if hook, ok := pModel.(PostInsert); ok {
		postHook = hook.PostInsertHook
	}
	return
}

// Update hooks ******************************************

// Pre update hook interface
type PreUpdate interface {
	PreUpdateHook() error
}

// Post update hook interface
type PostUpdate interface {
	PostUpdateHook() error
}

// Gets the hook function from the model if they exist
func updateHooks(pModel Model) (preHook ModelHook, postHook ModelHook) {
	if hook, ok := pModel.(PreUpdate); ok {
		preHook = hook.PreUpdateHook
	}
	if hook, ok := pModel.(PostUpdate); ok {
		postHook = hook.PostUpdateHook
	}
	return
}

// Delete hooks ******************************************

// Pre delete hook interface
type PreDelete interface {
	PreDeleteHook() error
}

// Post delete hook interface
type PostDelete interface {
	PostDeleteHook() error
}

// Gets the hook function from the model if they exist
func deleteHooks(pModel Model) (preHook ModelHook, postHook ModelHook) {
	if hook, ok := pModel.(PreDelete); ok {
		preHook = hook.PreDeleteHook
	}
	if hook, ok := pModel.(PostDelete); ok {
		postHook = hook.PostDeleteHook
	}
	return
}

// ******************************************** Arg retrievers

// Args interface allows you to retrieve the args from
// any type which implements this
type Args interface {
	// Retrieve args from the type
	Get() []interface{}
}

// Any function which takes a Model and returns an arg slice
type ModelArgs func(Model) []interface{}

// Gets all the bind args for a model.
// Does not include any join columns.
func BindArgs(pModel Model) []interface{} {
	return append(pModel.Keys(), pModel.Parameters()...)
}

// Gets the bind args required for a new Model
func insertArgs(pModel Model) []interface{} {
	return BindArgs(pModel)
}

// Gets the bind args required to update a Model
func updateArgs(pModel Model) []interface{} {
	return append(pModel.Parameters(), pModel.Keys()...)
}

// Gets the bind args required to delete Model
func deleteArgs(pModel Model) []interface{} {
	return BindArgs(pModel)
}

// TODO relationships and associations
type HasMany Association
type BelongsTo Association
type HasOne Association

type Association interface{}
type Transient bool
