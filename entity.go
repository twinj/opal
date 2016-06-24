package opal

import (
	"fmt"
)

type DomainModel struct {
	Domain
	Interface interface{}
}

// Domain is the skeleton interface for a Model. The metadata tied
// to a Model is based on its domain name. A ModelName is derived from
// the package path and the standard name used to access a type in a
// package.
type Domain interface {
	ModelName() ModelName
}

// The Entity interface is used to provide ActiveRecord
// and DAO access to a Model instance.
type Entity interface {

	// Should return an ActiveRecordDAO interface instance
	// of the main DAO
	ActiveRecord() ActiveRecordDAO

	// Returns the Entity's ModelName which represents its
	// domain or type
	ModelName() ModelName

	// Returns the model address so it can be used as
	// an ActiveRecord
	Model() Model

	// Each creation of a Model requires its own Entity
	// Creating a new Entity is based on the Domain's
	// default Entity for the Model which is created
	// through the NewEntity func passed at GEM startup.
	// The creation links the new Domain Entity with a
	// Model instance allowing database functions to be
	// performed on it.
	New(pModel Model) Entity

	// Insert passes the Model address to the the ModelDAO
	// service for persisting into the data-store
	Insert() Result

	// Save should update an already managed Model into
	// the data-store
	Save() Result

	// Delete should destroy and remove the Model from
	// the data-store
	Delete() Result

	// Metadata gets a copy of the Model's metadata
	Metadata() ModelMetadata

	// String returns a string representation of the Model
	String() string

	// TODO possible key interface

}

// OpalEntity implements the Entity interface
// If you require your own implementation you can supply Gem
// a function through which to create the Model embeddable
// instances of the Entity.
type OpalEntity struct {
	activeRecord ActiveRecordDAO
	modelName    *ModelName
	model        Model
	metadata     *ModelMetadata
}

// Pass this function into the Gem to create all base Entities
// for each Model
func NewEntity(pModelName ModelName) Entity {
	return &OpalEntity{currentGem.dao, &pModelName, nil, nil}
}

// TODO shrink use of instances here if possible heavy on performance
func (o OpalEntity) New(pModel Model) Entity {
	e := new(OpalEntity)
	e.activeRecord = o.activeRecord
	e.modelName = o.modelName
	e.model = pModel
	meta := currentGem.allModelsMetadata[*e.modelName]
	e.metadata = &meta
	return e
}

// TODO add entity to finds and query scanning in Gem and BaseDAO

func (o OpalEntity) ModelName() ModelName {
	return *o.modelName
}

func (o OpalEntity) ActiveRecord() ActiveRecordDAO {
	return o.activeRecord
}

func (o OpalEntity) Model() Model {
	return o.model
}

func (o *OpalEntity) Insert() Result {
	return o.activeRecord.Insert(o.model)
}

func (o *OpalEntity) Save() Result {
	return o.activeRecord.Save(o.model)
}

func (o *OpalEntity) Delete() Result {
	return o.activeRecord.Delete(o.model)
}

func (o *OpalEntity) Metadata() ModelMetadata {
	return *o.metadata
}

func (o *OpalEntity) String() string {
	return fmt.Sprint(BindArgs(o.model)...)
}
