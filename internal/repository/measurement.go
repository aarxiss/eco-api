package repository

import (
	"context"
	"eco-api/internal/model"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type MeasurementRepo struct {
	db *pgxpool.Pool
}

func NewMeasurementRepo(db *pgxpool.Pool) *MeasurementRepo {
	return &MeasurementRepo{db: db}
}

func (r *MeasurementRepo) GetAll() ([]model.Measurement, error) {
	query := `SELECT sensor_id, value, created_at FROM measurements ORDER BY created_at DESC LIMIT 100`
	rows, err := r.db.Query(context.Background(), query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []model.Measurement
	for rows.Next() {
		var m model.Measurement
		if err := rows.Scan(&m.SensorID, &m.Value, &m.CreatedAt); err == nil {
			results = append(results, m)
		}
	}
	return results, nil
}

func (r *MeasurementRepo) Create(m model.Measurement) error {
	query := `INSERT INTO measurements (sensor_id, value, created_at) VALUES ($1, $2, $3)`
	_, err := r.db.Exec(context.Background(), query, m.SensorID, m.Value, m.CreatedAt)
	return err
}

func (r *MeasurementRepo) Update(id string, value float64) error {
	query := `UPDATE measurements SET value = $1 WHERE sensor_id = $2`
	commandTag, err := r.db.Exec(context.Background(), query, value, id)
	if err != nil {
		return err
	}
	if commandTag.RowsAffected() == 0 {
		return fmt.Errorf("не знайдено")
	}
	return nil
}

func (r *MeasurementRepo) Delete(id string) error {
	query := `DELETE FROM measurements WHERE sensor_id = $1`
	commandTag, err := r.db.Exec(context.Background(), query, id)
	if err != nil {
		return err
	}
	if commandTag.RowsAffected() == 0 {
		return fmt.Errorf("не знайдено")
	}
	return nil
}

func (r *MeasurementRepo) GetByID(sensorID string) (*model.Measurement, error) {
	query := `SELECT sensor_id, value, created_at FROM measurements WHERE sensor_id = $1 ORDER BY created_at DESC LIMIT 1`

	var m model.Measurement
	err := r.db.QueryRow(context.Background(), query, sensorID).Scan(
		&m.SensorID,
		&m.Value,
		&m.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &m, nil
}
