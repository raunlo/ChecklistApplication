package repository

import (
	"context"
	"fmt"

	"com.raunlo.checklist/internal/core/domain"
	coreRepo "com.raunlo.checklist/internal/core/repository"
	"com.raunlo.checklist/internal/repository/connection"
	"com.raunlo.checklist/internal/repository/dbo"
	"com.raunlo.checklist/internal/repository/query"
	"com.raunlo.checklist/internal/util"
	"github.com/jackc/pgx/v5"
	"github.com/pkg/errors"
	"github.com/raunlo/pgx-with-automapper/mapper"
	"github.com/raunlo/pgx-with-automapper/pool"
)

type templateRepository struct {
	connection pool.Conn
}

func (repository *templateRepository) fetchWorkspaceIds(ctx context.Context, templateId uint64) ([]uint, domain.Error) {
	queryFunc := query.NewFindWorkspaceIdsByTemplateIdQueryFunction(templateId)
	dbos, err := connection.RunInTransaction(connection.TransactionProps[[]dbo.TemplateWorkspaceDBO]{
		Ctx:        ctx,
		Query:      queryFunc.GetTransactionalQueryFunction(),
		Connection: repository.connection,
		TxOptions:  connection.TxReadCommitted,
	})
	if err != nil {
		return nil, domain.Wrap(err, fmt.Sprintf("Failed to find workspace IDs for template(id=%d)", templateId), 500)
	}
	ids := make([]uint, len(dbos))
	for i, d := range dbos {
		ids[i] = uint(d.WorkspaceID)
	}
	return ids, nil
}

func (repository *templateRepository) SaveTemplate(ctx context.Context, template domain.Template) (domain.Template, domain.Error) {
	templateDBO := dbo.TemplateDBO{}
	templateDBO.FromDomain(template)

	rowDBOs := make([]dbo.TemplateRowDBO, len(template.Rows))
	for i, row := range template.Rows {
		rowDBOs[i].FromDomain(row)
	}

	queryFunc := query.NewSaveTemplateQueryFunction(templateDBO, rowDBOs)

	res, err := connection.RunInTransaction(connection.TransactionProps[dbo.TemplateDBO]{
		Ctx:        ctx,
		Query:      queryFunc.GetTransactionalQueryFunction(),
		Connection: repository.connection,
		TxOptions:  connection.TxReadCommitted,
	})

	if err != nil {
		return domain.Template{}, domain.Wrap(err, "Could not save template", 500)
	}

	template.Id = uint(res.ID)
	return template, nil
}

func (repository *templateRepository) FindTemplateById(ctx context.Context, id uint) (*domain.Template, domain.Error) {
	userId, _ := domain.GetUserIdFromContext(ctx)
	queryFunc := query.NewFindTemplateByIdQueryFunction(uint64(id), userId)

	res, err := connection.RunInTransaction(connection.TransactionProps[dbo.TemplateDBO]{
		Ctx:        ctx,
		Query:      queryFunc.GetTransactionalQueryFunction(),
		Connection: repository.connection,
		TxOptions:  connection.TxReadCommitted,
	})

	if err != nil {
		if errors.Is(err, mapper.ErrNoRows) {
			return nil, nil
		}
		return nil, domain.Wrap(err, fmt.Sprintf("Failed to find template(id=%d)", id), 500)
	}

	rowsQueryFunc := query.NewFindTemplateRowsByTemplateIdQueryFunction(uint64(id))
	rowDBOs, err := connection.RunInTransaction(connection.TransactionProps[[]dbo.TemplateRowDBO]{
		Ctx:        ctx,
		Query:      rowsQueryFunc.GetTransactionalQueryFunction(),
		Connection: repository.connection,
		TxOptions:  connection.TxReadCommitted,
	})

	if err != nil {
		return nil, domain.Wrap(err, fmt.Sprintf("Failed to find template rows for template(id=%d)", id), 500)
	}

	domainTemplate := res.ToDomain()
	for _, rowDBO := range rowDBOs {
		domainTemplate.Rows = append(domainTemplate.Rows, rowDBO.ToDomain())
	}

	wsIds, wsErr := repository.fetchWorkspaceIds(ctx, uint64(id))
	if wsErr != nil {
		return nil, wsErr
	}
	domainTemplate.WorkspaceIds = wsIds

	return util.AnyPointer(domainTemplate), nil
}

func (repository *templateRepository) FindTemplatesByUserId(ctx context.Context, userId string) ([]domain.Template, domain.Error) {
	queryFunc := query.NewFindAllTemplatesByUserIdQueryFunction(userId)

	templates, err := connection.RunInTransaction(connection.TransactionProps[[]dbo.TemplateDBO]{
		Ctx:        ctx,
		Query:      queryFunc.GetTransactionalQueryFunction(),
		Connection: repository.connection,
		TxOptions:  connection.TxReadCommitted,
	})

	if err != nil {
		return nil, domain.Wrap(err, "Failed to find templates", 500)
	}

	domainTemplates := make([]domain.Template, 0)
	for _, templateDBO := range templates {
		rowsQueryFunc := query.NewFindTemplateRowsByTemplateIdQueryFunction(templateDBO.ID)
		rowDBOs, err := connection.RunInTransaction(connection.TransactionProps[[]dbo.TemplateRowDBO]{
			Ctx:        ctx,
			Query:      rowsQueryFunc.GetTransactionalQueryFunction(),
			Connection: repository.connection,
			TxOptions:  connection.TxReadCommitted,
		})

		if err != nil {
			return nil, domain.Wrap(err, fmt.Sprintf("Failed to find template rows for template(id=%d)", templateDBO.ID), 500)
		}

		domainTemplate := templateDBO.ToDomain()
		for _, rowDBO := range rowDBOs {
			domainTemplate.Rows = append(domainTemplate.Rows, rowDBO.ToDomain())
		}

		wsIds, wsErr := repository.fetchWorkspaceIds(ctx, templateDBO.ID)
		if wsErr != nil {
			return nil, wsErr
		}
		domainTemplate.WorkspaceIds = wsIds

		domainTemplates = append(domainTemplates, domainTemplate)
	}

	return domainTemplates, nil
}

func (repository *templateRepository) UpdateTemplate(ctx context.Context, template domain.Template) (domain.Template, domain.Error) {
	templateDBO := dbo.TemplateDBO{}
	templateDBO.FromDomain(template)

	rowDBOs := make([]dbo.TemplateRowDBO, len(template.Rows))
	for i, row := range template.Rows {
		rowDBOs[i].FromDomain(row)
	}

	queryFunc := query.NewUpdateTemplateQueryFunction(templateDBO, rowDBOs)

	_, err := connection.RunInTransaction(connection.TransactionProps[bool]{
		Ctx: ctx,
		Query: func(tx pool.TransactionWrapper) (bool, error) {
			return true, queryFunc.GetTransactionalQueryFunction()(tx)
		},
		Connection: repository.connection,
		TxOptions:  connection.TxReadCommitted,
	})

	if err != nil {
		return domain.Template{}, domain.Wrap(err, "Could not update template", 500)
	}

	return template, nil
}

func (repository *templateRepository) DeleteTemplate(ctx context.Context, id uint) domain.Error {
	queryFunc := query.NewDeleteTemplateQueryFunction(uint64(id))

	_, err := connection.RunInTransaction(connection.TransactionProps[bool]{
		Ctx: ctx,
		Query: func(tx pool.TransactionWrapper) (bool, error) {
			return true, queryFunc.GetTransactionalQueryFunction()(tx)
		},
		Connection: repository.connection,
		TxOptions:  connection.TxReadCommitted,
	})

	if err != nil {
		return domain.Wrap(err, "Could not delete template", 500)
	}

	return nil
}

func (repository *templateRepository) CheckUserIsTemplateOwner(ctx context.Context, templateId uint, userId string) (bool, domain.Error) {
	queryFunc := query.NewCheckUserIsTemplateOwnerQueryFunction(uint64(templateId), userId)

	isOwner, err := connection.RunInTransaction(connection.TransactionProps[bool]{
		Ctx:        ctx,
		Query:      queryFunc.GetTransactionalQueryFunction(),
		Connection: repository.connection,
		TxOptions:  connection.TxReadCommitted,
	})

	if err != nil {
		return false, domain.Wrap(err, "Failed to check template ownership", 500)
	}

	return isOwner, nil
}

func (repository *templateRepository) CheckUserHasAccessToTemplate(ctx context.Context, templateId uint, userId string) (bool, domain.Error) {
	queryFunc := func(tx pool.TransactionWrapper) (bool, error) {
		var count int
		err := tx.QueryRow(ctx,
			`SELECT COUNT(*) FROM TEMPLATE t
			 WHERE t.ID = @templateId AND (
			   t.USER_ID = @userId
			   OR EXISTS (SELECT 1 FROM TEMPLATE_SHARE ts WHERE ts.TEMPLATE_ID = @templateId AND ts.SHARED_WITH_USER_ID = @userId)
			   OR EXISTS (
			         SELECT 1 FROM template_workspace tw
			         JOIN workspace_member wm ON wm.workspace_id = tw.workspace_id
			         WHERE tw.template_id = @templateId AND wm.user_id = @userId
			       )
			 )`,
			pgx.NamedArgs{
				"templateId": templateId,
				"userId":     userId,
			}).Scan(&count)
		return count > 0, err
	}

	hasAccess, err := connection.RunInTransaction(connection.TransactionProps[bool]{
		Ctx:        ctx,
		Query:      queryFunc,
		Connection: repository.connection,
		TxOptions:  connection.TxReadCommitted,
	})

	if err != nil {
		return false, domain.Wrap(err, "Failed to check template access", 500)
	}

	return hasAccess, nil
}

func (repository *templateRepository) CreateTemplateShare(ctx context.Context, templateId uint, sharedBy string, sharedWith string) domain.Error {
	queryFunc := func(tx pool.TransactionWrapper) (bool, error) {
		_, err := tx.Exec(ctx,
			`INSERT INTO TEMPLATE_SHARE(ID, TEMPLATE_ID, SHARED_BY_USER_ID, SHARED_WITH_USER_ID, CREATED_AT)
			 VALUES(nextval('template_share_id_sequence'), @templateId, @sharedBy, @sharedWith, CURRENT_TIMESTAMP)
			 ON CONFLICT (TEMPLATE_ID, SHARED_WITH_USER_ID) DO NOTHING`,
			pgx.NamedArgs{
				"templateId": templateId,
				"sharedBy":   sharedBy,
				"sharedWith": sharedWith,
			})
		return true, err
	}

	_, err := connection.RunInTransaction(connection.TransactionProps[bool]{
		Ctx:        ctx,
		Query:      queryFunc,
		Connection: repository.connection,
		TxOptions:  connection.TxReadCommitted,
	})

	if err != nil {
		return domain.Wrap(err, "Failed to create template share", 500)
	}

	return nil
}

func (repository *templateRepository) DeleteTemplateShare(ctx context.Context, templateId uint, userId string) domain.Error {
	queryFunc := func(tx pool.TransactionWrapper) (bool, error) {
		result, err := tx.Exec(ctx,
			`DELETE FROM TEMPLATE_SHARE WHERE TEMPLATE_ID = @templateId AND SHARED_WITH_USER_ID = @userId`,
			pgx.NamedArgs{
				"templateId": templateId,
				"userId":     userId,
			})
		return result.RowsAffected() == 1, err
	}

	success, err := connection.RunInTransaction(connection.TransactionProps[bool]{
		Ctx:        ctx,
		Query:      queryFunc,
		Connection: repository.connection,
		TxOptions:  connection.TxReadCommitted,
	})

	if err != nil {
		return domain.Wrap(err, "Failed to delete template share", 500)
	}

	if !success {
		return domain.NewError(fmt.Sprintf("No shared access found for template(id=%d)", templateId), 404)
	}

	return nil
}

func (repository *templateRepository) FindTemplatesByWorkspaceId(ctx context.Context, workspaceId uint) ([]domain.Template, domain.Error) {
	userId, _ := domain.GetUserIdFromContext(ctx)
	queryFunc := query.NewFindTemplatesByWorkspaceIdQueryFunction(uint64(workspaceId), userId)

	templates, err := connection.RunInTransaction(connection.TransactionProps[[]dbo.TemplateDBO]{
		Ctx:        ctx,
		Query:      queryFunc.GetTransactionalQueryFunction(),
		Connection: repository.connection,
		TxOptions:  connection.TxReadCommitted,
	})

	if err != nil {
		return nil, domain.Wrap(err, "Failed to find workspace templates", 500)
	}

	domainTemplates := make([]domain.Template, 0, len(templates))
	for _, templateDBO := range templates {
		rowsQueryFunc := query.NewFindTemplateRowsByTemplateIdQueryFunction(templateDBO.ID)
		rowDBOs, rowErr := connection.RunInTransaction(connection.TransactionProps[[]dbo.TemplateRowDBO]{
			Ctx:        ctx,
			Query:      rowsQueryFunc.GetTransactionalQueryFunction(),
			Connection: repository.connection,
			TxOptions:  connection.TxReadCommitted,
		})
		if rowErr != nil {
			return nil, domain.Wrap(rowErr, fmt.Sprintf("Failed to find template rows for template(id=%d)", templateDBO.ID), 500)
		}

		domainTemplate := templateDBO.ToDomain()
		for _, rowDBO := range rowDBOs {
			domainTemplate.Rows = append(domainTemplate.Rows, rowDBO.ToDomain())
		}

		wsIds, wsErr := repository.fetchWorkspaceIds(ctx, templateDBO.ID)
		if wsErr != nil {
			return nil, wsErr
		}
		domainTemplate.WorkspaceIds = wsIds

		domainTemplates = append(domainTemplates, domainTemplate)
	}

	return domainTemplates, nil
}

func (repository *templateRepository) AssignTemplateToWorkspace(ctx context.Context, templateId uint, workspaceId uint) domain.Error {
	queryFunc := query.NewAssignTemplateToWorkspaceQueryFunction(uint64(templateId), uint64(workspaceId))
	_, err := connection.RunInTransaction(connection.TransactionProps[bool]{
		Ctx:        ctx,
		Query:      queryFunc.GetTransactionalQueryFunction(),
		Connection: repository.connection,
		TxOptions:  connection.TxReadCommitted,
	})
	if err != nil {
		return domain.Wrap(err, fmt.Sprintf("Failed to assign template(id=%d) to workspace(id=%d)", templateId, workspaceId), 500)
	}
	return nil
}

func (repository *templateRepository) UnassignTemplateFromWorkspace(ctx context.Context, templateId uint, workspaceId uint) domain.Error {
	queryFunc := query.NewUnassignTemplateFromWorkspaceQueryFunction(uint64(templateId), uint64(workspaceId))
	found, err := connection.RunInTransaction(connection.TransactionProps[bool]{
		Ctx:        ctx,
		Query:      queryFunc.GetTransactionalQueryFunction(),
		Connection: repository.connection,
		TxOptions:  connection.TxReadCommitted,
	})
	if err != nil {
		return domain.Wrap(err, fmt.Sprintf("Failed to unassign template(id=%d) from workspace(id=%d)", templateId, workspaceId), 500)
	}
	if !found {
		return domain.NewError(fmt.Sprintf("Template(id=%d) is not assigned to workspace(id=%d)", templateId, workspaceId), 404)
	}
	return nil
}

func CreateTemplateRepository(connection pool.Conn) coreRepo.ITemplateRepository {
	return &templateRepository{
		connection: connection,
	}
}
