package main

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"net/url"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

const (
	attackTime   = time.Second * 10
	workersCount = 50
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	data, err := parseDataFile()
	if err != nil {
		return fmt.Errorf("failed to parse the data file: %w", err)
	}

	wg := &sync.WaitGroup{}
	var opsCount uint64

	startAt := time.Now()
	stopAt := startAt.Add(attackTime)

	attacker := func(stopAt time.Time) {
		for {
			if time.Now().After(stopAt) {
				return
			}
			name, err := data.getRandomName()
			if err != nil {
				continue
			}
			resp, err := http.Get(fmt.Sprintf("http://localhost:9000/users/name/%s", url.PathEscape(name)))
			if err != nil {
				continue
			}
			resp.Body.Close()
			atomic.AddUint64(&opsCount, 1)
		}
	}

	for i := 0; i < workersCount; i++ {
		wg.Add(1)
		go func() {
			attacker(stopAt)
			wg.Done()
		}()
	}
	wg.Wait()

	d := time.Since(startAt)
	fmt.Println("QPS: ", opsCount/uint64(d.Seconds()))
	return nil
}

type Data struct {
	Names []string      `json:"names"`
	mux   *sync.RWMutex `json:"-"`
}

func parseDataFile() (*Data, error) {
	const dataFileName = "../init-db/data.json"
	fileContents, err := os.ReadFile(dataFileName)
	if err != nil {
		return nil, fmt.Errorf("failed to read the data file %s: %w", dataFileName, err)
	}
	var data Data
	if err := json.Unmarshal(fileContents, &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal the data file: %w", err)
	}
	data.mux = &sync.RWMutex{}
	return &data, nil
}

func (d *Data) getRandomName() (string, error) {
	d.mux.RLock()
	defer d.mux.RUnlock()

	idx, err := getRandomSliceIdx(len(d.Names))
	if err != nil {
		return "", fmt.Errorf("failed to get a random name: %w", err)
	}
	return d.Names[idx], nil
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
