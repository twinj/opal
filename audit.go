/**
 * Date: 12/02/14
 * Time: 10:06 PM
 */
package opal

import (
	"time"
)

// Compile time check of Domain implementations
var _ Domain = &AuditableEntity{}

// Compile time check of Domain implementations
var _ Entity = &AuditableEntity{}


// **********************************************  SPECIAL TYPES

// User is a special type of Model
type User interface {
	Model
}

// **********************************************  SPECIAL TYPES

// Some types need to be auditable
// Expects a Model User type and  its Key Type
type Auditable interface {

	// Returns the User who created the auditable Model
	Creator() User

	// Sets the User who created the audible Model
	SetCreator(pCreatedBy User)

	// Returns the last User who modified the auditable Model
	Modifier() User

	// Sets the last User who modified the auditable Model
	SetModifier(pModifiedBy User)

	// Returns the time of creation
	Creation() time.Time

	// Sets the time of creation
	SetCreation(pCreated time.Time)

	// Returns the last time the model was modified
	Modification() time.Time

	// Sets the last time the model was modified
	SetModification(pLastModified time.Time)
}


// Auditable Entity *******************************************

// Implements Auditable
type AuditableEntity struct {
	Entity
	creator      User
	modifier     User
	creation     time.Time
	modification time.Time
}

// Returns the User who created the auditable Model
func (o AuditableEntity) Creator() User {
	return o.creator
}

// Sets the User who created the audible Model
func (o *AuditableEntity) SetCreator(pCreatedBy User) {
	o.creator = pCreatedBy
}

// Returns the last User who modified the auditable Model
func (o AuditableEntity) Modifier() User {
	return o.modifier
}

// Sets the last User who modified the auditable Model
func (o *AuditableEntity) SetModifier(pModifiedBy User) {
	o.modifier = pModifiedBy
}

// Returns the time of creation
func (o AuditableEntity) Creation() time.Time {
	return o.creation
}

// Sets the time of creation
func (o *AuditableEntity) SetCreation(pCreated time.Time) {
	o.creation = pCreated
}

// Returns the last time the model was modified
func (o AuditableEntity) Modification() time.Time {
	return o.modification
}

// Sets the last time the model was modified
func (o *AuditableEntity) SetModification(pLastModified time.Time) {
	o.modification = pLastModified
}
