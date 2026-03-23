package query

import (
	"context"

	"com.raunlo.checklist/internal/repository/dbo"
	"github.com/jackc/pgx/v5"
	"github.com/raunlo/pgx-with-automapper/pool"
)

// SaveTemplateQueryFunction saves a template and its items
type SaveTemplateQueryFunction struct {
	template dbo.TemplateDBO
	items    []dbo.TemplateItemDBO
}

func (q *SaveTemplateQueryFunction) GetTransactionalQueryFunction() func(tx pool.TransactionWrapper) (dbo.TemplateDBO, error) {
	return func(tx pool.TransactionWrapper) (dbo.TemplateDBO, error) {
		// Insert template
		err := tx.QueryRow(context.Background(),
			`INSERT INTO TEMPLATE(USER_ID, NAME, DESCRIPTION, CREATED_AT, UPDATED_AT)
			 VALUES(@userId, @name, @description, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
			 RETURNING ID`,
			pgx.NamedArgs{
				"userId":      q.template.UserID,
				"name":        q.template.Name,
				"description": q.template.Description,
			}).Scan(&q.template.ID)
		if err != nil {
			return dbo.TemplateDBO{}, err
		}

		// Insert template items
		for _, item := range q.items {
			item.TemplateID = q.template.ID
			err := tx.QueryRow(context.Background(),
				`INSERT INTO TEMPLATE_ITEM(TEMPLATE_ID, NAME, POSITION, CREATED_AT, UPDATED_AT)
				 VALUES(@templateId, @name, @position, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
				 RETURNING ID`,
				pgx.NamedArgs{
					"templateId": item.TemplateID,
					"name":       item.Name,
					"position":   item.Position,
				}).Scan(&item.ID)
			if err != nil {
				return dbo.TemplateDBO{}, err
			}
		}

		return q.template, nil
	}
}

func NewSaveTemplateQueryFunction(template dbo.TemplateDBO, items []dbo.TemplateItemDBO) *SaveTemplateQueryFunction {
	return &SaveTemplateQueryFunction{
		template: template,
		items:    items,
	}
}

// FindTemplateByIdQueryFunction finds a template by ID
type FindTemplateByIdQueryFunction struct {
	templateId uint64
}

func (q *FindTemplateByIdQueryFunction) GetTransactionalQueryFunction() func(tx pool.TransactionWrapper) (dbo.TemplateDBO, error) {
	return func(tx pool.TransactionWrapper) (dbo.TemplateDBO, error) {
		var template dbo.TemplateDBO
		err := tx.QueryRow(context.Background(),
			`SELECT ID, USER_ID, NAME, DESCRIPTION, CREATED_AT, UPDATED_AT
			 FROM TEMPLATE WHERE ID = @templateId`,
			pgx.NamedArgs{"templateId": q.templateId}).Scan(
			&template.ID, &template.UserID, &template.Name, &template.Description,
			&template.CreatedAt, &template.UpdatedAt)
		if err != nil {
			return dbo.TemplateDBO{}, err
		}
		return template, nil
	}
}

func NewFindTemplateByIdQueryFunction(templateId uint64) *FindTemplateByIdQueryFunction {
	return &FindTemplateByIdQueryFunction{templateId: templateId}
}

// FindTemplateItemsByTemplateIdQueryFunction finds all items for a template
type FindTemplateItemsByTemplateIdQueryFunction struct {
	templateId uint64
}

func (q *FindTemplateItemsByTemplateIdQueryFunction) GetTransactionalQueryFunction() func(tx pool.TransactionWrapper) ([]dbo.TemplateItemDBO, error) {
	return func(tx pool.TransactionWrapper) ([]dbo.TemplateItemDBO, error) {
		rows, err := tx.Query(context.Background(),
			`SELECT ID, TEMPLATE_ID, NAME, POSITION, CREATED_AT, UPDATED_AT
			 FROM TEMPLATE_ITEM WHERE TEMPLATE_ID = @templateId ORDER BY POSITION ASC`,
			pgx.NamedArgs{"templateId": q.templateId})
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		var items []dbo.TemplateItemDBO
		for rows.Next() {
			var item dbo.TemplateItemDBO
			err := rows.Scan(&item.ID, &item.TemplateID, &item.Name, &item.Position, &item.CreatedAt, &item.UpdatedAt)
			if err != nil {
				return nil, err
			}
			items = append(items, item)
		}
		return items, rows.Err()
	}
}

func NewFindTemplateItemsByTemplateIdQueryFunction(templateId uint64) *FindTemplateItemsByTemplateIdQueryFunction {
	return &FindTemplateItemsByTemplateIdQueryFunction{templateId: templateId}
}

// FindAllTemplatesByUserIdQueryFunction finds all templates for a user
type FindAllTemplatesByUserIdQueryFunction struct {
	userId string
}

func (q *FindAllTemplatesByUserIdQueryFunction) GetTransactionalQueryFunction() func(tx pool.TransactionWrapper) ([]dbo.TemplateDBO, error) {
	return func(tx pool.TransactionWrapper) ([]dbo.TemplateDBO, error) {
		rows, err := tx.Query(context.Background(),
			`SELECT ID, USER_ID, NAME, DESCRIPTION, CREATED_AT, UPDATED_AT
			 FROM TEMPLATE WHERE USER_ID = @userId ORDER BY UPDATED_AT DESC`,
			pgx.NamedArgs{"userId": q.userId})
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		var templates []dbo.TemplateDBO
		for rows.Next() {
			var template dbo.TemplateDBO
			err := rows.Scan(&template.ID, &template.UserID, &template.Name, &template.Description, &template.CreatedAt, &template.UpdatedAt)
			if err != nil {
				return nil, err
			}
			templates = append(templates, template)
		}
		return templates, rows.Err()
	}
}

func NewFindAllTemplatesByUserIdQueryFunction(userId string) *FindAllTemplatesByUserIdQueryFunction {
	return &FindAllTemplatesByUserIdQueryFunction{userId: userId}
}

// UpdateTemplateQueryFunction updates a template
type UpdateTemplateQueryFunction struct {
	template dbo.TemplateDBO
}

func (q *UpdateTemplateQueryFunction) GetTransactionalQueryFunction() func(tx pool.TransactionWrapper) error {
	return func(tx pool.TransactionWrapper) error {
		_, err := tx.Exec(context.Background(),
			`UPDATE TEMPLATE SET NAME = @name, DESCRIPTION = @description, UPDATED_AT = CURRENT_TIMESTAMP
			 WHERE ID = @id`,
			pgx.NamedArgs{
				"id":          q.template.ID,
				"name":        q.template.Name,
				"description": q.template.Description,
			})
		return err
	}
}

func NewUpdateTemplateQueryFunction(template dbo.TemplateDBO) *UpdateTemplateQueryFunction {
	return &UpdateTemplateQueryFunction{template: template}
}

// DeleteTemplateQueryFunction deletes a template (cascades to items)
type DeleteTemplateQueryFunction struct {
	templateId uint64
}

func (q *DeleteTemplateQueryFunction) GetTransactionalQueryFunction() func(tx pool.TransactionWrapper) error {
	return func(tx pool.TransactionWrapper) error {
		_, err := tx.Exec(context.Background(),
			`DELETE FROM TEMPLATE WHERE ID = @templateId`,
			pgx.NamedArgs{"templateId": q.templateId})
		return err
	}
}

func NewDeleteTemplateQueryFunction(templateId uint64) *DeleteTemplateQueryFunction {
	return &DeleteTemplateQueryFunction{templateId: templateId}
}

// CheckUserIsTemplateOwnerQueryFunction checks if user owns template
type CheckUserIsTemplateOwnerQueryFunction struct {
	templateId uint64
	userId     string
}

func (q *CheckUserIsTemplateOwnerQueryFunction) GetTransactionalQueryFunction() func(tx pool.TransactionWrapper) (bool, error) {
	return func(tx pool.TransactionWrapper) (bool, error) {
		var count int
		err := tx.QueryRow(context.Background(),
			`SELECT COUNT(*) FROM TEMPLATE WHERE ID = @templateId AND USER_ID = @userId`,
			pgx.NamedArgs{
				"templateId": q.templateId,
				"userId":     q.userId,
			}).Scan(&count)
		if err != nil {
			return false, err
		}
		return count > 0, nil
	}
}

func NewCheckUserIsTemplateOwnerQueryFunction(templateId uint64, userId string) *CheckUserIsTemplateOwnerQueryFunction {
	return &CheckUserIsTemplateOwnerQueryFunction{
		templateId: templateId,
		userId:     userId,
	}
}
