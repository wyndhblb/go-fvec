// The GoLang template for generation of go based objects

package main

var GOLANG_TEMPLATE string = `
// Auto generated by fvecgen on %s
// Edit at your own risk
//

package {{.PackageName}}

import (
	"github.com/wyndhblb/fvec"
	"strings"
	"fmt"
)

type {{.ClassName}} struct {
	keyspace string
	table string
	Name *fvec.VName
{{ range .Fields }}
	{{.FieldName}} *fvec.{{.FieldType}} ` + "`" + `json:"{{.ColumnName}}" cql:"{{.ColumnName}}" msg:"{{.ColumnName}}"` + "`" + `
{{ end }}

}

// New{{.ClassName}} a new object
func New{{.ClassName}}(keyspace, table string) *{{.ClassName}}{
	n := new({{.ClassName}})
	n.keyspace = keyspace
	n.table = table
	n.Name = new(fvec.VName)
	{{ range .Fields }}
	n.{{.FieldName}} = new(fvec.{{.FieldType}})
	{{ end }}
	return n
}

// Table name of the table
func (f *{{.ClassName}}) Table() string {
	return f.table
}

// Keyspace name of the keyspace
func (f *{{.ClassName}}) Keyspace() string {
	return f.keyspace
}

// DBColumns list of non counter db columns as strings
func (f *{{.ClassName}}) DBColumns() []string {
	return []string{
	{{ range .Fields }}
		{{ if .IsCounter }}
		{{ else }}
		"{{ .ColumnName }}",
		{{ end }}
	{{ end }}
	}
}


// DBCounterColumns list of db columns as strings that are counters
func (f *{{.ClassName}}) DBCounterColumns() []string {
	return []string{
	{{ range .Fields }}
		{{ if .IsCounter }}
		"{{ .ColumnName }}",
		{{ end }}
	{{ end }}
	}
}

// TypeStrings list of the column types as strings
func (f *{{.ClassName}}) TypeStrings() []string {
	return []string{
	{{ range .Fields }}
		"{{ .FieldType }}",
	{{ end }}
	}
}

// VarNameStrings the class fields as strings
func (f *{{.ClassName}}) VarNameStrings() []string {
	return []string{
	{{ range .Fields }}
		"{{ .FieldName }}",
	{{ end }}
	}
}

// NewVectorFromFieldName return a new vector of the type the column is
// an error will occur if the type itself is a Scalar
func (f *{{.ClassName}}) NewVectorFromFieldName(nm string) (fvec.Vector, error) {
	vmap := map[string]string{
	{{ range .Fields }}
		{{ if .IsVector }}
		"{{ .FieldName }}":"{{.FieldType}}",
		{{ end }}
	{{ end }}
	}

	if got, ok := vmap[nm]; ok{
		v := fvec.GetVectorFromString(got)
		if v == nil{
			return nil, fmt.Errorf("%s is not a vector", nm)
		}
		return v, nil
	}
	return nil, fmt.Errorf("%s is not a field", nm)
}


// NewScalarFromVarName return a new scalar of the type the column is
// an error will occur if the type itself is a Scalar
func (f *{{.ClassName}}) NewScalarFromFieldName(nm string) (fvec.Scalar, error) {
	vmap := map[string]string{
	{{ range .Fields }}
		{{ if .IsScalar }}
		"{{ .FieldName }}":"{{.FieldType}}",
		{{ end }}
	{{ end }}
	}

	if got, ok := vmap[nm]; ok{
		v := fvec.GetScalarFromString(got)
		if v == nil{
			return nil, fmt.Errorf("%s is not a scalar", nm)
		}
		return v, nil
	}
	return nil, fmt.Errorf("%s is not a field", nm)
}

// CassandraCreateStatement the list of cassandra create the table statement
// if there are counters in the mix, then there will be another table {table}_counters
// cassandra only allows counters in a table to itself aside from the primary key
func (f *{{.ClassName}}) CassandraCreateStatement() []string {

	queries := []string{}
	subs := []string{
		"id ascii",
		"slab ascii",
		"ord ascii",
		"key text",
		"tags map<text, text>",
	}
	cSubs := []string{
		"id ascii",
		"slab ascii",
		"ord ascii",
		"key text",
		"tags map<text, text>",
	}
	haveC := false
	haveA := false
	{{ range .Fields }}
	if len(f.{{.FieldName}}.CassandraCreateType(f.keyspace)) > 0{
		queries = append(queries, f.{{.FieldName}}.CassandraCreateType(f.keyspace))
	}
	{{ if .IsCounter }}
	cSubs = append(cSubs, "{{.ColumnName}} " + f.{{.FieldName}}.CassandraType())
	haveC = true
	{{ else }}
	subs = append(subs, "{{.ColumnName}} " + f.{{.FieldName}}.CassandraType())
	haveA = true
	{{ end }}
	{{ end }}

	createSQL := "CREATE TABLE IF NOT EXISTS " + f.keyspace + "." + f.table + "("
	createSQL += strings.Join(subs, ", ")
	createSQL += ` + "`" + `
		, PRIMARY KEY ((uid, slab), ord)
		) WITH CLUSTERING ORDER BY (ord ASC) AND
		compaction = {
		'class': 'TimeWindowCompactionStrategy',
		'compaction_window_unit': 'DAYS',
		'compaction_window_size': '1'
		}
		AND compression = {'sstable_compression': 'org.apache.cassandra.io.compress.LZ4Compressor'};
	` + "`" + `

	counterSQL := "CREATE TABLE IF NOT EXISTS " + f.keyspace + "." + f.table + "_counters("
	counterSQL += strings.Join(cSubs, ", ")
	counterSQL += ` + "`" + `
		, PRIMARY KEY ((uid, slab), ord)
		) WITH compaction = {'class': 'org.apache.cassandra.db.compaction.SizeTieredCompactionStrategy'}
		AND compression = {'sstable_compression': 'org.apache.cassandra.io.compress.LZ4Compressor'};
	` + "`" + `

	if haveA{
		queries = append(queries, createSQL)
	}
	if haveC{
		queries = append(queries, counterSQL)
	}
	return queries
}

// CassandraSelectQueries the set of queries to get the full object
// this does not include a where clause just the SELECT (stuff, stuff, ...) FROM {table}
// if there are counters there will be 2 queries for the counters table
func (f *{{.ClassName}}) CassandraSelectQueries() []string {
	return []string{
		"SELECT " + strings.Join(f.DBColumns(), ",") + " FROM " + f.table,
		{{ if .HaveCounters }}
		"SELECT " + strings.Join(f.DBCounterColumns(), ",") + " FROM " + f.table,
		{{ end }}
	}
}

// BuildObject given a set of where args (a string that is the WHERE col=? AND col=? and the param arges
// build the in total, as we may have to find things from a counter table
func (f *{{.ClassName}}) BuildObject(where string, args ...interface{}) []string {
	return []string{
		"SELECT " + strings.Join(f.DBColumns(), ",") + " FROM " + f.table,
		{{ if .HaveCounters }}
		"SELECT " + strings.Join(f.DBCounterColumns(), ",") + " FROM " + f.table,
		{{ end }}
	}
}

// GetName return the name object
func (f *{{.ClassName}}) GetName() *fvec.VName {
	return f.Name
}

// GetVectors get all the vectors as a list
func (f *{{.ClassName}}) GetVectors() []fvec.Vector{
	return []fvec.Vector{
	{{ range .Fields }}
		{{ if .IsVector }}
		f.{{.FieldName}},
		{{ end }}
	{{ end }}
	}
}

// GetScalars get all the scalars as a list
func (f *{{.ClassName}}) GetScalars() []fvec.Scalar{
	return []fvec.Scalar{
	{{ range .Fields }}
		{{ if .IsScalar }}
		f.{{.FieldName}},
		{{ end }}
	{{ end }}
	}
}


// GetCounters get all the counters as a list
func (f *{{.ClassName}}) GetCounters() []fvec.Scalar{
	return []fvec.Scalar{
	{{ range .Fields }}
		{{ if .IsCounter }}
		f.{{.FieldName}},
		{{ end }}
	{{ end }}
	}
}

// Key key name
func (f *{{.ClassName}}) Key() string{
	return f.Name.Key
}


// UniqueId unique id of the vector
func (f *{{.ClassName}}) UniqueId() uint64{
	return f.Name.UniqueId()
}

// UniqueIdString unique id as a base36 string of the vector
func (f *{{.ClassName}}) UniqueIdString() string{
	return f.Name.UniqueIdString()
}
`
