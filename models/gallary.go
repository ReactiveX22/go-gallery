package models

import (
	"database/sql"
	"errors"
	"fmt"
)

type Gallary struct {
	ID     int
	UserID int
	Title  string
}

type GalleryService struct {
	DB *sql.DB
}

func (gs *GalleryService) Create(title string, userID int) (*Gallary, error) {
	gallery := Gallary{
		Title:  title,
		UserID: userID,
	}

	row := gs.DB.QueryRow(`
	INSERT INTO galleries (title, user_id)
	VALUES ($1, $2) RETURNING id;
	`, gallery.Title, gallery.UserID)

	err := row.Scan(&gallery.ID)

	if err != nil {
		return nil, fmt.Errorf("create gallery: %w", err)
	}

	return &gallery, nil
}

func (gs *GalleryService) ByID(id int) (*Gallary, error) {
	gallery := Gallary{
		ID: id,
	}

	row := gs.DB.QueryRow(`
	SELECT title, user_id FROM galleries WHERE id=$1;
	`, id)

	err := row.Scan(&gallery.Title, &gallery.UserID)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("query gallary by id: %w", err)
	}

	return &gallery, nil
}

func (gs *GalleryService) ByUserID(userID int) ([]Gallary, error) {
	rows, err := gs.DB.Query(`SELECT id, title FROM galleries WHERE user_id=$1;`, userID)

	if err != nil {
		return nil, fmt.Errorf("query galleries by user: %w", err)
	}

	var galleries []Gallary

	for rows.Next() {
		gallery := Gallary{
			UserID: userID,
		}
		err := rows.Scan(&gallery.ID, &gallery.Title)

		if err != nil {
			return nil, fmt.Errorf("query galleries by user: %w", err)
		}
		galleries = append(galleries, gallery)
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf("query galleries by user: %w", err)
	}

	return galleries, nil
}

func (gs *GalleryService) Update(gallary Gallary) error {

	_, err := gs.DB.Exec(`
	UPDATE galleries 
	SET title=$2
	WHERE id=$1
	`, gallary.ID, gallary.Title)

	if err != nil {
		return fmt.Errorf("update gallery: %w", err)
	}

	return nil
}

func (gs *GalleryService) Delete(id int) error {

	_, err := gs.DB.Exec(`DELETE FROM galleries WHERE id=$1`, id)

	if err != nil {
		return fmt.Errorf("delete gallery: %w", err)
	}

	return nil
}
