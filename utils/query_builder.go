package utils

import (
    "fmt"
    "strings"
    "reflect"
)

// QueryBuilder is a struct that helps to build dynamic queries.
type QueryBuilder struct {
    SelectFields []string
    Conditions   map[string]interface{}
    SortFields   []string
    Limit        int
    Offset       int
}

// NewQueryBuilder initializes a new QueryBuilder instance.
func NewQueryBuilder() *QueryBuilder {
    return &QueryBuilder{
        Conditions: make(map[string]interface{}),
    }
}

// Select specifies the fields to be selected in the query.
func (qb *QueryBuilder) Select(fields ...string) *QueryBuilder {
    qb.SelectFields = fields
    return qb
}

// Where adds a condition to the query.
func (qb *QueryBuilder) Where(field string, value interface{}) *QueryBuilder {
    qb.Conditions[field] = value
    return qb
}

// WhereIn adds an IN condition to the query for multiple values.
func (qb *QueryBuilder) WhereIn(field string, values []interface{}) *QueryBuilder {
    qb.Conditions[field] = map[string]interface{}{"$in": values}
    return qb
}

// WhereLike adds a LIKE condition to the query.
func (qb *QueryBuilder) WhereLike(field, pattern string) *QueryBuilder {
    qb.Conditions[field] = map[string]string{"$like": pattern}
    return qb
}

// WhereBetween adds a BETWEEN condition to the query.
func (qb *QueryBuilder) WhereBetween(field string, from, to interface{}) *QueryBuilder {
    qb.Conditions[field] = map[string]interface{}{"$between": []interface{}{from, to}}
    return qb
}

// Sort specifies the fields to sort by. Prefix with "-" for descending order.
func (qb *QueryBuilder) Sort(fields ...string) *QueryBuilder {
    qb.SortFields = fields
    return qb
}

// SetLimit sets the number of records to return.
func (qb *QueryBuilder) SetLimit(limit int) *QueryBuilder {
    qb.Limit = limit
    return qb
}

// SetOffset sets the starting point for records to return.
func (qb *QueryBuilder) SetOffset(offset int) *QueryBuilder {
    qb.Offset = offset
    return qb
}

// ToSQL converts the QueryBuilder into a SQL WHERE clause and parameters.
func (qb *QueryBuilder) ToSQL() (string, []interface{}) {
    var conditions []string
    var params []interface{}
    counter := 1

    for field, value := range qb.Conditions {
        switch v := value.(type) {
        case map[string]interface{}:
            if inVals, ok := v["$in"]; ok {
                placeholders := []string{}
                for _, val := range inVals.([]interface{}) {
                    placeholders = append(placeholders, fmt.Sprintf("$%d", counter))
                    params = append(params, val)
                    counter++
                }
                conditions = append(conditions, fmt.Sprintf("%s IN (%s)", field, strings.Join(placeholders, ", ")))
            } else if betweenVals, ok := v["$between"]; ok {
                conditions = append(conditions, fmt.Sprintf("%s BETWEEN $%d AND $%d", field, counter, counter+1))
                params = append(params, betweenVals.([]interface{})...)
                counter += 2
            } else if likeVal, ok := v["$like"]; ok {
                conditions = append(conditions, fmt.Sprintf("%s LIKE $%d", field, counter))
                params = append(params, likeVal)
                counter++
            }
        default:
            conditions = append(conditions, fmt.Sprintf("%s = $%d", field, counter))
            params = append(params, value)
            counter++
        }
    }

    whereClause := ""
    if len(conditions) > 0 {
        whereClause = "WHERE " + strings.Join(conditions, " AND ")
    }

    orderBy := ""
    if len(qb.SortFields) > 0 {
        orderBy = "ORDER BY " + qb.buildSortClause()
    }

    limitClause := ""
    if qb.Limit > 0 {
        limitClause = fmt.Sprintf("LIMIT %d", qb.Limit)
    }

    offsetClause := ""
    if qb.Offset > 0 {
        offsetClause = fmt.Sprintf("OFFSET %d", qb.Offset)
    }

    return fmt.Sprintf("%s %s %s %s", whereClause, orderBy, limitClause, offsetClause), params
}

// buildSortClause generates the ORDER BY clause for SQL.
func (qb *QueryBuilder) buildSortClause() string {
    var sorts []string
    for _, field := range qb.SortFields {
        if strings.HasPrefix(field, "-") {
            sorts = append(sorts, fmt.Sprintf("%s DESC", strings.TrimPrefix(field, "-")))
        } else {
            sorts = append(sorts, fmt.Sprintf("%s ASC", field))
        }
    }
    return strings.Join(sorts, ", ")
}

// ToMongoFilter converts the QueryBuilder into a MongoDB filter.
func (qb *QueryBuilder) ToMongoFilter() map[string]interface{} {
    filter := make(map[string]interface{})
    for field, value := range qb.Conditions {
        switch v := value.(type) {
        case map[string]interface{}:
            if inVals, ok := v["$in"]; ok {
                filter[field] = map[string]interface{}{"$in": inVals}
            } else if betweenVals, ok := v["$between"]; ok {
                filter[field] = map[string]interface{}{"$gte": betweenVals.([]interface{})[0], "$lte": betweenVals.([]interface{})[1]}
            } else if likeVal, ok := v["$like"]; ok {
                filter[field] = map[string]interface{}{"$regex": likeVal, "$options": "i"}
            }
        default:
            filter[field] = value
        }
    }
    return filter
}

// GetMongoSort converts the QueryBuilder into a MongoDB sort specification.
func (qb *QueryBuilder) GetMongoSort() map[string]int {
    sortSpec := make(map[string]int)
    for _, field := range qb.SortFields {
        if strings.HasPrefix(field, "-") {
            sortSpec[strings.TrimPrefix(field, "-")] = -1
        } else {
            sortSpec[field] = 1
        }
    }
    return sortSpec
}

// ========================= EXAMPLE USAGE ============================

// qb := utils.NewQueryBuilder().
//     Where("name", "John").
//     Where("status", "active").
//     Sort("-created_at").
//     SetLimit(10).
//     SetOffset(20)

// sqlQuery, params := qb.ToSQL()
// fmt.Println("SQL Query:", sqlQuery)
// fmt.Println("Params:", params)

// // Output:
// // SQL Query: WHERE name = $1 AND status = $2 ORDER BY created_at DESC LIMIT 10 OFFSET 20
// // Params: [John active]

// qb := utils.NewQueryBuilder().
//     Where("name", "John").
//     Where("status", "active").
//     Sort("-created_at").
//     SetLimit(10).
//     SetOffset(0)

// mongoFilter := qb.ToMongoFilter()
// mongoSort := qb.GetMongoSort()

// fmt.Println("Mongo Filter:", mongoFilter)
// fmt.Println("Mongo Sort:", mongoSort)

// // Output:
// // Mongo Filter: map[name:John status:active]
// // Mongo Sort: map[created_at:-1]
