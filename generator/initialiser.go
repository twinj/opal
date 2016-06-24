package generator

import (
	"bitbucket.org/pkg/inflect"
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"reflect"
	"strconv"
	"strings"
	"text/template"
	"time"
	. "github.com/twinj/opal-old"
)

// Some rules
// 1: Primary keys if not one of the four base types will
// be an embedded type where only one param will be generated.
// 2: Compound keys which are expanded will be of the four base
//  base types and their primitives will be expanded into the
// parameters.
type ModelTemplate struct {
	Version string
	Date    string
	Time    string
	Types   map[string]*TemplateType
}

type TemplateType struct {
	Domain
	Version    string
	Date       string
	Time       string
	Receiver   string
	Model      string
	DAOName    string
	Path       string
	Package    string
	ImportName string
	Table      string
	Keys       []KeyField
	Columns    []TemplateField
}

type KeyField struct {
	Name      string
	TypeName  string
	Index     int
	Tag       string
	Kind      string
	Primitive string
}

type TemplateField struct {
	Name      string
	Index     int
	Tag       string
	Kind      string
	Primitive string
}

// INIT will scan each supplied Model/Domain object
// and generate the relevant Boilerplate code with which you can run
// TODO doco
func INIT(pBaseModel BaseModel) {
	gatherTemplateData(pBaseModel.Models())
}

// The system will run the code generation template
// based on the received skeletal Models and generate the code
func gatherTemplateData(pModels []Domain) {
	plate := ModelTemplate{}

	// Header information
	now := time.Now()
	//plate.Version = version.String()
	plate.Date = now.Format("2 January 2006")
	plate.Time = now.Format("Monday 3:04 AM")

	// Gather Table information:
	plate.Types = make(map[string]*TemplateType, len(pModels))

	// Check the first field
	// if it's an anonymous Entity extract its metadata
	// if not panic - invalid Domain Model
	// TODO does this have to be first
	for _, domain := range pModels {
		model := reflect.TypeOf(domain).Elem()

		temp := TemplateType{}
		//var useDefaultKey bool
		temp.Domain = domain
		temp.Version = plate.Version
		temp.Date = plate.Date
		temp.Time = plate.Time
		temp.Model = model.Name()
		temp.DAOName = inflect.Camelize(inflect.Tableize(temp.Model))

		temp.Receiver = strings.ToLower(temp.Model)
		temp.Path = model.PkgPath()
		temp.ImportName = importName(model)

		s := strings.Split(temp.Path, "/")
		temp.Package = s[len(s)-1]

		var keys map[string]bool = make(map[string]bool)
		var i int
		if model.Field(i).Type.Implements(reflect.TypeOf((*Entity)(nil)).Elem()) && model.Field(i).Anonymous {
			opalTags := ExtractOpalTags(model.Field(i).Tag)

			// Derive table name or get specific
			if opalTags.Get("Name") == "" {
				temp.Table = fmt.Sprintf("Name: %q", inflect.Tableize(temp.Model))
			} else {
				temp.Table = string(opalTags)
			}
			if opalTags.Get("Key") == "" {
				//useDefaultKey = true
				keys["Id"] = true
			} else {
				// TODO support compound keys
				key, _ := strconv.Unquote(opalTags.Get("Key"))
				keys[key] = true
			}
			i++
		} else {
			// TODO error
			log.Fatalf("Opal.RunTemplate: Model does not anonymously embed type which implements the BaseModel interface at field position 0.")
		}

		// Check for the Key field
		// if it's an anonymous BaseModel extract its metadata
		// if not panic - invalid Entity
		opal := reflect.TypeOf((*Opal)(nil)).Elem()

		field := model.Field(i)
		typ := field.Type
		if typ.Name() == "Key" && field.Anonymous {
			opalTags := ExtractOpalTags(field.Tag)

			// TODO support compound keys // TODO check field type
			if _, ok := model.FieldByName(string(opalTags)); ok {
				keys[string(opalTags)] = true
			} else {
				panic("Opal.runTemplate: Key metatag error")
			}
			i++
		}

		// Check each field if its an OPAL extract its metadata after we have the entity data
		for i < model.NumField() {
			field = model.Field(i)
			typ = field.Type
			if field.Type.Implements(opal) {
				opalTags := ExtractOpalTags(field.Tag)
				if opalTags == "" {
					opalTags = Tag(fmt.Sprintf("Name: %q", field.Name))
				}
				// Infer column name or get explicit
				if opalTags.Get("Name") == "" {
					opalTags = Tag(fmt.Sprintf("Name: %q, %s", field.Name, opalTags))
				}
				if keys[field.Name] {
					if typ.Name() == "AutoIncrement" {
						opalTags += ", AutoIncrement: true"
					}
					kind := reflect.Kind(reflect.New(typ).MethodByName("Kind").Call(nil)[0].Uint())

					key := KeyField{field.Name, typ.Name(), i, string(opalTags), getKind(typ.Name()), kind.String()}

					temp.Keys = append(temp.Keys, key)
				} else {
					kind := reflect.Kind(reflect.New(typ).MethodByName("Kind").Call(nil)[0].Uint())
					s := kind.String()
					if s == "slice" {
						s = "[]byte"
					}
					// TODO handle primitive types properly

					temp.Columns = append(temp.Columns, TemplateField{field.Name, i, string(opalTags), getKind(typ.Name()), s})
				}
			}
			i++
		}
		plate.Types[model.Name()] = &temp
	}

	// Gather relationship information
	for _, domain := range pModels {
		var i int
		t := reflect.TypeOf(domain).Elem()

		//		assoc := reflect.TypeOf((*Association)(nil)).Elem()
		dom := reflect.TypeOf((*Domain)(nil)).Elem()
		// Treat as a HasOne relationship
		for i < t.NumField() {
			if t.Field(i).Type.Implements(dom) {
				// TODO gather key information for each model first
				//
			}
			i++
		}
	}
	runTemplate(plate)
}

func runTemplate(pModels ModelTemplate) {
	// Create a template, add the function map, and parse the text.
	b, err := ioutil.ReadFile("src/github.com/twinj/opal/entity.template")
	if err != nil {
		log.Fatalf("Opal.runTemplate: reading file: %s", err)
	}
	tmpl, err := template.New("Model setup").Parse(string(b))
	if err != nil {
		log.Fatalf("Opal.runTemplate: parsing: %s", err)
	}

	// Run the template to verify the output.
	for _, li := range pModels.Types {
		buf := bytes.Buffer{}
		err = tmpl.Execute(&buf, li)
		if err != nil {
			log.Fatalf("Opal.runTemplate: execution: %s", err)
		}
		s := fmt.Sprintf("src/%s/%s_%s_init.go", li.Path, li.Package, li.Receiver)
		err = ioutil.WriteFile(s, buf.Bytes(), os.ModeTemporary)
		if err != nil {
			log.Fatalf("Opal.runTemplate: writing file %s: ", s, err)
		} // TODO parse multiple templates from one domain package into 1 file
	}
}

// A StructTag is the tag string in a struct field.
//
// By convention, tag strings are a concatenation of
// optionally space-separated key:"value" pairs.
// Each key is a non-empty string consisting of non-control
// characters other than space (U+0020 ' '), quote (U+0022 '"'),
// and colon (U+003A ':').  Each value is quoted using U+0022 '"'
// characters and Go string literal syntax.
type Tag string

func ExtractOpalTags(pStructTag reflect.StructTag) Tag {
	tag := string(pStructTag)
	if strings.Count(tag, "|") == 2 {
		tag = strings.Split(tag, "|")[1]
		return Tag(tag)
	}
	return Tag("")
}

// Get returns the value associated with key in the tag string.
// If there is no such key in the tag, Get returns the empty string.
// If the tag does not have the conventional format, the value
// returned by Get is unspecified.
func (tag Tag) Get(key string) string {

	for tag != "" {
		var quoted = true
		// skip leading space
		i := 0
		for i < len(tag) && tag[i] == ' ' {
			i++
		}
		tag = tag[i:]
		if tag == "" {
			break
		}

		// scan to colon.
		// a space or a quote is a syntax error
		i = 0
		for i < len(tag) && tag[i] != ' ' && tag[i] != ':' && tag[i] != '"' {
			i++
		}
		// check if the next rune is either a space or quote
		if i+1 >= len(tag) || tag[i] != ':' || !(tag[i+1] == '"' || tag[i+1] == ' ') {
			break
		}
		name := string(tag[:i])
		tag = tag[i+1:]
		i = 0
		// if a space is included skip it - only supports one
		if i < len(tag) && tag[i] == ' ' {
			i++
		}
		// Skips quote but what if not quoted value
		if i < len(tag) && tag[i] != '"' {
			quoted = false
		} else {
			i++
		}
		// Scan quoted string to find value if value is not quoted check for
		// comma
		for i < len(tag) && !((quoted && tag[i] == '"') || (!quoted && tag[i] == ',')) {
			if tag[i] == '\\' {
				i++
			}
			i++
		}
		if quoted {
			i++
		}
		if i >= len(tag) {
			i = len(tag) - 1
		}
		value := string(tag[:i+1])
		tag = tag[i+1:]
		if key == name {
			value = strings.Trim(value, " ")
			value = strings.Trim(value, ",")
			/*if ! strings.Contains(value, `"`) {
				return value
			}
			value, _ = strconv.Unquote(value)*/
			return value
		}

	}
	return ""
}

// Helper to retrieve Type name and package name with which we
// use to name a Model within the domain
func importName(pType reflect.Type) string {
	return fmt.Sprintf("%v", pType)
}

func getKind(pName string) string {
	switch pName {
	case "String":
		return "reflect.String"
	case "Int64":
		return "reflect.Int64"
	case "AutoIncrement":
		return "reflect.Int64"
	case "Float64":
		return "reflect.Float64"
	case "Bool":
		return "reflect.Bool"
	case "Key":
		return "PrimaryKey"
	case "Time":
		return "OpalTime"
	case "Slice":
		return "reflect.Slice"
	default:
		log.Panicln("Opal.Kind: non supported type:", pName)
	}
	return "paniced"
}
