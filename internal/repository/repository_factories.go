package repository

import (
	"com.raunlo.checklist/internal/core/repository"
	"github.com/raunlo/pgx-with-automapper/pool"
)

func CreateChecklistRepository(conn pool.Conn) repository.IChecklistRepository {
	return &checklistRepository{
		connection: conn,
	}
}

func CreateChecklistItemRepository(conn pool.Conn) repository.IChecklistItemsRepository {
	return &checklistItemRepository{
		conn: conn,
	}
}

func CreateChecklistItemTemplateRepository() repository.IChecklistItemTemplateRepository {
	return &checklistItemTemplateRepository{}
}
