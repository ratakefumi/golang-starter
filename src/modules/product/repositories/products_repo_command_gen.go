package repositories

import (
	"context"
	"database/sql"
	"fmt"
	productsmodel "golang-starter/src/modules/product/entities"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/nurcahyaari/sqlabst"
)

type RepositoryProductsCommand interface {
	InsertProductsList(ctx context.Context, productsList productsmodel.ProductsList) (*InsertResult, error)
	InsertProducts(ctx context.Context, products *productsmodel.Products) (*InsertResult, error)
	UpdateProductsByFilter(ctx context.Context, products *productsmodel.Products, filter Filter, updatedFields ...ProductsField) error
	UpdateProducts(ctx context.Context, products *productsmodel.Products, productid int32, updatedFields ...ProductsField) error
	DeleteProductsList(ctx context.Context, filter Filter) error
	DeleteProducts(ctx context.Context, productid int32) error
}

type RepositoryProductsCommandImpl struct {
	db *sqlabst.SqlAbst
}

func (repo *RepositoryProductsCommandImpl) InsertProductsList(ctx context.Context, productsList productsmodel.ProductsList) (*InsertResult, error) {
	command := `INSERT INTO products (product_category_fkid,
	admin_fkid,
	name,
	price,
	description,
	qty,
	image,
	label) VALUES
		`

	var (
		placeholders []string
		args         []interface{}
	)
	for _, products := range productsList {
		placeholders = append(placeholders, `(?,
	?,
	?,
	?,
	?,
	?,
	?,
	?)`)
		args = append(args,
			products.ProductCategoryFkid,
			products.AdminFkid,
			products.Name,
			products.Price,
			products.Description,
			products.Qty,
			products.Image,
			products.Label,
		)
	}
	command += strings.Join(placeholders, ",")

	sqlResult, err := repo.exec(ctx, command, args)
	if err != nil {
		return nil, err
	}

	return &InsertResult{Result: sqlResult}, nil
}

func (repo *RepositoryProductsCommandImpl) InsertProducts(ctx context.Context, products *productsmodel.Products) (*InsertResult, error) {
	return repo.InsertProductsList(ctx, productsmodel.ProductsList{products})
}

func (repo *RepositoryProductsCommandImpl) UpdateProductsByFilter(ctx context.Context, products *productsmodel.Products, filter Filter, updatedFields ...ProductsField) error {
	updatedFieldQuery, values := buildUpdateFieldsProductsQuery(updatedFields, products)
	command := fmt.Sprintf(`UPDATE products 
			SET %s 
		WHERE %s
		`, strings.Join(updatedFieldQuery, ","), filter.Query())
	values = append(values, filter.Values()...)
	_, err := repo.exec(ctx, command, values)
	return err
}

func (repo *RepositoryProductsCommandImpl) UpdateProducts(ctx context.Context, products *productsmodel.Products, productid int32, updatedFields ...ProductsField) error {
	updatedFieldQuery, values := buildUpdateFieldsProductsQuery(updatedFields, products)
	command := fmt.Sprintf(`UPDATE products 
			SET %s 
		WHERE product_id = ?
		`, strings.Join(updatedFieldQuery, ","))
	values = append(values, productid)
	_, err := repo.exec(ctx, command, values)
	return err
}

func (repo *RepositoryProductsCommandImpl) DeleteProductsList(ctx context.Context, filter Filter) error {
	command := "DELETE FROM products WHERE " + filter.Query()
	_, err := repo.exec(ctx, command, filter.Values())
	return err
}

func (repo *RepositoryProductsCommandImpl) DeleteProducts(ctx context.Context, productid int32) error {
	command := "DELETE FROM products WHERE product_id = ?"
	_, err := repo.exec(ctx, command, []interface{}{productid})
	return err
}

func NewRepoProductsCommand(db *sqlabst.SqlAbst) RepositoryProductsCommand {
	return &RepositoryProductsCommandImpl{
		db: db,
	}
}

func (repo *RepositoryProductsCommandImpl) exec(ctx context.Context, command string, args []interface{}) (sql.Result, error) {
	var (
		stmt *sqlx.Stmt
		err  error
	)
	stmt, err = repo.db.PreparexContext(ctx, command)

	if err != nil {
		return nil, err
	}

	return stmt.ExecContext(ctx, args...)
}

func buildUpdateFieldsProductsQuery(updatedFields ProductsFieldList, products *productsmodel.Products) ([]string, []interface{}) {
	var (
		updatedFieldsQuery []string
		args               []interface{}
	)

	for _, field := range updatedFields {
		switch field {
		case "product_id":
			updatedFieldsQuery = append(updatedFieldsQuery, "product_id = ?")
			args = append(args, products.ProductId)
		case "product_category_fkid":
			updatedFieldsQuery = append(updatedFieldsQuery, "product_category_fkid = ?")
			args = append(args, products.ProductCategoryFkid)
		case "admin_fkid":
			updatedFieldsQuery = append(updatedFieldsQuery, "admin_fkid = ?")
			args = append(args, products.AdminFkid)
		case "name":
			updatedFieldsQuery = append(updatedFieldsQuery, "name = ?")
			args = append(args, products.Name)
		case "price":
			updatedFieldsQuery = append(updatedFieldsQuery, "price = ?")
			args = append(args, products.Price)
		case "description":
			updatedFieldsQuery = append(updatedFieldsQuery, "description = ?")
			args = append(args, products.Description)
		case "qty":
			updatedFieldsQuery = append(updatedFieldsQuery, "qty = ?")
			args = append(args, products.Qty)
		case "image":
			updatedFieldsQuery = append(updatedFieldsQuery, "image = ?")
			args = append(args, products.Image)
		case "label":
			updatedFieldsQuery = append(updatedFieldsQuery, "label = ?")
			args = append(args, products.Label)
		}
	}

	return updatedFieldsQuery, args
}
