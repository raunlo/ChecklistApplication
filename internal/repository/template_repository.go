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
	"github.com/pkg/errors"
	"github.com/raunlo/pgx-with-automapper/mapper"
	"github.com/raunlo/pgx-with-automapper/pool"
)

type templateRepository struct {
	connection pool.Conn
}

func (repository *templateRepository) SaveTemplate(ctx context.Context, template domain.Template) (domain.Template, domain.Error) {
	templateDBO := dbo.TemplateDBO{}
	templateDBO.FromDomain(template)

	itemDBOs := make([]dbo.TemplateItemDBO, len(template.Items))
	for i, item := range template.Items {
		itemDBOs[i].FromDomain(item)
	}

	queryFunc := query.NewSaveTemplateQueryFunction(templateDBO, itemDBOs)

	res, err := connection.RunInTransaction(connection.TransactionProps[dbo.TemplateDBO]{
		Ctx:        ctx,
		Query:      queryFunc.GetTransactionalQueryFunction(),
		Connection: repository.connection,
		TxOptions:  connection.TxReadCommitted,
	})

	if err != nil {
		return domain.Template{}, domain.Wrap(err, "Could not save template", 500)
	}

	template.ID = uint(res.ID)
	return template, nil
}

func (repository *templateRepository) FindTemplateById(ctx context.Context, id uint) (*domain.Template, domain.Error) {
	queryFunc := query.NewFindTemplateByIdQueryFunction(uint64(id))

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

	// Fetch template items
	itemsQueryFunc := query.NewFindTemplateItemsByTemplateIdQueryFunction(uint64(id))
	itemDBOs, err := connection.RunInTransaction(connection.TransactionProps[[]dbo.TemplateItemDBO]{
		Ctx:        ctx,
		Query:      itemsQueryFunc.GetTransactionalQueryFunction(),
		Connection: repository.connection,
		TxOptions:  connection.TxReadCommitted,
	})

	if err != nil {
		return nil, domain.Wrap(err, fmt.Sprintf("Failed to find template items for template(id=%d)", id), 500)
	}

	domainTemplate := res.ToDomain()
	for _, itemDBO := range itemDBOs {
		domainTemplate.Items = append(domainTemplate.Items, itemDBO.ToDomain())
	}

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
		// Fetch items for each template
		itemsQueryFunc := query.NewFindTemplateItemsByTemplateIdQueryFunction(templateDBO.ID)
		itemDBOs, err := connection.RunInTransaction(connection.TransactionProps[[]dbo.TemplateItemDBO]{
			Ctx:        ctx,
			Query:      itemsQueryFunc.GetTransactionalQueryFunction(),
			Connection: repository.connection,
			TxOptions:  connection.TxReadCommitted,
		})

		if err != nil {
			return nil, domain.Wrap(err, fmt.Sprintf("Failed to find template items for template(id=%d)", templateDBO.ID), 500)
		}

		domainTemplate := templateDBO.ToDomain()
		for _, itemDBO := range itemDBOs {
			domainTemplate.Items = append(domainTemplate.Items, itemDBO.ToDomain())
		}
		domainTemplates = append(domainTemplates, domainTemplate)
	}

	return domainTemplates, nil
}

func (repository *templateRepository) UpdateTemplate(ctx context.Context, template domain.Template) (domain.Template, domain.Error) {
	templateDBO := dbo.TemplateDBO{}
	templateDBO.FromDomain(template)

	queryFunc := query.NewUpdateTemplateQueryFunction(templateDBO)

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

func CreateTemplateRepository(connection pool.Conn) coreRepo.ITemplateRepository {
	return &templateRepository{
		connection: connection,
	}
}
