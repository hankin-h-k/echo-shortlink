package redis

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/go-redis/redis"
	"github.com/mattheath/base62"
)

const (
	URLIDKEY           = "next.url.id"
	ShortlinkKey       = "shortlink:%s:url"
	URLHashKey         = "urlhash:%s:url"
	SHortlinkDetailKey = "shortlink:%s:detail"
)

type RedisCli struct {
	Cli *redis.Client
}

type URLDetail struct {
	URL                 string        `json:"url"`
	CreatedAt           string        `json:"created_at"`
	ExpirationInMinutes time.Duration `json:"expiration_in_minutes"`
}

func NewRedisCli(addr, passwd string, db int) *RedisCli {
	c := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: passwd,
		DB:       db,
	})

	if _, err := c.Ping().Result(); err != nil {
		panic(err)
	}

	return &RedisCli{
		Cli: c,
	}
}

func (r *RedisCli) Shorten(url string, exp int64) (string, error) {
	h := toShal(url)
	d, err := r.Cli.Get(fmt.Sprintf(URLHashKey, h)).Result()
	if err == redis.Nil {

	} else if err != nil {
		return "", err
	} else {
		if d == "{}" {

		} else {
			return d, nil
		}
	}

	err = r.Cli.Incr(URLIDKEY).Err()
	if err != nil {
		return "", err
	}

	id, err := r.Cli.Get(URLIDKEY).Int64()
	if err != nil {
		return "", err
	}
	eid := base62.EncodeInt64(id)

	err = r.Cli.Set(fmt.Sprintf(ShortlinkKey, eid), url, time.Minute*time.Duration(exp)).Err()
	if err != nil {
		return "", err
	}

	err = r.Cli.Set(fmt.Sprintf(URLHashKey, h), eid, time.Minute*time.Duration(exp)).Err()
	if err != nil {
		return "", err
	}

	detail, err := json.Marshal(
		&URLDetail{
			URL:                 url,
			CreatedAt:           time.Now().String(),
			ExpirationInMinutes: time.Duration(exp),
		},
	)
	if err != nil {
		return "", err
	}

	err = r.Cli.Set(fmt.Sprintf(SHortlinkDetailKey, eid), detail, time.Minute*time.Duration(exp)).Err()
	if err != nil {
		return "", err
	}

	return eid, nil
}
func (r *RedisCli) ShortlinkInfo(eid string) (interface{}, error) {
	d, err := r.Cli.Get(fmt.Sprintf(SHortlinkDetailKey, eid)).Result()
	if err == redis.Nil {
		return "", errors.New("Unknown short URL")
	} else if err != nil {
		return "", err
	} else {
		return d, nil
	}
}
func (r *RedisCli) Unshorten(eid string) (string, error) {
	url, err := r.Cli.Get(fmt.Sprintf(ShortlinkKey, eid)).Result()
	if err == redis.Nil {
		return "", errors.New("Unknown short URL")
	} else if err != nil {
		return "", err
	} else {
		return url, nil
	}
}

func toShal(urlStr string) string {
	u, err := url.Parse(urlStr)
	if err != nil {
		panic(err)
	}

	// 将URL转换为字节切片以计算哈希
	bytesToHash := []byte(u.String())

	// 计算字节切片的SHA256哈希
	hash := sha256.Sum256(bytesToHash)

	// 将哈希转换为十六进制字符串
	hashHex := hex.EncodeToString(hash[:])
	return hashHex
}
