package redis

import (
	"errors"
	"fmt"
	"github.com/forceu/gokapi/internal/helper"
	"github.com/forceu/gokapi/internal/models"
	redigo "github.com/gomodule/redigo/redis"
	"strconv"
	"strings"
	"time"
)

// DatabaseProvider contains the database instance
type DatabaseProvider struct {
	pool     *redigo.Pool
	dbPrefix string
}

// New returns an instance
func New(dbConfig models.DbConnection) (DatabaseProvider, error) {
	return DatabaseProvider{}.init(dbConfig)
}

// GetType returns 1, for being a Redis interface
func (p DatabaseProvider) GetType() int {
	return 1 // dbabstraction.Redis
}

// Init connects to the database and creates the table structure, if necessary
// IMPORTANT: The function returns itself, as Go does not allow this function to be pointer-based
// The resulting new reference must then be used.
func (p DatabaseProvider) init(config models.DbConnection) (DatabaseProvider, error) {
	if config.HostUrl == "" {
		return DatabaseProvider{}, errors.New("empty database url was provided")
	}
	p.dbPrefix = config.RedisPrefix
	p.pool = newPool(config)
	conn := p.pool.Get()
	defer conn.Close()
	_, err := redigo.String(conn.Do("PING"))
	return p, err
}

func getDialOptions(config models.DbConnection) []redigo.DialOption {
	dialOptions := []redigo.DialOption{redigo.DialClientName("gokapi")}
	if config.Username != "" {
		dialOptions = append(dialOptions, redigo.DialUsername(config.Username))
	}
	if config.Password != "" {
		dialOptions = append(dialOptions, redigo.DialPassword(config.Password))
	}
	if config.RedisUseSsl {
		dialOptions = append(dialOptions, redigo.DialUseTLS(true))
	}
	return dialOptions
}

func newPool(config models.DbConnection) *redigo.Pool {

	newRedisPool := &redigo.Pool{
		MaxIdle:     10,
		IdleTimeout: 2 * time.Minute,

		Dial: func() (redigo.Conn, error) {
			c, err := redigo.Dial("tcp", config.HostUrl, getDialOptions(config)...)
			if err != nil {
				fmt.Println("Error connecting to redis")
			}
			helper.Check(err)
			return c, err
		},

		TestOnBorrow: func(c redigo.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}
	return newRedisPool
}

// Upgrade migrates the DB to a new Gokapi version, if required
func (p DatabaseProvider) Upgrade(currentDbVersion int) {
	// Currently no upgrade necessary
	return
}

const keyDbVersion = "dbversion"

// GetDbVersion gets the version number of the database
func (p DatabaseProvider) GetDbVersion() int {
	key, _ := p.getKeyInt(keyDbVersion)
	return key
}

// SetDbVersion sets the version number of the database
func (p DatabaseProvider) SetDbVersion(currentVersion int) {
	p.setKey(keyDbVersion, currentVersion)
}

// Close the database connection
func (p DatabaseProvider) Close() {
	err := p.pool.Close()
	if err != nil {
		fmt.Println(err)
	}
}

// RunGarbageCollection runs the databases GC
func (p DatabaseProvider) RunGarbageCollection() {
	// No cleanup required
}

// Function to get all hashmaps with a given prefix
func (p DatabaseProvider) getAllValuesWithPrefix(prefix string) map[string]any {
	result := make(map[string]any)
	allKeys := p.getAllKeysWithPrefix(prefix)
	for _, key := range allKeys {
		value, err := p.getKeyRaw(key)
		if errors.Is(err, redigo.ErrNil) {
			continue
		}
		helper.Check(err)
		result[key] = value
	}
	return result
}

// Function to get all hashmaps with a given prefix
func (p DatabaseProvider) getAllHashesWithPrefix(prefix string) map[string][]any {
	result := make(map[string][]any)
	allKeys := p.getAllKeysWithPrefix(prefix)
	for _, key := range allKeys {
		hashMap, ok := p.getHashMap(key)
		if !ok {
			continue
		}
		result[key] = hashMap
	}
	return result
}

func (p DatabaseProvider) getAllKeysWithPrefix(prefix string) []string {
	var result []string
	conn := p.pool.Get()
	defer conn.Close()
	fullPrefix := p.dbPrefix + prefix
	cursor := 0
	for {
		reply, err := redigo.Values(conn.Do("SCAN", cursor, "MATCH", fullPrefix+"*", "COUNT", 100))
		helper.Check(err)

		cursor, _ = redigo.Int(reply[0], nil)
		keys, _ := redigo.Strings(reply[1], nil)
		for _, key := range keys {
			result = append(result, strings.Replace(key, p.dbPrefix, "", 1))
		}
		if cursor == 0 {
			break
		}
	}
	return result
}

func (p DatabaseProvider) setKey(id string, content any) {
	conn := p.pool.Get()
	defer conn.Close()
	_, err := conn.Do("SET", p.dbPrefix+id, content)
	helper.Check(err)
}

func (p DatabaseProvider) getKeyRaw(id string) (any, error) {
	conn := p.pool.Get()
	defer conn.Close()
	return conn.Do("GET", p.dbPrefix+id)
}

func (p DatabaseProvider) getKeyString(id string) (string, bool) {
	result, err := redigo.String(p.getKeyRaw(id))
	if result == "" {
		return "", false
	}
	helper.Check(err)
	return result, true
}

func (p DatabaseProvider) getKeyInt(id string) (int, bool) {
	result, err := p.getKeyRaw(id)
	if result == nil {
		return 0, false
	}
	resultInt, err2 := redigo.Int(result, err)
	helper.Check(err2)
	return resultInt, true
}
func (p DatabaseProvider) getKeyBytes(id string) ([]byte, bool) {
	result, err := p.getKeyRaw(id)
	if result == nil {
		return nil, false
	}
	resultInt, err2 := redigo.Bytes(result, err)
	helper.Check(err2)
	return resultInt, true
}

func (p DatabaseProvider) getHashMap(id string) ([]any, bool) {
	conn := p.pool.Get()
	defer conn.Close()
	result, err := redigo.Values(conn.Do("HGETALL", p.dbPrefix+id))
	helper.Check(err)
	if len(result) == 0 {
		return nil, false
	}
	return result, true
}

func (p DatabaseProvider) buildArgs(id string) redigo.Args {
	return redigo.Args{}.Add(p.dbPrefix + id)
}

func (p DatabaseProvider) setHashMap(content redigo.Args) {
	conn := p.pool.Get()
	defer conn.Close()
	_, err := conn.Do("HMSET", content...)
	helper.Check(err)
}

func (p DatabaseProvider) setExpiryAt(id string, expiry int64) {
	conn := p.pool.Get()
	defer conn.Close()
	_, err := conn.Do("EXPIREAT", p.dbPrefix+id, strconv.FormatInt(expiry, 10))
	helper.Check(err)
}
func (p DatabaseProvider) setExpiryInSeconds(id string, expiry int64) {
	conn := p.pool.Get()
	defer conn.Close()
	_, err := conn.Do("EXPIRE", p.dbPrefix+id, strconv.FormatInt(expiry, 10))
	helper.Check(err)
}

func (p DatabaseProvider) deleteKey(id string) {
	conn := p.pool.Get()
	defer conn.Close()
	_, err := conn.Do("DEL", p.dbPrefix+id)
	helper.Check(err)
}

func (p DatabaseProvider) runEval(cmd string) {
	conn := p.pool.Get()
	defer conn.Close()
	_, err := conn.Do("EVAL", cmd, "0")
	helper.Check(err)
}

func (p DatabaseProvider) deleteAllWithPrefix(prefix string) {
	p.runEval("for _,k in ipairs(redis.call('keys','" + p.dbPrefix + prefix + "*')) do redis.call('del',k) end")
}
