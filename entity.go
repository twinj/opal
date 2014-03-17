package opal

import "fmt"

// Entity *****************************************************

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

	// Returns the models ModelName which represents its
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

	// The string method should print the entity in which ever format
	// you prefer  // TODO ? better placement
	String() string

}

// OpalEntity implements the Entity interface
// If you require your own implementation you can supply Gem
// a function through which to create the Model embeddable
// instances of the Entity.
type OpalEntity struct {
	activeRecord ActiveRecordDAO
	modelName *ModelName
	model Model
}

// Pass this function into the GEm to create all base Entities
// for each Model
func NewEntity(pModelName ModelName) Entity {
	return &OpalEntity{currentGem.dao, &pModelName, nil}
}

func (o OpalEntity) New(pModel Model) Entity {
	e := new(OpalEntity)
	e.activeRecord = o.activeRecord
	e.modelName = o.modelName
	e.model = pModel
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

// TODO temporary insert
func (o *OpalEntity) Insert() Result {
	return o.activeRecord.Persist(o.model)
}

func (o *OpalEntity) Save() Result {
	return o.activeRecord.Merge(o.model)
}

func (o *OpalEntity) Delete() Result {
	// TODO should there be a memory nil here?
	return o.activeRecord.Remove(o.model)
}

func (o *OpalEntity) String() string {
	return fmt.Sprint(BindArgs(o.model)...)
}
