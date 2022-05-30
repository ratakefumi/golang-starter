package repositories

import (
	"context"
	"database/sql"
	"fmt"
	usersmodel "golang-starter/src/modules/user/entities"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/nurcahyaari/sqlabst"
)

type RepositoryUsersCommand interface {
	InsertUsersList(ctx context.Context, usersList usersmodel.UsersList) (*InsertResult, error)
	InsertUsers(ctx context.Context, users *usersmodel.Users) (*InsertResult, error)
	UpdateUsersByFilter(ctx context.Context, users *usersmodel.Users, filter Filter, updatedFields ...UsersField) error
	UpdateUsers(ctx context.Context, users *usersmodel.Users, userid int32, updatedFields ...UsersField) error
	DeleteUsersList(ctx context.Context, filter Filter) error
	DeleteUsers(ctx context.Context, userid int32) error
}

type RepositoryUsersCommandImpl struct {
	db *sqlabst.SqlAbst
}

func (repo *RepositoryUsersCommandImpl) InsertUsersList(ctx context.Context, usersList usersmodel.UsersList) (*InsertResult, error) {
	command := `INSERT INTO users (photo,
	username,
	email,
	password,
	name,
	created_at,
	updated_at) VALUES
		`

	var (
		placeholders []string
		args         []interface{}
	)
	for _, users := range usersList {
		placeholders = append(placeholders, `(?,
	?,
	?,
	?,
	?,
	?,
	?)`)
		args = append(args,
			users.Photo,
			users.Username,
			users.Email,
			users.Password,
			users.Name,
			users.CreatedAt,
			users.UpdatedAt,
		)
	}
	command += strings.Join(placeholders, ",")

	sqlResult, err := repo.exec(ctx, command, args)
	if err != nil {
		return nil, err
	}

	return &InsertResult{Result: sqlResult}, nil
}

func (repo *RepositoryUsersCommandImpl) InsertUsers(ctx context.Context, users *usersmodel.Users) (*InsertResult, error) {
	return repo.InsertUsersList(ctx, usersmodel.UsersList{users})
}

func (repo *RepositoryUsersCommandImpl) UpdateUsersByFilter(ctx context.Context, users *usersmodel.Users, filter Filter, updatedFields ...UsersField) error {
	updatedFieldQuery, values := buildUpdateFieldsUsersQuery(updatedFields, users)
	command := fmt.Sprintf(`UPDATE users 
			SET %s 
		WHERE %s
		`, strings.Join(updatedFieldQuery, ","), filter.Query())
	values = append(values, filter.Values()...)
	_, err := repo.exec(ctx, command, values)
	return err
}

func (repo *RepositoryUsersCommandImpl) UpdateUsers(ctx context.Context, users *usersmodel.Users, userid int32, updatedFields ...UsersField) error {
	updatedFieldQuery, values := buildUpdateFieldsUsersQuery(updatedFields, users)
	command := fmt.Sprintf(`UPDATE users 
			SET %s 
		WHERE user_id = ?
		`, strings.Join(updatedFieldQuery, ","))
	values = append(values, userid)
	_, err := repo.exec(ctx, command, values)
	return err
}

func (repo *RepositoryUsersCommandImpl) DeleteUsersList(ctx context.Context, filter Filter) error {
	command := "DELETE FROM users WHERE " + filter.Query()
	_, err := repo.exec(ctx, command, filter.Values())
	return err
}

func (repo *RepositoryUsersCommandImpl) DeleteUsers(ctx context.Context, userid int32) error {
	command := "DELETE FROM users WHERE user_id = ?"
	_, err := repo.exec(ctx, command, []interface{}{userid})
	return err
}

func NewRepoUsersCommand(db *sqlabst.SqlAbst) RepositoryUsersCommand {
	return &RepositoryUsersCommandImpl{
		db: db,
	}
}

func (repo *RepositoryUsersCommandImpl) exec(ctx context.Context, command string, args []interface{}) (sql.Result, error) {
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

func buildUpdateFieldsUsersQuery(updatedFields UsersFieldList, users *usersmodel.Users) ([]string, []interface{}) {
	var (
		updatedFieldsQuery []string
		args               []interface{}
	)

	for _, field := range updatedFields {
		switch field {
		case "user_id":
			updatedFieldsQuery = append(updatedFieldsQuery, "user_id = ?")
			args = append(args, users.UserId)
		case "photo":
			updatedFieldsQuery = append(updatedFieldsQuery, "photo = ?")
			args = append(args, users.Photo)
		case "username":
			updatedFieldsQuery = append(updatedFieldsQuery, "username = ?")
			args = append(args, users.Username)
		case "email":
			updatedFieldsQuery = append(updatedFieldsQuery, "email = ?")
			args = append(args, users.Email)
		case "password":
			updatedFieldsQuery = append(updatedFieldsQuery, "password = ?")
			args = append(args, users.Password)
		case "name":
			updatedFieldsQuery = append(updatedFieldsQuery, "name = ?")
			args = append(args, users.Name)
		case "created_at":
			updatedFieldsQuery = append(updatedFieldsQuery, "created_at = ?")
			args = append(args, users.CreatedAt)
		case "updated_at":
			updatedFieldsQuery = append(updatedFieldsQuery, "updated_at = ?")
			args = append(args, users.UpdatedAt)
		}
	}

	return updatedFieldsQuery, args
}
