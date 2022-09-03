package repo

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"sync"
	"unicode"

	"github.com/Deny7676yar/observability/Prometheus/promitheus-go/app/internal/entities/linkentity"
	"github.com/google/uuid"

	log "github.com/sirupsen/logrus"
)

type LinkeStore interface {
	Create(ctx context.Context, l linkentity.Link) (*uuid.UUID, error)
	ReadLinkRank(ctx context.Context, uid uuid.UUID) (*linkentity.Link, error)
	Delete(ctx context.Context, uid uuid.UUID) error
	SearchLinks(ctx context.Context, s string) (chan linkentity.Link, error)
	GetLongURL(ctx context.Context, sh string) (string, error)
	RankCounter(ctx context.Context, uid uuid.UUID, rank int) error
}

const (
	lenghtURL = 5
	buf       = 100
)

type Links struct {
	lstore LinkeStore
	mu     sync.Mutex
}

func NewLinks(lstore LinkeStore) *Links {
	return &Links{
		lstore: lstore,
	}
}

//Create - создание ссылки в виде json
func (ls *Links) Create(ctx context.Context, l linkentity.Link) (*linkentity.Link, error) {
	//linkentity.Link - определяется на слое entites
	var err error
	csl, err := createShortURL()
	if err != nil {
		return nil, fmt.Errorf("create link error: %w", err)
	}

	l.LinkID = uuid.New()
	id, err := ls.lstore.Create(ctx, l)
	if err != nil {
		return nil, fmt.Errorf("create link error: %w", err)
	}

	l.LinkID = *id
	l.ResultLink = csl

	return &l, nil
}

func createShortURL() (string, error) {
	var result string
	for len(result) < lenghtURL {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(127))) //nolint
		if err != nil {
			return "", err
		}
		n := num.Int64()
		if unicode.IsLetter(rune(n)) {
			result += string(rune(n))
		}
	}
	return result, nil
}

func (ls *Links) ReadLinkRank(ctx context.Context, uid uuid.UUID) (*linkentity.Link, error) {
	l, err := ls.lstore.ReadLinkRank(ctx, uid)
	if err != nil {
		return nil, fmt.Errorf("read user error: %w", err)
	}
	return l, nil
}

func (ls *Links) Delete(ctx context.Context, uid uuid.UUID) (*linkentity.Link, error) {
	l, err := ls.lstore.ReadLinkRank(ctx, uid)
	if err != nil {
		return nil, fmt.Errorf("search user error: %w", err)
	}
	return l, ls.lstore.Delete(ctx, uid)
}

func (ls *Links) SearchLinks(ctx context.Context, s string) (chan linkentity.Link, error) {
	chin, err := ls.lstore.SearchLinks(ctx, s)
	if err != nil {
		return nil, err
	}
	chout := make(chan linkentity.Link, buf) //промежуточный буферезированый канал ссылок
	go func() {
		defer close(chout)
		for {
			select {
			case <-ctx.Done():
				return
			case l, ok := <-chin: //получаем очередной обьект ссылки из канала
				if !ok {
					return
				}
				chout <- l
			}
		}
	}()
	return chout, nil
}

func (ls *Links) GetLongURL(ctx context.Context, sh string) (string, error) {
	var (
		longURL string
		rankURL int
	)
	chin, err := ls.lstore.SearchLinks(ctx, sh)
	if err != nil {
		return "", err
	}
	//промежуточный буферезированый канал ссылок
	chout := make(chan linkentity.Link)
	go func() {
		defer close(chout)
		for {
			select {
			case <-ctx.Done():
				return
			case l, ok := <-chin: //получаем очередной обьект ссылки из канала
				if !ok {
					return
				}
				chout <- l
				lu := <-chout
				longURL = lu.OriginLink

				ls.mu.Lock()
				rankURL = lu.Rank + 1
				ls.lstore.RankCounter(ctx, lu.LinkID, rankURL) //nolint
				if err != nil {
					log.WithFields(log.Fields{
						"Get RankCounter:": err,
					}).Errorf("Do Not RankCounter")
					return
				}
				ls.mu.Unlock()
			}
		}
	}()

	return longURL, nil
}

func (ls *Links) RankCounter(ctx context.Context, uid uuid.UUID, rank int) {

}
