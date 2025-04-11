package repo

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/kirkegaard/terminal-pet/pkg/db"
	"github.com/kirkegaard/terminal-pet/pkg/db/models"
	"github.com/kirkegaard/terminal-pet/pkg/pet"
)

type PetRepository struct {
	db *db.DB
}

func NewPetRepository(database *db.DB) *PetRepository {
	if database == nil {
		database = db.GetInstance()
	}

	return &PetRepository{
		db: database,
	}
}

func (r *PetRepository) FindByParentPublicKey(ctx context.Context, publicKey string) (*pet.Pet, error) {
	if r.db == nil {
		r.db = db.GetInstance()
		if r.db == nil {
			return nil, fmt.Errorf("no database connection available")
		}
	}

	var model models.Pet

	var userID int
	err := r.db.QueryRow("SELECT id FROM users WHERE public_key = ?", publicKey).Scan(&userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("find user by public key: %w", err)
	}

	err = r.db.QueryRow(`
		SELECT id, name, birthday, parent_id, hunger, happiness, discipline, health, weight, is_sick, has_pooped
		FROM pets WHERE parent_id = ? LIMIT 1
	`, userID).Scan(
		&model.ID,
		&model.Name,
		&model.BirthDate,
		&model.ParentID,
		&model.Hunger,
		&model.Happiness,
		&model.Discipline,
		&model.Health,
		&model.Weight,
		&model.IsSick,
		&model.HasPooped,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("find pet by parent id: %w", err)
	}

	parent := pet.NewParent(model.ParentID, "Player")

	petModel := pet.NewPet(model.Name, model.BirthDate, parent)
	petModel.Hunger = model.Hunger
	petModel.Happiness = model.Happiness
	petModel.Discipline = model.Discipline
	petModel.Health = model.Health
	petModel.Weight = model.Weight
	petModel.IsSick = model.IsSick
	petModel.HasPooped = model.HasPooped

	return petModel, nil
}

func (r *PetRepository) Save(ctx context.Context, p *pet.Pet, publicKey string) error {
	if r.db == nil {
		r.db = db.GetInstance()
		if r.db == nil {
			return fmt.Errorf("no database connection available")
		}
	}

	var userID int
	err := r.db.QueryRow("SELECT id FROM users WHERE public_key = ?", publicKey).Scan(&userID)
	if err != nil {
		if err == sql.ErrNoRows {
			res, err := r.db.Exec("INSERT INTO users (name, public_key) VALUES (?, ?)", p.Parent.Name, publicKey)
			if err != nil {
				return fmt.Errorf("create user: %w", err)
			}

			id, err := res.LastInsertId()
			if err != nil {
				return fmt.Errorf("get last insert id: %w", err)
			}

			userID = int(id)
			p.Parent.ID = userID
		} else {
			return fmt.Errorf("find user by public key: %w", err)
		}
	}

	// Check if the pet already exists
	var petID int
	err = r.db.QueryRow("SELECT id FROM pets WHERE parent_id = ?", userID).Scan(&petID)
	if err != nil {
		if err == sql.ErrNoRows {
			_, err := r.db.Exec(`
				INSERT INTO pets (
					name, birthday, parent_id, hunger, happiness, discipline, health, weight, is_sick, has_pooped
				) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
			`,
				p.Name,
				p.BirthDate,
				userID,
				p.Hunger,
				p.Happiness,
				p.Discipline,
				p.Health,
				p.Weight,
				p.IsSick,
				p.HasPooped,
			)
			if err != nil {
				return fmt.Errorf("create pet: %w", err)
			}
		} else {
			return fmt.Errorf("find pet by parent id: %w", err)
		}
	} else {
		_, err := r.db.Exec(`
			UPDATE pets SET 
				name = ?,
				birthday = ?,
				hunger = ?, 
				happiness = ?, 
				discipline = ?, 
				health = ?, 
				weight = ?, 
				is_sick = ?,
				has_pooped = ?,
				updated_at = ?
			WHERE id = ?
		`,
			p.Name,
			p.BirthDate,
			p.Hunger,
			p.Happiness,
			p.Discipline,
			p.Health,
			p.Weight,
			p.IsSick,
			p.HasPooped,
			time.Now(),
			petID,
		)
		if err != nil {
			return fmt.Errorf("update pet: %w", err)
		}
	}

	return nil
}
