package fvec

import (
	"bytes"
	"fmt"
	"github.com/wyndhblb/go-utils/pools"
	"github.com/wyndhblb/timeslab"
	"reflect"
	"testing"
	"time"
)

func Test_Slab_Formatting(t *testing.T) {

	ti := time.Date(2009, time.November, 10, 23, 1, 2, 0, time.UTC)

	tData := make(map[timeslab.Resolution]string)
	tData[timeslab.Resolution_MIN] = "200911102301"
	tData[timeslab.Resolution_MIN10] = "2009111023I100"
	tData[timeslab.Resolution_MIN30] = "2009111023I300"
	tData[timeslab.Resolution_HOUR] = "2009111023"
	tData[timeslab.Resolution_DAY] = "20091110"
	tData[timeslab.Resolution_WEEK] = "200946"
	tData[timeslab.Resolution_MONTH] = "200911"
	tData[timeslab.Resolution_MONTH2] = "2009M25"
	tData[timeslab.Resolution_MONTH3] = "2009M33"
	tData[timeslab.Resolution_MONTH6] = "2009M61"
	tData[timeslab.Resolution_YEAR] = "2009"

	for res, st := range tData {
		nm.Resolution = res
		onSl := nm.ToSlab(ti)
		if onSl != st {
			t.Fatalf("Invalid time slab: got: %s, wanted: %s for resolution %s", onSl, st, timeslab.Resolution_name[int32(res)])
		}
	}

	ti = time.Date(2009, time.May, 30, 6, 45, 2, 0, time.UTC)
	tData = make(map[timeslab.Resolution]string)
	tData[timeslab.Resolution_MIN] = "200905300645"
	tData[timeslab.Resolution_MIN10] = "2009053006I104"
	tData[timeslab.Resolution_MIN30] = "2009053006I301"
	tData[timeslab.Resolution_HOUR] = "2009053006"
	tData[timeslab.Resolution_DAY] = "20090530"
	tData[timeslab.Resolution_WEEK] = "200922"
	tData[timeslab.Resolution_MONTH] = "200905"
	tData[timeslab.Resolution_MONTH2] = "2009M22"
	tData[timeslab.Resolution_MONTH3] = "2009M31"
	tData[timeslab.Resolution_MONTH6] = "2009M60"
	tData[timeslab.Resolution_YEAR] = "2009"

	for res, st := range tData {
		nm.Resolution = res
		onSl := nm.ToSlab(ti)
		if onSl != st {
			t.Fatalf("Invalid time slab: got: %s, wanted: %s for resolution %s", onSl, st, timeslab.Resolution_name[int32(res)])
		}
	}
}

func getTypeName(v interface{}) string {
	if t := reflect.TypeOf(v); t.Kind() == reflect.Ptr {
		return "*" + t.Elem().Name()
	} else {
		return t.Name()
	}
}

// run through all the permutations to make sure things are assigned and really available
// from the generate stages
func Test_GetVector(t *testing.T) {

	basetypes := [...]byte{'l', 's', 'm'}
	stypes := [...]string{"s", "i", "d"}
	for _, b := range basetypes {
		for _, f := range stypes {
			v := GetVector(b, f)
			t.Logf("got: %T for %s", v, string(b)+f)
			if v == nil && b != 'm' {
				t.Fatalf("Failed to get vector for %s + %s", string(b), f)
			}
			if v != nil {
				tMap := SHORT_NAME_MAP[string(b)+f]
				if tMap == "" || "*"+tMap != getTypeName(v) {
					t.Fatalf("Failed to get vector for %s + %s :: got wrong type (%s->%s)", string(b), f, tMap, getTypeName(v))
				}
			}

			for _, f2 := range stypes {
				v := GetVector(b, f+f2)
				t.Logf("got: %T for %s", v, string(b)+f+f2)
				if f != "d" && v == nil {
					t.Fatalf("Failed to get vector for %s + %s%s", string(b), f, f2)
				}

				if v != nil {
					tMap := SHORT_NAME_MAP[string(b)+f+f2]
					if tMap == "" || "*"+tMap != getTypeName(v) {
						t.Fatalf("Failed to get vector for %s + %s%s :: got wrong type  (%s->%s)", string(b), f, f2, tMap, getTypeName(v))
					}
				}

				if b == 'm' && f != "d" {
					for _, f3 := range stypes {
						v := GetVector(b, f+f2+f3)
						t.Logf("got: %T for %s", v, string(b)+f+f2+f3)
						if v == nil {
							t.Fatalf("Failed to get vector for %s + %s%s%s", string(b), f, f2, f3)
						}
						tMap := SHORT_NAME_MAP[string(b)+f+f2+f3]
						if tMap == "" || "*"+tMap != getTypeName(v) {
							t.Fatalf("Failed to get vector for %s + %s%s%s :: got wrong type  (%s->%s)", string(b), f, f, f3, tMap, getTypeName(v))
						}
					}
				}
			}

		}
	}
}

// run through all the permutations to make sure things are assigned and really available
// from the generate stages
func Test_GetVectorFromCassType(t *testing.T) {
	for k, v := range SHORT_NAME_MAP {
		vec := GetVectorFromString(k)
		if vec == nil {
			t.Fatalf("Invalid vector type: from %s, %v", k, v)
		}
		vec = GetVectorFromString(v)
		if vec == nil {
			t.Fatalf("Invalid vector type: from %s, %v", v, v)
		}
	}
	nope := GetVectorFromString("mnkey")
	if nope != nil {
		t.Fatal("Should NOT have gotten nil for `mnkey`")
	}
}

// run through all the permutations to make sure things are assigned and really available
// from the generate stages
func Test_GetVectorFromString(t *testing.T) {
	for k, v := range CASSANDRA_TYPE_MAP {
		vec := GetVectorFromCassType(k)
		if vec.TypeName() != v {
			t.Fatalf("Invalid vector type: %v", v)
		}
	}
}

type testWideCass struct {
	Vlii  VLIntInt
	Vssd  VSDbl
	Vmss  VMStrStr
	Vmiss VMIntTPStrStr
	Vmiii VMIntTPIntInt
}

// simple test to make sure a struct of vectors gets the types properly created
func Test_CassandraWideRow(t *testing.T) {
	n := new(testWideCass)
	keysp := "test"

	should := []string{
		"CREATE TYPE IF NOT EXISTS test.VTIntInt ( k bigint, v bigint );",
		"",
		"",
		"CREATE TYPE IF NOT EXISTS test.VTStrStr ( k varchar, v varchar );",
		"CREATE TYPE IF NOT EXISTS test.VTIntInt ( k bigint, v bigint );",
	}

	data := []string{
		n.Vlii.CassandraCreateType(keysp),
		n.Vssd.CassandraCreateType(keysp),
		n.Vmss.CassandraCreateType(keysp),
		n.Vmiss.CassandraCreateType(keysp),
		n.Vmiii.CassandraCreateType(keysp),
	}

	for i, str := range data {
		if str != should[i] {
			t.Fatalf("Invalid create type %s != %s", str, should[i])
		}
	}

	data = []string{
		n.Vlii.CassandraType(),
		n.Vssd.CassandraType(),
		n.Vmss.CassandraType(),
		n.Vmiss.CassandraType(),
		n.Vmiii.CassandraType(),
	}
	should = []string{
		"list<frozen<VTIntInt>>",
		"set<double>",
		"map<varchar,varchar>",
		"map<bigint,frozen<VTStrStr>>",
		"map<bigint,frozen<VTIntInt>>",
	}
	for i, str := range data {
		fmt.Println(i, str)
		if str != should[i] {
			t.Fatalf("Invalid create type %s != %s", str, should[i])
		}
	}
}

// benching
var key string = "stats.cadent-all.all-1-stats-infra-integ.mfpaws.com.reader.cassandra.rawrender.get-time-ns.count"
var tgs string = "moo=goo,loo=goo,houst=all-1-stats-infra-integ.mfpaws.com"

var nm VName = VName{
	Key:  key,
	Tags: TagsFromString(tgs),
}

// benching fmt.Fprintf(buf ...) vs []byte(s + s + s)
func Benchmark__FmtF_String(b *testing.B) {

	b.ResetTimer()
	b.ReportAllocs()
	buf := bytes.NewBuffer(nil)
	for i := 0; i < b.N; i++ {
		fmt.Fprintf(buf, "%s:%s:%s", key, tgs)
	}
}

// benching fmt.Fprintf(buf ...) vs []byte(s + s + s)
func Benchmark__Add_String(b *testing.B) {

	b.ResetTimer()
	b.ReportAllocs()
	buf := bytes.NewBuffer(nil)
	for i := 0; i < b.N; i++ {
		buf.Write([]byte(key + ":" + tgs))
	}
}

func Benchmark__UniqueIdAllTagsAdd(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		buf := pools.GetFnv64a()
		fmt.Fprintf(buf, "%s:%s", nm.Key, nm.SortedTags())
		_ = buf.Sum64()
		pools.PutFnv64a(buf)
	}
}
