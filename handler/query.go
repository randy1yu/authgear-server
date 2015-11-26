package handler

import (
	"errors"
	"fmt"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/oursky/skygear/skydb"
	"github.com/oursky/skygear/skydb/skydbconv"
	"github.com/oursky/skygear/skyerr"
)

func sortFromRaw(rawSort []interface{}, sort *skydb.Sort) {
	var (
		keyPath   string
		funcExpr  skydb.Func
		sortOrder skydb.SortOrder
	)
	switch v := rawSort[0].(type) {
	case map[string]interface{}:
		if err := (*skydbconv.MapKeyPath)(&keyPath).FromMap(v); err != nil {
			panic(err)
		}
	case []interface{}:
		var err error
		funcExpr, err = parseFunc(v)
		if err != nil {
			panic(err)
		}
	default:
		panic(fmt.Errorf("unexpected type of sort expression = %T", rawSort[0]))
	}

	orderStr, _ := rawSort[1].(string)
	if orderStr == "" {
		panic(errors.New("empty sort order in sort descriptor"))
	}
	switch orderStr {
	case "asc":
		sortOrder = skydb.Asc
	case "desc":
		sortOrder = skydb.Desc
	default:
		panic(fmt.Errorf("unknown sort order: %v", orderStr))
	}

	sort.KeyPath = keyPath
	sort.Func = funcExpr
	sort.Order = sortOrder
}

func sortsFromRaw(rawSorts []interface{}) []skydb.Sort {
	length := len(rawSorts)
	sorts := make([]skydb.Sort, length, length)

	for i := range rawSorts {
		sortSlice, _ := rawSorts[i].([]interface{})
		if len(sortSlice) != 2 {
			panic(fmt.Errorf("got len(sort descriptor) = %v, want 2", len(sortSlice)))
		}
		sortFromRaw(sortSlice, &sorts[i])
	}

	return sorts
}

func predicateOperatorFromString(operatorString string) skydb.Operator {
	switch operatorString {
	case "and":
		return skydb.And
	case "or":
		return skydb.Or
	case "not":
		return skydb.Not
	case "eq":
		return skydb.Equal
	case "gt":
		return skydb.GreaterThan
	case "lt":
		return skydb.LessThan
	case "gte":
		return skydb.GreaterThanOrEqual
	case "lte":
		return skydb.LessThanOrEqual
	case "neq":
		return skydb.NotEqual
	case "like":
		return skydb.Like
	case "ilike":
		return skydb.ILike
	case "in":
		return skydb.In
	default:
		panic(fmt.Errorf("unrecognized operator = %s", operatorString))
	}
}

func predicateFromRaw(rawPredicate []interface{}) skydb.Predicate {
	if len(rawPredicate) < 2 {
		panic(fmt.Errorf("got len(predicate) = %v, want at least 2", len(rawPredicate)))
	}

	rawOperator, ok := rawPredicate[0].(string)
	if !ok {
		panic(fmt.Errorf("got predicate[0]'s type = %T, want string", rawPredicate[0]))
	}

	operator := predicateOperatorFromString(rawOperator)
	children := make([]interface{}, len(rawPredicate)-1)
	for i := 1; i < len(rawPredicate); i++ {
		if operator.IsCompound() {
			subRawPredicate, ok := rawPredicate[i].([]interface{})
			if !ok {
				panic(fmt.Errorf("got non-dict in subpredicate at %v", i-1))
			}
			children[i-1] = predicateFromRaw(subRawPredicate)
		} else {
			expr := parseExpression(rawPredicate[i])
			if expr.Type == skydb.KeyPath && strings.Contains(expr.Value.(string), ".") {

				panic(fmt.Errorf("Key path `%s` is not supported.", expr.Value))
			}
			children[i-1] = expr
		}
	}

	if operator.IsBinary() && len(children) != 2 {
		panic(fmt.Errorf("Expected number of expressions be 2, got %v", len(children)))
	}

	predicate := skydb.Predicate{
		Operator: operator,
		Children: children,
	}
	return predicate
}

func parseExpression(i interface{}) skydb.Expression {
	switch v := i.(type) {
	case map[string]interface{}:
		var keyPath string
		if err := skydbconv.MapFrom(i, (*skydbconv.MapKeyPath)(&keyPath)); err == nil {
			return skydb.Expression{
				Type:  skydb.KeyPath,
				Value: keyPath,
			}
		}
	case []interface{}:
		if len(v) > 0 {
			if f, err := parseFunc(v); err == nil {
				return skydb.Expression{
					Type:  skydb.Function,
					Value: f,
				}
			}
		}
	}

	return skydb.Expression{
		Type:  skydb.Literal,
		Value: skydbconv.ParseInterface(i),
	}
}

func parseFunc(s []interface{}) (f skydb.Func, err error) {
	keyword, _ := s[0].(string)
	if keyword != "func" {
		return nil, errors.New("not a function")
	}

	funcName, _ := s[1].(string)
	switch funcName {
	case "distance":
		f, err = parseDistanceFunc(s[2:])
	case "":
		return nil, errors.New("empty function name")
	default:
		return nil, fmt.Errorf("got unrecgonized function name = %s", funcName)
	}

	return
}

func parseDistanceFunc(s []interface{}) (*skydb.DistanceFunc, error) {
	if len(s) != 2 {
		return nil, fmt.Errorf("want 2 arguments for distance func, got %d", len(s))
	}

	var field string
	if err := skydbconv.MapFrom(s[0], (*skydbconv.MapKeyPath)(&field)); err != nil {
		return nil, fmt.Errorf("invalid key path: %v", err)
	}

	var location skydb.Location
	if err := skydbconv.MapFrom(s[1], (*skydbconv.MapLocation)(&location)); err != nil {
		return nil, fmt.Errorf("invalid location: %v", err)
	}

	return &skydb.DistanceFunc{
		Field:    field,
		Location: &location,
	}, nil
}

func queryFromRaw(rawQuery map[string]interface{}, query *skydb.Query) (err skyerr.Error) {
	defer func() {
		// use panic to escape from inner error
		if r := recover(); r != nil {
			if queryErr, ok := r.(error); ok {
				log.WithField("rawQuery", rawQuery).Debugln("failed to construct query")
				err = skyerr.NewFmt(skyerr.RequestInvalidErr, "failed to construct query: %v", queryErr.Error())
			} else {
				log.WithField("recovered", r).Errorln("panic recovered while constructing query")
				err = skyerr.New(skyerr.RequestInvalidErr, "error occurred while constructing query")
			}
		}
	}()
	recordType, _ := rawQuery["record_type"].(string)
	if recordType == "" {
		return skyerr.New(skyerr.RequestInvalidErr, "recordType cannot be empty")
	}
	query.Type = recordType

	mustDoSlice(rawQuery, "predicate", func(rawPredicate []interface{}) error {
		predicate := predicateFromRaw(rawPredicate)
		if err := predicate.Validate(); err != nil {
			return skyerr.NewRequestInvalidErr(fmt.Errorf("invalid predicate: %v", err))
		}
		query.Predicate = &predicate
		return nil
	})

	mustDoSlice(rawQuery, "sort", func(rawSorts []interface{}) error {
		query.Sorts = sortsFromRaw(rawSorts)
		return nil
	})

	if transientIncludes, ok := rawQuery["include"].(map[string]interface{}); ok {
		query.ComputedKeys = map[string]skydb.Expression{}
		for key, value := range transientIncludes {
			query.ComputedKeys[key] = parseExpression(value)
		}
	}

	mustDoSlice(rawQuery, "desired_keys", func(desiredKeys []interface{}) error {
		query.DesiredKeys = make([]string, len(desiredKeys))
		for i, key := range desiredKeys {
			key, ok := key.(string)
			if !ok {
				return skyerr.New(skyerr.RequestInvalidErr, "unexpected value in desired_keys")
			}
			query.DesiredKeys[i] = key
		}
		return nil
	})

	if getCount, ok := rawQuery["count"].(bool); ok {
		query.GetCount = getCount
	}

	if offset, _ := rawQuery["offset"].(float64); offset > 0 {
		query.Offset = uint64(offset)
	}

	if limit, ok := rawQuery["limit"].(float64); ok {
		query.Limit = new(uint64)
		*query.Limit = uint64(limit)
	}
	return nil
}

// execute do when if the value of key in m is []interface{}. If value exists
// for key but its type is not []interface{} or do returns an error, it panics.
func mustDoSlice(m map[string]interface{}, key string, do func(value []interface{}) error) {
	vi, ok := m[key]
	if ok && vi != nil {
		v, ok := vi.([]interface{})
		if ok {
			if err := do(v); err != nil {
				panic(err)
			}
		} else {
			panic(fmt.Errorf("%#s has to be an array", key))
		}
	}
}
