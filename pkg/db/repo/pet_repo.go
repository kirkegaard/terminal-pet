package repo

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/charmbracelet/log"
	"github.com/kirkegaard/terminal-pet/pkg/db"
	"github.com/kirkegaard/terminal-pet/pkg/db/models"
	"github.com/kirkegaard/terminal-pet/pkg/pet"
)

type PetRepository struct {
	db       *db.DB
	userRepo *UserRepository
}

func NewPetRepository(database *db.DB) *PetRepository {
	if database == nil {
		database = db.GetInstance()
	}

	return &PetRepository{
		db:       database,
		userRepo: NewUserRepository(database),
	}
}

// Create creates a new pet in the database
func (r *PetRepository) Create(ctx context.Context, p *pet.Pet) error {
	if r.db == nil {
		return fmt.Errorf("no database connection available")
	}

	result, err := r.db.Exec(`
		INSERT INTO pets (
			name, birthday, parent_id, hunger, happiness, discipline, health, weight, is_sick, has_pooped, lights_on
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		p.Name,
		p.BirthDate,
		p.Parent.ID,
		p.Hunger,
		p.Happiness,
		p.Discipline,
		p.Health,
		p.Weight,
		p.IsSick,
		p.HasPooped,
		p.LightsOn,
	)
	if err != nil {
		return fmt.Errorf("create pet: %w", err)
	}

	// Get the last inserted ID and set it on the pet
	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("get last insert id: %w", err)
	}

	p.ID = int(id)
	return nil
}

// GetByParentID retrieves a pet by its parent ID
func (r *PetRepository) GetByParentID(ctx context.Context, parentID int) (*pet.Pet, error) {
	if r.db == nil {
		return nil, fmt.Errorf("no database connection available")
	}

	var model models.Pet

	err := r.db.QueryRow(`
		SELECT id, name, birthday, parent_id, hunger, happiness, discipline, health, weight, is_sick, has_pooped, lights_on, updated_at
		FROM pets WHERE parent_id = ? ORDER BY created_at DESC LIMIT 1
	`, parentID).Scan(
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
		&model.LightsOn,
		&model.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get pet by parent id: %w", err)
	}

	parent := pet.NewParent(model.ParentID, "Player")

	petModel := pet.NewPet(model.Name, model.BirthDate, parent)
	petModel.ID = model.ID
	petModel.Hunger = model.Hunger
	petModel.Happiness = model.Happiness
	petModel.Discipline = model.Discipline
	petModel.Health = model.Health
	petModel.Weight = model.Weight
	petModel.IsSick = model.IsSick
	petModel.HasPooped = model.HasPooped
	petModel.LightsOn = model.LightsOn
	petModel.LastVisit = model.UpdatedAt

	return petModel, nil
}

// Update updates an existing pet in the database
func (r *PetRepository) Update(ctx context.Context, p *pet.Pet) error {
	if r.db == nil {
		return fmt.Errorf("no database connection available")
	}

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
			lights_on = ?,
			updated_at = ?
		WHERE id = ? AND parent_id = ?
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
		p.LightsOn,
		time.Now(),
		p.ID,
		p.Parent.ID,
	)
	if err != nil {
		return fmt.Errorf("update pet: %w", err)
	}

	return nil
}

// Delete removes a pet from the database
func (r *PetRepository) Delete(ctx context.Context, id int) error {
	if r.db == nil {
		return fmt.Errorf("no database connection available")
	}

	_, err := r.db.Exec("DELETE FROM pets WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("delete pet: %w", err)
	}

	return nil
}

// FindByParentPublicKey is a helper method that retrieves a pet by the parent's public key
func (r *PetRepository) FindByParentPublicKey(ctx context.Context, publicKey string) (*pet.Pet, error) {
	if r.db == nil {
		return nil, fmt.Errorf("no database connection available")
	}

	userID, err := r.userRepo.GetByPublicKey(ctx, publicKey)
	if err != nil {
		return nil, err
	}

	if userID == 0 {
		return nil, nil
	}

	pet, err := r.GetByParentID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if pet != nil {
		log.Debug("Found pet by public key", "id", pet.ID, "name", pet.Name)
	} else {
		log.Debug("No pet found for user", "userID", userID)
	}

	return pet, nil
}

// Save is a helper method that either creates or updates a pet
func (r *PetRepository) Save(ctx context.Context, p *pet.Pet, publicKey string) error {
	if r.db == nil {
		return fmt.Errorf("no database connection available")
	}

	log.Debug("Saving pet", "id", p.ID, "name", p.Name)

	userID, err := r.userRepo.GetByPublicKey(ctx, publicKey)
	if err != nil {
		return err
	}

	if userID == 0 {
		userID, err = r.userRepo.Create(ctx, p.Parent.Name, publicKey)
		if err != nil {
			return err
		}
		p.Parent.ID = userID
	}

	// Check if the pet already exists
	existingPet, err := r.GetByParentID(ctx, userID)
	if err != nil {
		return err
	}

	if existingPet == nil {
		return r.Create(ctx, p)
	} else {
		return r.Update(ctx, p)
	}
}
