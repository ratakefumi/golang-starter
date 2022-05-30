package repositories

import (
	"context"
	"errors"
	"fmt"
	usersmodel "golang-starter/src/modules/user/entities"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/nurcahyaari/sqlabst"
)

type RepositoryUsersQuery interface {
	SelectUsers(fields ...UsersField) RepositoryUsersQuery
	ExcludeUsers(excludedFields ...UsersField) RepositoryUsersQuery
	FilterUsers(filter Filter) RepositoryUsersQuery
	PaginationUsers(pagination Pagination) RepositoryUsersQuery
	OrderByUsers(orderBy []Order) RepositoryUsersQuery
	GetUsersCount(ctx context.Context) (int, error)
	GetUsers(ctx context.Context) (*usersmodel.Users, error)
	GetUsersList(ctx context.Context) (usersmodel.UsersList, error)
}

type RepositoryUsersQueryImpl struct {
	db         *sqlabst.SqlAbst
	query      string
	filter     Filter
	orderBy    []Order
	pagination Pagination
	fields     UsersFieldList
}

func (repo *RepositoryUsersQueryImpl) SelectUsers(fields ...UsersField) RepositoryUsersQuery {
	return &RepositoryUsersQueryImpl{
		db:         repo.db,
		filter:     repo.filter,
		orderBy:    repo.orderBy,
		pagination: repo.pagination,
		fields:     fields,
	}
}

func (repo *RepositoryUsersQueryImpl) ExcludeUsers(excludedFields ...UsersField) RepositoryUsersQuery {
	selectedFieldsStr := excludeFields(UsersFieldList(excludedFields).toString(),
		UsersSelectFields{}.All().toString())

	var selectedFields []UsersField
	for _, sel := range selectedFieldsStr {
		selectedFields = append(selectedFields, UsersField(sel))
	}

	return &RepositoryUsersQueryImpl{
		db:         repo.db,
		filter:     repo.filter,
		orderBy:    repo.orderBy,
		pagination: repo.pagination,
		fields:     selectedFields,
	}
}

func (repo *RepositoryUsersQueryImpl) FilterUsers(filter Filter) RepositoryUsersQuery {
	return &RepositoryUsersQueryImpl{
		db:         repo.db,
		filter:     filter,
		orderBy:    repo.orderBy,
		pagination: repo.pagination,
		fields:     repo.fields,
	}
}

func (repo *RepositoryUsersQueryImpl) PaginationUsers(pagination Pagination) RepositoryUsersQuery {
	return &RepositoryUsersQueryImpl{
		db:         repo.db,
		filter:     repo.filter,
		orderBy:    repo.orderBy,
		pagination: pagination,
		fields:     repo.fields,
	}
}

func (repo *RepositoryUsersQueryImpl) OrderByUsers(orderBy []Order) RepositoryUsersQuery {
	return &RepositoryUsersQueryImpl{
		db:         repo.db,
		filter:     repo.filter,
		orderBy:    orderBy,
		pagination: repo.pagination,
		fields:     repo.fields,
	}
}

func (repo *RepositoryUsersQueryImpl) GetUsersList(ctx context.Context) (usersmodel.UsersList, error) {
	var (
		usersList usersmodel.UsersList
		values    []interface{}
	)

	if len(repo.fields) == 0 {
		repo.fields = UsersSelectFields{}.All()
	}

	query := fmt.Sprintf("SELECT %s FROM users", strings.Join(repo.fields.toString(), ","))
	if repo.filter != nil {
		query += " WHERE " + repo.filter.Query()
		values = append(values, repo.filter.Values()...)
	}

	if len(repo.orderBy) > 0 {
		var orderStr []string
		for _, order := range repo.orderBy {
			orderStr = append(orderStr, order.Value()+" "+order.Direction())
		}
		query += fmt.Sprintf(" ORDER BY %s", strings.Join(orderStr, ","))
	}

	if repo.pagination != nil {
		offset := (repo.pagination.GetPage() - 1) * repo.pagination.GetSize()
		query += fmt.Sprintf(" LIMIT %d OFFSET %d", repo.pagination.GetSize(), offset)
	}

	err := repo.db.SelectContext(ctx, &usersList, query, values...)
	if err != nil {
		return nil, err
	}
	return usersList, nil
}

func (repo *RepositoryUsersQueryImpl) GetUsersCount(ctx context.Context) (int, error) {
	var values []interface{}
	query := fmt.Sprintf("SELECT count(1) FROM users")
	if repo.filter != nil {
		query += " WHERE " + repo.filter.Query()
		values = append(values, repo.filter.Values()...)
	}

	var count int
	err := repo.db.QueryRowContext(ctx, query, values...).Scan(&count)
	return count, err
}

func (repo *RepositoryUsersQueryImpl) GetUsers(ctx context.Context) (*usersmodel.Users, error) {
	usersList, err := repo.GetUsersList(ctx)
	if err != nil {
		return nil, err
	}

	if len(usersList) == 0 {
		return nil, errors.New("users not found")
	}

	return usersList[0], nil
}

func NewRepoUsersQuery(db *sqlabst.SqlAbst) RepositoryUsersQuery {
	return &RepositoryUsersQueryImpl{
		db: db,
	}
}

type UsersField string
type UsersFieldList []UsersField

func (fieldList UsersFieldList) toString() []string {
	var fieldsStr []string
	for _, field := range fieldList {
		fieldsStr = append(fieldsStr, string(field))
	}
	return fieldsStr
}

type UsersSelectFields struct {
}

func (UsersSelectFields) UserId() UsersField {
	return UsersField("user_id")
}
func (UsersSelectFields) Photo() UsersField {
	return UsersField("photo")
}
func (UsersSelectFields) Username() UsersField {
	return UsersField("username")
}
func (UsersSelectFields) Email() UsersField {
	return UsersField("email")
}
func (UsersSelectFields) Password() UsersField {
	return UsersField("password")
}
func (UsersSelectFields) Name() UsersField {
	return UsersField("name")
}
func (UsersSelectFields) CreatedAt() UsersField {
	return UsersField("created_at")
}
func (UsersSelectFields) UpdatedAt() UsersField {
	return UsersField("updated_at")
}

func (UsersSelectFields) All() UsersFieldList {
	return []UsersField{
		UsersField("user_id"),
		UsersField("photo"),
		UsersField("username"),
		UsersField("email"),
		UsersField("password"),
		UsersField("name"),
		UsersField("created_at"),
		UsersField("updated_at"),
	}
}

func NewUsersSelectFields() UsersSelectFields {
	return UsersSelectFields{}
}

type UsersFilter struct {
	operator string
	query    []string
	values   []interface{}
}

func NewUsersFilter(operator string) UsersFilter {
	if operator == "" {
		operator = "AND"
	}
	return UsersFilter{
		operator: operator,
	}
}

func (f UsersFilter) SetFilterByUserId(value interface{}, operator string) UsersFilter {
	query := "user_id " + operator + " (?)"
	var values []interface{}
	if value == nil {
		query = "user_id " + operator
	} else {
		switch strings.ToUpper(operator) {
		case "IN", "NOT IN":
			query, values, _ = sqlx.In(query, value)
		default:
			values = append(values, value)
		}
	}
	return UsersFilter{
		operator: f.operator,
		query:    append(f.query, query),
		values:   append(f.values, values...),
	}
}
func (f UsersFilter) SetFilterByPhoto(value interface{}, operator string) UsersFilter {
	query := "photo " + operator + " (?)"
	var values []interface{}
	if value == nil {
		query = "photo " + operator
	} else {
		switch strings.ToUpper(operator) {
		case "IN", "NOT IN":
			query, values, _ = sqlx.In(query, value)
		default:
			values = append(values, value)
		}
	}
	return UsersFilter{
		operator: f.operator,
		query:    append(f.query, query),
		values:   append(f.values, values...),
	}
}
func (f UsersFilter) SetFilterByUsername(value interface{}, operator string) UsersFilter {
	query := "username " + operator + " (?)"
	var values []interface{}
	if value == nil {
		query = "username " + operator
	} else {
		switch strings.ToUpper(operator) {
		case "IN", "NOT IN":
			query, values, _ = sqlx.In(query, value)
		default:
			values = append(values, value)
		}
	}
	return UsersFilter{
		operator: f.operator,
		query:    append(f.query, query),
		values:   append(f.values, values...),
	}
}
func (f UsersFilter) SetFilterByEmail(value interface{}, operator string) UsersFilter {
	query := "email " + operator + " (?)"
	var values []interface{}
	if value == nil {
		query = "email " + operator
	} else {
		switch strings.ToUpper(operator) {
		case "IN", "NOT IN":
			query, values, _ = sqlx.In(query, value)
		default:
			values = append(values, value)
		}
	}
	return UsersFilter{
		operator: f.operator,
		query:    append(f.query, query),
		values:   append(f.values, values...),
	}
}
func (f UsersFilter) SetFilterByPassword(value interface{}, operator string) UsersFilter {
	query := "password " + operator + " (?)"
	var values []interface{}
	if value == nil {
		query = "password " + operator
	} else {
		switch strings.ToUpper(operator) {
		case "IN", "NOT IN":
			query, values, _ = sqlx.In(query, value)
		default:
			values = append(values, value)
		}
	}
	return UsersFilter{
		operator: f.operator,
		query:    append(f.query, query),
		values:   append(f.values, values...),
	}
}
func (f UsersFilter) SetFilterByName(value interface{}, operator string) UsersFilter {
	query := "name " + operator + " (?)"
	var values []interface{}
	if value == nil {
		query = "name " + operator
	} else {
		switch strings.ToUpper(operator) {
		case "IN", "NOT IN":
			query, values, _ = sqlx.In(query, value)
		default:
			values = append(values, value)
		}
	}
	return UsersFilter{
		operator: f.operator,
		query:    append(f.query, query),
		values:   append(f.values, values...),
	}
}
func (f UsersFilter) SetFilterByCreatedAt(value interface{}, operator string) UsersFilter {
	query := "created_at " + operator + " (?)"
	var values []interface{}
	if value == nil {
		query = "created_at " + operator
	} else {
		switch strings.ToUpper(operator) {
		case "IN", "NOT IN":
			query, values, _ = sqlx.In(query, value)
		default:
			values = append(values, value)
		}
	}
	return UsersFilter{
		operator: f.operator,
		query:    append(f.query, query),
		values:   append(f.values, values...),
	}
}
func (f UsersFilter) SetFilterByUpdatedAt(value interface{}, operator string) UsersFilter {
	query := "updated_at " + operator + " (?)"
	var values []interface{}
	if value == nil {
		query = "updated_at " + operator
	} else {
		switch strings.ToUpper(operator) {
		case "IN", "NOT IN":
			query, values, _ = sqlx.In(query, value)
		default:
			values = append(values, value)
		}
	}
	return UsersFilter{
		operator: f.operator,
		query:    append(f.query, query),
		values:   append(f.values, values...),
	}
}

func (f UsersFilter) Query() string {
	return strings.Join(f.query, " "+f.operator+" ")
}

func (f UsersFilter) Values() []interface{} {
	return f.values
}

type UsersUserIdOrder struct {
	direction string
}

func (o UsersUserIdOrder) SetDirection(direction string) UsersUserIdOrder {
	return UsersUserIdOrder{
		direction: direction,
	}
}
func (o UsersUserIdOrder) Value() string {
	return "user_id"
}
func (o UsersUserIdOrder) Direction() string {
	return o.direction
}
func NewUsersUserIdOrder() UsersUserIdOrder {
	return UsersUserIdOrder{}
}

type UsersPhotoOrder struct {
	direction string
}

func (o UsersPhotoOrder) SetDirection(direction string) UsersPhotoOrder {
	return UsersPhotoOrder{
		direction: direction,
	}
}
func (o UsersPhotoOrder) Value() string {
	return "photo"
}
func (o UsersPhotoOrder) Direction() string {
	return o.direction
}
func NewUsersPhotoOrder() UsersPhotoOrder {
	return UsersPhotoOrder{}
}

type UsersUsernameOrder struct {
	direction string
}

func (o UsersUsernameOrder) SetDirection(direction string) UsersUsernameOrder {
	return UsersUsernameOrder{
		direction: direction,
	}
}
func (o UsersUsernameOrder) Value() string {
	return "username"
}
func (o UsersUsernameOrder) Direction() string {
	return o.direction
}
func NewUsersUsernameOrder() UsersUsernameOrder {
	return UsersUsernameOrder{}
}

type UsersEmailOrder struct {
	direction string
}

func (o UsersEmailOrder) SetDirection(direction string) UsersEmailOrder {
	return UsersEmailOrder{
		direction: direction,
	}
}
func (o UsersEmailOrder) Value() string {
	return "email"
}
func (o UsersEmailOrder) Direction() string {
	return o.direction
}
func NewUsersEmailOrder() UsersEmailOrder {
	return UsersEmailOrder{}
}

type UsersPasswordOrder struct {
	direction string
}

func (o UsersPasswordOrder) SetDirection(direction string) UsersPasswordOrder {
	return UsersPasswordOrder{
		direction: direction,
	}
}
func (o UsersPasswordOrder) Value() string {
	return "password"
}
func (o UsersPasswordOrder) Direction() string {
	return o.direction
}
func NewUsersPasswordOrder() UsersPasswordOrder {
	return UsersPasswordOrder{}
}

type UsersNameOrder struct {
	direction string
}

func (o UsersNameOrder) SetDirection(direction string) UsersNameOrder {
	return UsersNameOrder{
		direction: direction,
	}
}
func (o UsersNameOrder) Value() string {
	return "name"
}
func (o UsersNameOrder) Direction() string {
	return o.direction
}
func NewUsersNameOrder() UsersNameOrder {
	return UsersNameOrder{}
}

type UsersCreatedAtOrder struct {
	direction string
}

func (o UsersCreatedAtOrder) SetDirection(direction string) UsersCreatedAtOrder {
	return UsersCreatedAtOrder{
		direction: direction,
	}
}
func (o UsersCreatedAtOrder) Value() string {
	return "created_at"
}
func (o UsersCreatedAtOrder) Direction() string {
	return o.direction
}
func NewUsersCreatedAtOrder() UsersCreatedAtOrder {
	return UsersCreatedAtOrder{}
}

type UsersUpdatedAtOrder struct {
	direction string
}

func (o UsersUpdatedAtOrder) SetDirection(direction string) UsersUpdatedAtOrder {
	return UsersUpdatedAtOrder{
		direction: direction,
	}
}
func (o UsersUpdatedAtOrder) Value() string {
	return "updated_at"
}
func (o UsersUpdatedAtOrder) Direction() string {
	return o.direction
}
func NewUsersUpdatedAtOrder() UsersUpdatedAtOrder {
	return UsersUpdatedAtOrder{}
}
