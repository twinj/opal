This was a tech spike. Do not use this code. This is being re coded from scratch at myesui/opal

opal 
====

A Go ORM which implements the ActiveRecord pattern

In development API in no way guaranteed.

Currently only supports Sqlite3

#Use

Each domain must first create and initialise its models.

To achieve this you must implement the BaseModel interface and pass it into Gemma's
StartArgs.

	package domain

	var (
		domains = [...]Domain{
		new(Person),
		new(Pet),
		new(Order),
		new(Country),
	}
	)

	type YourBaseModel struct {}

	func (YourBaseModel) Models() []Domain {
		return domains[:]
	}

Before you can start using your domain structs you must first generate
your domain access code. To do this create a separate go project with
the same relative path to your domain objects and simply run the INIT
function. Looking to make this step redundant. It is best to keep this separate
so the files it generates can be used in your project.

	package main

	import (
		. "github.com/twinj/opal"
		"myesui/domain"
	)

	func main() {
	INIT(new(domain.YourBaseModel))
	}


You are now ready to start using Opal.

	import (
		"database/sql"
		"your/domain"
		_ "github.com/mattn/sqlite3"
		. "github.com/twinj/opal"
	)

	var (
		Em *Gem
	)

	func wherever() {
	 db, err := sql.Open("sqlite3", "file:./person_test.db?cache=shared&mode=rwc")

		args := StartArgs{
			BaseModel: new(domain.YourBaseModel),
			DB: db,
			Dialect: &dialect.Sqlite3{},
		}
		Em = GEM(args)
	}

#Features thus far

Models:

	type Person struct {
		Entity
		Id       AutoIncrement
		Name     String
	}

	type Country struct {
		Entity
		Code String
		Description String
	}

Model CRUD:

	person := domain.InitPerson()
	person.Id.A(190)
	person.Name.A("Tom Baker")
	domain.People.Insert(person)

	person.Name.A("Tom Fred Baker")
	domain.People.Update(person)

	domain.People.Delete(person)

	person = domain.People.Find(170)

Model ActiveRecord:

	person := domain.NewPerson{
			Id: 400,
			Name: "Frank Cheese",
		}.Save()

    person.Name.A("Frank Cheesy").Save()

    person.Delete()

Model transactions:

	result, ok := GEM.Begin((func(t Transaction) Result {
			person := domain.InitPerson()
			person.Id.A(190)
			person.Name.A("Tom Baker")
			t.Insert(person3)
			person = domain.InitPerson()
			person.Id.A(191)
			person.Name.A("Peter Parker")
			return person.Insert()
		})).Go()

#Other Features

* Convention over configuration
* Uses Go Inflect library to name tables

#Planned Features

* Dialects to support different Sql databases
* Validation interfaces
* Relational mapping with maps and interfaces
* Compound and embedded keys and embedded struct
