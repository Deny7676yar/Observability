package main

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"net/url"
	"os"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

const (
	PGUserName   = "gopher"
	PGUserSecret = "P@ssw0rd"
	PGHost       = "127.0.0.1"
	PGPort       = "5432"
	PGDatabase   = "app"

	usersToCreate    = 4000
	articlesToCreate = 10000
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
	log.Println("All done!")
}

func run() error {
	data, err := parseDataFile()
	if err != nil {
		return fmt.Errorf("failed to parse the data file: %w", err)
	}

	pool, err := pgxpool.Connect(context.Background(), composeConnString())
	if err != nil {
		return fmt.Errorf("failed to initialize a connection pool: %w", err)
	}
	defer pool.Close()

	if err := fillUsersTable(context.Background(), pool, usersToCreate, data.Names); err != nil {
		return fmt.Errorf("failed to fill the users table: %w", err)
	}
	if err := fillArticlesTable(context.Background(), pool, articlesToCreate, data); err != nil {
		return fmt.Errorf("failed to fill the articles table: %w", err)
	}
	return nil
}

type Data struct {
	Names            []string `json:"names"`
	ChemicalElements []string `json:"chemical_elements"`
	Shoes            []string `json:"shoes"`
}

func parseDataFile() (*Data, error) {
	const dataFileName = "data.json"
	fileContents, err := os.ReadFile(dataFileName)
	if err != nil {
		return nil, fmt.Errorf("failed to read the data file %s: %w", dataFileName, err)
	}
	var data Data
	if err := json.Unmarshal(fileContents, &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal the data file: %w", err)
	}
	return &data, nil
}

func composeConnString() string {
	return fmt.Sprintf(
		"postgresql://%s:%s@%s:%s/%s",
		url.PathEscape(PGUserName),
		url.PathEscape(PGUserSecret),
		url.PathEscape(PGHost),
		url.PathEscape(PGPort),
		url.PathEscape(PGDatabase),
	)
}

func fillUsersTable(ctx context.Context, conn *pgxpool.Pool, usersCount int, names []string) (err error) {
	getRequest := func() (string, []interface{}, error) {
		const request = `INSERT INTO users(name) VALUES($1)`
		nameIdx, err := getRandomSliceIdx(len(names))
		if err != nil {
			return "", nil, fmt.Errorf("failed to get a random name idx: %w", err)
		}
		return request, []interface{}{names[nameIdx]}, nil
	}
	if err := writeBatches(ctx, conn, usersCount, getRequest); err != nil {
		return err
	}
	return nil
}

func fillArticlesTable(ctx context.Context, conn *pgxpool.Pool, articlesCount int, data *Data) error {
	userIDs, err := getUsersIDs(ctx, conn)
	if err != nil {
		return err
	}
	getRequest := func() (string, []interface{}, error) {
		const request = `INSERT INTO articles(user_id, title) VALUES($1, $2)`
		uidIdx, err := getRandomSliceIdx(len(userIDs))
		if err != nil {
			return "", nil, fmt.Errorf("failed to get a user ID idx: %w", err)
		}
		shoeNameIdx, err := getRandomSliceIdx(len(data.Shoes))
		if err != nil {
			return "", nil, fmt.Errorf("failed to get a shoe name idx: %w", err)
		}
		chemicalElIdx, err := getRandomSliceIdx(len(data.ChemicalElements))
		if err != nil {
			return "", nil, fmt.Errorf("failed to get a chemical element name idx: %w", err)
		}
		return request, []interface{}{*userIDs[uidIdx], fmt.Sprintf("%s %s", data.Shoes[shoeNameIdx], data.ChemicalElements[chemicalElIdx])}, nil
	}
	if err := writeBatches(ctx, conn, articlesCount, getRequest); err != nil {
		return err
	}
	return nil
}

func getUsersIDs(ctx context.Context, conn *pgxpool.Pool) ([]*pgtype.UUID, error) {
	rows, err := conn.Query(context.Background(), `SELECT id FROM users`)
	if err != nil {
		return nil, fmt.Errorf("failed to get the users IDs from the DB: %w", err)
	}
	defer rows.Close()

	uuids := make([]*pgtype.UUID, 0)
	for rows.Next() {
		var uuid pgtype.UUID
		if err := rows.Scan(&uuid); err != nil {
			return nil, fmt.Errorf("failed to scan a user's UUID: %w", err)
		}
		uuids = append(uuids, &uuid)
	}
	return uuids, nil
}

func writeBatches(ctx context.Context, conn *pgxpool.Pool, count int, getRequest func() (string, []interface{}, error)) error {
	tx, err := conn.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to start a DB transaction: %w", err)
	}
	defer func() {
		if err != nil {
			if rollbackErr := tx.Rollback(ctx); rollbackErr != nil {
				log.Println("[ERR]: failed to rollback the transaction: %w", rollbackErr)
			}
			return
		}
		if commitErr := tx.Commit(ctx); commitErr != nil {
			err = fmt.Errorf("failed to finish the transaction: %w", commitErr)
		}
	}()

	sendBatch := func(ctx context.Context, tx pgx.Tx, b *pgx.Batch) (*pgx.Batch, error) {
		batchResults := tx.SendBatch(ctx, b)
		if err := batchResults.Close(); err != nil {
			return nil, fmt.Errorf("failed to send the batch: %w", err)
		}
		b = &pgx.Batch{}
		return b, nil
	}

	ctr := 0
	b := &pgx.Batch{}
	for {
		if ctr >= count {
			break
		}
		request, requestArgs, err := getRequest()
		if err != nil {
			return fmt.Errorf("failed to generate a request: %w", err)
		}
		b.Queue(request, requestArgs...)
		ctr++
		if ctr%1000 == 0 {
			b, err = sendBatch(ctx, tx, b)
			if err != nil {
				return err
			}
			continue
		}
	}
	if _, err = sendBatch(ctx, tx, b); err != nil {
		return err
	}
	return nil
}

func getRandomSliceIdx(sliceLen int) (int, error) {
	if sliceLen == 0 {
		return 0, nil
	}
	randInt, err := rand.Int(rand.Reader, big.NewInt(int64(sliceLen)))
	if err != nil {
		return -1, fmt.Errorf("failed to generate a random slice index: %w", err)
	}
	return int(randInt.Int64()), nil
}
