package data

import (
	"github.com/dimall-id/lumos/data/builder"
	"github.com/dimall-id/lumos/misc"
	"gorm.io/gorm"
	"net/http"
	"regexp"
	"strconv"
)

type QueryBuilder interface {
	IsValid (value string) bool
	ApplyQuery (db *gorm.DB, field string, condition string) *gorm.DB
}

type Query struct {
	db *gorm.DB
	builders map[string]QueryBuilder
}

func New (db *gorm.DB) Query  {
	datas := make(map[string]QueryBuilder)
	datas["date"] = &builder.DateBuilder{}
	datas["list"] = &builder.ListBuilder{}
	datas["numeric"] = &builder.NumericBuilder{}
	datas["order"] = &builder.OrderBuilder{}
	datas["select"] = &builder.SelectBuilder{}
	datas["string"] = &builder.StringBuilder{}
	datas["with"] = &builder.WithBuilder{}
	tx := db
	return Query{
		db: tx,
		builders: datas,
	}
}

func (q *Query) AddBuilder (key string, builder QueryBuilder) {
	q.builders[key] = builder
}

func (q *Query) RemoveBuilder (key string) {
	delete(q.builders, key)
}

func (q *Query) GetBuilder (value string) QueryBuilder {
	for _, b := range q.builders {
		if b.IsValid(value) {
			return b
		}
	}
	return nil
}

func (q *Query) Query (queries string) map[string]string {
	r := regexp.MustCompile(`(?:(?P<key>[\w\d\_]+)=(?P<value>[\w\d\:\[\]\,\;\_\%]+))+`)
	exps := r.FindAllStringSubmatch(queries, -1)
	var results = make(map[string]string)
	var keys = make(map[string]int)
	for i, key := range r.SubexpNames() {
		if key != "" {
			keys[key] = i
		}
	}
	for _, exp := range exps {
		results[exp[keys["key"]]] = exp[keys["value"]]
	}
	return results
}

func (q *Query) BuildResponse (r *http.Request, result interface{}) builder.HttpResponse {
	queries := q.Query(r.URL.RawQuery)
	for field, condition := range queries {
		b := q.GetBuilder(condition)
		if b != nil {
			q.db = b.ApplyQuery(q.db, field, condition)
		}
	}

	if val, ok := queries["paging"]; ok {
		if builder.IsPagingValid(val) {
			paging := misc.BuildToMap(builder.PagingPattern, val)
			page,_ := strconv.ParseInt(paging["page"], 10, 32)
			perPage,_ := strconv.ParseInt(paging["per_page"], 10, 32)
			return *builder.Paging(&builder.Param{
				DB: q.db,
				Page: int(page),
				PerPage: int(perPage),
				ShowSQL: true,
				Path: r.URL.Path,
			}, result)
		}
	}

	var count int64
	q.db.Model(result).Count(&count)
	q.db.Find(result)
	return builder.HttpResponse{
		Data: result,
		Total: int(count),
	}
}