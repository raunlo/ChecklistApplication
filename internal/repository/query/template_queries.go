package query

import (
	"context"

	"com.raunlo.checklist/internal/repository/dbo"
	"github.com/jackc/pgx/v5"
	"github.com/raunlo/pgx-with-automapper/pool"
)

// SaveTemplateQueryFunction saves a template and its rows
type SaveTemplateQueryFunction struct {
	template dbo.TemplateDBO
	rows     []dbo.TemplateRowDBO
}

func (q *SaveTemplateQueryFunction) GetTransactionalQueryFunction() func(tx pool.TransactionWrapper) (dbo.TemplateDBO, error) {
	return func(tx pool.TransactionWrapper) (dbo.TemplateDBO, error) {
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

		for _, row := range q.rows {
			row.TemplateID = q.template.ID
			err := tx.QueryRow(context.Background(),
				`INSERT INTO TEMPLATE_ROW(TEMPLATE_ID, NAME, POSITION, CREATED_AT, UPDATED_AT)
				 VALUES(@templateId, @name, @position, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
				 RETURNING ID`,
				pgx.NamedArgs{
					"templateId": row.TemplateID,
					"name":       row.Name,
					"position":   row.Position,
				}).Scan(&row.ID)
			if err != nil {
				return dbo.TemplateDBO{}, err
			}
		}

		return q.template, nil
	}
}

func NewSaveTemplateQueryFunction(template dbo.TemplateDBO, rows []dbo.TemplateRowDBO) *SaveTemplateQueryFunction {
	return &SaveTemplateQueryFunction{
		template: template,
		rows:     rows,
	}
}

// FindTemplateByIdQueryFunction finds a template by ID
type FindTemplateByIdQueryFunction struct {
	templateId uint64
	userId     string
}

func (q *FindTemplateByIdQueryFunction) GetTransactionalQueryFunction() func(tx pool.TransactionWrapper) (dbo.TemplateDBO, error) {
	return func(tx pool.TransactionWrapper) (dbo.TemplateDBO, error) {
		var template dbo.TemplateDBO
		err := tx.QueryRow(context.Background(),
			`SELECT ID, USER_ID, NAME, DESCRIPTION, CREATED_AT, UPDATED_AT, (USER_ID = @userId) AS IS_OWNER
			 FROM TEMPLATE WHERE ID = @templateId`,
			pgx.NamedArgs{"templateId": q.templateId, "userId": q.userId}).Scan(
			&template.ID, &template.UserID, &template.Name, &template.Description,
			&template.CreatedAt, &template.UpdatedAt, &template.IsOwner)
		if err != nil {
			return dbo.TemplateDBO{}, err
		}
		return template, nil
	}
}

func NewFindTemplateByIdQueryFunction(templateId uint64, userId string) *FindTemplateByIdQueryFunction {
	return &FindTemplateByIdQueryFunction{templateId: templateId, userId: userId}
}

// FindTemplateRowsByTemplateIdQueryFunction finds all rows for a template
type FindTemplateRowsByTemplateIdQueryFunction struct {
	templateId uint64
}

func (q *FindTemplateRowsByTemplateIdQueryFunction) GetTransactionalQueryFunction() func(tx pool.TransactionWrapper) ([]dbo.TemplateRowDBO, error) {
	return func(tx pool.TransactionWrapper) ([]dbo.TemplateRowDBO, error) {
		rows, err := tx.Query(context.Background(),
			`SELECT ID, TEMPLATE_ID, NAME, POSITION, CREATED_AT, UPDATED_AT
			 FROM TEMPLATE_ROW WHERE TEMPLATE_ID = @templateId
			 ORDER BY POSITION ASC`,
			pgx.NamedArgs{"templateId": q.templateId})
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		var result []dbo.TemplateRowDBO
		for rows.Next() {
			var row dbo.TemplateRowDBO
			err := rows.Scan(&row.ID, &row.TemplateID, &row.Name, &row.Position, &row.CreatedAt, &row.UpdatedAt)
			if err != nil {
				return nil, err
			}
			result = append(result, row)
		}
		return result, rows.Err()
	}
}

func NewFindTemplateRowsByTemplateIdQueryFunction(templateId uint64) *FindTemplateRowsByTemplateIdQueryFunction {
	return &FindTemplateRowsByTemplateIdQueryFunction{templateId: templateId}
}

// FindAllTemplatesByUserIdQueryFunction finds all templates visible to a user
type FindAllTemplatesByUserIdQueryFunction struct {
	userId string
}

func (q *FindAllTemplatesByUserIdQueryFunction) GetTransactionalQueryFunction() func(tx pool.TransactionWrapper) ([]dbo.TemplateDBO, error) {
	return func(tx pool.TransactionWrapper) ([]dbo.TemplateDBO, error) {
		rows, err := tx.Query(context.Background(),
			`SELECT DISTINCT t.ID, t.USER_ID, t.NAME, t.DESCRIPTION, t.CREATED_AT, t.UPDATED_AT,
			        (t.USER_ID = @userId) AS IS_OWNER
			 FROM TEMPLATE t
			 LEFT JOIN template_workspace tw ON tw.template_id = t.ID
			 LEFT JOIN workspace_member wm ON wm.workspace_id = tw.workspace_id AND wm.user_id = @userId
			 WHERE t.USER_ID = @userId OR wm.user_id = @userId
			 ORDER BY t.UPDATED_AT DESC`,
			pgx.NamedArgs{"userId": q.userId})
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		var templates []dbo.TemplateDBO
		for rows.Next() {
			var template dbo.TemplateDBO
			err := rows.Scan(&template.ID, &template.UserID, &template.Name, &template.Description, &template.CreatedAt, &template.UpdatedAt, &template.IsOwner)
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

// UpdateTemplateQueryFunction updates a template and replaces its rows
type UpdateTemplateQueryFunction struct {
	template dbo.TemplateDBO
	rows     []dbo.TemplateRowDBO
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
		if err != nil {
			return err
		}

		_, err = tx.Exec(context.Background(),
			`DELETE FROM TEMPLATE_ROW WHERE TEMPLATE_ID = @templateId`,
			pgx.NamedArgs{"templateId": q.template.ID})
		if err != nil {
			return err
		}

		for _, row := range q.rows {
			_, err = tx.Exec(context.Background(),
				`INSERT INTO TEMPLATE_ROW(TEMPLATE_ID, NAME, POSITION, CREATED_AT, UPDATED_AT)
				 VALUES(@templateId, @name, @position, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`,
				pgx.NamedArgs{
					"templateId": q.template.ID,
					"name":       row.Name,
					"position":   row.Position,
				})
			if err != nil {
				return err
			}
		}

		return nil
	}
}

func NewUpdateTemplateQueryFunction(template dbo.TemplateDBO, rows []dbo.TemplateRowDBO) *UpdateTemplateQueryFunction {
	return &UpdateTemplateQueryFunction{template: template, rows: rows}
}

// DeleteTemplateQueryFunction deletes a template (cascades to rows)
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

// FindTemplatesByWorkspaceIdQueryFunction finds all templates for a workspace via join table
type FindTemplatesByWorkspaceIdQueryFunction struct {
	workspaceId uint64
	userId      string
}

func (q *FindTemplatesByWorkspaceIdQueryFunction) GetTransactionalQueryFunction() func(tx pool.TransactionWrapper) ([]dbo.TemplateDBO, error) {
	return func(tx pool.TransactionWrapper) ([]dbo.TemplateDBO, error) {
		rows, err := tx.Query(context.Background(),
			`SELECT t.ID, t.USER_ID, t.NAME, t.DESCRIPTION, t.CREATED_AT, t.UPDATED_AT,
			        (t.USER_ID = @userId) AS IS_OWNER
			 FROM TEMPLATE t
			 JOIN template_workspace tw ON tw.template_id = t.ID
			 WHERE tw.workspace_id = @workspaceId
			 ORDER BY t.UPDATED_AT DESC`,
			pgx.NamedArgs{"workspaceId": q.workspaceId, "userId": q.userId})
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		var templates []dbo.TemplateDBO
		for rows.Next() {
			var t dbo.TemplateDBO
			err := rows.Scan(&t.ID, &t.UserID, &t.Name, &t.Description, &t.CreatedAt, &t.UpdatedAt, &t.IsOwner)
			if err != nil {
				return nil, err
			}
			templates = append(templates, t)
		}
		return templates, rows.Err()
	}
}

func NewFindTemplatesByWorkspaceIdQueryFunction(workspaceId uint64, userId string) *FindTemplatesByWorkspaceIdQueryFunction {
	return &FindTemplatesByWorkspaceIdQueryFunction{workspaceId: workspaceId, userId: userId}
}

// FindWorkspaceIdsByTemplateIdQueryFunction finds all workspace IDs for a template
type FindWorkspaceIdsByTemplateIdQueryFunction struct {
	templateId uint64
}

func (q *FindWorkspaceIdsByTemplateIdQueryFunction) GetTransactionalQueryFunction() func(tx pool.TransactionWrapper) ([]dbo.TemplateWorkspaceDBO, error) {
	return func(tx pool.TransactionWrapper) ([]dbo.TemplateWorkspaceDBO, error) {
		rows, err := tx.Query(context.Background(),
			`SELECT workspace_id FROM template_workspace WHERE template_id = @templateId`,
			pgx.NamedArgs{"templateId": q.templateId})
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		var result []dbo.TemplateWorkspaceDBO
		for rows.Next() {
			var item dbo.TemplateWorkspaceDBO
			if err := rows.Scan(&item.WorkspaceID); err != nil {
				return nil, err
			}
			result = append(result, item)
		}
		return result, rows.Err()
	}
}

func NewFindWorkspaceIdsByTemplateIdQueryFunction(templateId uint64) *FindWorkspaceIdsByTemplateIdQueryFunction {
	return &FindWorkspaceIdsByTemplateIdQueryFunction{templateId: templateId}
}

// AssignTemplateToWorkspaceQueryFunction inserts a template-workspace association
type AssignTemplateToWorkspaceQueryFunction struct {
	templateId  uint64
	workspaceId uint64
}

func (q *AssignTemplateToWorkspaceQueryFunction) GetTransactionalQueryFunction() func(tx pool.TransactionWrapper) (bool, error) {
	return func(tx pool.TransactionWrapper) (bool, error) {
		_, err := tx.Exec(context.Background(),
			`INSERT INTO template_workspace(template_id, workspace_id, assigned_at)
			 VALUES(@templateId, @workspaceId, CURRENT_TIMESTAMP)
			 ON CONFLICT DO NOTHING`,
			pgx.NamedArgs{
				"templateId":  q.templateId,
				"workspaceId": q.workspaceId,
			})
		return err == nil, err
	}
}

func NewAssignTemplateToWorkspaceQueryFunction(templateId uint64, workspaceId uint64) *AssignTemplateToWorkspaceQueryFunction {
	return &AssignTemplateToWorkspaceQueryFunction{templateId: templateId, workspaceId: workspaceId}
}

// UnassignTemplateFromWorkspaceQueryFunction removes a template-workspace association
type UnassignTemplateFromWorkspaceQueryFunction struct {
	templateId  uint64
	workspaceId uint64
}

func (q *UnassignTemplateFromWorkspaceQueryFunction) GetTransactionalQueryFunction() func(tx pool.TransactionWrapper) (bool, error) {
	return func(tx pool.TransactionWrapper) (bool, error) {
		result, err := tx.Exec(context.Background(),
			`DELETE FROM template_workspace WHERE template_id = @templateId AND workspace_id = @workspaceId`,
			pgx.NamedArgs{
				"templateId":  q.templateId,
				"workspaceId": q.workspaceId,
			})
		if err != nil {
			return false, err
		}
		return result.RowsAffected() > 0, nil
	}
}

func NewUnassignTemplateFromWorkspaceQueryFunction(templateId uint64, workspaceId uint64) *UnassignTemplateFromWorkspaceQueryFunction {
	return &UnassignTemplateFromWorkspaceQueryFunction{templateId: templateId, workspaceId: workspaceId}
}
