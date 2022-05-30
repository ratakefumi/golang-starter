package repositories

import (
	"context"
	"errors"
	"fmt"
	productsmodel "golang-starter/src/modules/product/entities"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/nurcahyaari/sqlabst"
)

type RepositoryProductsQuery interface {
	SelectProducts(fields ...ProductsField) RepositoryProductsQuery
	ExcludeProducts(excludedFields ...ProductsField) RepositoryProductsQuery
	FilterProducts(filter Filter) RepositoryProductsQuery
	PaginationProducts(pagination Pagination) RepositoryProductsQuery
	OrderByProducts(orderBy []Order) RepositoryProductsQuery
	GetProductsCount(ctx context.Context) (int, error)
	GetProducts(ctx context.Context) (*productsmodel.Products, error)
	GetProductsList(ctx context.Context) (productsmodel.ProductsList, error)
}

type RepositoryProductsQueryImpl struct {
	db         *sqlabst.SqlAbst
	query      string
	filter     Filter
	orderBy    []Order
	pagination Pagination
	fields     ProductsFieldList
}

func (repo *RepositoryProductsQueryImpl) SelectProducts(fields ...ProductsField) RepositoryProductsQuery {
	return &RepositoryProductsQueryImpl{
		db:         repo.db,
		filter:     repo.filter,
		orderBy:    repo.orderBy,
		pagination: repo.pagination,
		fields:     fields,
	}
}

func (repo *RepositoryProductsQueryImpl) ExcludeProducts(excludedFields ...ProductsField) RepositoryProductsQuery {
	selectedFieldsStr := excludeFields(ProductsFieldList(excludedFields).toString(),
		ProductsSelectFields{}.All().toString())

	var selectedFields []ProductsField
	for _, sel := range selectedFieldsStr {
		selectedFields = append(selectedFields, ProductsField(sel))
	}

	return &RepositoryProductsQueryImpl{
		db:         repo.db,
		filter:     repo.filter,
		orderBy:    repo.orderBy,
		pagination: repo.pagination,
		fields:     selectedFields,
	}
}

func (repo *RepositoryProductsQueryImpl) FilterProducts(filter Filter) RepositoryProductsQuery {
	return &RepositoryProductsQueryImpl{
		db:         repo.db,
		filter:     filter,
		orderBy:    repo.orderBy,
		pagination: repo.pagination,
		fields:     repo.fields,
	}
}

func (repo *RepositoryProductsQueryImpl) PaginationProducts(pagination Pagination) RepositoryProductsQuery {
	return &RepositoryProductsQueryImpl{
		db:         repo.db,
		filter:     repo.filter,
		orderBy:    repo.orderBy,
		pagination: pagination,
		fields:     repo.fields,
	}
}

func (repo *RepositoryProductsQueryImpl) OrderByProducts(orderBy []Order) RepositoryProductsQuery {
	return &RepositoryProductsQueryImpl{
		db:         repo.db,
		filter:     repo.filter,
		orderBy:    orderBy,
		pagination: repo.pagination,
		fields:     repo.fields,
	}
}

func (repo *RepositoryProductsQueryImpl) GetProductsList(ctx context.Context) (productsmodel.ProductsList, error) {
	var (
		productsList productsmodel.ProductsList
		values       []interface{}
	)

	if len(repo.fields) == 0 {
		repo.fields = ProductsSelectFields{}.All()
	}

	query := fmt.Sprintf("SELECT %s FROM products", strings.Join(repo.fields.toString(), ","))
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

	err := repo.db.SelectContext(ctx, &productsList, query, values...)
	if err != nil {
		return nil, err
	}
	return productsList, nil
}

func (repo *RepositoryProductsQueryImpl) GetProductsCount(ctx context.Context) (int, error) {
	var values []interface{}
	query := fmt.Sprintf("SELECT count(1) FROM products")
	if repo.filter != nil {
		query += " WHERE " + repo.filter.Query()
		values = append(values, repo.filter.Values()...)
	}

	var count int
	err := repo.db.QueryRowContext(ctx, query, values...).Scan(&count)
	return count, err
}

func (repo *RepositoryProductsQueryImpl) GetProducts(ctx context.Context) (*productsmodel.Products, error) {
	productsList, err := repo.GetProductsList(ctx)
	if err != nil {
		return nil, err
	}

	if len(productsList) == 0 {
		return nil, errors.New("products not found")
	}

	return productsList[0], nil
}

func NewRepoProductsQuery(db *sqlabst.SqlAbst) RepositoryProductsQuery {
	return &RepositoryProductsQueryImpl{
		db: db,
	}
}

type ProductsField string
type ProductsFieldList []ProductsField

func (fieldList ProductsFieldList) toString() []string {
	var fieldsStr []string
	for _, field := range fieldList {
		fieldsStr = append(fieldsStr, string(field))
	}
	return fieldsStr
}

type ProductsSelectFields struct {
}

func (ProductsSelectFields) ProductId() ProductsField {
	return ProductsField("product_id")
}
func (ProductsSelectFields) ProductCategoryFkid() ProductsField {
	return ProductsField("product_category_fkid")
}
func (ProductsSelectFields) AdminFkid() ProductsField {
	return ProductsField("admin_fkid")
}
func (ProductsSelectFields) Name() ProductsField {
	return ProductsField("name")
}
func (ProductsSelectFields) Price() ProductsField {
	return ProductsField("price")
}
func (ProductsSelectFields) Description() ProductsField {
	return ProductsField("description")
}
func (ProductsSelectFields) Qty() ProductsField {
	return ProductsField("qty")
}
func (ProductsSelectFields) Image() ProductsField {
	return ProductsField("image")
}
func (ProductsSelectFields) Label() ProductsField {
	return ProductsField("label")
}

func (ProductsSelectFields) All() ProductsFieldList {
	return []ProductsField{
		ProductsField("product_id"),
		ProductsField("product_category_fkid"),
		ProductsField("admin_fkid"),
		ProductsField("name"),
		ProductsField("price"),
		ProductsField("description"),
		ProductsField("qty"),
		ProductsField("image"),
		ProductsField("label"),
	}
}

func NewProductsSelectFields() ProductsSelectFields {
	return ProductsSelectFields{}
}

type ProductsFilter struct {
	operator string
	query    []string
	values   []interface{}
}

func NewProductsFilter(operator string) ProductsFilter {
	if operator == "" {
		operator = "AND"
	}
	return ProductsFilter{
		operator: operator,
	}
}

func (f ProductsFilter) SetFilterByProductId(value interface{}, operator string) ProductsFilter {
	query := "product_id " + operator + " (?)"
	var values []interface{}
	if value == nil {
		query = "product_id " + operator
	} else {
		switch strings.ToUpper(operator) {
		case "IN", "NOT IN":
			query, values, _ = sqlx.In(query, value)
		default:
			values = append(values, value)
		}
	}
	return ProductsFilter{
		operator: f.operator,
		query:    append(f.query, query),
		values:   append(f.values, values...),
	}
}
func (f ProductsFilter) SetFilterByProductCategoryFkid(value interface{}, operator string) ProductsFilter {
	query := "product_category_fkid " + operator + " (?)"
	var values []interface{}
	if value == nil {
		query = "product_category_fkid " + operator
	} else {
		switch strings.ToUpper(operator) {
		case "IN", "NOT IN":
			query, values, _ = sqlx.In(query, value)
		default:
			values = append(values, value)
		}
	}
	return ProductsFilter{
		operator: f.operator,
		query:    append(f.query, query),
		values:   append(f.values, values...),
	}
}
func (f ProductsFilter) SetFilterByAdminFkid(value interface{}, operator string) ProductsFilter {
	query := "admin_fkid " + operator + " (?)"
	var values []interface{}
	if value == nil {
		query = "admin_fkid " + operator
	} else {
		switch strings.ToUpper(operator) {
		case "IN", "NOT IN":
			query, values, _ = sqlx.In(query, value)
		default:
			values = append(values, value)
		}
	}
	return ProductsFilter{
		operator: f.operator,
		query:    append(f.query, query),
		values:   append(f.values, values...),
	}
}
func (f ProductsFilter) SetFilterByName(value interface{}, operator string) ProductsFilter {
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
	return ProductsFilter{
		operator: f.operator,
		query:    append(f.query, query),
		values:   append(f.values, values...),
	}
}
func (f ProductsFilter) SetFilterByPrice(value interface{}, operator string) ProductsFilter {
	query := "price " + operator + " (?)"
	var values []interface{}
	if value == nil {
		query = "price " + operator
	} else {
		switch strings.ToUpper(operator) {
		case "IN", "NOT IN":
			query, values, _ = sqlx.In(query, value)
		default:
			values = append(values, value)
		}
	}
	return ProductsFilter{
		operator: f.operator,
		query:    append(f.query, query),
		values:   append(f.values, values...),
	}
}
func (f ProductsFilter) SetFilterByDescription(value interface{}, operator string) ProductsFilter {
	query := "description " + operator + " (?)"
	var values []interface{}
	if value == nil {
		query = "description " + operator
	} else {
		switch strings.ToUpper(operator) {
		case "IN", "NOT IN":
			query, values, _ = sqlx.In(query, value)
		default:
			values = append(values, value)
		}
	}
	return ProductsFilter{
		operator: f.operator,
		query:    append(f.query, query),
		values:   append(f.values, values...),
	}
}
func (f ProductsFilter) SetFilterByQty(value interface{}, operator string) ProductsFilter {
	query := "qty " + operator + " (?)"
	var values []interface{}
	if value == nil {
		query = "qty " + operator
	} else {
		switch strings.ToUpper(operator) {
		case "IN", "NOT IN":
			query, values, _ = sqlx.In(query, value)
		default:
			values = append(values, value)
		}
	}
	return ProductsFilter{
		operator: f.operator,
		query:    append(f.query, query),
		values:   append(f.values, values...),
	}
}
func (f ProductsFilter) SetFilterByImage(value interface{}, operator string) ProductsFilter {
	query := "image " + operator + " (?)"
	var values []interface{}
	if value == nil {
		query = "image " + operator
	} else {
		switch strings.ToUpper(operator) {
		case "IN", "NOT IN":
			query, values, _ = sqlx.In(query, value)
		default:
			values = append(values, value)
		}
	}
	return ProductsFilter{
		operator: f.operator,
		query:    append(f.query, query),
		values:   append(f.values, values...),
	}
}
func (f ProductsFilter) SetFilterByLabel(value interface{}, operator string) ProductsFilter {
	query := "label " + operator + " (?)"
	var values []interface{}
	if value == nil {
		query = "label " + operator
	} else {
		switch strings.ToUpper(operator) {
		case "IN", "NOT IN":
			query, values, _ = sqlx.In(query, value)
		default:
			values = append(values, value)
		}
	}
	return ProductsFilter{
		operator: f.operator,
		query:    append(f.query, query),
		values:   append(f.values, values...),
	}
}

func (f ProductsFilter) Query() string {
	return strings.Join(f.query, " "+f.operator+" ")
}

func (f ProductsFilter) Values() []interface{} {
	return f.values
}

type ProductsProductIdOrder struct {
	direction string
}

func (o ProductsProductIdOrder) SetDirection(direction string) ProductsProductIdOrder {
	return ProductsProductIdOrder{
		direction: direction,
	}
}
func (o ProductsProductIdOrder) Value() string {
	return "product_id"
}
func (o ProductsProductIdOrder) Direction() string {
	return o.direction
}
func NewProductsProductIdOrder() ProductsProductIdOrder {
	return ProductsProductIdOrder{}
}

type ProductsProductCategoryFkidOrder struct {
	direction string
}

func (o ProductsProductCategoryFkidOrder) SetDirection(direction string) ProductsProductCategoryFkidOrder {
	return ProductsProductCategoryFkidOrder{
		direction: direction,
	}
}
func (o ProductsProductCategoryFkidOrder) Value() string {
	return "product_category_fkid"
}
func (o ProductsProductCategoryFkidOrder) Direction() string {
	return o.direction
}
func NewProductsProductCategoryFkidOrder() ProductsProductCategoryFkidOrder {
	return ProductsProductCategoryFkidOrder{}
}

type ProductsAdminFkidOrder struct {
	direction string
}

func (o ProductsAdminFkidOrder) SetDirection(direction string) ProductsAdminFkidOrder {
	return ProductsAdminFkidOrder{
		direction: direction,
	}
}
func (o ProductsAdminFkidOrder) Value() string {
	return "admin_fkid"
}
func (o ProductsAdminFkidOrder) Direction() string {
	return o.direction
}
func NewProductsAdminFkidOrder() ProductsAdminFkidOrder {
	return ProductsAdminFkidOrder{}
}

type ProductsNameOrder struct {
	direction string
}

func (o ProductsNameOrder) SetDirection(direction string) ProductsNameOrder {
	return ProductsNameOrder{
		direction: direction,
	}
}
func (o ProductsNameOrder) Value() string {
	return "name"
}
func (o ProductsNameOrder) Direction() string {
	return o.direction
}
func NewProductsNameOrder() ProductsNameOrder {
	return ProductsNameOrder{}
}

type ProductsPriceOrder struct {
	direction string
}

func (o ProductsPriceOrder) SetDirection(direction string) ProductsPriceOrder {
	return ProductsPriceOrder{
		direction: direction,
	}
}
func (o ProductsPriceOrder) Value() string {
	return "price"
}
func (o ProductsPriceOrder) Direction() string {
	return o.direction
}
func NewProductsPriceOrder() ProductsPriceOrder {
	return ProductsPriceOrder{}
}

type ProductsDescriptionOrder struct {
	direction string
}

func (o ProductsDescriptionOrder) SetDirection(direction string) ProductsDescriptionOrder {
	return ProductsDescriptionOrder{
		direction: direction,
	}
}
func (o ProductsDescriptionOrder) Value() string {
	return "description"
}
func (o ProductsDescriptionOrder) Direction() string {
	return o.direction
}
func NewProductsDescriptionOrder() ProductsDescriptionOrder {
	return ProductsDescriptionOrder{}
}

type ProductsQtyOrder struct {
	direction string
}

func (o ProductsQtyOrder) SetDirection(direction string) ProductsQtyOrder {
	return ProductsQtyOrder{
		direction: direction,
	}
}
func (o ProductsQtyOrder) Value() string {
	return "qty"
}
func (o ProductsQtyOrder) Direction() string {
	return o.direction
}
func NewProductsQtyOrder() ProductsQtyOrder {
	return ProductsQtyOrder{}
}

type ProductsImageOrder struct {
	direction string
}

func (o ProductsImageOrder) SetDirection(direction string) ProductsImageOrder {
	return ProductsImageOrder{
		direction: direction,
	}
}
func (o ProductsImageOrder) Value() string {
	return "image"
}
func (o ProductsImageOrder) Direction() string {
	return o.direction
}
func NewProductsImageOrder() ProductsImageOrder {
	return ProductsImageOrder{}
}

type ProductsLabelOrder struct {
	direction string
}

func (o ProductsLabelOrder) SetDirection(direction string) ProductsLabelOrder {
	return ProductsLabelOrder{
		direction: direction,
	}
}
func (o ProductsLabelOrder) Value() string {
	return "label"
}
func (o ProductsLabelOrder) Direction() string {
	return o.direction
}
func NewProductsLabelOrder() ProductsLabelOrder {
	return ProductsLabelOrder{}
}
