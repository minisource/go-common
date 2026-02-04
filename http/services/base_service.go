package services

// import (
// 	"context"
// 	"database/sql"
// 	"fmt"
// 	"math"
// 	"reflect"
// 	"strings"
// 	"time"

// 	"github.com/minisource/go-common/common"
// 	"github.com/minisource/go-common/constants"
// 	"github.com/minisource/go-common/dto"
// 	"github.com/minisource/go-common/logging"
// 	"github.com/minisource/go-common/metrics"
// 	"github.com/minisource/go-common/service_errors"
// )

// type BaseService[T any, Tc any, Tu any, Tr any] struct {
// 	DB     *sqlc.Queries // SQLC-generated queries
// 	Logger logging.Logger
// }

// func NewBaseService[T any, Tc any, Tu any, Tr any](cfg *logging.LoggerConfig, db *sql.DB) *BaseService[T, Tc, Tu, Tr] {
// 	return &BaseService[T, Tc, Tu, Tr]{
// 		DB:     sqlc.New(db), // Initialize SQLC client
// 		Logger: logging.NewLogger(cfg),
// 	}
// }

// func (s *BaseService[T, Tc, Tu, Tr]) Create(ctx context.Context, req *Tc) (*Tr, error) {
// 	model, err := common.TypeConverter[T](req)
// 	if err != nil {
// 		return nil, err
// 	}

// 	// Convert the model to SQLC's input type
// 	input, err := common.TypeConverter[sqlc.CreateModelParams](model)
// 	if err != nil {
// 		return nil, err
// 	}

// 	// Execute the SQL query using SQLC
// 	createdModel, err := s.DB.CreateModel(ctx, *input)
// 	if err != nil {
// 		s.Logger.Error(logging.Postgres, logging.Insert, err.Error(), nil)
// 		metrics.DbCall.WithLabelValues(reflect.TypeOf(*model).String(), "Create", "Failed").Inc()
// 		return nil, err
// 	}

// 	metrics.DbCall.WithLabelValues(reflect.TypeOf(*model).String(), "Create", "Success").Inc()
// 	return common.TypeConverter[Tr](&createdModel)
// }

// func (s *BaseService[T, Tc, Tu, Tr]) Update(ctx context.Context, id int, req *Tu) (*Tr, error) {
// 	updateMap, err := common.TypeConverter[map[string]interface{}](req)
// 	if err != nil {
// 		return nil, err
// 	}

// 	// Convert the update map to SQLC's input type
// 	input := sqlc.UpdateModelParams{
// 		ID:         int32(id),
// 		ModifiedBy: sql.NullInt64{Int64: int64(ctx.Value(constants.UserIdKey).(float64)), Valid: true},
// 		ModifiedAt: sql.NullTime{Time: time.Now().UTC(), Valid: true},
// 	}

// 	// Apply updates from the map
// 	for k, v := range *updateMap {
// 		switch k {
// 		case "field1":
// 			input.Field1 = v.(string)
// 		case "field2":
// 			input.Field2 = v.(int32)
// 		// Add more fields as needed
// 		}
// 	}

// 	// Execute the SQL query using SQLC
// 	updatedModel, err := s.DB.UpdateModel(ctx, input)
// 	if err != nil {
// 		s.Logger.Error(logging.Postgres, logging.Update, err.Error(), nil)
// 		metrics.DbCall.WithLabelValues(reflect.TypeOf(*new(T)).String(), "Update", "Failed").Inc()
// 		return nil, err
// 	}

// 	metrics.DbCall.WithLabelValues(reflect.TypeOf(*new(T)).String(), "Update", "Success").Inc()
// 	return common.TypeConverter[Tr](&updatedModel)
// }

// func (s *BaseService[T, Tc, Tu, Tr]) Delete(ctx context.Context, id int) error {
// 	if ctx.Value(constants.UserIdKey) == nil {
// 		return &service_errors.ServiceError{EndUserMessage: service_errors.PermissionDenied}
// 	}

// 	// Execute the SQL query using SQLC
// 	err := s.DB.DeleteModel(ctx, sqlc.DeleteModelParams{
// 		ID:         int32(id),
// 		DeletedBy:  sql.NullInt64{Int64: int64(ctx.Value(constants.UserIdKey).(float64)), Valid: true},
// 		DeletedAt:  sql.NullTime{Time: time.Now().UTC(), Valid: true},
// 	})
// 	if err != nil {
// 		s.Logger.Error(logging.Postgres, logging.Update, service_errors.RecordNotFound, nil)
// 		metrics.DbCall.WithLabelValues(reflect.TypeOf(*new(T)).String(), "Delete", "Failed").Inc()
// 		return err
// 	}

// 	metrics.DbCall.WithLabelValues(reflect.TypeOf(*new(T)).String(), "Delete", "Success").Inc()
// 	return nil
// }

// func (s *BaseService[T, Tc, Tu, Tr]) GetById(ctx context.Context, id int) (*Tr, error) {
// 	model, err := s.DB.GetModel(ctx, int32(id))
// 	if err != nil {
// 		metrics.DbCall.WithLabelValues(reflect.TypeOf(*new(T)).String(), "GetById", "Failed").Inc()
// 		return nil, err
// 	}

// 	metrics.DbCall.WithLabelValues(reflect.TypeOf(*new(T)).String(), "GetById", "Success").Inc()
// 	return common.TypeConverter[Tr](&model)
// }

// func (s *BaseService[T, Tc, Tu, Tr]) GetByFilter(ctx context.Context, req *dto.PaginationInputWithFilter) (*dto.PagedList[Tr], error) {
// 	query := getQuery[T](&req.DynamicFilter)
// 	sort := getSort[T](&req.DynamicFilter)

// 	// Execute the SQL query using SQLC
// 	models, err := s.DB.ListModels(ctx, sqlc.ListModelsParams{
// 		Query:  query,
// 		Sort:   sort,
// 		Offset: int32(req.GetOffset()),
// 		Limit:  int32(req.GetPageSize()),
// 	})
// 	if err != nil {
// 		metrics.DbCall.WithLabelValues(reflect.TypeOf(*new(T)).String(), "GetByFilter", "Failed").Inc()
// 		return nil, err
// 	}

// 	// Convert the result to the response type
// 	rItems, err := common.TypeConverter[[]Tr](&models)
// 	if err != nil {
// 		return nil, err
// 	}

// 	// Get the total count of rows
// 	totalRows, err := s.DB.CountModels(ctx, query)
// 	if err != nil {
// 		return nil, err
// 	}

// 	metrics.DbCall.WithLabelValues(reflect.TypeOf(*new(T)).String(), "GetByFilter", "Success").Inc()
// 	return NewPagedList(rItems, totalRows, req.PageNumber, int64(req.PageSize)), nil
// }

// func NewPagedList[T any](items *[]T, count int64, pageNumber int, pageSize int64) *dto.PagedList[T] {
// 	pl := &dto.PagedList[T]{
// 		PageNumber: pageNumber,
// 		TotalRows:  count,
// 		Items:      items,
// 	}
// 	pl.TotalPages = int(math.Ceil(float64(count) / float64(pageSize)))
// 	pl.HasNextPage = pl.PageNumber < pl.TotalPages
// 	pl.HasPreviousPage = pl.PageNumber > 1

// 	return pl
// }

// // Paginate
// func Paginate[T any, Tr any](pagination *dto.PaginationInputWithFilter, db *sqlc.Queries) (*dto.PagedList[Tr], error) {
// 	query := getQuery[T](&pagination.DynamicFilter)
// 	sort := getSort[T](&pagination.DynamicFilter)

// 	// Execute the SQL query using SQLC
// 	items, err := db.ListModels(context.Background(), sqlc.ListModelsParams{
// 		Query:  query,
// 		Sort:   sort,
// 		Offset: int32(pagination.GetOffset()),
// 		Limit:  int32(pagination.GetPageSize()),
// 	})
// 	if err != nil {
// 		return nil, err
// 	}

// 	// Convert the result to the response type
// 	rItems, err := common.TypeConverter[[]Tr](&items)
// 	if err != nil {
// 		return nil, err
// 	}

// 	// Get the total count of rows
// 	totalRows, err := db.CountModels(context.Background(), query)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return NewPagedList(rItems, totalRows, pagination.PageNumber, int64(pagination.PageSize)), nil
// }

// // getQuery
// func getQuery[T any](filter *dto.DynamicFilter) string {
// 	t := new(T)
// 	typeT := reflect.TypeOf(*t)
// 	query := make([]string, 0)
// 	query = append(query, "deleted_by is null")
// 	if filter.Filter != nil {
// 		for name, filter := range filter.Filter {
// 			fld, ok := typeT.FieldByName(name)
// 			if ok {
// 				fld.Name = common.ToSnakeCase(fld.Name)
// 				switch filter.Type {
// 				case "contains":
// 					query = append(query, fmt.Sprintf("%s ILike '%%%s%%'", fld.Name, filter.From))
// 				case "notContains":
// 					query = append(query, fmt.Sprintf("%s not ILike '%%%s%%'", fld.Name, filter.From))
// 				case "startsWith":
// 					query = append(query, fmt.Sprintf("%s ILike '%s%%'", fld.Name, filter.From))
// 				case "endsWith":
// 					query = append(query, fmt.Sprintf("%s ILike '%%%s'", fld.Name, filter.From))
// 				case "equals":
// 					query = append(query, fmt.Sprintf("%s = '%s'", fld.Name, filter.From))
// 				case "notEqual":
// 					query = append(query, fmt.Sprintf("%s != '%s'", fld.Name, filter.From))
// 				case "lessThan":
// 					query = append(query, fmt.Sprintf("%s < %s", fld.Name, filter.From))
// 				case "lessThanOrEqual":
// 					query = append(query, fmt.Sprintf("%s <= %s", fld.Name, filter.From))
// 				case "greaterThan":
// 					query = append(query, fmt.Sprintf("%s > %s", fld.Name, filter.From))
// 				case "greaterThanOrEqual":
// 					query = append(query, fmt.Sprintf("%s >= %s", fld.Name, filter.From))
// 				case "inRange":
// 					if fld.Type.Kind() == reflect.String {
// 						query = append(query, fmt.Sprintf("%s >= '%s'", fld.Name, filter.From))
// 						query = append(query, fmt.Sprintf("%s <= '%s'", fld.Name, filter.To))
// 					} else {
// 						query = append(query, fmt.Sprintf("%s >= %s", fld.Name, filter.From))
// 						query = append(query, fmt.Sprintf("%s <= %s", fld.Name, filter.To))
// 					}
// 				}
// 			}
// 		}
// 	}
// 	return strings.Join(query, " AND ")
// }

// // getSort
// func getSort[T any](filter *dto.DynamicFilter) string {
// 	t := new(T)
// 	typeT := reflect.TypeOf(*t)
// 	sort := make([]string, 0)
// 	if filter.Sort != nil {
// 		for _, tp := range *filter.Sort {
// 			fld, ok := typeT.FieldByName(tp.ColId)
// 			if ok && (tp.Sort == "asc" || tp.Sort == "desc") {
// 				fld.Name = common.ToSnakeCase(fld.Name)
// 				sort = append(sort, fmt.Sprintf("%s %s", fld.Name, tp.Sort))
// 			}
// 		}
// 	}
// 	return strings.Join(sort, ", ")
// }