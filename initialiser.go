/**
 * Date: 9/02/14
 * Time: 7:06 PM
 */
package opal

import (
	"bytes"
	"fmt"
	"github.com/twinj/version"
	"bitbucket.org/pkg/inflect"
	"io/ioutil"
	"log"
	"os"
	"reflect"
	"strconv"
	"strings"
	"text/template"
	"time"
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
	plate.Version = version.String()
	plate.Date = now.Format("2 January 2006")
	plate.Time = now.Format("Monday 3:04 AM")

	// Gather Table information:
	plate.Types = make(map[string]*TemplateType, len(pModels))

	// Check the first field
	// if it's an anonymous Entity extract its metadata
	// if not panic - invalid Domain Model
	for _, domain := range pModels {
		t := reflect.TypeOf(domain).Elem()
		var dType TemplateType
		//var useDefaultKey bool
		dType.Domain = domain
		dType.Version = plate.Version
		dType.Date = plate.Date
		dType.Time = plate.Time
		dType.Model = t.Name()
		dType.DAOName = inflect.Camelize(inflect.Tableize(dType.Model))

		dType.Receiver = strings.ToLower(t.Name())
		dType.Path = t.PkgPath()
		dType.ImportName = importName(t)

		s := strings.Split(dType.Path, "/")
		dType.Package = s[len(s)-1]

		var keys map[string]bool = make(map[string]bool)
		var i int
		if t.Field(i).Type.Implements(reflect.TypeOf((*Entity)(nil)).Elem()) && t.Field(i).Anonymous {
			opalTags := ExtractOpalTags(t.Field(i).Tag)

			// Derive table name or get specific
			if opalTags.Get("Name") == "" {
				dType.Table = fmt.Sprintf("Name: %q", inflect.Tableize(dType.Model))
			} else {
				dType.Table = string(opalTags)
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
		opal := reflect.TypeOf((*OPAL)(nil)).Elem()

		if t.Field(i).Type.Name() == "Key" && t.Field(i).Anonymous {
			opalTags := ExtractOpalTags(t.Field(i).Tag)

			// TODO support compound keys // TODO check field type
			if _, ok := t.FieldByName(string(opalTags)); ok {
				keys[string(opalTags)] = true
			} else {
				panic("Opal.runTemplate: Key metatag error")
			}
			i++
		}

		// Check each field if its an OPAL extract its metadata
		for i < t.NumField() {
			if t.Field(i).Type.Implements(opal) {
				opalTags := ExtractOpalTags(t.Field(i).Tag)
				if opalTags == "" {
					opalTags = Tag(fmt.Sprintf("Name: %q", t.Field(i).Name))
				}
				// Infer column name or get explicit
				if opalTags.Get("Name") == "" {
					opalTags = Tag(fmt.Sprintf("Name: %q, %s", t.Field(i).Name, opalTags))
				}
				if keys[t.Field(i).Name] {
					if t.Field(i).Type.Name() == "AutoIncrement" {
						opalTags += ", AutoIncrement: true"
					}
					kind := reflect.Kind(reflect.New(t.Field(i).Type).MethodByName("Kind").Call(nil)[0].Uint())
					key := KeyField{t.Field(i).Name, t.Field(i).Type.Name(), i, string(opalTags), getKind(t.Field(i).Type.Name()), kind.String()}

					dType.Keys = append(dType.Keys, key)
				} else {
					kind := reflect.Kind(reflect.New(t.Field(i).Type).MethodByName("Kind").Call(nil)[0].Uint())
					s := kind.String()
					if s == "slice" {
						s = "[]byte"
					} // TODO handle primitive types properly
					dType.Columns = append(dType.Columns, TemplateField{t.Field(i).Name, i, string(opalTags), getKind(t.Field(i).Type.Name()), s})
				}
			}
			i++
		}
		plate.Types[t.Name()] = &dType
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

// Helper to retrieve Type name and package name with which we
// use to name a Model within the domain
func importName(pType reflect.Type) string {
	s := strings.Split(pType.PkgPath(), "/")[0]
	s = strings.TrimLeft(pType.PkgPath(), s+"/")
	s = fmt.Sprintf("%s.%s", s, pType.Name())
	return s
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
		panic("Opal.Kind: non supported type")
	}
}

func ExtractOpalTags(pStructTag reflect.StructTag) Tag {
	tag := string(pStructTag)
	if strings.Count(tag, "|") == 2 {
		tag = strings.Split(tag, "|")[1] // TODO check proper form
		return Tag(tag)
	}
	return Tag("")
}

// TODO Could map all tags

func runTemplate(e ModelTemplate) {
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
	for _, li := range e.Types {
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
		for i < len(tag) && !( (quoted && tag[i] == '"') || ( ! quoted && tag[i] == ',')) {
			if tag[i] == '\\' {
				i++
			}
			i++
		}
		if quoted {
			i++
		}
		if i >= len(tag) {
			i = len(tag)-1
		}
		value := string(tag[:i + 1])
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
